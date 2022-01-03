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
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	securityv1 "github.com/openshift/api/security/v1"
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
	"k8s.io/client-go/util/workqueue"
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

// define labels that users need to add to CSI secrets.
var secretsLabels = map[string]string{
	config.LabelProduct: string(config.Product),
}

type reconciler func(instance *csiscaleoperator.CSIScaleOperator) error

var crStatus = csiv1.CSIScaleOperatorStatus{}

// +kubebuilder:rbac:groups=csi.ibm.com,resources=*,verbs=*

// +kubebuilder:rbac:groups="",resources={pods,persistentvolumeclaims,services,endpoints,events,configmaps,secrets,secrets/status,services/finalizers,serviceaccounts},verbs=*
// TODO: Does the operator need to access to all the resources mentioned above?
// TODO: Does all resources mentioned above required delete/patch/update permissions?

// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources={clusterroles,clusterrolebindings},verbs=*
// +kubebuilder:rbac:groups="apps",resources={deployments,daemonsets,replicasets,statefulsets},verbs=*
// +kubebuilder:rbac:groups="apps",resourceNames=ibm-spectrum-scale-csi-operator,resources=deployments/finalizers,verbs=get;update
// +kubebuilder:rbac:groups="storage.k8s.io",resources={volumeattachments,storageclasses,csidrivers},verbs=*
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=*

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
		logger.Error(err, "failed to get CSIScaleOperator.")
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

	logger.Info("adding Finalizer")
	if err := r.addFinalizerIfNotPresent(instance); err != nil {
		logger.Error(err, "couldn't add Finalizer")
		return ctrl.Result{}, err
	}

	logger.Info("checking if CSIScaleOperator object got deleted")
	if !instance.GetDeletionTimestamp().IsZero() {

		logger.Info("attempting cleanup of CSI driver")
		isFinalizerExists, err := r.hasFinalizer(instance)
		if err != nil {
			logger.Error(err, "finalizer check failed")
			return ctrl.Result{}, err
		}

		if !isFinalizerExists {
			logger.Error(err, "no finalizer was found")
			return ctrl.Result{}, nil
		}

		if err := r.deleteClusterRolesAndBindings(instance); err != nil {
			logger.Error(err, "failed to delete ClusterRoles and ClusterRolesBindings")
			return ctrl.Result{}, err
		}

		if err := r.deleteCSIDriver(instance); err != nil {
			logger.Error(err, "failed to delete CSIDriver")
			return ctrl.Result{}, err
		}

		if err := r.removeFinalizer(instance); err != nil {
			logger.Error(err, "failed to remove Finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Removed CSI driver successfully")
		return ctrl.Result{}, nil
	}

	logger.Info("create resources")
	// create the resources which never change if not exist
	for _, rec := range []reconciler{
		r.reconcileCSIDriver,
		r.reconcileServiceAccount,
		r.reconcileClusterRole,
		r.reconcileClusterRoleBinding,
		r.reconcileSecurityContextConstraint,
	} {
		if err = rec(instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info("Creation of the resources which never change is successful")

	// Synchronizing the resources which change over time.
	// Resource list:
	// 1. Cluster configMap
	// 2. Attacher statefulset
	// 3. Provisioner statefulset
	// 4. Snapshotter statefulset
	// 5. Resizer statefulset
	// 6. Driver daemonset
	logger.Info("Synchronizing the resources which change over time.")

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

	// Synchronizing attacher statefulset
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

	// Synchronizing provisioner statefulset
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

	// Synchronizing snapshotter statefulset
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

	// Synchronizing resizer statefulset
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

	csiNodeSyncer := clustersyncer.GetCSIDaemonsetSyncer(r.Client, r.Scheme, instance, daemonSetRestartedKey, daemonSetRestartedValue)
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

// SetupWithManager sets up the controller with the Manager.
func (r *CSIScaleOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {

	logger := csiLog.WithName("SetupWithManager")

	logger.Info("running IBM Spectrum Scale CSI operator", "version", config.OperatorVersion)
	logger.Info("setting up the controller with the manager.")
	p, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: secretsLabels,
	})
	if err != nil {
		logger.Error(err, "Unable to create label selector predicate. Controller instance will not be created.")
		return err
	}
	preds := builder.WithPredicates(p)

	CSIReconcileRequestFunc := func() []reconcile.Request {
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

	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1.CSIScaleOperator{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.Secret{}},
			handler.Funcs{
				CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
					for _, request := range CSIReconcileRequestFunc() {
						q.Add(request)
					}
				},
				UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
					// TODO: Check if data in e.ObjectNew and e.ObjectOld are same or modified.
					// Daemon set should update only when data in secret is modified.
					// Daemon set should not update when secret is updated but not the data.
					// e.g Secret type, resource version etc is modified.
					// TODO: Update only those daemon set which are in the same namespace as the secret that triggered the event.
					for _, daemonSet := range CSIDaemonListFunc() {
						logger.Info("Secrets were modified. Daemon Set will be updated. Restarting node specific pods.")
						err = r.rolloutRestartNode(&daemonSet)
						if err != nil {
							logger.Error(err, "Unable to update daemon set. Please restart node specific pods manually.")
						} else {
							daemonSetRestartedKey, daemonSetRestartedValue = r.getRestartedAtAnnotation(daemonSet.Spec.Template.ObjectMeta.Annotations)
						}
					}

					for _, request := range CSIReconcileRequestFunc() {
						q.Add(request)
					}
				},
				DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
					for _, request := range CSIReconcileRequestFunc() {
						q.Add(request)
					}
				},
			}, preds).
		Owns(&appsv1.Deployment{}).
		Complete(r)
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
		logger.Error(err, "no finalizer found")
		return false, err
	}

	logger.Info("returning with finalizer", "name", finalizerName)
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
		logger.Error(err, "couldn't get finalizer")
		return err
	}

	accessor.SetFinalizers(removeListEntry(accessor.GetFinalizers(), finalizerName))
	if err := r.Client.Update(context.TODO(), instance.Unwrap()); err != nil {
		logger.Error(err, "failed to remove", "finalizer", finalizerName, "from", accessor.GetName())
		return err
	}

	logger.Info("finalizer was removed")
	return nil
}

func (r *CSIScaleOperatorReconciler) addFinalizerIfNotPresent(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("addFinalizerIfNotPresent")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		logger.Error(err, "failed to get finalizer name")
		return err
	}

	if !Contains(accessor.GetFinalizers(), finalizerName) {
		logger.Info("adding", "finalizer", finalizerName, "on", accessor.GetName())
		accessor.SetFinalizers(append(accessor.GetFinalizers(), finalizerName))

		if err := r.Client.Update(context.TODO(), instance.Unwrap()); err != nil {
			logger.Error(err, "failed to add", "finalizer", finalizerName, "on", accessor.GetName())
			return err
		}
	}
	logger.Info("finalizer was added with", "name", finalizerName)
	return nil
}

func (r *CSIScaleOperatorReconciler) getAccessorAndFinalizerName(instance *csiscaleoperator.CSIScaleOperator) (metav1.Object, string, error) {
	logger := csiLog.WithName("getAccessorAndFinalizerName")

	finalizerName := config.CSIFinalizer

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "failed to get meta information of instance")
		return nil, "", err
	}

	logger.Info("got finalizer with", "name", finalizerName)
	return accessor, finalizerName, nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRolesAndBindings(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteClusterRolesAndBindings")

	logger.Info("calling deleteClusterRoleBindings()")
	if err := r.deleteClusterRoleBindings(instance); err != nil {
		logger.Error(err, "deletion of ClusterRoleBindings failed")
		return err
	}

	logger.Info("calling deleteClusterRoles()")
	if err := r.deleteClusterRoles(instance); err != nil {
		logger.Error(err, "deletion of ClusterRoles failed")
		return err
	}

	logger.Info("deletion of ClusterRoles and ClusterRoleBindings succeeded")
	return nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRoles(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteClusterRoles")

	logger.Info("deleting ClusterRoles")
	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("continuing working on ClusterRoles for deletion")
			continue
		} else if err != nil {
			logger.Error(err, "failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			logger.Info("deleting ClusterRole", "Name", cr.GetName())
			if err := r.Client.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}
	logger.Info("exiting deleteClusterRoles()")
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
	logger.V(1).Info("Exiting reconcileCSIDriver() method.")
	return nil
}

func (r *CSIScaleOperatorReconciler) reconcileServiceAccount(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("reconcileServiceAccount")
	logger.Info("Creating the required ServiceAccount resources.")

	controller := instance.GenerateControllerServiceAccount()
	node := instance.GenerateNodeServiceAccount()
	attacher := instance.GenerateAttacherServiceAccount()
	provisioner := instance.GenerateProvisionerServiceAccount()
	snapshotter := instance.GenerateSnapshotterServiceAccount()
	resizer := instance.GenerateResizerServiceAccount()

	controllerServiceAccountName := config.GetNameForResource(config.CSIControllerServiceAccount, instance.Name)
	nodeServiceAccountName := config.GetNameForResource(config.CSINodeServiceAccount, instance.Name)
	// attacherServiceAccountName := config.GetNameForResource(config.CSIAttacherServiceAccount, instance.Name)
	// provisionerServiceAccountName := config.GetNameForResource(config.CSIProvisionerServiceAccount, instance.Name)
	// snapshotterServiceAccountName := config.GetNameForResource(config.CSISnapshotterServiceAccount, instance.Name)

	for _, sa := range []*corev1.ServiceAccount{
		controller,
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
			logger.Info("Creation of ServiceAccount " + sa.GetName() + "is successful")

			if controllerServiceAccountName == sa.Name {
				rErr := r.restartControllerPod(logger, instance)

				if rErr != nil {
					message := "Failed to restart controller pod."
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
			}
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
			logger.Info("ServiceAccount already exists.", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
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

func (r *CSIScaleOperatorReconciler) restartControllerPod(logger logr.Logger, instance *csiscaleoperator.CSIScaleOperator) error {

	logger.Info("restarting Controller Pod")
	controllerPod := &corev1.Pod{}
	controllerDeployment, err := r.getControllerDeployment(instance)
	if err != nil {
		logger.Error(err, "failed to get controller deployment")
		return err
	}

	logger.Info("controller requires restart",
		"ReadyReplicas", controllerDeployment.Status.ReadyReplicas,
		"Replicas", controllerDeployment.Status.Replicas)
	logger.Info("restarting csi controller")

	err = r.getControllerPod(controllerDeployment, controllerPod)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		logger.Error(err, "failed to get controller pod")
		return err
	}

	return r.restartControllerPodfromDeployment(logger, controllerDeployment, controllerPod)
}

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

func (r *CSIScaleOperatorReconciler) restartControllerPodfromDeployment(logger logr.Logger,
	controllerDeployment *appsv1.Deployment, controllerPod *corev1.Pod) error {
	logger.Info("controller requires restart",
		"ReadyReplicas", controllerDeployment.Status.ReadyReplicas,
		"Replicas", controllerDeployment.Status.Replicas)
	logger.Info("restarting csi controller")

	return r.Client.Delete(context.TODO(), controllerPod)
}

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

func (r *CSIScaleOperatorReconciler) getControllerDeployment(instance *csiscaleoperator.CSIScaleOperator) (*appsv1.Deployment, error) {
	controllerDeployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.GetNameForResource(config.CSIController, instance.Name),
		Namespace: instance.Namespace,
	}, controllerDeployment)

	return controllerDeployment, err
}

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
			logger.Info("update ClusterRoleBinding with", "Name", crb.GetName())
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

// reconcileSecurityContextConstraint handles creation/updating of SecurityContextConstraints resource in cluster.
// It returns error if fails to create/update SecurityContextConstraints resource.
func (r *CSIScaleOperatorReconciler) reconcileSecurityContextConstraint(instance *csiscaleoperator.CSIScaleOperator) error {

	logger := csiLog.WithName("reconcileSecurityContextConstraint")
	_, isOpenShift := os.LookupEnv(config.ENVIsOpenShift)
	if !isOpenShift {
		logger.Info("This is not an OpenShift cluster, so skipping reconciliation of SecurityContextConstraints")
		return nil
	}

	logger.Info("Creating required SecurityContextConstraints resource.")
	csiaccess_users_new := []string{
		"system:serviceaccount:" + instance.Namespace + ":" + config.GetNameForResource(config.CSIAttacherServiceAccount, instance.Name),
		"system:serviceaccount:" + instance.Namespace + ":" + config.GetNameForResource(config.CSIProvisionerServiceAccount, instance.Name),
		"system:serviceaccount:" + instance.Namespace + ":" + config.GetNameForResource(config.CSINodeServiceAccount, instance.Name),
		"system:serviceaccount:" + instance.Namespace + ":" + config.GetNameForResource(config.CSISnapshotterServiceAccount, instance.Name),
		"system:serviceaccount:" + instance.Namespace + ":" + config.GetNameForResource(config.CSIResizerServiceAccount, instance.Name),
	}

	// Check if  SCC "spectrum-scale-csiaccess" exists in cluster
	SCC := &securityv1.SecurityContextConstraints{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.CSISCC,
		Namespace: instance.Namespace,
	}, SCC)
	// If SCC does not exist, create a new SCC resource.
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating SecurityContextConstraints.")
		SCC := instance.GenerateSecurityContextConstraint(csiaccess_users_new)
		err = r.Client.Create(context.TODO(), SCC)
		if err != nil {
			message := "Failed to create SecurityContextConstraints." + SCC.GetName()
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
		logger.Info("SecurityContextConstraint created successfully.")
	} else if err != nil { // If fetching SCC fails with error, return error.
		logger.Error(err, "Failed to fetch SecurityContextConstraints from the cluster. Ignore if it's not Redhat Openshift Container Platform.")
		// Discuss: Should log level be changed to type Info?
		// return err
	} else { // If SCC already exists.
		// Get list of existing users
		logger.Info("SecurityContextConstraint already exists in cluster. Fetching users and service accounts details.")
		csiaccess_users := SCC.Users
		// Append new users, if it doesn't already exist.
		updateRequired := false
		for _, user := range csiaccess_users_new {
			if !containsString(csiaccess_users, user) {
				csiaccess_users = append(csiaccess_users, user)
				updateRequired = true
			}
		}
		// Update SCC resource
		if updateRequired {
			logger.Info("updating SecurityContextConstraint with new users.")
			newSCC := instance.GenerateSecurityContextConstraint(csiaccess_users)
			newSCC.SetResourceVersion(SCC.GetResourceVersion())
			err = r.Client.Update(context.TODO(), newSCC)
			if err != nil {
				message := "Failed to update SecurityContextConstraints " + newSCC.GetName()
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
			logger.Info("SecurityContextConstraint updated successfully.", "Users", csiaccess_users)
		} else {
			logger.Info("SecurityContextConstraint is up to date, no updates required.")
		}
	}

	logger.V(1).Info("Reconciliation of SecurityContextConstraints is successful")
	return nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRoleBindings(instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.WithName("deleteClusterRoleBindings")

	logger.Info("deleting ClusterRoleBindings")
	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("continue looking for ClusterRoleBindings", "Name", crb.GetName())
			continue
		} else if err != nil {
			logger.Error(err, "failed to get ClusterRoleBinding", "Name", crb.GetName())
			return err
		} else {
			logger.Info("deleting ClusterRoleBinding", "Name", crb.GetName())
			if err := r.Client.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRoleBinding", "Name", crb.GetName())
				return err
			}
		}
	}
	logger.Info("exiting deleteClusterRoleBindings()")
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

	crStatus.Version = config.DriverVersion

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

	logger.Info("deleting CSIDriver")
	csiDriver := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      csiDriver.Name,
		Namespace: csiDriver.Namespace,
	}, found)
	if err == nil {
		logger.Info("deleting CSIDriver", "Name", csiDriver.GetName())
		if err := r.Client.Delete(context.TODO(), found); err != nil {
			logger.Error(err, "failed to delete CSIDriver", "Name", csiDriver.GetName())
			return err
		}
	} else if errors.IsNotFound(err) {
		logger.Info("CSIDriver not found for deletion")
		return nil
	} else {
		logger.Error(err, "failed to get CSIDriver", "Name", csiDriver.GetName())
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

	service := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      "controller-manager",
		Namespace: "openshift-controller-manager",
	}, service)
	if err == nil {
		logger.Info("CSI Operator is running on an OpenShift cluster.")
		os.Setenv(config.ENVIsOpenShift, "True")
	}
}
