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
	"encoding/json"
	"fmt"
	"os"
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

	csiv1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	config "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	csiscaleoperator "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/internal/csiscaleoperator"
	clustersyncer "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/syncer"
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
	err := r.Client.Get(ctx, req.NamespacedName, instanceUnwrap)
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

	watchResources[corev1.ResourceConfigMaps.String()][config.CSIEnvVarConfigMap] = true

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
	// The CSIConfigMap has the data which is last applied, so
	// compare the clusters data of current driver instance with
	// CSIConfigMap data and rollout restart if there is a difference.
	configMap := &corev1.ConfigMap{}
	cerr := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.CSIConfigMap,
		Namespace: req.Namespace,
	}, configMap)
	if cerr != nil {
		if !errors.IsNotFound(cerr) {
			message := "Failed to get ConfigMap resource " + config.CSIConfigMap
			logger.Error(cerr, message)
			// TODO: Add event.
			meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
				Type:    string(config.StatusConditionSuccess),
				Status:  metav1.ConditionFalse,
				Reason:  string(csiv1.ResourceReadError),
				Message: message,
			})
			return ctrl.Result{}, cerr
		}
	} else {
		clustersBytes, err := json.Marshal(&instance.Spec.Clusters)
		if err != nil {
			logger.Error(err, "Failed to marshal clusters data of this instance")
			return ctrl.Result{}, err
		}
		clustersString := string(clustersBytes)
		configMapDataBytes, err := json.Marshal(&configMap.Data)
		if err != nil {
			logger.Error(err, "Failed to marshal data of ConfigMap "+config.CSIConfigMap)
			return ctrl.Result{}, err
		}
		configMapDataString := string(configMapDataBytes)
		configMapDataString = strings.Replace(configMapDataString, " ", "", -1)
		configMapDataString = strings.Replace(configMapDataString, "\\\"", "\"", -1)

		if !strings.Contains(configMapDataString, clustersString) {
			logger.Info("Some of the cluster fields of CSIScaleOperator instance are changed, so restarting node plugin pods")
			daemonSet, derr := r.getNodeDaemonSet(instance)
			if derr != nil {
				if !errors.IsNotFound(derr) {
					message := "Failed to get node plugin Daemonset"
					logger.Error(derr, message)
					// TODO: Add event.
					meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
						Type:    string(config.StatusConditionSuccess),
						Status:  metav1.ConditionFalse,
						Reason:  string(csiv1.ResourceReadError),
						Message: message,
					})
					return ctrl.Result{}, derr
				}
			} else {
				err = r.rolloutRestartNode(daemonSet)
				if err != nil {
					logger.Error(err, "Failed to rollout restart of node plugin pods")
				} else {
					daemonSetRestartedKey, daemonSetRestartedValue = r.getRestartedAtAnnotation(daemonSet.Spec.Template.ObjectMeta.Annotations)
				}
			}
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

	cmData := map[string]string{}
	cm, err := r.getConfigMap(instance, config.CSIEnvVarConfigMap)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if err == nil && len(cm.Data) != 0 {
		cmData = parseConfigMap(cm)
	} else {
		logger.Info("Optional configmap is either not found or is empty, skipped parsing it", "configmap", config.CSIEnvVarConfigMap)
	}

	csiNodeSyncer := clustersyncer.GetCSIDaemonsetSyncer(r.Client, r.Scheme, instance, daemonSetRestartedKey, daemonSetRestartedValue, CGPrefix, cmData)
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

	//update daemonset and restart its pods(driver pods) when optional configmap ibm-spectrum-scale-csi having valid env vars
	//ie. in the format VAR_DRIVER_ENV: VAL, is created/deleted
	//Do not restart driver pods when the configmap contains invalid Envs.
	implicitRestartOnCreateDelete := func(cfgmapData map[string]string) bool {
		for key := range cfgmapData {
			if strings.HasPrefix(strings.ToUpper(key), config.CSIEnvVarPrefix) {
				return true
			}
		}
		return false
	}

	//Allow implicit restart of driver pods when returns true
	//implicit restart occurs automatically based on daemonset updateStretegy when a daemonset gets updated
	implicitRestartOnUpdate := func(oldCfgMapData, newCfgMapData map[string]string) bool {
		for key, newVal := range newCfgMapData {
			//Allow restart of driver pods when a new valid env var is found or the value of existing valid env var is updated
			if oldVal, ok := oldCfgMapData[key]; !ok {
				if strings.HasPrefix(strings.ToUpper(key), config.CSIEnvVarPrefix) {
					return true
				}
			} else if oldVal != newVal && strings.HasPrefix(strings.ToUpper(key), config.CSIEnvVarPrefix) {
				return true
			}
		}

		for key := range oldCfgMapData {
			//look for deleted valid env vars of the old configmap in the new configmap
			//if deleted restart driver pods
			if _, ok := newCfgMapData[key]; !ok {
				if strings.HasPrefix(strings.ToUpper(key), config.CSIEnvVarPrefix) {
					return true
				}
			}
		}
		return false
	}

	predicateFuncs := func(resourceKind string) predicate.Funcs {
		return predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				if isCSIResource(e.Object.GetName(), resourceKind) {
					if resourceKind == corev1.ResourceConfigMaps.String() && e.Object.GetName() == config.CSIEnvVarConfigMap {
						return implicitRestartOnCreateDelete(e.Object.(*corev1.ConfigMap).Data)
					} else {
						r.restartDriverPods(mgr, "created", resourceKind, e.Object.GetName())
					}
					return true
				}
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				if isCSIResource(e.ObjectNew.GetName(), resourceKind) {
					if resourceKind == corev1.ResourceSecrets.String() && !reflect.DeepEqual(e.ObjectOld.(*corev1.Secret).Data, e.ObjectNew.(*corev1.Secret).Data) {
						r.restartDriverPods(mgr, "updated", resourceKind, e.ObjectOld.GetName())
					} else if resourceKind == corev1.ResourceConfigMaps.String() {
						if e.ObjectNew.GetName() == config.CSIEnvVarConfigMap && !reflect.DeepEqual(e.ObjectOld.(*corev1.ConfigMap).Data, e.ObjectNew.(*corev1.ConfigMap).Data) {
							return implicitRestartOnUpdate(e.ObjectOld.(*corev1.ConfigMap).Data, e.ObjectNew.(*corev1.ConfigMap).Data)
						}
					}
					return true
				}
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				if isCSIResource(e.Object.GetName(), resourceKind) {
					if resourceKind == corev1.ResourceConfigMaps.String() && e.Object.GetName() == config.CSIEnvVarConfigMap {
						return implicitRestartOnCreateDelete(e.Object.(*corev1.ConfigMap).Data)
					} else {
						r.restartDriverPods(mgr, "deleted", resourceKind, e.Object.GetName())
					}
					return true
				}
				return false
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

// getConfigMap fetches data from the "ibm-spectrum-scale-csi-config" configmap from the cluster
// and returns a configmap reference.
func (r *CSIScaleOperatorReconciler) getConfigMap(instance *csiscaleoperator.CSIScaleOperator, name string) (*corev1.ConfigMap, error) {

	logger := csiLog.WithName("getConfigMap").WithValues("Kind", corev1.ResourceConfigMaps, "Name", name)
	logger.Info("Reading optional CSI configmap resource from the cluster.")

	cm := &corev1.ConfigMap{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: instance.Namespace,
	}, cm)
	if err != nil && errors.IsNotFound(err) {
		message := fmt.Sprintf("Optional configmap resource %s not found.", name)
		logger.Info(message)
	} else if err != nil {
		message := fmt.Sprintf("Failed to get configmap %s information from cluster.", name)
		logger.Error(err, message)
		// TODO: Add event.
		meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
			Type:    string(config.StatusConditionSuccess),
			Status:  metav1.ConditionFalse,
			Reason:  string(csiv1.ResourceReadError),
			Message: message,
		})
	}
	return cm, err
}

// parseConfigMap parses the data in the configMap in the desired format(VAR_DRIVER_ENV_NAME: VALUE to ENV_NAME: VALUE).
func parseConfigMap(cm *corev1.ConfigMap) map[string]string {

	logger := csiLog.WithName("parseConfigMap").WithValues("Name", config.CSIEnvVarConfigMap)
	logger.Info("Parsing the data from the optional configmap.", "configmap", config.CSIEnvVarConfigMap)

	data := map[string]string{}
	invalidEnv := []string{}
	for key, value := range cm.Data {
		if strings.HasPrefix(strings.ToUpper(key), config.CSIEnvVarPrefix) {
			data[strings.ToUpper(key[11:])] = value
		} else {
			invalidEnv = append(invalidEnv, key)
		}
	}
	logger.Info("Invalid environment variables in the configmap, only the valid ones will be set on driver pods", "Invalid Env Vars", invalidEnv)
	logger.Info("Parsing the data from the optional configmap is successful", "configmap", config.CSIEnvVarConfigMap)
	return data
}
