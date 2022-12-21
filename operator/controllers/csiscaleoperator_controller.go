/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	csiv1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	v1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	config "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	csiscaleoperator "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/internal/csiscaleoperator"
	clustersyncer "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/syncer"

	//TODO: This is temporary change, once the SpectrumRestV2 is exported
	//in driver code and merged in some IBM branch, change this line and
	//adjust the dependencies.
	"github.com/amdabhad/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
)

// CSIScaleOperatorReconciler reconciles a CSIScaleOperator object
type CSIScaleOperatorReconciler struct {
	Client        client.Client
	Scheme        *runtime.Scheme
	recorder      record.EventRecorder
	serverVersion string
}

const MinControllerReplicas = 1

var daemonSetRestartedKey = ""
var daemonSetRestartedValue = ""

var csiLog = log.Log.WithName("csiscaleoperator_controller")

type reconciler func(instance *csiscaleoperator.CSIScaleOperator) error

var crStatus = csiv1.CSIScaleOperatorStatus{}

//a map of connectors to make REST calls to GUI
var scaleConnMap = make(map[string]connectors.SpectrumScaleConnector)

// watchResources stores resource kind and resource names of the resources
// that the controller is going to watch.
// Namespace information is not stored in the variable
// as the operator is namespace scoped and watches only within the given namespaces.
// Reference: getWatchNamespace() in main.go
var watchResources = map[string]map[string]bool{corev1.ResourceConfigMaps.String(): {}, corev1.ResourceSecrets.String(): {}}

// +kubebuilder:rbac:groups=csi.ibm.com,resources=*,verbs=*

// +kubebuilder:rbac:groups="",resources={pods,persistentvolumeclaims,services,endpoints,events,configmaps,secrets,secrets/status,services/finalizers,serviceaccounts},verbs=*
// TODO: Does the operator need to access to all the resources mentioned above?
// TODO: Does all resources mentioned above required delete/patch/update permissions?

// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources={clusterroles,clusterrolebindings},verbs=*
// +kubebuilder:rbac:groups="apps",resources={deployments,daemonsets,replicasets,statefulsets},verbs=create;delete;get;list;update;watch
// +kubebuilder:rbac:groups="apps",resourceNames=ibm-spectrum-scale-csi-operator,resources=deployments/finalizers,verbs=get;update
// +kubebuilder:rbac:groups="storage.k8s.io",resources={volumeattachments,storageclasses,csidrivers},verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=*
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch

// TODO: In case of multiple controllers, define role and rolebinding separately for leases.
// +kubebuilder:rbac:groups="coordination.k8s.io",resources={leases},verbs=create;delete;get;list;update;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CSIScaleOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CSIScaleOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := csiLog.WithName("Reconcile")
	logger.Info("CSI setup started.")

	setENVIsOpenShift(r)

	// Fetch the CSIScaleOperator instance
	logger.Info("Fetching CSIScaleOperator instance.")
	instance := csiscaleoperator.New(&csiv1.CSIScaleOperator{})

	instanceUnwrap := instance.Unwrap()
	var err error
	err = r.Client.Get(ctx, req.NamespacedName, instanceUnwrap)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Error(err, "CSIScaleOperator resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get CSIScaleOperator.")
		return ctrl.Result{}, err
	}

	meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
		Type:    string(config.StatusConditionSuccess),
		Status:  metav1.ConditionUnknown,
		Reason:  string(csiv1.Unknown),
		Message: "",
	})

	// Update status conditions after reconcile completed
	// crStatus := csiv1.CSIScaleOperatorStatus{}
	defer func() {
		cr := &csiv1.CSIScaleOperator{
			TypeMeta: metav1.TypeMeta{
				Kind:       config.Kind,
				APIVersion: config.APIGroup + "/" + config.APIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		}
		if err := r.SetStatus(instance); err != nil {
			logger.Error(err, "Assigning values to status sub-resource object failed.")
		}
		cr.Status = crStatus
		err := r.Client.Status().Patch(ctx, cr, client.Merge, client.FieldOwner("CSIScaleOperator"))
		if err != nil {
			logger.Error(err, "Deferred update of resource status failed.", "Status", cr.Status)
		}
		logger.V(1).Info("Updated resource status.", "Status", cr.Status)
	}()

	r.Scheme.Default(instanceUnwrap)
	err = r.Client.Update(context.TODO(), instanceUnwrap)
	if err != nil {
		logger.Error(err, "Reconciler Client.Update() failed")
		return ctrl.Result{}, err
	}

	for _, cluster := range instance.Spec.Clusters {
		if cluster.Cacert != "" {
			watchResources[corev1.ResourceConfigMaps.String()][cluster.Cacert] = true
		}
		if cluster.Secrets != "" {
			watchResources[corev1.ResourceSecrets.String()][cluster.Secrets] = true
		}
	}

	logger.Info("Adding Finalizer")
	if err := r.addFinalizerIfNotPresent(instance); err != nil {
		message := "Couldn't add Finalizer"
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceUpdateError),
			Message: message,
		})
		return ctrl.Result{}, err
	}

	logger.Info("Checking if CSIScaleOperator object got deleted")
	if !instance.GetDeletionTimestamp().IsZero() {

		logger.Info("Attempting cleanup of CSI driver")
		isFinalizerExists, err := r.hasFinalizer(instance)
		if err != nil {
			message := "Finalizer check failed"
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return ctrl.Result{}, err
		}

		if !isFinalizerExists {
			logger.Error(err, "No finalizer was found")
			return ctrl.Result{}, nil
		}

		if err := r.deleteClusterRolesAndBindings(instance); err != nil {
			message := "Failed to delete ClusterRoles and ClusterRolesBindings"
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceDeleteError),
				Message: message,
			})
			return ctrl.Result{}, err
		}

		if err := r.deleteCSIDriver(instance); err != nil {
			message := "Failed to delete CSIDriver"
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceDeleteError),
				Message: message,
			})
			return ctrl.Result{}, err
		}

		if err := r.removeFinalizer(instance); err != nil {
			message := "Failed to remove Finalizer"
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceUpdateError),
				Message: message,
			})
			return ctrl.Result{}, err
		}
		logger.Info("Removed CSI driver successfully")
		return ctrl.Result{}, nil
	}

	if pass, err := r.checkPrerequisite(instance); !pass {
		logger.Error(err, "Pre-requisite check failed.")
		return ctrl.Result{}, err
	} else {
		logger.Info("Pre-requisite check passed.")
	}

	cmExists, clustersStanzaModified, err := r.isClusterStanzaModified(req.Namespace, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !cmExists || clustersStanzaModified {
		err = ValidateCRParams(instance)
		if err != nil {
			logger.Error(fmt.Errorf("CR validation for driver configuration params failed"), "")
			return ctrl.Result{}, err
		}
		logger.Info("Driver configuration params are validated successfully")
	}

	if len(instance.Spec.Clusters) != 0 {
		err = r.getSpectrumScaleConnectors(instance, cmExists, clustersStanzaModified)
		if err != nil {
			message := "Error in getting connectors"
			logger.Error(err, message)
			return ctrl.Result{}, err
		}
	}

	//Test connectors and REST calls
	//TODO: Remove the test calls and definitions once evrything
	//is working fine.
	testClusterID()
	testRESTcalls(instance.Spec.Clusters)

	//If first pass or cluster stanza modified handle primary FS and fileset
	if !cmExists || clustersStanzaModified {
		err = r.handlePrimaryFSandFileset(instance)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info("Create resources")
	// create the resources which never change if not exist
	for _, rec := range []reconciler{
		r.reconcileCSIDriver,
		r.reconcileServiceAccount,
		r.reconcileClusterRole,
		r.reconcileClusterRoleBinding,
	} {
		if err = rec(instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info("Successfully created resources like ServiceAccount, ClusterRoles and ClusterRoleBinding")

	// Rollout restart of node plugin pods, if modified driver
	// manifest file is applied.

	if cmExists && clustersStanzaModified {
		logger.Info("Some of the cluster fields of CSIScaleOperator instance are changed, so restarting node plugin pods")
		err = r.handleDriverRestart(instance)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Synchronizing the resources which change over time.
	// Resource list:
	// 1. Cluster configMap
	// 2. Attacher deployment
	// 3. Provisioner deployment
	// 4. Snapshotter deployment
	// 5. Resizer deployment
	// 6. Driver daemonset

	// Synchronizing cluster configMap
	csiConfigmapSyncer := clustersyncer.CSIConfigmapSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiConfigmapSyncer, r.recorder); err != nil {
		message := "Synchronization of " + config.CSIConfigMap + " ConfigMap failed."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceSyncError),
			Message: message,
		})
		return ctrl.Result{}, err
	}
	logger.Info("Synchronization of ConfigMap is successful")

	// Synchronizing attacher deployment
	if err := r.removeDeprecatedStatefulset(instance, config.GetNameForResource(config.CSIControllerAttacher, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}
	csiControllerSyncer := clustersyncer.GetAttacherSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncer, r.recorder); err != nil {
		message := "Synchronization of attacher interface failed."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceSyncError),
			Message: message,
		})
		return ctrl.Result{}, err
	}
	logger.Info("Synchronization of attacher interface is successful")

	// Synchronizing provisioner deployment
	if err := r.removeDeprecatedStatefulset(instance, config.GetNameForResource(config.CSIControllerProvisioner, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}
	csiControllerSyncerProvisioner := clustersyncer.GetProvisionerSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncerProvisioner, r.recorder); err != nil {
		message := "Synchronization of provisioner interface failed."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceSyncError),
			Message: message,
		})
		return ctrl.Result{}, err
	}
	logger.Info("Synchronization of provisioner interface is successful")

	// Synchronizing snapshotter deployment
	if err := r.removeDeprecatedStatefulset(instance, config.GetNameForResource(config.CSIControllerSnapshotter, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}
	csiControllerSyncerSnapshotter := clustersyncer.GetSnapshotterSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncerSnapshotter, r.recorder); err != nil {
		message := "Synchronization of snapshotter interface failed."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceSyncError),
			Message: message,
		})
		return ctrl.Result{}, err
	}
	logger.Info("Synchronization of snapshotter interface is successful")

	// Synchronizing resizer deployment
	if err := r.removeDeprecatedStatefulset(instance, config.GetNameForResource(config.CSIControllerResizer, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}
	csiControllerSyncerResizer := clustersyncer.GetResizerSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncerResizer, r.recorder); err != nil {
		message := "Synchronization of resizer interface failed."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceSyncError),
			Message: message,
		})
		return ctrl.Result{}, err
	}
	logger.Info("Synchronization of resizer interface is successful")

	// Synchronizing node/driver daemonset
	CGPrefix := r.GetConsistencyGroupPrefix(instance)

	if instance.Spec.CGPrefix == "" {
		logger.Info("Updating consistency group prefix in CSIScaleOperator resource.")
		instance.Spec.CGPrefix = CGPrefix
		err := r.Client.Update(ctx, instance.Unwrap())
		if err != nil {
			logger.Error(err, "Reconciler Client.Update() failed.")
			return ctrl.Result{}, err
		}
		logger.Info("Successfully updated consistency group prefix in CSIScaleOperator resource.")

	}

	if err != nil {
		return ctrl.Result{}, err
	}

	csiNodeSyncer := clustersyncer.GetCSIDaemonsetSyncer(r.Client, r.Scheme, instance, daemonSetRestartedKey, daemonSetRestartedValue, CGPrefix)
	if err := syncer.Sync(context.TODO(), csiNodeSyncer, r.recorder); err != nil {
		message := "Synchronization of node/driver interface failed."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceSyncError),
			Message: message,
		})
		return ctrl.Result{}, err
	}
	logger.Info("Synchronization of node/driver interface is successful")

	message := "The CSI driver resources have been created/updated successfully."
	logger.Info(message)

	// if err := r.SetStatus(instance); err != nil {
	// 	logger.Error(err, "Assigning values to status sub-resource object failed.")
	//	return ctrl.Result{}, err
	// }
	// TODO: Add event.
	meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
		Type:    string(config.StatusConditionSuccess),
		Status:  metav1.ConditionTrue,
		Reason:  string(csiv1.CSIConfigured),
		Message: message,
	})

	logger.Info("CSI setup completed successfully.")
	return ctrl.Result{}, nil
}

//handleDriverRestart gets a driver daemoset from the cluster and
//restarts driver pods, returns error if there is any.
func (r *CSIScaleOperatorReconciler) handleDriverRestart(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("handleDriverRestart")
	var err error
	var daemonSet *appsv1.DaemonSet
	daemonSet, err = r.getNodeDaemonSet(instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			message := "Failed to get the driver Daemonset"
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
		} else {
			message := "The driver Daemonset is not found on cluster"
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
		}
	} else {
		err = r.rolloutRestartNode(daemonSet)
		if err != nil {
			message := "Failed to rollout restart of driver pods"
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceUpdateError),
				Message: message,
			})
		} else {
			daemonSetRestartedKey, daemonSetRestartedValue = r.getRestartedAtAnnotation(daemonSet.Spec.Template.ObjectMeta.Annotations)
		}
	}
	return err
}

//isClusterStanzaModified checks if spectrum-scale-config configmap exists
//and if it exists checks if the clusters stanza is modified by
//comparing it with the configmap data.
//It returns 1st value (cmExists) which indicates if clusters configmap exists,
//2nd value (clustersStanzaModified) which idicates whether clusters stanza is
//modified in case the configmap exists, and 3rd value as an error if any.
func (r *CSIScaleOperatorReconciler) isClusterStanzaModified(namespace string, instance *csiscaleoperator.CSIScaleOperator) (bool, bool, error) {
	logger := csiLog.WithName("isClusterStanzaModified")
	cmExists := false
	clustersStanzaModified := false
	configMap := &corev1.ConfigMap{}
	cerr := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.CSIConfigMap,
		Namespace: namespace,
	}, configMap)
	if cerr != nil {
		if !errors.IsNotFound(cerr) {
			message := "Failed to get ConfigMap resource " + config.CSIConfigMap
			logger.Error(cerr, message)

			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return cmExists, clustersStanzaModified, cerr
		} else {
			//configmap not found - first pass
			return cmExists, clustersStanzaModified, nil
		}
	} else {
		cmExists = true
		clustersBytes, err := json.Marshal(&instance.Spec.Clusters)
		if err != nil {
			logger.Error(err, "Failed to marshal clusters data of this instance")
			return cmExists, clustersStanzaModified, err
		}
		clustersString := string(clustersBytes)
		configMapDataBytes, err := json.Marshal(&configMap.Data)
		if err != nil {
			logger.Error(err, "Failed to marshal ConfigMap data"+config.CSIConfigMap)
			return cmExists, clustersStanzaModified, err
		}
		configMapDataString := string(configMapDataBytes)
		configMapDataString = strings.Replace(configMapDataString, " ", "", -1)
		configMapDataString = strings.Replace(configMapDataString, "\\\"", "\"", -1)

		if !strings.Contains(configMapDataString, clustersString) {
			logger.Info("Clusters stanza in driver manifest is changed")
			clustersStanzaModified = true
		}
	}
	return cmExists, clustersStanzaModified, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CSIScaleOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {

	logger := csiLog.WithName("SetupWithManager")

	logger.Info("Running IBM Spectrum Scale CSI operator", "version", config.OperatorVersion)
	logger.Info("Setting up the controller with the manager.")

	CSIReconcileRequestFunc := func(obj client.Object) []reconcile.Request {
		var requests = []reconcile.Request{}
		var CSIScales = csiv1.CSIScaleOperatorList{}
		_ = mgr.GetClient().List(context.TODO(), &CSIScales)
		//Note:  All CSIScaleOperator objects present in cluster are added to request.
		for _, CSIScale := range CSIScales.Items {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      CSIScale.Name,
					Namespace: CSIScale.Namespace,
				},
			})
		}
		return requests
	}

	isCSIResource := func(resourceName string, resourceKind string) bool {
		return watchResources[resourceKind][resourceName]
	}

	predicateFuncs := func(resourceKind string) predicate.Funcs {
		return predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				if isCSIResource(e.Object.GetName(), resourceKind) {
					r.restartDriverPods(mgr, "created", resourceKind, e.Object.GetName())
				} else {
					return false
				}
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				if isCSIResource(e.ObjectNew.GetName(), resourceKind) {
					if !reflect.DeepEqual(e.ObjectOld.(*corev1.Secret).Data, e.ObjectNew.(*corev1.Secret).Data) {
						r.restartDriverPods(mgr, "updated", resourceKind, e.ObjectOld.GetName())
					}
				} else {
					return false
				}
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				if isCSIResource(e.Object.GetName(), resourceKind) {
					r.restartDriverPods(mgr, "deleted", resourceKind, e.Object.GetName())
				} else {
					return false
				}
				return true
			},
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1.CSIScaleOperator{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(CSIReconcileRequestFunc),
			builder.WithPredicates(predicateFuncs(corev1.ResourceSecrets.String())),
		).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(CSIReconcileRequestFunc),
			builder.WithPredicates(predicateFuncs(corev1.ResourceConfigMaps.String())),
		).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *CSIScaleOperatorReconciler) restartDriverPods(mgr ctrl.Manager, event string, resourceKind string, resourceName string) {

	logger := csiLog.WithName("restartDriverPods").WithValues("Kind", resourceKind, "Name", resourceName, "Event", event)

	CSIDaemonListFunc := func() []appsv1.DaemonSet {
		var CSIDaemonSets = []appsv1.DaemonSet{}
		var DaemonSets = appsv1.DaemonSetList{}
		_ = mgr.GetClient().List(context.TODO(), &DaemonSets)

		for _, DaemonSet := range DaemonSets.Items {
			if DaemonSet.Labels[config.LabelProduct] == config.Product {
				CSIDaemonSets = append(CSIDaemonSets, DaemonSet)
			}
		}
		return CSIDaemonSets
	}

	daemonSets := CSIDaemonListFunc()
	for i := range daemonSets {
		logger.Info("Watched resource is modified. Driver pods will be restarted.")
		err := r.rolloutRestartNode(&daemonSets[i])
		if err != nil {
			logger.Error(err, "Unable to restart driver pods. Please restart node specific pods manually.")
		} else {
			daemonSetRestartedKey, daemonSetRestartedValue = r.getRestartedAtAnnotation(daemonSets[i].Spec.Template.ObjectMeta.Annotations)
		}
	}

}

func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func (r *CSIScaleOperatorReconciler) hasFinalizer(instance *csiscaleoperator.CSIScaleOperator) (bool, error) {

	logger := csiLog.WithName("hasFinalizer")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		logger.Error(err, "No finalizer found")
		return false, err
	}

	logger.Info("Returning with finalizer", "name", finalizerName)
	return Contains(accessor.GetFinalizers(), finalizerName), nil
}

// This removes an entry from list of strings
func removeListEntry(list []string, s string) []string {
	var newList []string
	for _, v := range list {
		if v != s {
			newList = append(newList, v)
		}
	}
	return newList
}

func (r *CSIScaleOperatorReconciler) removeFinalizer(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("removeFinalizer")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		logger.Error(err, "Couldn't get finalizer")
		return err
	}

	accessor.SetFinalizers(removeListEntry(accessor.GetFinalizers(), finalizerName))
	if err := r.Client.Update(context.TODO(), instance.Unwrap()); err != nil {
		logger.Error(err, "Failed to remove", "finalizer", finalizerName, "from", accessor.GetName())
		return err
	}

	logger.Info("Finalizer was removed.")
	return nil
}

func (r *CSIScaleOperatorReconciler) addFinalizerIfNotPresent(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("addFinalizerIfNotPresent")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		logger.Error(err, "Failed to get finalizer name")
		return err
	}

	if !Contains(accessor.GetFinalizers(), finalizerName) {
		logger.Info("Adding", "finalizer", finalizerName, "on", accessor.GetName())
		accessor.SetFinalizers(append(accessor.GetFinalizers(), finalizerName))

		if err := r.Client.Update(context.TODO(), instance.Unwrap()); err != nil {
			logger.Error(err, "Failed to add", "finalizer", finalizerName, "on", accessor.GetName())
			return err
		}
	}
	logger.Info("Finalizer was added with", "name", finalizerName)
	return nil
}

func (r *CSIScaleOperatorReconciler) getAccessorAndFinalizerName(instance *csiscaleoperator.CSIScaleOperator) (metav1.Object, string, error) {
	logger := csiLog.WithName("getAccessorAndFinalizerName")

	finalizerName := config.CSIFinalizer

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "Failed to get meta information of instance")
		return nil, "", err
	}

	logger.Info("Got finalizer with", "name", finalizerName)
	return accessor, finalizerName, nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRolesAndBindings(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteClusterRolesAndBindings")

	logger.Info("Calling deleteClusterRoleBindings()")
	if err := r.deleteClusterRoleBindings(instance); err != nil {
		logger.Error(err, "Deletion of ClusterRoleBindings failed")
		return err
	}

	logger.Info("Calling deleteClusterRoles()")
	if err := r.deleteClusterRoles(instance); err != nil {
		logger.Error(err, "Deletion of ClusterRoles failed")
		return err
	}

	logger.Info("Deletion of ClusterRoles and ClusterRoleBindings succeeded.")
	return nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRoles(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteClusterRoles")

	logger.Info("Deleting ClusterRoles")
	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Continuing working on ClusterRoles for deletion")
			continue
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			logger.Info("Deleting ClusterRole", "Name", cr.GetName())
			if err := r.Client.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "Failed to delete ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}
	logger.Info("Exiting deleteClusterRoles method.")
	return nil
}

// reconcileCSIDriver creates a new CSIDriver object in the cluster.
// It returns nil if CSIDriver object is created successfully or it already exists.
func (r *CSIScaleOperatorReconciler) reconcileCSIDriver(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("reconcileCSIDriver").WithValues("Name", config.DriverName)
	logger.Info("Creating a new CSIDriver resource.")

	cd := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      cd.Name,
		Namespace: "",
	}, found)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), cd)
		if err != nil {
			message := "Failed to create a new CSIDriver resource."
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceCreateError),
				Message: message,
			})
			return err
		}
	} else if err != nil {
		message := "Failed to get the CSIDriver object from the cluster."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceReadError),
			Message: message,
		})
		return err
	} else {
		// Resource already exists - don't requeue
		logger.Info("Resource CSIDriver already exists.")
	}
	logger.V(1).Info("Exiting reconcileCSIDriver method.")
	return nil
}

func (r *CSIScaleOperatorReconciler) reconcileServiceAccount(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("reconcileServiceAccount")
	logger.Info("Creating the required ServiceAccount resources.")

	// controller := instance.GenerateControllerServiceAccount()
	node := instance.GenerateNodeServiceAccount()
	attacher := instance.GenerateAttacherServiceAccount()
	provisioner := instance.GenerateProvisionerServiceAccount()
	snapshotter := instance.GenerateSnapshotterServiceAccount()
	resizer := instance.GenerateResizerServiceAccount()

	// controllerServiceAccountName := config.GetNameForResource(config.CSIControllerServiceAccount, instance.Name)
	nodeServiceAccountName := config.GetNameForResource(config.CSINodeServiceAccount, instance.Name)
	// attacherServiceAccountName := config.GetNameForResource(config.CSIAttacherServiceAccount, instance.Name)
	// provisionerServiceAccountName := config.GetNameForResource(config.CSIProvisionerServiceAccount, instance.Name)
	// snapshotterServiceAccountName := config.GetNameForResource(config.CSISnapshotterServiceAccount, instance.Name)

	for _, sa := range []*corev1.ServiceAccount{
		// controller,
		node,
		attacher,
		provisioner,
		snapshotter,
		resizer,
	} {
		if err := controllerutil.SetControllerReference(instance.Unwrap(), sa, r.Scheme); err != nil {
			message := "Failed to set controller reference for ServiceAccount " + sa.GetName()
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.CSINotConfigured),
				Message: message,
			})
			return err
		}
		found := &corev1.ServiceAccount{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      sa.Name,
			Namespace: sa.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ServiceAccount.", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
			err = r.Client.Create(context.TODO(), sa)
			if err != nil {
				message := "Failed to create ServiceAccount resource " + sa.GetName() + "."
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceCreateError),
					Message: message,
				})
				return err
			}
			logger.Info("Creation of ServiceAccount " + sa.GetName() + " is successful")

			//if controllerServiceAccountName == sa.Name {
			//	rErr := r.restartControllerPod(logger, instance)
			//	if rErr != nil {
			//		message := "Failed to restart controller pod."
			//		logger.Error(rErr, message)
			//		// TODO: Add event.
			//		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			//			Type:    string(config.StatusConditionSuccess),
			//			Status:  metav1.ConditionFalse,
			//			Reason:  string(csiv1.CSINotConfigured),
			//			Message: message,
			//		})
			//		return rErr
			//	}
			//}

			if nodeServiceAccountName == sa.Name {

				nodeDaemonSet, err := r.getNodeDaemonSet(instance)
				if err != nil && errors.IsNotFound(err) {
					logger.Info("Daemonset doesn't exist. Restart not required.")
				} else if err != nil {
					message := "Failed to check if DaemonSet exists."
					logger.Error(err, message)
					// TODO: Add event.
					meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
						Type:    string(config.StatusConditionSuccess),
						Status:  metav1.ConditionFalse,
						Reason:  string(csiv1.ResourceReadError),
						Message: message,
					})
					return err
				} else {
					logger.Info("DaemonSet exists, node rollout requires restart",
						"DesiredNumberScheduled", nodeDaemonSet.Status.DesiredNumberScheduled,
						"NumberAvailable", nodeDaemonSet.Status.NumberAvailable)

					rErr := r.rolloutRestartNode(nodeDaemonSet)
					if rErr != nil {
						message := "Failed to rollout restart node DaemonSet"
						logger.Error(rErr, message)
						// TODO: Add event.
						meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
							Type:    string(config.StatusConditionSuccess),
							Status:  metav1.ConditionFalse,
							Reason:  string(csiv1.CSINotConfigured),
							Message: message,
						})
						return rErr
					}

					daemonSetRestartedKey, daemonSetRestartedValue = r.getRestartedAtAnnotation(nodeDaemonSet.Spec.Template.ObjectMeta.Annotations)
					logger.Info("Rollout restart of node DaemonSet is successful")
				}
				// TODO: Should restart sidecar pods if respective ServiceAccount is created afterwards?
			}
		} else if err != nil {
			message := "Failed to get ServiceAccount " + sa.GetName()
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return err
		} else {
			// Cannot update the service account of an already created pod.
			// Reference: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
			logger.Info("ServiceAccount " + sa.GetName() + " already exists.")
		}
	}
	logger.V(1).Info("Reconciliation of all the ServiceAccounts is successful")
	return nil
}

func (r *CSIScaleOperatorReconciler) getNodeDaemonSet(instance *csiscaleoperator.CSIScaleOperator) (*appsv1.DaemonSet, error) {
	node := &appsv1.DaemonSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.GetNameForResource(config.CSINode, instance.Name),
		Namespace: instance.Namespace,
	}, node)

	return node, err
}

/*
func (r *CSIScaleOperatorReconciler) restartControllerPod(logger logr.Logger, instance *csiscaleoperator.CSIScaleOperator) error {

	logger.Info("Restarting Controller Pod")
	controllerPod := &corev1.Pod{}
	controllerDeployment, err := r.getControllerDeployment(instance)
	if err != nil {
		logger.Error(err, "Failed to get controller deployment")
		return err
	}

	logger.Info("Controller requires restart",
		"ReadyReplicas", controllerDeployment.Status.ReadyReplicas,
		"Replicas", controllerDeployment.Status.Replicas)
	logger.Info("Restarting csi controller")

	err = r.getControllerPod(controllerDeployment, controllerPod)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		logger.Error(err, "Failed to get controller pod")
		return err
	}

	return r.restartControllerPodfromDeployment(logger, controllerDeployment, controllerPod)
}
*/
/*
func (r *CSIScaleOperatorReconciler) getControllerPod(controllerDeployment *appsv1.Deployment, controllerPod *corev1.Pod) error {
	controllerPodName := controllerPod.Name
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      controllerPodName,
		Namespace: controllerDeployment.Namespace,
	}, controllerPod)
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}
*/
/*
func (r *CSIScaleOperatorReconciler) restartControllerPodfromDeployment(logger logr.Logger,
	controllerDeployment *appsv1.Deployment, controllerPod *corev1.Pod) error {
	logger.Info("Controller requires restart",
		"ReadyReplicas", controllerDeployment.Status.ReadyReplicas,
		"Replicas", controllerDeployment.Status.Replicas)
	logger.Info("Restarting csi controller")

	return r.Client.Delete(context.TODO(), controllerPod)
}
*/

func (r *CSIScaleOperatorReconciler) rolloutRestartNode(node *appsv1.DaemonSet) error {
	restartedAt := fmt.Sprintf("%s/restartedAt", config.APIGroup)
	timestamp := time.Now().String()
	node.Spec.Template.ObjectMeta.Annotations[restartedAt] = timestamp
	return r.Client.Update(context.TODO(), node)
}

func (r *CSIScaleOperatorReconciler) getRestartedAtAnnotation(Annotations map[string]string) (string, string) {
	restartedAt := fmt.Sprintf("%s/restartedAt", config.APIGroup)
	for key, element := range Annotations {
		if key == restartedAt {
			return key, element
		}
	}
	return "", ""
}

/*
func (r *CSIScaleOperatorReconciler) getControllerDeployment(instance *csiscaleoperator.CSIScaleOperator) (*appsv1.Deployment, error) {
	controllerDeployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.GetNameForResource(config.CSIController, instance.Name),
		Namespace: instance.Namespace,
	}, controllerDeployment)

	return controllerDeployment, err
}
*/

func (r *CSIScaleOperatorReconciler) reconcileClusterRole(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("reconcileClusterRole")
	logger.Info("Creating the required ClusterRole resources.")

	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRole", "Name", cr.GetName())
			err = r.Client.Create(context.TODO(), cr)
			if err != nil {
				message := "Failed to create ClusterRole " + cr.GetName()
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceCreateError),
					Message: message,
				})
				return err
			}
		} else if err != nil {
			message := "Failed to get ClusterRole " + cr.GetName()
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return err
		} else {
			logger.Info("Clusterrole " + cr.GetName() + " already exists. Updating clusterrole.")
			err = r.Client.Update(context.TODO(), cr)
			if err != nil {
				message := "Failed to update ClusterRole " + cr.GetName()
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceUpdateError),
					Message: message,
				})
				return err
			}
		}
	}
	logger.V(1).Info("Reconciliation of ClusterRoles is successful")
	return nil
}

func (r *CSIScaleOperatorReconciler) getClusterRoles(instance *csiscaleoperator.CSIScaleOperator) []*rbacv1.ClusterRole {
	externalProvisioner := instance.GenerateProvisionerClusterRole()
	externalAttacher := instance.GenerateAttacherClusterRole()
	externalSnapshotter := instance.GenerateSnapshotterClusterRole()
	externalResizer := instance.GenerateResizerClusterRole()
	nodePlugin := instance.GenerateNodePluginClusterRole()
	// controllerSCC := instance.GenerateSCCForControllerClusterRole()
	// nodeSCC := instance.GenerateSCCForNodeClusterRole()

	return []*rbacv1.ClusterRole{
		externalProvisioner,
		externalAttacher,
		externalSnapshotter,
		externalResizer,
		nodePlugin,
		// controllerSCC,
		// nodeSCC,
	}
}

func (r *CSIScaleOperatorReconciler) getClusterRoleBindings(instance *csiscaleoperator.CSIScaleOperator) []*rbacv1.ClusterRoleBinding {
	externalProvisioner := instance.GenerateProvisionerClusterRoleBinding()
	externalAttacher := instance.GenerateAttacherClusterRoleBinding()
	externalSnapshotter := instance.GenerateSnapshotterClusterRoleBinding()
	externalResizer := instance.GenerateResizerClusterRoleBinding()
	nodePlugin := instance.GenerateNodePluginClusterRoleBinding()
	// controllerSCC := instance.GenerateSCCForControllerClusterRoleBinding()
	// nodeSCC := instance.GenerateSCCForNodeClusterRoleBinding()

	return []*rbacv1.ClusterRoleBinding{
		externalProvisioner,
		externalAttacher,
		externalSnapshotter,
		externalResizer,
		nodePlugin,
		//controllerSCC,
		//nodeSCC,
	}
}

func (r *CSIScaleOperatorReconciler) reconcileClusterRoleBinding(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("reconcileClusterRoleBinding")
	logger.Info("Creating the required ClusterRoleBinding resources.")

	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRoleBinding.", "Name", crb.GetName())
			err = r.Client.Create(context.TODO(), crb)
			if err != nil {
				message := "Failed to create ClusterRoleBinding resource " + crb.GetName()
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceCreateError),
					Message: message,
				})
				return err
			}
		} else if err != nil {
			message := "Failed to get ClusterRoleBinding " + crb.GetName()
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return err
		} else {
			// Resource already exists - don't requeue

			logger.Info("Clusterrolebinding " + crb.GetName() + " already exists. Updating clusterolebinding.")
			err = r.Client.Update(context.TODO(), crb)
			if err != nil {
				message := "Failed to update ClusterRoleBinding " + crb.GetName()
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceUpdateError),
					Message: message,
				})
				return err
			}
		}
	}
	logger.V(1).Info("Reconciliation of ClusterRoleBindings is successful")
	return nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRoleBindings(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteClusterRoleBindings")

	logger.Info("Deleting ClusterRoleBindings")
	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Continue looking for ClusterRoleBindings", "Name", crb.GetName())
			continue
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRoleBinding", "Name", crb.GetName())
			return err
		} else {
			logger.Info("Deleting ClusterRoleBinding", "Name", crb.GetName())
			if err := r.Client.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "Failed to delete ClusterRoleBinding", "Name", crb.GetName())
				return err
			}
		}
	}
	logger.Info("Exiting deleteClusterRoleBindings method.")
	return nil
}

// TODO: Status should show state of the driver.
// SetStatus() function assigns values to following fields of status sub-resource.
// Phase: ["", Creating, Running, Failed]
// ControllerReady: True/False
// NodeReady: True/False
// Version: Driver version picked from ibm-spectrum-scale-csi\controllers\config\constants.go
func (r *CSIScaleOperatorReconciler) SetStatus(instance *csiscaleoperator.CSIScaleOperator) error {

	logger := csiLog.WithName("SetStatus")
	logger.Info("Assigning values to status sub-resource object.")

	/*
		controllerPod := &corev1.Pod{}
		controllerDeployment, err := r.getControllerDeployment(instance)
		if err != nil {

			logger.Error(err, "failed to get controller deployment")
			return err
		}

		nodeDaemonSet, err := r.getNodeDaemonSet(instance)
		if err != nil {
			logger.Error(err, "Failed to get node daemonSet.")
			return err
		}

		crStatus.ControllerReady = r.isControllerReady(controllerDeployment)
		crStatus.NodeReady = r.isNodeReady(nodeDaemonSet)

		phase := csiv1.DriverPhaseNone
		if instance.Status.ControllerReady && instance.Status.NodeReady {
			phase = csiv1.DriverPhaseRunning
		} else {
			if !instance.Status.ControllerReady {
				err := r.getControllerPod(controllerDeployment, controllerPod)
				if err != nil {
					logger.Error(err, "failed to get controller pod")
					return err
				}

				if !r.areAllPodImagesSynced(controllerDeployment, controllerPod) {
					r.restartControllerPodfromDeployment(logger, controllerDeployment, controllerPod)

				}
			}
			phase = csiv1.DriverPhaseCreating
		}
		crStatus.Phase = phase
	*/

	crStatus.Versions = []csiv1.Version{
		{
			Name:    instance.Name,
			Version: config.DriverVersion,
		},
	}

	logger.V(1).Info("Setting status of CSIScaleOperator is successful")
	return nil
}

/*
func (r *CSIScaleOperatorReconciler) isControllerReady(controller *appsv1.Deployment) bool {
	logger := csiLog.WithName("isControllerReady")
	logMessage := "Controller status"
	logKey := "ReadyReplicas == MinControllerReplicas"
	logValueTrue := "True"
	logValueFalse := "False"
	logger.Info(logMessage+":", "ReadyReplicas:", controller.Status.ReadyReplicas)
	if controller.Status.ReadyReplicas == MinControllerReplicas {
		logger.Info(logMessage+":", logKey+":", logValueTrue)
		return true
	} else {
		logger.Info(logMessage+":", logKey+":", logValueFalse)
		return false
	}
}
*/
/*
func (r *CSIScaleOperatorReconciler) isNodeReady(node *appsv1.DaemonSet) bool {
	return node.Status.DesiredNumberScheduled == node.Status.NumberAvailable
}

*/
/*
func (r *CSIScaleOperatorReconciler) areAllPodImagesSynced(controllerDeployment *appsv1.Deployment, controllerPod *corev1.Pod) bool {

	logger := csiLog.WithName("areAllPodImagesSynced")
	statefulSetContainers := controllerDeployment.Spec.Template.Spec.Containers
	podContainers := controllerPod.Spec.Containers
	if len(statefulSetContainers) != len(podContainers) {
		return false
	}
	for i := 0; i < len(statefulSetContainers); i++ {
		statefulSetImage := statefulSetContainers[i].Image
		podImage := podContainers[i].Image

		if statefulSetImage != podImage {
			logger.Info("csi controller image not in sync",
				"statefulSetImage", statefulSetImage, "podImage", podImage)
			return false
		}
	}
	return true
}
*/

// TODO: Unused code. Remove if not required.
// Helper function to check for a string in a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func (r *CSIScaleOperatorReconciler) deleteCSIDriver(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteCSIDriver")

	logger.Info("Deleting CSIDriver")
	csiDriver := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      csiDriver.Name,
		Namespace: csiDriver.Namespace,
	}, found)
	if err == nil {
		logger.Info("Deleting CSIDriver", "Name", csiDriver.GetName())
		if err := r.Client.Delete(context.TODO(), found); err != nil {
			logger.Error(err, "Failed to delete CSIDriver", "Name", csiDriver.GetName())
			return err
		}
	} else if errors.IsNotFound(err) {
		logger.Info("CSIDriver not found for deletion")
		return nil
	} else {
		logger.Error(err, "Failed to get CSIDriver", "Name", csiDriver.GetName())
		return err
	}
	logger.Info("Deletion of CSIDriver is successful")
	return nil
}

// setENVIsOpenShift checks for an OpenShift service to identify whether
// CSI operator is running on an OpenShift cluster. If running on an
// OpenShift cluster, an environment variable is set, which is later
// used to reconcile resources needed only for OpenShift.
func setENVIsOpenShift(r *CSIScaleOperatorReconciler) {
	logger := csiLog.WithName("setENVIsOpenShift")
	_, isOpenShift := os.LookupEnv(config.ENVIsOpenShift)
	if !isOpenShift {
		service := &corev1.Service{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      "controller-manager",
			Namespace: "openshift-controller-manager",
		}, service)
		if err == nil {
			logger.Info("CSI Operator is running on an OpenShift cluster.")
			setEnvErr := os.Setenv(config.ENVIsOpenShift, "True")
			if setEnvErr != nil {
				logger.Error(err, "Error setting environment variable ENVIsOpenShift")
			}
		}
	}
}

// GetConsistencyGroupPrefix returns a universal unique ideintiier(UUID) of string format.
// For Redhat Openshift Cluster Platform, Cluster ID as string is returned.
// For Vanilla kubernetes cluster, generated UUID is returned.
func (r *CSIScaleOperatorReconciler) GetConsistencyGroupPrefix(instance *csiscaleoperator.CSIScaleOperator) string {
	logger := csiLog.WithName("GetConsistencyGroupPrefix")

	logger.Info("Checking if consistency group prefix is passed in CSIScaleOperator specs.")
	if instance.Spec.CGPrefix != "" {
		logger.Info("Consistency group prefix found in CSIScaleOperator specs.")
		return instance.Spec.CGPrefix
	}

	logger.Info("Consistency group prefix is not found in CSIScaleOperator specs.")
	logger.Info("Fetching cluster information.")
	_, isOpenShift := os.LookupEnv(config.ENVIsOpenShift)
	if !isOpenShift {
		logger.Info("Cluster is a Kubernetes Platform.")
		UUID := r.GenerateUUID()
		return UUID.String()
	}

	logger.Info("Cluster is Redhat Openshift Cluster Platform.")
	logger.Info("Fetching cluster ID from ClusterVersion resource.")
	CV := &configv1.ClusterVersion{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name: "version",
	}, CV)
	if err != nil {
		logger.Info("Unable to fetch the cluster scoped resource.")
		UUID := r.GenerateUUID()
		return UUID.String()
	}
	UUID := string(CV.Spec.ClusterID)
	return UUID

}

// GenerateUUID returns a new random UUID.
func (r *CSIScaleOperatorReconciler) GenerateUUID() uuid.UUID {
	logger := csiLog.WithName("GenerateUUID")
	logger.Info("Generating a unique cluster ID.")
	UUID := uuid.New()
	return UUID
}

func (r *CSIScaleOperatorReconciler) removeDeprecatedStatefulset(instance *csiscaleoperator.CSIScaleOperator, name string) error {
	logger := csiLog.WithName("removeDeprecatedStatefulset").WithValues("Name", name)
	logger.Info("Removing deprecated statefulset resource from the cluster.")

	STS := &appsv1.StatefulSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: instance.Namespace,
	}, STS)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("Statefulset resource not found in the cluster.")
	} else if err != nil {
		message := "Failed to get statefulset information from the cluster."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceReadError),
			Message: message,
		})
		return err
	} else {
		logger.Info("Found statefulset resource. Sidecar controllers as statefulsets are replaced by deployments in CSI >= 2.6.0. Removing statefulset.")
		if err := r.Client.Delete(context.TODO(), STS); err != nil {
			message := "Unable to delete " + name + " statefulset."
			logger.Error(err, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceDeleteError),
				Message: message,
			})
			return err
		}
	}
	return nil
}

func (r *CSIScaleOperatorReconciler) checkPrerequisite(instance *csiscaleoperator.CSIScaleOperator) (bool, error) {

	logger := csiLog.WithName("checkPrerequisite")
	logger.Info("Checking pre-requisites.")

	// get list of secrets from custom resource
	secrets := []string{}
	for _, cluster := range instance.Spec.Clusters {
		if len(cluster.Secrets) != 0 {
			secrets = append(secrets, cluster.Secrets)
		}
	}

	// get list of configMaps from custom resource
	configMaps := []string{}
	for _, cluster := range instance.Spec.Clusters {
		if len(cluster.Cacert) != 0 {
			configMaps = append(configMaps, cluster.Cacert)
		}
	}

	if len(secrets) != 0 {
		for _, secret := range secrets {
			if exists, err := r.resourceExists(instance, secret, corev1.ResourceSecrets.String()); !exists {
				return false, err
			}
			logger.Info(fmt.Sprintf("Secret resource %s found.", secret))
		}
	}

	if len(configMaps) != 0 {
		for _, configMap := range configMaps {
			if exists, err := r.resourceExists(instance, configMap, corev1.ResourceConfigMaps.String()); !exists {
				return false, err
			}
			logger.Info(fmt.Sprintf("ConfigMap resource %s found.", configMap))
		}
	}

	return true, nil
}

func (r *CSIScaleOperatorReconciler) resourceExists(instance *csiscaleoperator.CSIScaleOperator, name string, kind string) (bool, error) {

	logger := csiLog.WithName("resourceExists").WithValues("Kind", kind, "Name", name)
	logger.Info("Checking resource exists")

	var err error

	if kind == corev1.ResourceSecrets.String() {
		found := &corev1.Secret{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      name,
			Namespace: instance.Namespace,
		}, found)
	}

	if kind == corev1.ResourceConfigMaps.String() {
		found := &corev1.ConfigMap{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      name,
			Namespace: instance.Namespace,
		}, found)
	}

	if err != nil && errors.IsNotFound(err) {
		message := "Resource not found."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceNotFoundError),
			Message: message,
		})
		return false, err
	} else if err != nil {
		message := "Failed to get resource information from cluster."
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceReadError),
			Message: message,
		})
		return false, err
	} else {
		return true, nil
	}
}

func testClusterID() error {
	logger := csiLog.WithName("testClusterID")
	logger.Info("TEST: Getting ClusterID of primary cluster")

	clusterID, err := scaleConnMap[config.Primary].GetClusterId()
	if err != nil {
		logger.Error(err, "TEST: error in getting clusterID")
	} else {
		logger.Info("TEST: got clusterID successfully", "clusterID", clusterID)
	}
	return nil
}

func testRESTcalls(clusters []csiv1.CSICluster) error {
	logger := csiLog.WithName("testRESTcalls")
	for _, cluster := range clusters {
		var err error
		fsList, err := scaleConnMap[cluster.Id].ListFilesystems()
		if err != nil {
			logger.Error(err, "TEST: error in ListFilesystems")
		}
		logger.Info("TEST: ListFilesystems", "clusterID", cluster.Id, "fsList", fsList)

		fsetExists, err := scaleConnMap[cluster.Id].CheckIfFilesetExist("fs1", "fset1")
		if err != nil {
			logger.Error(err, "TEST: error in CheckIfFilesetExist")
		}
		logger.Info("TEST: CheckIfFilesetExist", "clusterID", cluster.Id, "fsetExists", fsetExists)

		fsetExists, err = scaleConnMap[cluster.Id].CheckIfFilesetExist("fs2", "fset1")
		if err != nil {
			logger.Error(err, "TEST: error in CheckIfFilesetExist")
		}
		logger.Info("TEST: CheckIfFilesetExist", "clusterID", cluster.Id, "fsetExists", fsetExists)
	}

	filesetList, err := scaleConnMap[config.Primary].ListFileset("fs1", "fset1")
	if err != nil {
		logger.Error(err, "TEST: error in fileset info")
	} else {
		logger.Info("TEST: got fileset fset1", "fset1", filesetList)
	}
	return nil
}

//newConnector creates and return a new connector to make REST calls for the passed cluster
func (r *CSIScaleOperatorReconciler) newConnector(instance *csiscaleoperator.CSIScaleOperator,
	cluster csiv1.CSICluster) (connectors.SpectrumScaleConnector, error) {
	logger := csiLog.WithName("newSpectrumScaleConnector")
	logger.Info("creating new SpectrumScaleConnector for cluster with", "ID", cluster.Id)

	var rest *connectors.SpectrumRestV2
	var tr *http.Transport
	username := ""
	password := ""

	if cluster.Secrets != "" {
		secret := &corev1.Secret{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      cluster.Secrets,
			Namespace: instance.Namespace,
		}, secret)
		if err != nil && errors.IsNotFound(err) {
			message := fmt.Sprintf("Secret %v not found", cluster.Secrets)
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceNotFoundError),
				Message: message,
			})
			return &connectors.SpectrumRestV2{}, err
		} else if err != nil {
			message := fmt.Sprintf("Failed to get secret %v", cluster.Secrets)
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return nil, err
		}
		username = string(secret.Data[config.SecretUsername])
		password = string(secret.Data[config.SecretPassword])
	}

	if cluster.SecureSslMode == true && cluster.Cacert != "" {
		configMap := &corev1.ConfigMap{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      cluster.Cacert,
			Namespace: instance.Namespace,
		}, configMap)

		if err != nil && errors.IsNotFound(err) {
			message := fmt.Sprintf("ConfigMap %v not found", cluster.Cacert)
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceNotFoundError),
				Message: message,
			})
			return nil, err
		} else if err != nil {
			message := fmt.Sprintf("Failed to get ConfigMap %v", cluster.Cacert)
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return nil, err
		}
		cacertValue := []byte(configMap.Data[cluster.Cacert])
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(cacertValue); !ok {
			return nil, fmt.Errorf("Parsing CA cert %v failed", cluster.Cacert)
		}
		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool, MinVersion: tls.VersionTLS12}}
		logger.Info("Created Spectrum Scale connector with SSL mode for guiHost(s)")

	} else {
		//#nosec G402 InsecureSkipVerify was requested by user.
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}} //nolint:gosec
		logger.Info("Created Spectrum Scale connector without SSL mode for guiHost(s)")
	}

	rest = &connectors.SpectrumRestV2{
		HTTPclient: &http.Client{
			Transport: tr,
			Timeout:   time.Second * config.HTTPClientTimeout,
		},
		User:          username,
		Password:      password,
		EndPointIndex: 0, //Use first GUI as primary by default
	}

	for i := range cluster.RestApi {
		guiHost := cluster.RestApi[i].GuiHost
		guiPort := cluster.RestApi[i].GuiPort
		if guiPort == 0 {
			guiPort = settings.DefaultGuiPort
		}
		endpoint := fmt.Sprintf("%s://%s:%d/", settings.GuiProtocol, guiHost, guiPort)
		rest.Endpoint = append(rest.Endpoint, endpoint)
	}
	return rest, nil
}

//getSpectrumScaleConnectors gets the connectors for all the clusters in driver
//manifest and sets those in scaleConnMap also checks if GUI is reachable and
//cluster ID is valid.
func (r *CSIScaleOperatorReconciler) getSpectrumScaleConnectors(instance *csiscaleoperator.CSIScaleOperator, cmExists bool, clustersStanzaModified bool) error {
	logger := csiLog.WithName("getSpectrumScaleConnectors")
	logger.Info("getting spectrum scale connectors")

	operatorRestarted := (len(scaleConnMap) == 0) && cmExists
	for _, cluster := range instance.Spec.Clusters {
		isPrimaryCluster := cluster.Primary != nil
		if !cmExists || clustersStanzaModified || operatorRestarted {
			//First pass or driver CR modified or connector map is empty (due
			//to operator restart) -> process all the clusters.
			//1.a First pass/cluster stanza modified: GUI for all clusters must
			//be reachable and cluster ID must be valid.
			//1.b Opeartor restarted: Primary cluster GUI must be reachable and
			//clusterID must match but these are not mandetory for non-primary
			//clusters. For non-primary clusters if any issue in connecting to
			//GUI or cluster ID validation -> log the error and go ahead.
			//TODO: Add event when non-primary cluster GUI is not reachable or
			//invalid cluster ID.
			//2. For other passes: only check for primary cluster.
			logger.Info("getting connector for the cluster", "ID", cluster.Id)
			_, connectorExists := scaleConnMap[cluster.Id]
			if !connectorExists || clustersStanzaModified {
				connector, err := r.newConnector(instance, cluster)
				if err != nil {
					return err
				}
				scaleConnMap[cluster.Id] = connector
				if isPrimaryCluster {
					scaleConnMap[config.Primary] = connector
				}
			}
			//check if GUI is reachable
			id, err := scaleConnMap[cluster.Id].GetClusterId()
			if err != nil {
				message := fmt.Sprintf("Error in connecting to GUI. Error: %v", err.Error())
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceReadError),
					Message: message,
				})
				if operatorRestarted && !isPrimaryCluster {
					//if operator is restarted and there is an error while
					//connecting to GUI of non-primary cluster -> do not return
					// error and continue with other clusters.
					continue
				}
				return err
			}
			//check if cluster ID from manifest matches with the one obtained from GUI
			if id != cluster.Id {
				message := fmt.Sprintf("The cluster ID %v in driver manifest does not match with the cluster ID %v obtained from cluster", cluster.Id, id)
				logger.Error(err, message)
				// TODO: Add event.
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceConfigError),
					Message: message,
				})
				if operatorRestarted && !isPrimaryCluster {
					//if operator is restarted and cluster ID validation
					//fails for non-primary cluster -> do not return
					//error and continue with other clusters.
					continue
				}
				return fmt.Errorf(message)
			}
		} else {
			//pass 2 or later (configmap exists) and cluster stanza is not modified and
			//connector map is not empty (operator is not restarted)  -> process only
			//primary cluster.
			if isPrimaryCluster {
				if _, connectorExists := scaleConnMap[cluster.Id]; !connectorExists {
					logger.Info("getting connector for the primary cluster", "ID", cluster.Id)
					connector, err := r.newConnector(instance, cluster)
					if err != nil {
						return err
					}
					scaleConnMap[cluster.Id] = connector
					scaleConnMap[config.Primary] = connector
				}
				//check if GUI is reachable
				_, err := scaleConnMap[cluster.Id].GetClusterId()
				if err != nil {
					message := fmt.Sprintf("Error in connecting to GUI. Error: %v", err.Error())
					logger.Error(err, message)
					// TODO: Add event.
					meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
						Type:    string(config.StatusConditionSuccess),
						Status:  metav1.ConditionFalse,
						Reason:  string(csiv1.ResourceReadError),
						Message: message,
					})
					return err
				}
				//Cluster ID validation is not required again as it is already done
				//in 1st pass.
				//Return from here as primary cluster is already processed
				return nil
			} else {
				//if not primary - don't process this cluster for 2nd or later pass
				continue
			}
		}
	}
	return nil
}

//handlePrimaryFSandFileset checks if primary FS exists, also checkes if primary fileset exists.
//If primary fileset does not exist, it is created and also if a driectory
//to store symlinks is created if it does not exist.
func (r *CSIScaleOperatorReconciler) handlePrimaryFSandFileset(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("handlePrimaryFSandFileset")
	primary := r.getPrimaryCluster(instance)
	if primary == nil {
		message := "No primary is defined in driver manifest"
		err := fmt.Errorf(message)
		logger.Error(err, "")
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceConfigError),
			Message: message,
		})
		return err
	}

	sc := scaleConnMap[config.Primary]

	// check if primary filesystem exists
	fsMountInfo, err := sc.GetFilesystemMountDetails(primary.PrimaryFs)
	if err != nil {
		message := fmt.Sprintf("Error in getting filesystem mount details for %s. Error %v", primary.PrimaryFs, err.Error())
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceConfigError),
			Message: message,
		})
		return err

	}

	// In case primary fset value is not specified in configuation then use default
	if primary.PrimaryFset == "" {
		primary.PrimaryFset = config.DefaultPrimaryFileset
		logger.Info("Primary fileset is not specified", "using default primary fileset %s", config.DefaultPrimaryFileset)
	}

	primaryFSMount := fsMountInfo.MountPoint

	// Get FS name on owning cluster
	// Examples of remoteDeviceName:
	// 1. Local FS fs2 -
	//		"remoteDeviceName" : "<scale local cluster name>:fs2"
	// 2. Remote FS fs1 which can be mounted locally with a different name
	//		"remoteDeviceName" : "<scale remote cluster name>:fs1"
	remoteDeviceName := strings.Split(fsMountInfo.RemoteDeviceName, ":")
	fsNameOnOwningCluster := remoteDeviceName[len(remoteDeviceName)-1]

	// //check if multiple GUIs are passed
	// if len(cluster.RestAPI) > 1 {
	// 	err := driver.cs.checkGuiHASupport(sc)
	// 	if err != nil {
	// 		return nil, scaleConfig, cluster.Primary, err
	// 	}
	// }

	if primary.RemoteCluster != "" {
		//if remote cluster is present, use connector of remote cluster
		sc = scaleConnMap[primary.RemoteCluster]
		if fsNameOnOwningCluster == "" {
			message := "Failed to get the name of remote filesystem"
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return fmt.Errorf(message)
		}
	}

	//check if primary filesystem exists on remote cluster and mounted on atleast one node
	fsMountInfo, err = sc.GetFilesystemMountDetails(fsNameOnOwningCluster)
	if err != nil {
		message := fmt.Sprintf("Error in getting filesystem details for %s. Error: %v", fsNameOnOwningCluster, err.Error())
		logger.Error(err, message)
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceReadError),
			Message: message,
		})
		return err
	}

	fsMountPoint := fsMountInfo.MountPoint

	fsetLinkPath, err := createPrimaryFileset(sc, fsNameOnOwningCluster, fsMountPoint, primary.PrimaryFset, primary.InodeLimit)
	if err != nil {
		message := fmt.Sprintf("Error in creating primary fileset %s. Error: %v", primary.PrimaryFset, err.Error())
		logger.Error(err, message)
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceCreateError),
			Message: message,
		})
		return err
	}

	// In case primary FS is remotely mounted, run fileset refresh task on primary cluster
	if primary.RemoteCluster != "" {
		_, err := scaleConnMap[config.Primary].ListFileset(primary.PrimaryFs, primary.PrimaryFset)
		if err != nil {
			logger.Info("Primary fileset is not visible on primary cluster. Running fileset refresh task", "fileset name", primary.PrimaryFset)
			err = scaleConnMap[config.Primary].FilesetRefreshTask()
			if err != nil {
				message := fmt.Sprintf("Error in fileset refresh task. Error: %v", err.Error())
				logger.Error(err, message)
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceSyncError),
					Message: message,
				})
				return err
			}
		}

		// retry listing fileset again after some time after refresh
		time.Sleep(8 * time.Second)
		_, err = scaleConnMap[config.Primary].ListFileset(primary.PrimaryFs, primary.PrimaryFset)
		if err != nil {
			message := fmt.Sprintf("Primary fileset %s not visible on primary cluster even after running fileset refresh task. Error: %v", primary.PrimaryFset, err.Error())
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceSyncError),
				Message: message,
			})
			return err
		}
	}

	//A directory can be created from accessing cluster, so get the path on accessing cluster
	if fsMountPoint != primaryFSMount {
		fsetLinkPath = strings.Replace(fsetLinkPath, fsMountPoint, primaryFSMount, 1)
	}

	// Create directory where volume symlinks will reside
	symlinkDirPath, symlinkDirRelativePath, err := createSymlinksDir(scaleConnMap[config.Primary], primary.PrimaryFs, primaryFSMount, fsetLinkPath)
	if err != nil {
		message := fmt.Sprintf("Error in creating volumes directory %s. Error: %v", config.SymlinkDir, err.Error())
		logger.Error(err, message)
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceCreateError),
			Message: message,
		})
		return err
	}

	logger.Info("symlink directory paths:", "symlinkDirPath", symlinkDirPath, "symlinkDirRelativePath", symlinkDirRelativePath)
	//TODO: add symlinkDirPath and symlinkDirRelativePath in  driver pod env vars as required
	logger.Info("Primary FS and fileset are processed successfully")
	return nil
}

//getPrimaryCluster returns primary cluster of the passed instance.
func (r *CSIScaleOperatorReconciler) getPrimaryCluster(instance *csiscaleoperator.CSIScaleOperator) *v1.CSIFilesystem {
	var primary *v1.CSIFilesystem
	for _, cluster := range instance.Spec.Clusters {
		if cluster.Primary != nil {
			primary = cluster.Primary
		}
	}
	return primary
}

//createPrimaryFileset creates a primary fileset and returns it's the
//path where it is linked. If primary fileset exists and is already linked,
//the link path is returned. If primary fileset already exists and not linked,
//it is linked and link path is returned.
func createPrimaryFileset(sc connectors.SpectrumScaleConnector, fsNameOnOwningCluster string,
	fsMountPoint string, filesetName string, inodeLimit string) (string, error) {

	logger := csiLog.WithName("createPrimaryFileset")
	logger.Info("Creating primary fileset", " primaryFS", fsNameOnOwningCluster,
		"mount point", fsMountPoint, "filesetName", filesetName)

	newLinkPath := path.Join(fsMountPoint, filesetName) //Link path to set if the fileset is not linked

	// create primary fileset if not already created
	fsetResponse, err := sc.ListFileset(fsNameOnOwningCluster, filesetName)
	if err != nil {
		logger.Info("Primary fileset not found, so creating it", "fileseName", filesetName)
		opts := make(map[string]interface{})
		if inodeLimit != "" {
			opts[connectors.UserSpecifiedInodeLimit] = inodeLimit
		}

		err = sc.CreateFileset(fsNameOnOwningCluster, filesetName, opts)
		if err != nil {
			message := fmt.Sprintf("Unable to create primary fileset %s. Error: %v", filesetName, err.Error())
			logger.Error(err, message)
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceCreateError),
				Message: message,
			})
			return "", err
		}
	} else {
		linkPath := fsetResponse.Config.Path
		if linkPath == "" || linkPath == "--" {
			logger.Info("Primary fileset not linked. Linking it", "filesetName", filesetName)
			err = sc.LinkFileset(fsNameOnOwningCluster, filesetName, newLinkPath)
			if err != nil {
				message := fmt.Sprintf("Unable to link primary fileset %s. Error: %v", filesetName, err.Error())
				logger.Error(err, message)
				meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
					Type:    string(config.StatusConditionSuccess),
					Status:  metav1.ConditionFalse,
					Reason:  string(csiv1.ResourceUpdateError),
					Message: message,
				})
				return "", err
			} else {
				logger.Info("Linked primary fileset", "filesetName", filesetName, "linkpath", newLinkPath)
			}
		} else {
			logger.Info("Primary fileset exists and linked", "filesetName", filesetName, "linkpath", linkPath)
		}
	}
	return newLinkPath, nil
}

//createSymlinksDir creates a .volumes directory on the fileset path fsetLinkPath,
//and returns absolute, relative paths and error if there is any.
func createSymlinksDir(sc connectors.SpectrumScaleConnector, fs string, fsMountPath string,
	fsetLinkPath string) (string, string, error) {

	logger := csiLog.WithName("createSymlinkPath")
	logger.Info("Creating a directory for symlinks", "directory", config.SymlinkDir,
		"filesystem", fs, "fsMountPath", fsMountPath, "filesetlinkpath", fsetLinkPath)

	fsetRelativePath := strings.Replace(fsetLinkPath, fsMountPath, "", 1)
	fsetRelativePath = strings.Trim(fsetRelativePath, "!/")
	fsetLinkPath = strings.TrimSuffix(fsetLinkPath, "/")

	symlinkDirPath := fmt.Sprintf("%s/%s", fsetLinkPath, config.SymlinkDir)
	symlinkDirRelativePath := fmt.Sprintf("%s/%s", fsetRelativePath, config.SymlinkDir)

	err := sc.MakeDirectory(fs, symlinkDirRelativePath, config.DefaultUID, config.DefaultGID)
	if err != nil {
		message := fmt.Sprintf("Directory creation failed. Filesystem: %s, relative path: %s. Error: %v", fs, symlinkDirRelativePath, err.Error())
		logger.Error(err, message)
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceCreateError),
			Message: message,
		})
		return symlinkDirPath, symlinkDirRelativePath, err
	}

	return symlinkDirPath, symlinkDirRelativePath, nil
}

//ValidateCRParams validates driver configuration parameters in the operator instance, returns error if any validation fails
func ValidateCRParams(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("ValidateCRParams")
	logger.Info("Validating CR for driver config params")

	if len(instance.Spec.Clusters) == 0 {
		return fmt.Errorf("Missing cluster information in Spectrum Scale configuration")
	}

	primaryClusterFound, issueFound := false, false
	rClusterForPrimaryFS := ""
	var nonPrimaryClusters = make([]string, len(instance.Spec.Clusters))

	for i := 0; i < len(instance.Spec.Clusters); i++ {
		cluster := instance.Spec.Clusters[i]

		if cluster.Id == "" {
			issueFound = true
			logger.Error(fmt.Errorf("Mandatory parameter 'id' is not specified"), "")
		}
		if len(cluster.RestApi) == 0 {
			issueFound = true
			logger.Error(fmt.Errorf("Mandatory section 'restApi' is not specified for cluster %v", cluster.Id), "")
		}
		if len(cluster.RestApi) != 0 && cluster.RestApi[0].GuiHost == "" {
			issueFound = true
			logger.Error(fmt.Errorf("Mandatory parameter 'guiHost' is not specified for cluster %v", cluster.Id), "")
		}

		if cluster.Primary != nil && *cluster.Primary != (v1.CSIFilesystem{}) {
			if primaryClusterFound {
				issueFound = true
				logger.Error(fmt.Errorf("More than one primary clusters specified"), "")
			}

			primaryClusterFound = true
			if cluster.Primary.PrimaryFs == "" {
				issueFound = true
				logger.Error(fmt.Errorf("Mandatory parameter 'primaryFs' is not specified for primary cluster %v", cluster.Id), "")
			}

			rClusterForPrimaryFS = cluster.Primary.RemoteCluster
		} else {
			//when its a not primary cluster
			nonPrimaryClusters[i] = cluster.Id
		}

		if cluster.Secrets == "" {
			issueFound = true
			logger.Error(fmt.Errorf("Mandatory parameter 'secrets' is not specified for cluster %v", cluster.Id), "")
		}

		if cluster.SecureSslMode && cluster.Cacert == "" {
			issueFound = true
			logger.Error(fmt.Errorf("CA certificate not specified in secure SSL mode for cluster %v", cluster.Id), "")
		}
	}

	if !primaryClusterFound {
		issueFound = true
		logger.Error(fmt.Errorf("No primary clusters specified"), "")
	}

	if rClusterForPrimaryFS != "" && stringInSlice(rClusterForPrimaryFS, nonPrimaryClusters) {
		issueFound = true
		logger.Error(fmt.Errorf("Remote cluster specified for primary filesystem: %s, but no definition found for it in config", rClusterForPrimaryFS), "")
	}

	if issueFound {
		message := "one or more issues found in Spectrum scale csi driver configuration, check Spectrum Scale csi operator logs"
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceConfigError),
			Message: message,
		})
		return fmt.Errorf(message)
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(b, a) {
			return true
		}
	}
	return false
}
