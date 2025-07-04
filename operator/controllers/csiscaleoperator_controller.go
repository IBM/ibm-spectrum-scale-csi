/*
Copyright 2022, 2024 IBM Corp.

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
	"strconv"
	"strings"
	"time"
	"unicode"

	uuid "github.com/google/uuid"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/presslabs/controller-util/pkg/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
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
	csiLog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	csiv1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"

	//v1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	config "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	csiscaleoperator "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/internal/csiscaleoperator"
	clustersyncer "github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/syncer"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CSIScaleOperatorReconciler reconciles a CSIScaleOperator object
type CSIScaleOperatorReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	//serverVersion string
	CSIEnvConfig clustersyncer.CSIEnvConfigs
}

const MinControllerReplicas = 1
const (
	CSIScaleOperatorControllerName = config.CSIScaleOperatorControllerName
)

var restartedAtKey = ""
var restartedAtValue = ""

//var csiLog = log.Log.WithName("csiscaleoperator_controller")

type reconciler func(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error

var crStatus = csiv1.CSIScaleOperatorStatus{}

// A map of changed clusters, used to process only changed
// clusters in case of clusters stanza is modified
var changedClusters = make(map[string]bool)

// a map of connectors to make REST calls to GUI
var scaleConnMap = make(map[string]connectors.SpectrumScaleConnector)

var cmData = make(map[string]string)
var cmDataCopy = make(map[string]string)

var clusterTypeData = make(map[string]string)

//var symlinkDirPath = ""

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
	//_ = log.FromContext(ctx)

	logger := csiLog.FromContext(ctx).WithName("Reconcile")
	logger.Info("CSI setup started.")

	//setENVIsOpenShift(r)

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
		if err := r.SetStatus(ctx, instance); err != nil {
			logger.Error(err, "Assigning values to status sub-resource object failed.")
		}
		cr.Status = crStatus
		err := r.Client.Status().Patch(ctx, cr, client.Merge, client.FieldOwner("CSIScaleOperator"))
		if err != nil {
			logger.Error(err, "Deferred update of resource status failed.", "Status", cr.Status)
		}
		logger.V(1).Info("Updated resource status.", "Status", cr.Status)
	}()

	// all older secrets not to be watched for reconcillation
	for k := range watchResources[corev1.ResourceSecrets.String()] {
		watchResources[corev1.ResourceSecrets.String()][k] = false
	}

	for _, cluster := range instance.Spec.Clusters {
		if cluster.Cacert != "" {
			watchResources[corev1.ResourceConfigMaps.String()][cluster.Cacert] = true
		}
		if cluster.Secrets != "" {
			watchResources[corev1.ResourceSecrets.String()][cluster.Secrets] = true
		}
	}
	logger.Info("CSI driver watchResources ", "watchResources :", watchResources)
	watchResources[corev1.ResourceConfigMaps.String()][config.EnvVarConfigMap] = true

	logger.Info("Adding Finalizer")
	if err := r.addFinalizerIfNotPresent(ctx, instance); err != nil {
		message := fmt.Sprintf("Failed to add the finalizer %s to the CSISCaleOperator instance %s", config.CSIFinalizer, instance.Name)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}

	logger.Info("Checking if CSIScaleOperator object got deleted")
	if !instance.GetDeletionTimestamp().IsZero() {

		logger.Info("Attempting cleanup of CSI driver")
		isFinalizerExists, err := r.hasFinalizer(ctx, instance)
		if err != nil {
			message := fmt.Sprintf("Failed to get the finalizer %s for the CSISCaleOperator instance %s", config.CSIFinalizer, instance.Name)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return ctrl.Result{}, err
		}

		if !isFinalizerExists {
			logger.Error(err, "No finalizer was found")
			return ctrl.Result{}, nil
		}

		if err := r.deleteClusterRolesAndBindings(ctx, instance); err != nil {
			message := fmt.Sprintf("Failed to delete the ClusterRoles and ClusterRoleBindings for the CSISCaleOperator instance %s."+
				" To get the list of the ClusterRoles and ClusterRoleBindings, use the selector as --selector='product=%s'", instance.Name, config.Product,
			)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.DeleteFailed), message,
			)
			return ctrl.Result{}, err
		}

		if err := r.deleteCSIDriver(ctx, instance); err != nil {
			message := fmt.Sprintf("Failed to delete the CSIDriver %s for the CSISCaleOperator instance %s", config.DriverName, instance.Name)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.DeleteFailed), message,
			)
			return ctrl.Result{}, err
		}

		if err := r.removeFinalizer(ctx, instance); err != nil {
			message := fmt.Sprintf("Failed to remove the finalizer %s for the CSISCaleOperator instance %s", config.CSIFinalizer, instance.Name)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
			)
			return ctrl.Result{}, err
		}
		logger.Info("Removed CSI driver successfully")
		return ctrl.Result{}, nil
	}

	if pass, err := r.checkPrerequisite(ctx, instance); !pass {
		logger.Error(err, "Pre-requisite check failed.")
		return ctrl.Result{}, err
	} else {
		logger.Info("Pre-requisite check passed.")
	}

	cmExists, clustersStanzaModified, err := r.isClusterStanzaModified(ctx, req.Namespace, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !cmExists || clustersStanzaModified {
		err = ValidateCRParams(ctx, instance)
		if err != nil {
			message := "Failed to validate IBM Storage Scale CSI configurations." +
				" Please check the cluster stanza under the Spec.Clusters section in the CSISCaleOperator instance " + instance.Name
			logger.Error(fmt.Errorf("%s", message), "")
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.ValidationFailed), message,
			)
			return ctrl.Result{}, err
		}
		logger.Info("The IBM Storage Scale CSI configurations are validated successfully")
	}

	if len(instance.Spec.Clusters) != 0 {
		requeAfterDelay, err := r.handleSpectrumScaleConnectors(ctx, instance, cmExists, clustersStanzaModified)
		if err != nil {
			message := "Error in getting connectors"
			logger.Error(err, message)
			if requeAfterDelay == 0 {
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: requeAfterDelay}, nil
			}
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
		if err = rec(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info("Successfully created resources like ServiceAccount, ClusterRoles and ClusterRoleBinding")

	// Rollout restart of node plugin pods, if modified driver
	// manifest file is applied.
	if cmExists && clustersStanzaModified {
		logger.Info("Some of the cluster fields of CSIScaleOperator instance are changed, so restarting node plugin pods")
		err = r.handleDriverRestart(ctx, instance)
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

	// Calling the function to get platform type and check whether CNSA is present or not in the cluster
	_, clusterConfigTypeExists := clusterTypeData[config.ENVClusterConfigurationType]
	_, cnsaOperatorPresenceExists := clusterTypeData[config.ENVClusterCNSAPresenceCheck]

	// If the setup is a cnsa dev setup, then we are not going to set any clusterTypeData as cnsa operator is not present in the cluster and platform is also local.
	inClusterScaleGui := os.Getenv("incluster_gui_host")
	if inClusterScaleGui == "" {
		if !clusterConfigTypeExists || !cnsaOperatorPresenceExists {
			logger.Info("Checking the clusterType and presence of CNSA")
			err := getClusterTypeAndCNSAOperatorPresence(ctx, config.CNSAOperatorNamespace)
			if err != nil {
				logger.Error(err, "Failed to check cluster platform and cnsa presence")
				return ctrl.Result{}, err
			}
		}
	}
	logger.Info("CSI environment variables are found successfully", "CSIConfig", r.CSIEnvConfig)
	/*	// Get CSI env images from the ClusterManagerConfig for cnsa
		r.GetCSIConfig(ctx)
	*/
	// Synchronizing optional configMap
	cm, err := r.getConfigMap(ctx, instance, config.EnvVarConfigMap)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if errors.IsNotFound(err) {
		cmData = map[string]string{}
		cmDataCopy = map[string]string{}
		//this means cm is deleted, so set defaults
		logger.Info("Optional ConfigMap is not found", "ConfigMap", config.EnvVarConfigMap)
		// setting default values if values are empty
		setDefaultDriverEnvValues(ctx, cmData)
		logger.Info("Final optional configmap values ", "when the optional configmap is absent", cmData)
	} else {
		isValidationNeeded := false
		// for the first iteration or if there is change in cm data, then validation is required
		if len(cmDataCopy) == 0 || !reflect.DeepEqual(cm.Data, cmDataCopy) {
			isValidationNeeded = true
		}
		cmDataCopy = cm.Data

		if isValidationNeeded {
			if err == nil && len(cm.Data) != 0 {
				cmData = r.parseConfigMap(ctx, instance, cm)
				logger.Info("Final optional configmap values ", "when the optional configmap is present", cmData)

			} else {
				cmData = map[string]string{}
				logger.Info("Optional ConfigMap is either not found or is empty, skipped parsing it", "ConfigMap", config.EnvVarConfigMap)
				// setting default values if values are empty
				setDefaultDriverEnvValues(ctx, cmData)
				logger.Info("Final optional configmap values ", "when the optional configmap is absent", cmData)
			}
		}
	}

	// Setting the clusterType and presence of CNSA in the final cmData before driver sync
	if _, exists := cmData[config.ENVClusterConfigurationType]; !exists {
		logger.Info("Setting the clusterType and presence of CNSA in final cmData")

		cmData[config.ENVClusterConfigurationType] = clusterTypeData[config.ENVClusterConfigurationType]
		cmData[config.ENVClusterCNSAPresenceCheck] = clusterTypeData[config.ENVClusterCNSAPresenceCheck]
	}

	logger.Info("Final optional configmap values", "when the sent to syncer is ", cmData)

	ifPrimaryDisable := false
	if strings.ToUpper(cmData["PRIMARY_FILESYSTEM"]) == "DISABLED" {
		ifPrimaryDisable = true
	}
	logger.Info("PRIMARY_FILESYSTEM is", " DISABLED ", ifPrimaryDisable)

	//For first pass handle primary FS and fileset
	if !cmExists && !ifPrimaryDisable {
		requeAfterDelay, err := r.handlePrimaryFSandFileset(ctx, instance)
		if err != nil {
			if requeAfterDelay == 0 {
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: requeAfterDelay}, nil
			}
		}
	}

	// Synchronizing node/driver daemonset
	CGPrefix := r.GetConsistencyGroupPrefix(ctx, instance)

	if instance.Spec.CGPrefix == "" {
		logger.Info("Updating consistency group prefix in CSIScaleOperator resource.")
		instance.Spec.CGPrefix = CGPrefix
		err := r.Client.Update(ctx, instance.Unwrap())
		if err != nil {
			logger.Error(err, "Reconciler Client.Update() failed.")
			message := "Failed to update the consistency group prefix in CSIScaleOperator resource " + instance.Name
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
			)
			return ctrl.Result{}, err
		}
		logger.Info("Successfully updated consistency group prefix in CSIScaleOperator resource.")

	}

	csiNodeSyncer := clustersyncer.GetCSIDaemonsetSyncer(ctx, r.Client, r.Scheme, instance, restartedAtKey, restartedAtValue, CGPrefix, cmData, r.CSIEnvConfig)
	if err := syncer.Sync(context.TODO(), csiNodeSyncer, nil); err != nil {
		message := "Synchronization of node/driver " + config.GetNameForResource(config.CSINode, instance.Name) + " DaemonSet failed for the CSISCaleOperator instance " + instance.Name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("Synchronization of node/driver %s DaemonSet is successful", config.GetNameForResource(config.CSINode, instance.Name)))

	// Setting resources limit values for sidecars
	cpuLimits := cmData[config.SidecarCPULimits]
	memoryLimits := cmData[config.SidecarMemoryLimits]

	// volNamePrefix for provisioner sidecar
	volNamePrefix := cmData[config.EnvVolNamePrefixKey]
	// Synchronizing attacher deployment
	if err := r.removeDeprecatedStatefulset(ctx, instance, config.GetNameForResource(config.CSIControllerAttacher, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}

	csiControllerSyncer := clustersyncer.GetAttacherSyncer(ctx, r.Client, r.Scheme, instance, restartedAtKey, restartedAtValue, cpuLimits, memoryLimits, r.CSIEnvConfig)
	if err := syncer.Sync(context.TODO(), csiControllerSyncer, nil); err != nil {
		message := "Synchronization of " + config.GetNameForResource(config.CSIControllerAttacher, instance.Name) + " Deployment failed for the CSISCaleOperator instance " + instance.Name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("Synchronization of %s Deployment is successful", config.GetNameForResource(config.CSIControllerAttacher, instance.Name)))

	// Synchronizing provisioner deployment
	if err := r.removeDeprecatedStatefulset(ctx, instance, config.GetNameForResource(config.CSIControllerProvisioner, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}

	csiControllerSyncerProvisioner := clustersyncer.GetProvisionerSyncer(ctx, r.Client, r.Scheme, instance, restartedAtKey, restartedAtValue, cpuLimits, memoryLimits, volNamePrefix, r.CSIEnvConfig)
	if err := syncer.Sync(context.TODO(), csiControllerSyncerProvisioner, nil); err != nil {
		message := "Synchronization of " + config.GetNameForResource(config.CSIControllerProvisioner, instance.Name) + " Deployment failed for the CSISCaleOperator instance " + instance.Name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("Synchronization of %s Deployment is successful", config.GetNameForResource(config.CSIControllerProvisioner, instance.Name)))

	// Synchronizing snapshotter deployment
	if err := r.removeDeprecatedStatefulset(ctx, instance, config.GetNameForResource(config.CSIControllerSnapshotter, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}
	csiControllerSyncerSnapshotter := clustersyncer.GetSnapshotterSyncer(ctx, r.Client, r.Scheme, instance, restartedAtKey, restartedAtValue, cpuLimits, memoryLimits, r.CSIEnvConfig)
	if err := syncer.Sync(context.TODO(), csiControllerSyncerSnapshotter, nil); err != nil {
		message := "Synchronization of " + config.GetNameForResource(config.CSIControllerSnapshotter, instance.Name) + " Deployment failed for the CSISCaleOperator instance " + instance.Name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("Synchronization of %s Deployment is successful", config.GetNameForResource(config.CSIControllerSnapshotter, instance.Name)))

	// Synchronizing resizer deployment
	if err := r.removeDeprecatedStatefulset(ctx, instance, config.GetNameForResource(config.CSIControllerResizer, instance.Name)); err != nil {
		return ctrl.Result{}, err
	}
	csiControllerSyncerResizer := clustersyncer.GetResizerSyncer(ctx, r.Client, r.Scheme, instance, restartedAtKey, restartedAtValue, cpuLimits, memoryLimits, r.CSIEnvConfig)
	if err := syncer.Sync(context.TODO(), csiControllerSyncerResizer, nil); err != nil {
		message := "Synchronization of " + config.GetNameForResource(config.CSIControllerResizer, instance.Name) + " Deployment failed for the CSISCaleOperator instance " + instance.Name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("Synchronization of %s Deployment is successful", config.GetNameForResource(config.CSIControllerResizer, instance.Name)))

	// Synchronizing cluster configMap
	csiConfigmapSyncer := clustersyncer.CSIConfigmapSyncer(ctx, r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiConfigmapSyncer, nil); err != nil {
		message := "Synchronization of " + config.CSIConfigMap + " ConfigMap failed for the CSISCaleOperator instance " + instance.Name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
		)
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("Synchronization of ConfigMap %s is successful", config.CSIConfigMap))

	message := "The CSI driver resources have been created/updated successfully"
	logger.Info(message)

	SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeNormal, string(config.StatusConditionSuccess),
		metav1.ConditionTrue, string(csiv1.CSIConfigured), message,
	)

	logger.Info("CSI setup completed successfully.")

	if len(instance.Spec.Clusters) != 0 {
		logger.Info("Checking GUI password expiry")
		// For validating gui password expires in every 24hours.
		passwordExipredClusters := r.listGUIPasswdExpiredClusters(ctx, instance, instance.Spec.Clusters)
		if len(passwordExipredClusters) > 0 {
			message := fmt.Sprintf("Either the username/password is incorrect or the password has been expired for the Scale GUI clusterIds: %v", passwordExipredClusters)
			logger.Info(message)

			RaiseCSOEvent(instance, r.Recorder, corev1.EventTypeWarning,
				string(csiv1.AuthError), message,
			)
		}
		logger.Info("Done with checking GUI password expiry")
		return ctrl.Result{RequeueAfter: 24 * time.Hour}, nil
	}
	return ctrl.Result{}, nil
}

// handleDriverRestart gets a driver daemonset from the cluster and
// restarts driver pods, returns error if there is any.
func (r *CSIScaleOperatorReconciler) handleDriverRestart(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("handleDriverRestart")
	var err error
	var daemonSet *appsv1.DaemonSet
	daemonSet, err = r.getNodeDaemonSet(instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			message := "Failed to get the driver DaemonSet: " + config.GetNameForResource(config.CSINode, instance.Name)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
		} else {
			message := "The driver DaemonSet " + config.GetNameForResource(config.CSINode, instance.Name) + " is not found"
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
		}
	} else {
		err = r.rolloutRestartNode(daemonSet)
		if err != nil {
			message := "Failed to rollout restart of driver pods"
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
			)
		} else {
			restartedAtKey, restartedAtValue = r.getRestartedAtAnnotation(daemonSet.Spec.Template.ObjectMeta.Annotations)
		}
	}
	return err
}

// isClusterStanzaModified checks if spectrum-scale-config configmap exists
// and if it exists checks if the clusters stanza is modified by
// comparing it with the configmap data.
// It returns 1st value (cmExists) which indicates if clusters configmap exists,
// 2nd value (clustersStanzaModified) which idicates whether clusters stanza is
// modified in case the configmap exists, and 3rd value as an error if any.
func (r *CSIScaleOperatorReconciler) isClusterStanzaModified(ctx context.Context, namespace string, instance *csiscaleoperator.CSIScaleOperator) (bool, bool, error) {
	logger := csiLog.FromContext(ctx).WithName("isClusterStanzaModified")
	cmExists := false
	clustersStanzaModified := false
	currentCMDataString := ""
	configMap := &corev1.ConfigMap{}
	cerr := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.CSIConfigMap,
		Namespace: namespace,
	}, configMap)
	if cerr != nil {
		if !errors.IsNotFound(cerr) {
			message := "Failed to get the ConfigMap: " + config.CSIConfigMap
			logger.Error(cerr, message)

			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
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
			logger.Error(err, "Failed to marshal ConfigMap data "+config.CSIConfigMap)
			return cmExists, clustersStanzaModified, err
		}
		currentCMDataString = string(configMapDataBytes)
		currentCMDataString = strings.Replace(currentCMDataString, " ", "", -1)
		currentCMDataString = strings.Replace(currentCMDataString, "\\\"", "\"", -1)

		if !strings.Contains(currentCMDataString, clustersString) {
			logger.Info("Clusters stanza in driver manifest is changed")
			clustersStanzaModified = true
		}

		//if isClusterStanzaModified, check and update modified clusters in changedClusters
		if clustersStanzaModified {
			logger.Info("The clusters stanza is modified ")
			changedClusters = make(map[string]bool)
			err := r.updateChangedClusters(ctx, instance, currentCMDataString, instance.Spec.Clusters)
			if err != nil {
				return cmExists, clustersStanzaModified, err
			}
		}
	}
	return cmExists, clustersStanzaModified, nil
}

// updateChangedClusters updates var changedClusters and also returns
// error if primary stanza of the primary cluster is also modified.
// It also deletes unnecessary cluster entries from connector map, for
// which clusterID is present in current configmap data but not in new CR data.
func (r *CSIScaleOperatorReconciler) updateChangedClusters(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, currentCMcmString string, newCRClusters []csiv1.CSICluster) error {
	logger := csiLog.FromContext(ctx).WithName("updateChangedClusters")
	currentCMclusters := []csiv1.CSICluster{}
	prefix := "{\"" + config.CSIConfigMap + ".json\":\"{\"clusters\":"
	postfix := "}\"}"
	currentCMcmString = strings.Replace(currentCMcmString, prefix, "", 1)
	currentCMcmString = strings.Replace(currentCMcmString, postfix, "", 1)

	configMapDataBytes := []byte(currentCMcmString)
	err := json.Unmarshal(configMapDataBytes, &currentCMclusters)
	if err != nil {
		message := fmt.Sprintf("Failed to unmarshal data of ConfigMap: %v", config.CSIConfigMap)
		err := fmt.Errorf("%s", message)
		logger.Error(err, "")
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.UnmarshalFailed), message,
		)
		return err
	}

	//This is a map to track which clusters from current configmap are also
	//present in new CR, so that the connectors for the ones which are not
	//present in new CR but are present in current configmap can be deleted.
	currentCMProcessedClusters := make(map[string]bool)

	for _, crCluster := range newCRClusters {
		//For the cluster ID of each clusters of updated CR, get the clusters
		//data of the current configmap and compare that with new CR data
		oldCMCluster := r.getClusterByID(crCluster.Id, currentCMclusters)
		if reflect.DeepEqual(oldCMCluster, csiv1.CSICluster{}) {
			//case 1: new cluster is added in CR
			//no matching cluster is found, that means it is a new
			//cluster added in CR --> add entry in changedClusters,
			//so that the new clusters data can be validated later.
			changedClusters[crCluster.Id] = true
		} else {
			//exisiting cluster in current configmap
			if !reflect.DeepEqual(crCluster, oldCMCluster) {
				//case 2: clusters data of current configmap is different than
				//the new CR --> add entry in changedClusters, so that the new
				//clusters data can be validated later.
				changedClusters[crCluster.Id] = true
				currentCMProcessedClusters[crCluster.Id] = true

				//Check if the primary stanza from current configmap is changed
				//and return err if it is changed, as we don't want to change
				//the primary after first successful iteration.
				if oldCMCluster.Primary != nil && !reflect.DeepEqual(oldCMCluster.Primary, crCluster.Primary) {
					primaryString := fmt.Sprintf("{filesystem:%v, fileset:%v",
						oldCMCluster.Primary.PrimaryFs, oldCMCluster.Primary.PrimaryFset)
					if oldCMCluster.Primary.RemoteCluster != "" {
						primaryString += fmt.Sprintf(", remote cluster: %v}", oldCMCluster.Primary.RemoteCluster)
					} else {
						primaryString += "}"
					}
					message := fmt.Sprintf("Primary stanza is modified for cluster with ID %s. Use the orignal primary %s and try again",
						crCluster.Id, primaryString)
					err := fmt.Errorf("%s", message)
					logger.Error(err, "")
					SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
						metav1.ConditionFalse, string(csiv1.PrimaryClusterStanzaModified), message,
					)
					return err
				}
			}
		}
	}

	//case 3: clusters data in current configmap and new CR mataches, nothing to be done here.
	//case 4: delete - current configmap has an entry, which is not there in new CR --> delete
	//the connector for that cluster as we no longer need it.
	for _, cluster := range currentCMclusters {
		if _, processed := currentCMProcessedClusters[cluster.Id]; !processed {
			delete(scaleConnMap, cluster.Id)
		}
	}
	logger.Info("The changed clusters to process", "changedClusters", changedClusters)
	return nil
}

// getClusterByID returns a cluster matching the passed clusterID
// from the passed list of clusters.
func (r *CSIScaleOperatorReconciler) getClusterByID(id string, clusters []csiv1.CSICluster) csiv1.CSICluster {
	for _, cluster := range clusters {
		if id == cluster.Id {
			return cluster
		}
	}
	return csiv1.CSICluster{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CSIScaleOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {

	logger := csiLog.Log.WithName("SetupWithManager")
	logger.Info("Running IBM Storage Scale CSI operator", "version", config.OperatorVersion)
	logger.Info("Setting up the controller with the manager.")

	CSIReconcileRequestFunc := func(ctx context.Context, obj client.Object) []reconcile.Request {
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
	shouldRequeueOnCreateOrDelete := func(cfgmapData map[string]string) bool {
		for key := range cfgmapData {
			if containsStringInSlice(config.CSIOptionalConfigMapKeys, strings.ToUpper(key)) {
				return true
			}
		}
		logger.Info(fmt.Sprintf("No env vars found with prefix %s in the configmap %s, skipping proccessing them", config.EnvVarPrefix, config.EnvVarConfigMap))
		return false
	}

	//Allow implicit restart of driver pods when returns true
	//implicit restart occurs automatically based on daemonset updateStretegy when a daemonset gets updated
	shouldRequeueOnUpdate := func(oldCfgMapData, newCfgMapData map[string]string) bool {
		for key, newVal := range newCfgMapData {
			//Allow restart of driver pods when a new valid env var is found or the value of existing valid env var is updated
			if oldVal, ok := oldCfgMapData[key]; !ok {
				if containsStringInSlice(config.CSIOptionalConfigMapKeys, strings.ToUpper(key)) {
					return true
				}
			} else if oldVal != newVal && containsStringInSlice(config.CSIOptionalConfigMapKeys, strings.ToUpper(key)) {
				return true
			}
		}

		for key := range oldCfgMapData {
			//look for deleted valid env vars of the old configmap in the new configmap
			//if deleted restart driver pods
			if _, ok := newCfgMapData[key]; !ok {
				if containsStringInSlice(config.CSIOptionalConfigMapKeys, strings.ToUpper(key)) {
					return true
				}
			}
		}
		return false
	}

	predicateFuncs := func(resourceKind string) predicate.Funcs {
		logger := csiLog.Log.WithName("predicateFuncs")

		return predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				if isCSIResource(e.Object.GetName(), resourceKind) {
					if resourceKind == corev1.ResourceConfigMaps.String() && e.Object.GetName() == config.EnvVarConfigMap {
						if shouldRequeueOnCreateOrDelete(e.Object.(*corev1.ConfigMap).Data) {
							r.setRestartedAtValues()
							logger.Info("Restarting driver and sidecar pods due to creation of", "Resource", resourceKind, "Name", e.Object.GetName())
							return true
						}
					} else {
						logger.Info("Restarting driver and sidecar pods due to creation of", "Resource", resourceKind, "Name", e.Object.GetName())
						r.setRestartedAtValues()
						return true
					}
				}
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				if isCSIResource(e.ObjectNew.GetName(), resourceKind) {
					/*if resourceKind == corev1.ResourceSecrets.String() && !reflect.DeepEqual(e.ObjectOld.(*corev1.Secret).Data, e.ObjectNew.(*corev1.Secret).Data) {
						r.setRestartedAtValues()
						logger.Info("Restarting driver and sidecar pods due to update of", "Resource", resourceKind, "Name", e.ObjectOld.GetName())
						scaleGuiSecretChanged = true
						return true
					} else */
					if resourceKind == corev1.ResourceConfigMaps.String() {
						if e.ObjectNew.GetName() == config.EnvVarConfigMap && !reflect.DeepEqual(e.ObjectOld.(*corev1.ConfigMap).Data, e.ObjectNew.(*corev1.ConfigMap).Data) {
							if shouldRequeueOnUpdate(e.ObjectOld.(*corev1.ConfigMap).Data, e.ObjectNew.(*corev1.ConfigMap).Data) {
								r.setRestartedAtValues()
								logger.Info("Restarting driver and sidecar pods due to update of", "Resource", resourceKind, "Name", e.ObjectOld.GetName())
								return true
							}
						}
					}
				}
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				if isCSIResource(e.Object.GetName(), resourceKind) {
					if resourceKind == corev1.ResourceConfigMaps.String() && e.Object.GetName() == config.EnvVarConfigMap {
						if shouldRequeueOnCreateOrDelete(e.Object.(*corev1.ConfigMap).Data) {
							r.setRestartedAtValues()
							logger.Info("Restarting driver and sidecar pods due to deletion of", "Resource", resourceKind, "Name", e.Object.GetName())
							return true
						}
					} else {
						r.setRestartedAtValues()
						logger.Info("Restarting driver and sidecar pods due to deletion of", "Resource", resourceKind, "Name", e.Object.GetName())
						return true
					}
				}
				return false
			},
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1.CSIScaleOperator{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(CSIReconcileRequestFunc),
			builder.WithPredicates(predicateFuncs(corev1.ResourceSecrets.String())),
		).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(CSIReconcileRequestFunc),
			builder.WithPredicates(predicateFuncs(corev1.ResourceConfigMaps.String())),
		).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}

func (r *CSIScaleOperatorReconciler) setRestartedAtValues() {

	restartedAtKey = fmt.Sprintf("%s/restartedAt", config.APIGroup)
	restartedAtValue = time.Now().String()

}

func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func (r *CSIScaleOperatorReconciler) hasFinalizer(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) (bool, error) {

	logger := csiLog.FromContext(ctx).WithName("hasFinalizer")
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(ctx, instance)
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

func (r *CSIScaleOperatorReconciler) removeFinalizer(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("removeFinalizer")
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(ctx, instance)
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

func (r *CSIScaleOperatorReconciler) addFinalizerIfNotPresent(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("addFinalizerIfNotPresent")
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(ctx, instance)
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

func (r *CSIScaleOperatorReconciler) getAccessorAndFinalizerName(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) (metav1.Object, string, error) {
	logger := csiLog.FromContext(ctx).WithName("getAccessorAndFinalizerName")
	finalizerName := config.CSIFinalizer

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "Failed to get meta information of instance")
		return nil, "", err
	}

	logger.Info("Got finalizer with", "name", finalizerName)
	return accessor, finalizerName, nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRolesAndBindings(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("deleteClusterRolesAndBindings")
	logger.Info("Calling deleteClusterRoleBindings()")
	if err := r.deleteClusterRoleBindings(ctx, instance); err != nil {
		logger.Error(err, "Deletion of ClusterRoleBindings failed")
		return err
	}

	logger.Info("Calling deleteClusterRoles()")
	if err := r.deleteClusterRoles(ctx, instance); err != nil {
		logger.Error(err, "Deletion of ClusterRoles failed")
		return err
	}

	logger.Info("Deletion of ClusterRoles and ClusterRoleBindings succeeded.")
	return nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRoles(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("deleteClusterRoles")

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
func (r *CSIScaleOperatorReconciler) reconcileCSIDriver(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("reconcileCSIDriver").WithValues("Name", config.DriverName)
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
			message := fmt.Sprintf("Failed to create the CSIDriver %s for the CSISCaleOperator instance %s", config.DriverName, instance.Name)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.CreateFailed), message,
			)
			return err
		}
	} else if err != nil {
		message := fmt.Sprintf("Failed to get the CSIDriver %s for the CSISCaleOperator instance %s", config.DriverName, instance.Name)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
		return err
	} else {
		// Resource already exists - don't requeue
		logger.Info("Resource CSIDriver already exists.")
	}
	logger.V(1).Info("Exiting reconcileCSIDriver method.")
	return nil
}

func (r *CSIScaleOperatorReconciler) reconcileServiceAccount(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("reconcileServiceAccount")
	logger.Info("Creating the required ServiceAccount resources.")

	// controller := instance.GenerateControllerServiceAccount()
	node := instance.GenerateNodeServiceAccount()
	attacher := instance.GenerateAttacherServiceAccount()
	provisioner := instance.GenerateProvisionerServiceAccount()
	snapshotter := instance.GenerateSnapshotterServiceAccount()
	resizer := instance.GenerateResizerServiceAccount()

	// controllerServiceAccountName := config.GetNameForResource(config.CSIControllerServiceAccount, instance.Name)
	nodeServiceAccountName := config.GetNameForResource(config.CSINodeServiceAccount, instance.Name)
	attacherServiceAccountName := config.GetNameForResource(config.CSIAttacherServiceAccount, instance.Name)
	provisionerServiceAccountName := config.GetNameForResource(config.CSIProvisionerServiceAccount, instance.Name)
	snapshotterServiceAccountName := config.GetNameForResource(config.CSISnapshotterServiceAccount, instance.Name)
	resizerServiceAccountName := config.GetNameForResource(config.CSIResizerServiceAccount, instance.Name)

	for _, sa := range []*corev1.ServiceAccount{
		// controller,
		node,
		attacher,
		provisioner,
		snapshotter,
		resizer,
	} {
		if err := controllerutil.SetControllerReference(instance.Unwrap(), sa, r.Scheme); err != nil {
			message := "Failed to set the controller reference for ServiceAccount: " + sa.GetName()
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
			)
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
				message := "Failed to create the ServiceAccount: " + sa.GetName()
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.CreateFailed), message,
				)
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

				updateErr := r.updateCSINodeDaemonSet(ctx, instance)
				if updateErr != nil {
					return updateErr
				}
				// TODO: Should restart sidecar pods if respective ServiceAccount is created afterwards?
			} else if attacherServiceAccountName == sa.Name || provisionerServiceAccountName == sa.Name || snapshotterServiceAccountName == sa.Name || resizerServiceAccountName == sa.Name {

				updateErr := r.updateCSISideCars(ctx, instance, sa.Name)
				if updateErr != nil {
					return updateErr
				}
				// TODO: Should restart sidecar pods if respective ServiceAccount is created afterwards?
			}
		} else if err != nil {
			message := "Failed to get the ServiceAccount: " + sa.GetName()
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return err
		} else {
			// Cannot update the service account of an already created pod.
			// Reference: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/

			logger.Info("service-account " + sa.GetName() + " already exists. Updating.")
			err = r.Client.Update(context.TODO(), sa)
			if err != nil {
				message := "Failed to update the service-account: " + sa.GetName()
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
				)
				return err
			}
			logger.Info("ServiceAccount " + sa.GetName() + " has been updated.")

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

func (r *CSIScaleOperatorReconciler) getDeployment(instance *csiscaleoperator.CSIScaleOperator, name config.ResourceName) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      config.GetNameForResource(name, instance.Name),
		Namespace: instance.Namespace,
	}, deployment)

	return deployment, err
}

func (r *CSIScaleOperatorReconciler) updateCSISideCars(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, serviceAccountName string) error {
	logger := csiLog.FromContext(ctx).WithName("updateCSISideCars")
	logger.Info("Restarting Sidecars when the ServiceAccount is updated.")
	sideCarName, err := r.getDeployment(instance, config.ResourceName(serviceAccountName))
	if err != nil && errors.IsNotFound(err) {
		logger.Info("SideCar doesn't exist. Restart not required.", "SideCar", serviceAccountName)
	} else if err != nil {
		message := "Failed to get the SideCar Name: " + serviceAccountName
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
		return err
	} else {
		logger.Info("SideCar pods exists, deployment rollout requires restart",
			"DesiredNumberScheduled", sideCarName.Status.Replicas,
			"NumberAvailable", sideCarName.Status.AvailableReplicas)

		rErr := r.rolloutRestartDeployment(sideCarName)
		if rErr != nil {
			message := "Failed to rollout restart of the SideCar: " + serviceAccountName
			logger.Error(rErr, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
			)
			return rErr
		}
		restartedAtKey, restartedAtValue = r.getRestartedAtAnnotation(sideCarName.Spec.Template.ObjectMeta.Annotations)
		logger.Info("Rollout restart of the Sidecar is successful", "SideCar", serviceAccountName)
	}
	return nil
}

func (r *CSIScaleOperatorReconciler) updateCSINodeDaemonSet(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("updateCSINodeDaemonSet")
	logger.Info("Restarting NodeDaemonSet when the ServiceAccount is updated.")
	nodeDaemonSet, err := r.getNodeDaemonSet(instance)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Daemonset doesn't exist. Restart not required.")
	} else if err != nil {
		message := "Failed to get the driver DaemonSet: " + config.GetNameForResource(config.CSINode, instance.Name)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
		return err
	} else {
		logger.Info("DaemonSet exists, node rollout requires restart",
			"DesiredNumberScheduled", nodeDaemonSet.Status.DesiredNumberScheduled,
			"NumberAvailable", nodeDaemonSet.Status.NumberAvailable)

		rErr := r.rolloutRestartNode(nodeDaemonSet)
		if rErr != nil {
			message := "Failed to rollout restart of node DaemonSet: " + config.GetNameForResource(config.CSINode, instance.Name)
			logger.Error(rErr, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
			)
			return rErr
		}

		restartedAtKey, restartedAtValue = r.getRestartedAtAnnotation(nodeDaemonSet.Spec.Template.ObjectMeta.Annotations)
		logger.Info("Rollout restart of node DaemonSet is successful")
	}
	return nil
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

func (r *CSIScaleOperatorReconciler) rolloutRestartDeployment(deploy *appsv1.Deployment) error {
	restartedAt := fmt.Sprintf("%s/restartedAt", config.APIGroup)
	timestamp := time.Now().String()
	deploy.Spec.Template.ObjectMeta.Annotations[restartedAt] = timestamp
	return r.Client.Update(context.TODO(), deploy)
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

func (r *CSIScaleOperatorReconciler) reconcileClusterRole(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("reconcileClusterRole")
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
				message := "Failed to create the ClusterRole: " + cr.GetName()
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.CreateFailed), message,
				)
				return err
			}
		} else if err != nil {
			message := "Failed to get the ClusterRole: " + cr.GetName()
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return err
		} else {
			logger.Info("Clusterrole " + cr.GetName() + " already exists. Updating clusterrole.")
			err = r.Client.Update(context.TODO(), cr)
			if err != nil {
				message := "Failed to update the ClusterRole: " + cr.GetName()
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
				)
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

func (r *CSIScaleOperatorReconciler) reconcileClusterRoleBinding(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("reconcileClusterRoleBinding")
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
				message := "Failed to create the ClusterRoleBinding: " + crb.GetName()
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.CreateFailed), message,
				)
				return err
			}
		} else if err != nil {
			message := "Failed to get the ClusterRoleBinding: " + crb.GetName()
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return err
		} else {
			// Resource already exists - don't requeue

			logger.Info("Clusterrolebinding " + crb.GetName() + " already exists. Updating clusterolebinding.")
			err = r.Client.Update(context.TODO(), crb)
			if err != nil {
				message := "Failed to update the ClusterRoleBinding: " + crb.GetName()
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.UpdateFailed), message,
				)
				return err
			}
		}
	}
	logger.V(1).Info("Reconciliation of ClusterRoleBindings is successful")
	return nil
}

func (r *CSIScaleOperatorReconciler) deleteClusterRoleBindings(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("deleteClusterRoleBindings")

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
func (r *CSIScaleOperatorReconciler) SetStatus(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {

	logger := csiLog.FromContext(ctx).WithName("SetStatus")
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
	logger := csiLog.FromContext(ctx).WithName("isControllerReady")
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

	logger := csiLog.FromContext(ctx).WithName("areAllPodImagesSynced")
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
/*func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}*/

func (r *CSIScaleOperatorReconciler) deleteCSIDriver(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("deleteCSIDriver")

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
/*func setENVIsOpenShift(r *CSIScaleOperatorReconciler) {
	logger := csiLog.FromContext(ctx).WithName("setENVIsOpenShift")
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
}*/

// GetConsistencyGroupPrefix returns a universal unique ideintiier(UUID) of string format.
// For Redhat Openshift Cluster Platform, Cluster ID as string is returned.
// For Vanilla kubernetes cluster, generated UUID is returned.
func (r *CSIScaleOperatorReconciler) GetConsistencyGroupPrefix(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) string {
	logger := csiLog.FromContext(ctx).WithName("GetConsistencyGroupPrefix")
	logger.Info("Checking if consistency group prefix is passed in CSIScaleOperator specs.")
	if instance.Spec.CGPrefix != "" {
		logger.Info("Consistency group prefix found in CSIScaleOperator specs.")
		return instance.Spec.CGPrefix
	}

	logger.Info("Consistency group prefix is not found in CSIScaleOperator specs.")
	logger.Info("Fetching cluster information.")

	if clusterTypeData[config.ENVClusterConfigurationType] != config.ENVClusterTypeOpenshift {
		logger.Info("Cluster is a Kubernetes Platform.")
		UUID := r.GenerateUUID(ctx)
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
		UUID := r.GenerateUUID(ctx)
		return UUID.String()
	}
	UUID := string(CV.Spec.ClusterID)
	return UUID

}

// GenerateUUID returns a new random UUID.
func (r *CSIScaleOperatorReconciler) GenerateUUID(ctx context.Context) uuid.UUID {
	logger := csiLog.FromContext(ctx).WithName("GenerateUUID")
	logger.Info("Generating a unique cluster ID.")
	UUID := uuid.New()
	return UUID
}

func (r *CSIScaleOperatorReconciler) removeDeprecatedStatefulset(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, name string) error {
	logger := csiLog.FromContext(ctx).WithName("removeDeprecatedStatefulset").WithValues("Name", name)
	logger.Info("Removing deprecated statefulset resource from the cluster.")

	STS := &appsv1.StatefulSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: instance.Namespace,
	}, STS)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("Statefulset resource not found in the cluster.")
	} else if err != nil {
		message := "Failed to get the StatefulSet: " + name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
		return err
	} else {
		logger.Info("Found statefulset resource. Sidecar controllers as statefulsets are replaced by deployments in CSI >= 2.6.0. Removing statefulset.")
		if err := r.Client.Delete(context.TODO(), STS); err != nil {
			message := "Failed to delete the StatefulSet: " + name
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.DeleteFailed), message,
			)
			return err
		}
	}
	return nil
}

func (r *CSIScaleOperatorReconciler) checkPrerequisite(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) (bool, error) {

	logger := csiLog.FromContext(ctx).WithName("checkPrerequisite")
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
			if exists, err := r.resourceExists(ctx, instance, secret, string(config.Secret)); !exists {
				return false, err
			}
			logger.Info(fmt.Sprintf("Secret resource %s found.", secret))
		}
	}

	if len(configMaps) != 0 {
		for _, configMap := range configMaps {
			if exists, err := r.resourceExists(ctx, instance, configMap, string(config.ConfigMap)); !exists {
				return false, err
			}
			logger.Info(fmt.Sprintf("ConfigMap resource %s found.", configMap))
		}
	}

	return true, nil
}

func (r *CSIScaleOperatorReconciler) resourceExists(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, name string, kind string) (bool, error) {
	logger := csiLog.FromContext(ctx).WithName("resourceExists").WithValues("Kind", kind, "Name", name)
	logger.Info("Checking resource exists")

	var err error
	if kind == string(config.Secret) {
		found := &corev1.Secret{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      name,
			Namespace: instance.Namespace,
		}, found)
	}

	if kind == string(config.ConfigMap) {
		found := &corev1.ConfigMap{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      name,
			Namespace: instance.Namespace,
		}, found)
	}

	if err != nil && errors.IsNotFound(err) {
		message := fmt.Sprintf("The %s %s is not found. Please make sure to create %s named %s", kind, name, kind, name)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
		return false, err
	} else if err != nil {
		message := "Failed to get the " + kind + ": " + name
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
		return false, err
	} else {
		return true, nil
	}
}

// newConnector creates and return a new connector to make REST calls for the passed cluster
func (r *CSIScaleOperatorReconciler) newConnector(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator,
	cluster csiv1.CSICluster) (connectors.SpectrumScaleConnector, error) {
	logger := csiLog.FromContext(ctx).WithName("newConnector")
	logger.Info("Creating new IBM Storage Scale Connector for cluster with", "ID", cluster.Id)

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
			message := fmt.Sprintf("The Secret %v is not found. Please create a basic-auth Secret with your GUI host credentials", cluster.Secrets)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return &connectors.SpectrumRestV2{}, err
		} else if err != nil {
			message := fmt.Sprintf("Failed to get the Secret: %v", cluster.Secrets)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return nil, err
		}
		username = strings.TrimSpace(string(secret.Data[config.SecretUsername]))
		password = strings.TrimSuffix(string(secret.Data[config.SecretPassword]), "\n")
	}

	if cluster.SecureSslMode && cluster.Cacert != "" {
		configMap := &corev1.ConfigMap{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Name:      cluster.Cacert,
			Namespace: instance.Namespace,
		}, configMap)

		if err != nil && errors.IsNotFound(err) {
			message := fmt.Sprintf("The ConfigMap %v to specify GUI certificates, is not found. Please create one", cluster.Cacert)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return nil, err
		} else if err != nil {
			message := fmt.Sprintf("Failed to get the ConfigMap %v storing GUI certificates", cluster.Cacert)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetFailed), message,
			)
			return nil, err
		}
		cacertValue := []byte(configMap.Data[cluster.Cacert])
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(cacertValue); !ok {
			return nil, fmt.Errorf("parsing CA cert %v failed", cluster.Cacert)
		}
		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool, MinVersion: tls.VersionTLS12}}
		logger.Info("Created IBM Storage Scale connector with SSL mode for guiHost(s)")

	} else {
		//#nosec G402 InsecureSkipVerify was requested by user.
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}} //nolint:gosec
		logger.Info("Created IBM Storage Scale connector without SSL mode for guiHost(s)")
	}
	logger.Info("Created IBM Storage Scale connector rest : ", " username ", username)
	rest = &connectors.SpectrumRestV2{
		HTTPclient: &http.Client{
			Transport: tr,
			Timeout:   time.Second * config.HTTPClientTimeout,
		},
		ClusterConfig: settings.Clusters{
			MgmtUsername: username,
			MgmtPassword: password,
		},
		EndPointIndex:   0, //Use first GUI as primary by default
		RequestCalledBy: "operator",
	}

	for i := range cluster.RestApi {
		guiHost := cluster.RestApi[i].GuiHost
		guiPort := cluster.RestApi[i].GuiPort
		if guiPort == 0 {
			guiPort = settings.DefaultGuiPort
		}
		// For CNSA Dev setup, if the GUI host is set to localroute env
		// use exposed route
		inClusterScaleGui := os.Getenv("incluster_gui_host")
		if inClusterScaleGui != "" && isLocalGUIHost(guiHost) {
			guiHost = inClusterScaleGui
			guiPort = 443
		}
		//

		endpoint := fmt.Sprintf("%s://%s:%d/", settings.GuiProtocol, guiHost, guiPort)
		rest.Endpoint = append(rest.Endpoint, endpoint)
	}
	return rest, nil
}

// isLocalGUIHost returns true if the host is a GUI host of a local cluster
// and returns false if the host is a GUI host of a remote cluster.
func isLocalGUIHost(host string) bool {
	var guiRoutePrefix = config.ScaleGUIRoute + "-" + config.ScaleProduct
	// the GUI service
	var guiService = config.ScaleGUIService
	if strings.HasPrefix(host, guiRoutePrefix) || strings.HasPrefix(host, guiService) {
		return true
	}
	return false
}

// handleSpectrumScaleConnectors gets the connectors for all the clusters in driver
// manifest and sets those in scaleConnMap also checks if GUI is reachable and
// cluster ID is valid.
func (r *CSIScaleOperatorReconciler) handleSpectrumScaleConnectors(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, cmExists bool, clustersStanzaModified bool) (time.Duration, error) {
	logger := csiLog.FromContext(ctx).WithName("handleSpectrumScaleConnectors")
	logger.Info("Checking IBM Storage Scale connectors")

	//scaleGuiSecretChanged := false
	requeAfterDelay := time.Duration(0)
	operatorRestarted := (len(scaleConnMap) == 0) && cmExists
	for _, cluster := range instance.Spec.Clusters {
		isPrimaryCluster := cluster.Primary != nil
		if !cmExists || clustersStanzaModified || operatorRestarted {
			//These are the prerequisite checks and preprocessing done at
			//multiple passes of operator/driver:
			//1st pass: check all clusters - check if GUI is reachable and clusterID is valid for all clusters.
			//Pass no. 2+ (without clusters stanza modification in manifest): check no cluster.
			//Pass no. 2+ (with applying modified clusters stanza in manifest): check GUI of changed cluster is reachable + clusterID is valid.
			//Operator restarted: check if only primary GUI is reachable.
			//Driver started/restarted: check if only primary GUI is reachable.
			_, connectorExists := scaleConnMap[cluster.Id]
			_, isClusterChanged := changedClusters[cluster.Id]

			//Create a new connector if it does not exists already or
			//if it exists but cluster stanza is modified and this cluster
			//data is changed
			if !connectorExists || (clustersStanzaModified && isClusterChanged) {
				connector, err := r.newConnector(ctx, instance, cluster)
				if err != nil {
					return requeAfterDelay, err
				}
				scaleConnMap[cluster.Id] = connector
				if isPrimaryCluster {
					scaleConnMap[config.Primary] = connector
				}
			}

			//Validate GUI connection and cluster ID in CR.
			//1. Check if GUI is reachable for the 1st pass.
			// For pass no. 2+ if clusterstanza modified, check for only changed cluster
			if !cmExists || (clustersStanzaModified && isClusterChanged) {
				if operatorRestarted && !isPrimaryCluster {
					//if operator is restarted and this is not a primary cluster,
					//no need to check if GUI is reachable or clusterID is valid.
					continue
				}
				id, err := scaleConnMap[cluster.Id].GetClusterId(context.TODO())
				if err != nil {
					message := fmt.Sprintf("Failed to connect to the GUI of the cluster with ID: %s", cluster.Id)
					if strings.Contains(err.Error(), config.ErrorUnauthorized) {
						message += ". " + config.ErrorUnauthorized + ", the Secret " + cluster.Secrets +
							" has incorrect credentials, please correct the credentials"
						requeAfterDelay = 1 * time.Minute
					} else if strings.Contains(err.Error(), config.ErrorForbidden) {
						message += ". " + config.ErrorForbidden +
							", GUI user specified in the Secret " + cluster.Secrets +
							" is locked due to multiple connection attempts with incorrect credentials," +
							" please contact IBM Storage Scale Administrator"
						requeAfterDelay = 1 * time.Minute
					}

					logger.Error(err, message)
					SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
						metav1.ConditionFalse, string(csiv1.GUIConnFailed), message,
					)
					//remove the connector if GUI connection fails
					delete(scaleConnMap, cluster.Id)
					return requeAfterDelay, err
				} else {
					logger.Info("The GUI connection for the cluster is successful", "Cluster ID", cluster.Id)
				}
				//2. Check if cluster ID from manifest matches with the one obtained from GUI
				if operatorRestarted {
					//If this is the operator restart case, no need to validate cluster ID again for any cluster,
					//as it gets validated in 1st pass already.
					continue
				}
				if id != cluster.Id {
					message := fmt.Sprintf("The cluster ID %v in IBM Storage Scale CSI configurations does not match with the cluster ID %v obtained from cluster."+
						" Please check the Spec.Clusters section in the resource %s/%s", cluster.Id, id, instance.Kind, instance.Name,
					)
					logger.Error(err, message)
					SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
						metav1.ConditionFalse, string(csiv1.ClusterIDMismatch), message,
					)
					return requeAfterDelay, fmt.Errorf("%s", message)
				} else {
					logger.Info(fmt.Sprintf("The cluster ID %s is validated successfully", cluster.Id))
				}
			}
		}
	}
	return requeAfterDelay, nil
}

// handlePrimaryFSandFileset checks if primary FS exists, also checkes if primary fileset exists.
// If primary fileset does not exist, it is created and also if a directory
// to store symlinks is created if it does not exist. It returns the absolute path of symlink
// directory and error if there is any.
func (r *CSIScaleOperatorReconciler) handlePrimaryFSandFileset(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) (time.Duration, error) {
	logger := csiLog.FromContext(ctx).WithName("handlePrimaryFSandFileset")
	requeAfterDelay := time.Duration(0)
	primaryReference := r.getPrimaryCluster(instance)
	if primaryReference == nil {
		message := fmt.Sprintf("No primary cluster is defined in the IBM Storage Scale CSI configurations under Spec.Clusters section in the CSISCaleOperator instance %s/%s", instance.Kind, instance.Name)
		err := fmt.Errorf("%s", message)
		logger.Error(err, "")
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.PrimaryClusterUndefined), message,
		)
		return requeAfterDelay, err
	}

	primary := *primaryReference
	sc := scaleConnMap[config.Primary]

	// check if primary filesystem exists
	fsMountInfo, err := sc.GetFilesystemMountDetails(context.TODO(), primary.PrimaryFs)
	if err != nil {
		requeAfterDelay = 2 * time.Minute
		message := fmt.Sprintf("Failed to get the details of the primary filesystem: %s, retrying after 2 minutes", primary.PrimaryFs)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFileSystemFailed), message,
		)
		return requeAfterDelay, err
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
	// 		return "", err
	// 	}
	// }

	if primary.RemoteCluster != "" {
		//if remote cluster is present, use connector of remote cluster
		sc = scaleConnMap[primary.RemoteCluster]
		if fsNameOnOwningCluster == "" {
			message := "failed to get the name of the remote filesystem from the cluster"
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.GetRemoteFileSystemFailed), message,
			)
			return requeAfterDelay, fmt.Errorf("%s", message)
		}
	}

	//check if primary filesystem exists on remote cluster and mounted on atleast one node
	fsMountInfo, err = sc.GetFilesystemMountDetails(context.TODO(), fsNameOnOwningCluster)
	if err != nil {
		message := fmt.Sprintf("Failed to the get details of the filesystem: %s", fsNameOnOwningCluster)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFileSystemFailed), message,
		)
		return requeAfterDelay, err
	}

	fsMountPoint := fsMountInfo.MountPoint

	fsetLinkPath, err := r.createPrimaryFileset(ctx, instance, sc, fsNameOnOwningCluster, fsMountPoint, primary.PrimaryFset, primary.InodeLimit)
	if err != nil {
		message := fmt.Sprintf("Failed to create the primary fileset %s on the primary filesystem %s", primary.PrimaryFset, primary.PrimaryFs)
		logger.Error(err, message)
		return requeAfterDelay, err
	}

	// In case primary FS is remotely mounted, run fileset refresh task on primary cluster
	if primary.RemoteCluster != "" {
		filesetInfo, err := scaleConnMap[config.Primary].ListFileset(context.TODO(), primary.PrimaryFs, primary.PrimaryFset)
		if err != nil || reflect.ValueOf(filesetInfo).IsZero() {
			logger.Info("Primary fileset is not visible on primary cluster. Running fileset refresh task", "fileset name", primary.PrimaryFset)
			err = scaleConnMap[config.Primary].FilesetRefreshTask(context.TODO())
			if err != nil {
				message := "error in fileset refresh task"
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.FilesetRefreshFailed), message,
				)
				return requeAfterDelay, err
			}

			// retry listing fileset again after some time after refresh
			time.Sleep(8 * time.Second)
			filesetInfo, err = scaleConnMap[config.Primary].ListFileset(context.TODO(), primary.PrimaryFs, primary.PrimaryFset)
			if err != nil || reflect.ValueOf(filesetInfo).IsZero() {
				message := fmt.Sprintf("Primary fileset %s is not visible on primary cluster even after running fileset refresh task", primary.PrimaryFset)
				logger.Error(err, message)
				SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
					metav1.ConditionFalse, string(csiv1.GetFilesetFailed), message,
				)
				return requeAfterDelay, err
			}
		}
	}

	//A directory can be created from accessing cluster, so get the path on accessing cluster
	if fsMountPoint != primaryFSMount {
		fsetLinkPath = strings.Replace(fsetLinkPath, fsMountPoint, primaryFSMount, 1)
	}

	// Create directory where volume symlinks will reside
	symlinkDirPath, _, err := r.createSymlinksDir(ctx, instance, scaleConnMap[config.Primary], primary.PrimaryFs, primaryFSMount, fsetLinkPath)
	if err != nil {
		message := fmt.Sprintf("Failed to create the directory %s on the primary filesystem %s", config.SymlinkDir, primary.PrimaryFs)
		logger.Error(err, message)
		return requeAfterDelay, err
	}
	logger.Info("The symlinks directory path is:", "symlinkDirPath", symlinkDirPath)
	return requeAfterDelay, nil
}

// getPrimaryCluster returns primary cluster of the passed instance.
func (r *CSIScaleOperatorReconciler) getPrimaryCluster(instance *csiscaleoperator.CSIScaleOperator) *csiv1.CSIFilesystem {
	var primary *csiv1.CSIFilesystem
	for _, cluster := range instance.Spec.Clusters {
		if cluster.Primary != nil {
			primary = cluster.Primary
		}
	}
	return primary
}

// createPrimaryFileset creates a primary fileset and returns it's path
// where it is linked. If primary fileset exists and is already linked,
// the link path is returned. If primary fileset already exists and not linked,
// it is linked and link path is returned.
func (r *CSIScaleOperatorReconciler) createPrimaryFileset(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, sc connectors.SpectrumScaleConnector, fsNameOnOwningCluster string,
	fsMountPoint string, filesetName string, inodeLimit string) (string, error) {
	logger := csiLog.FromContext(ctx).WithName("createPrimaryFileset")
	logger.Info("Creating primary fileset", " primaryFS", fsNameOnOwningCluster,
		"mount point", fsMountPoint, "filesetName", filesetName)

	newLinkPath := path.Join(fsMountPoint, filesetName) //Link path to set if the fileset is not linked

	// create primary fileset if not already created
	fsetResponse, err := sc.ListFileset(context.TODO(), fsNameOnOwningCluster, filesetName)
	if err != nil {
		message := fmt.Sprintf("Failed to list fileset in filesystem %s", fsNameOnOwningCluster)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFilesetFailed), message,
		)
		return "", err
	} else if reflect.ValueOf(fsetResponse).IsZero() {
		logger.Info("Primary fileset not found, so creating it", "fileseName", filesetName)
		opts := make(map[string]interface{})
		if inodeLimit != "" {
			opts[connectors.UserSpecifiedInodeLimit] = inodeLimit
		}

		err = sc.CreateFileset(context.TODO(), fsNameOnOwningCluster, filesetName, opts)
		if err != nil {
			message := fmt.Sprintf("Failed to create the primary fileset %s on the filesystem %s", filesetName, fsNameOnOwningCluster)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.CreateFilesetFailed), message,
			)
			return "", err
		}
		logger.Info("Primary fileset is created successfully", "filesetName", filesetName)

		fsetResponse, err = sc.ListFileset(context.TODO(), fsNameOnOwningCluster, filesetName)
		if err != nil {
			// fileset got created but listing failed, return without cleanup
			message := fmt.Sprintf("unable to list newly created primary fileset [%v] in filesystem [%v]. Error: %v", filesetName, fsNameOnOwningCluster, err)
			logger.Error(err, message)
			return "", err
		}
	} else {
		if !strings.Contains(fsetResponse.Config.Comment, connectors.FilesetComment) {
			message := fmt.Sprintf("Primary fileset [%s] is not created by IBM Storage Scale CSI driver. Cannot use it.", filesetName)
			err := fmt.Errorf("%s", message)
			logger.Error(err, "")
			return "", err
		}
	}

	linkPath := fsetResponse.Config.Path
	if linkPath == "" || linkPath == "--" {
		logger.Info("Primary fileset not linked. Linking it", "filesetName", filesetName)
		err = sc.LinkFileset(context.TODO(), fsNameOnOwningCluster, filesetName, newLinkPath)
		if err != nil {
			message := fmt.Sprintf("Failed to link the primary fileset %s to the linkpath %s on the filesystem %s", filesetName, newLinkPath, fsNameOnOwningCluster)
			logger.Error(err, message)
			SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
				metav1.ConditionFalse, string(csiv1.LinkFilesetFailed), message,
			)
			return "", err
		} else {
			logger.Info("Linked primary fileset", "filesetName", filesetName, "linkpath", newLinkPath)
		}
	} else {
		logger.Info("Primary fileset exists and linked", "filesetName", filesetName, "linkpath", linkPath)
	}
	return newLinkPath, nil
}

// createSymlinksDir creates a .volumes directory on the fileset path fsetLinkPath,
// and returns absolute, relative paths and error if there is any.
func (r *CSIScaleOperatorReconciler) createSymlinksDir(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, sc connectors.SpectrumScaleConnector, fs string, fsMountPath string,
	fsetLinkPath string) (string, string, error) {
	logger := csiLog.FromContext(ctx).WithName("createSymlinkPath")
	logger.Info("Creating a directory for symlinks", "directory", config.SymlinkDir,
		"filesystem", fs, "fsMountPath", fsMountPath, "filesetlinkpath", fsetLinkPath)

	fsetRelativePath, symlinkDirPath := getSymlinkDirPath(fsetLinkPath, fsMountPath)
	symlinkDirRelativePath := fmt.Sprintf("%s/%s", fsetRelativePath, config.SymlinkDir)

	err := sc.MakeDirectory(context.TODO(), fs, symlinkDirRelativePath, config.DefaultUID, config.DefaultGID) //MakeDirectory doesn't return error if the directory already exists
	if err != nil {
		message := fmt.Sprintf("Failed to create a symlink directory with relative path %s on filesystem %s", symlinkDirRelativePath, fs)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.CreateDirFailed), message,
		)
		return symlinkDirPath, symlinkDirRelativePath, err
	}

	return symlinkDirPath, symlinkDirRelativePath, nil
}

// getSymlinkDirPath formats and returns the paths of the directory,
// where symlinks are stored for version 1 volumes.
func getSymlinkDirPath(fsetLinkPath string, fsMountPath string) (string, string) {
	fsetRelativePath := strings.Replace(fsetLinkPath, fsMountPath, "", 1)
	fsetRelativePath = strings.Trim(fsetRelativePath, "!/")
	fsetLinkPath = strings.TrimSuffix(fsetLinkPath, "/")

	symlinkDirPath := fmt.Sprintf("%s/%s", fsetLinkPath, config.SymlinkDir)
	return fsetRelativePath, symlinkDirPath
}

// ValidateCRParams validates driver configuration parameters and returns error if any validation fails
func ValidateCRParams(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator) error {
	logger := csiLog.FromContext(ctx).WithName("ValidateCRParams")
	logger.Info(fmt.Sprintf("Validating the IBM Storage Scale CSI configurations of the resource %s/%s", instance.Kind, instance.Name))

	if len(instance.Spec.Clusters) == 0 {
		return fmt.Errorf("missing cluster information in IBM Storage Scale configuration")
	}

	primaryClusterFound, issueFound := false, false
	remoteClusterID := ""
	var nonPrimaryClusters = make(map[string]bool)

	for i := 0; i < len(instance.Spec.Clusters); i++ {
		cluster := instance.Spec.Clusters[i]

		if cluster.Id == "" {
			issueFound = true
			logger.Error(fmt.Errorf("mandatory parameter 'id' is not specified"), "")
		}
		if len(cluster.RestApi) == 0 {
			issueFound = true
			logger.Error(fmt.Errorf("mandatory section 'restApi' is not specified for cluster %v", cluster.Id), "")
		}
		if len(cluster.RestApi) != 0 && cluster.RestApi[0].GuiHost == "" {
			issueFound = true
			logger.Error(fmt.Errorf("mandatory parameter 'guiHost' is not specified for cluster %v", cluster.Id), "")
		}

		if cluster.Primary != nil && *cluster.Primary != (csiv1.CSIFilesystem{}) {
			if primaryClusterFound {
				issueFound = true
				logger.Error(fmt.Errorf("more than one primary clusters specified"), "")
			}

			primaryClusterFound = true
			if cluster.Primary.PrimaryFs == "" {
				issueFound = true
				logger.Error(fmt.Errorf("mandatory parameter 'primaryFs' is not specified for primary cluster %v", cluster.Id), "")
			}

			remoteClusterID = cluster.Primary.RemoteCluster
		} else {
			//when its a not primary cluster
			nonPrimaryClusters[cluster.Id] = true
		}

		if cluster.Secrets == "" {
			issueFound = true
			logger.Error(fmt.Errorf("mandatory parameter 'secrets' is not specified for cluster %v", cluster.Id), "")
		}

		if cluster.SecureSslMode && cluster.Cacert == "" {
			issueFound = true
			logger.Error(fmt.Errorf("ca certificate not specified in secure SSL mode for cluster %v", cluster.Id), "")
		}
	}

	if !primaryClusterFound {
		issueFound = true
		logger.Error(fmt.Errorf("no primary clusters specified"), "")
	}
	_, nonPrimaryClusterExists := nonPrimaryClusters[remoteClusterID]
	if remoteClusterID != "" && !nonPrimaryClusterExists {
		issueFound = true
		logger.Error(fmt.Errorf("remote cluster specified for primary filesystem: %s, but no entry found for it in driver manifest", remoteClusterID), "")
	}

	if issueFound {
		message := "one or more issues found while validating driver manifest, check operator logs for details"
		return fmt.Errorf("%s", message)
	}
	return nil
}

// getConfigMap fetches data from the "ibm-spectrum-scale-csi-config" configmap from the cluster
// and returns a configmap reference.
func (r *CSIScaleOperatorReconciler) getConfigMap(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, name string) (*corev1.ConfigMap, error) {
	logger := csiLog.FromContext(ctx).WithName("getConfigMap").WithValues("Kind", corev1.ResourceConfigMaps, "Name", name)
	logger.Info("Reading optional CSI configmap resource from the cluster.")

	cm := &corev1.ConfigMap{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: instance.Namespace,
	}, cm)
	if err != nil && errors.IsNotFound(err) {
		message := fmt.Sprintf("Optional ConfigMap resource %s not found", name)
		logger.Info(message)
	} else if err != nil {
		message := fmt.Sprintf("Failed to get the optional ConfigMap: %s", name)
		logger.Error(err, message)
		SetStatusAndRaiseEvent(instance, r.Recorder, corev1.EventTypeWarning, string(config.StatusConditionSuccess),
			metav1.ConditionFalse, string(csiv1.GetFailed), message,
		)
	}
	return cm, err
}

// parseConfigMap parses the data in the configMap in the desired format(VAR_DRIVER_ENV_NAME: VALUE to ENV_NAME: VALUE).
func (r *CSIScaleOperatorReconciler) parseConfigMap(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, cm *corev1.ConfigMap) map[string]string {
	logger := csiLog.FromContext(ctx).WithName("parseConfigMap").WithValues("Name", config.EnvVarConfigMap)
	logger.Info("Parsing the data from the optional configmap.", "configmap", config.EnvVarConfigMap)

	validEnvMap := map[string]string{}
	invalidEnvKeys := []string{}
	invalidEnvValueMap := map[string]string{}
	for key, value := range cm.Data {
		keyUpper := strings.ToUpper(key)
		if containsStringInSlice(config.CSIOptionalConfigMapKeys[:], keyUpper) {
			switch keyUpper {
			case config.EnvLogLevelKeyPrefixed:
				validateEnvVarValue(config.EnvLogLevelValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.EnvPersistentLogKeyPrefixed:
				validateEnvVarValue(config.EnvPersistentLogValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.EnvNodePublishMethodKeyPrefixed:
				validateEnvVarValue(config.EnvNodePublishMethodValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.EnvVolumeStatsCapabilityKeyPrefixed:
				validateEnvVarValue(config.EnvVolumeStatsCapabilityValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.EnvDiscoverCGFilesetKeyPrefixed:
				validateEnvVarValue(config.EnvDiscoverCGFilesetValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.EnvPrimaryFilesystemKeyPrefixed:
				validateEnvVarValue(config.EnvPrimaryFilesystemValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.EnvVolNamePrefixKeyPrefixed:
				validateVolNamePrefix(ctx, keyUpper, strings.ToLower(value), validEnvMap, invalidEnvValueMap)
			case config.DaemonSetUpgradeMaxUnavailableKey:
				validateMaxUnavailableValue(ctx, keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.HostNetworkKey:
				validateHostNetworkValue(ctx, config.EnvHostNetworkValues[:], keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.DriverCPULimits:
				validateCPULimitsValue(ctx, keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.DriverMemoryLimits:
				validateMemoryLimitsValue(ctx, keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.SidecarCPULimits:
				validateCPULimitsValue(ctx, keyUpper, value, validEnvMap, invalidEnvValueMap)
			case config.SidecarMemoryLimits:
				validateMemoryLimitsValue(ctx, keyUpper, value, validEnvMap, invalidEnvValueMap)
			}
		} else {
			invalidEnvKeys = append(invalidEnvKeys, key)
		}
	}
	// setting default values if values are empty/wrong
	setDefaultDriverEnvValues(ctx, validEnvMap)
	logger.Info("Final accepted value ", "from the optional configmap", validEnvMap)
	var message string = ""
	if len(invalidEnvKeys) > 0 && len(invalidEnvValueMap) > 0 {
		message = fmt.Sprintf("There are few entries %v with wrong key which will not be processed and few entries having wrong values %v in the configmap %s, default values will be used", invalidEnvKeys, invalidEnvValueMap, config.EnvVarConfigMap)

	} else if len(invalidEnvKeys) > 0 {
		message = fmt.Sprintf("There are few entries %v with wrong key in the configmap %s which will not be processed", invalidEnvKeys, config.EnvVarConfigMap)

	} else if len(invalidEnvValueMap) > 0 {
		message = fmt.Sprintf("There are few entries having wrong values %v in the configmap %s, default values will be used", invalidEnvValueMap, config.EnvVarConfigMap)
	}

	if len(message) > 0 {
		logger.Info(message)
		RaiseCSOEvent(instance, r.Recorder, corev1.EventTypeWarning,
			string(csiv1.ValidationWarning), message,
		)
	}

	logger.Info("Parsing the data from the optional configmap is successful", "configmap", config.EnvVarConfigMap)
	return validEnvMap
}

func SetStatusAndRaiseEvent(instance runtime.Object, rec record.EventRecorder,
	eventType string, conditionType string, status metav1.ConditionStatus, reason string, msg string) {
	meta.SetStatusCondition(&crStatus.Conditions, metav1.Condition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: msg,
	})
	rec.Event(instance, eventType, reason, msg)
}

func RaiseCSOEvent(instance runtime.Object, rec record.EventRecorder,
	eventType string, reason string, msg string) {
	rec.Event(instance, eventType, reason, msg)
}

func validateMaxUnavailableValue(ctx context.Context, key string, value string, data map[string]string, invalidEnvValue map[string]string) {
	logger := csiLog.FromContext(ctx).WithName("validateMaxUnavailableValue")
	logger.Info("Validating daemonset maxunavailable input ", "inputMaxunavailable", value)
	input := strings.TrimSuffix(value, "%")
	if s, err := strconv.Atoi(input); err == nil && (s > 0 && s < 100) && strings.HasSuffix(value, "%") {
		logger.Info("daemonset maxunavailable parsed integer ", "inputMaxunavailableInt", s)
		data[key] = value
	} else {
		logger.Error(err, " Failed to parse the input maxunvaialble value")
		invalidEnvValue[key] = value
	}
}

func validateVolNamePrefix(ctx context.Context, key string, value string, data map[string]string, invalidEnvValue map[string]string) {
	logger := csiLog.FromContext(ctx).WithName("validateVolumeNamePrefix")
	logger.Info("Validating volume name prefix input ", "volumeNamePrefix", value)

	if len(value) > 2 && len(value) < 6 && isLetter(value) {
		logger.Info("validateVolNamePrefix parsed :", "volumeNamePrefix", value)
		data[key[11:]] = value
	} else {
		logger.Error(fmt.Errorf("the input volume name prefix is not right,the prefix value must be between [3 to 5] length, doesn't contains any special characters and numbers,  volumeNamePrefix : %v", value), "Volume Name Prefix Error")
		invalidEnvValue[key] = value
	}
}

func isLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func validateHostNetworkValue(ctx context.Context, inputSlice []string, key, value string, envMap, invalidEnvValue map[string]string) {
	logger := csiLog.FromContext(ctx).WithName("validateHostNetworkValue")
	logger.Info("Validating host network input", "inputHostNetwork", value)

	if containsStringInSlice(inputSlice, strings.ToUpper(value)) {
		envMap[key] = value
	} else {
		invalidEnvValue[key] = value
	}
}

// validateCPULimitsValue: To set only accepted CPU limit from configMap
func validateCPULimitsValue(ctx context.Context, key string, value string, data map[string]string, invalidEnvValue map[string]string) {
	logger := csiLog.FromContext(ctx).WithName("validateCPULimitsValue")
	logger.Info("Validating CPU limits Value input ", "cpuLimits", value)

	isResourceRequestValid, err := checkMinResourceRequirement(ctx, value, config.PodsCPULimitsLowerValue)
	if isResourceRequestValid {
		logger.Info("Validation of CPU limits Value successful ", "cpuLimits", value)
		data[key] = value
	} else {
		logger.Error(fmt.Errorf("failed to parse [inputCPULimitValue=%s] or wrong value passed. The value must be greater than or equal to %s, error : %v", value, config.PodsCPULimitsLowerValue, err), "CPU limits validation error")
		invalidEnvValue[key] = value
	}
}

// validateMemoryLimitsValue: To set only accepted memory limit from configMap
func validateMemoryLimitsValue(ctx context.Context, key string, value string, data map[string]string, invalidEnvValue map[string]string) {
	logger := csiLog.FromContext(ctx).WithName("validateMemoryLimitsValue")
	logger.Info("Validating memory limits Value input ", "memoryLimits", value)

	isResourceRequestValid, err := checkMinResourceRequirement(ctx, value, config.PodsMemoryLimitsLowerValue)
	if isResourceRequestValid {
		logger.Info("Validation of memomry limits Value successful ", "memoryLimits", value)
		data[key] = value
	} else {
		logger.Error(fmt.Errorf("failed to parse [inputMemoryLimitValue=%s] or wrong value passed. The value must be greater than or equal to %s, error : %v", value, config.PodsMemoryLimitsLowerValue, err), "Memory limits validation error")
		invalidEnvValue[key] = value
	}
}

// checkMinResourceRequirement: Requested resource must be greater or equal to it's lower limit.
// Return true when requestedresource is greater than min required resource
func checkMinResourceRequirement(ctx context.Context, resourceRequested string, minResourceRequired string) (bool, error) {
	logger := csiLog.FromContext(ctx).WithName("checkMinResourceRequirement")
	logger.Info("Validating resource limit requirement ", "resourceRequested", resourceRequested, "minResourceRequired", minResourceRequired)
	var minResourceSizeMilli, resourceRequestedMilliValue int64
	minResourceSize, err := resource.ParseQuantity(minResourceRequired)
	if err != nil {
		return false, err
	} else {
		minResourceSizeMilli = minResourceSize.MilliValue()
	}
	resourceSize, err := resource.ParseQuantity(resourceRequested)
	if err != nil {
		return false, err
	} else {
		resourceRequestedMilliValue = resourceSize.MilliValue()
	}
	if minResourceSizeMilli <= resourceRequestedMilliValue {
		return true, nil
	} else {
		return false, nil
	}
}

// containsStringInSlice checks if a string is present in a slice
func containsStringInSlice(inputSlice []string, stringToFind string) bool {
	for _, v := range inputSlice {
		if v == stringToFind {
			return true
		}
	}
	return false
}

// validateEnvVarValue checks if the input environment variable value is valid (one of the allowed values), and if
// it is valid then removes predefined prefix from the key which is associated only for driver pod env var
// and adds the env var to validEnvMap. If the value is not valid, then adds the env var to invalidEnvMap
func validateEnvVarValue(inputSlice []string, key string, value string, validEnvMap map[string]string, invalidEnvMap map[string]string) {
	if containsStringInSlice(inputSlice, strings.ToUpper(value)) {
		validEnvMap[key[11:]] = value
	} else {
		invalidEnvMap[key] = value
	}
}

func setDefaultDriverEnvValues(ctx context.Context, envMap map[string]string) {
	logger := csiLog.FromContext(ctx).WithName("setDefaultDriverEnvValues")
	// Set default LogLevel when it is not present in envMap
	if _, ok := envMap[config.EnvLogLevelKey]; !ok {
		logger.Info("logger level is empty or incorrect.", "Defaulting logLevel to", config.EnvLogLevelDefaultValue)
		envMap[config.EnvLogLevelKey] = config.EnvLogLevelDefaultValue
	}
	// Set default PersistentLog when it is not present in envMap
	if _, ok := envMap[config.EnvPersistentLogKey]; !ok {
		logger.Info("PersistentLog is empty or incorrect.", "Defaulting PersistentLog to", config.EnvPersistentLogDefaultValue)
		envMap[config.EnvPersistentLogKey] = config.EnvPersistentLogDefaultValue
	}
	// Set default NodePublishMethod when it is not present in envMap
	if _, ok := envMap[config.EnvNodePublishMethodKey]; !ok {
		logger.Info("NodePublishMethod is empty or incorrect.", "Defaulting NodePublishMethod to", config.EnvNodePublishMethodDefaultValue)
		envMap[config.EnvNodePublishMethodKey] = config.EnvNodePublishMethodDefaultValue
	}
	// Set default VolumeStatsCapability when it is not present in envMap
	if _, ok := envMap[config.EnvVolumeStatsCapabilityKey]; !ok {
		logger.Info("VolumeStatsCapability is empty or incorrect.", "Defaulting VolumeStatsCapability to", config.EnvVolumeStatsCapabilityDefaultValue)
		envMap[config.EnvVolumeStatsCapabilityKey] = config.EnvVolumeStatsCapabilityDefaultValue
	}
	// Set default VolNamePrefix when it is not present in envMap
	if _, ok := envMap[config.EnvVolNamePrefixKey]; !ok {
		logger.Info("VolNamePrefix is empty or incorrect.", "Defaulting VolNamePrefix to", config.EnvVolNamePrefixDefaultValue)
		envMap[config.EnvVolNamePrefixKey] = config.EnvVolNamePrefixDefaultValue
	}
	// Make empty DaemonSetUpgradeMaxUnavailable when it is not present in envMap
	if _, ok := envMap[config.DaemonSetUpgradeMaxUnavailableKey]; !ok {
		logger.Info("DaemonSetUpgradeMaxUnavailable is empty or incorrect.", "Defaulting DaemonSetUpgradeMaxUnavailable to", "1")
		envMap[config.DaemonSetUpgradeMaxUnavailableKey] = ""
	}
	// Set default DiscoverCGFileset when it is not present in envMap
	if _, ok := envMap[config.EnvDiscoverCGFilesetKey]; !ok {
		envDiscoverCGFilesetDefaultValue := getDiscoverCGFilesetDefaultValue(ctx)
		logger.Info("DiscoverCGFileset is empty or incorrect.", "Defaulting DiscoverCGFileset to", envDiscoverCGFilesetDefaultValue)
		envMap[config.EnvDiscoverCGFilesetKey] = envDiscoverCGFilesetDefaultValue
	}

	// set default HostNetwork env when it is not present in envMap
	if _, ok := envMap[config.HostNetworkKey]; !ok {
		logger.Info("Host Network is empty or incorrect.", "Defaulting Host Network to", config.EnvHostNetworkDefaultValue)
		envMap[config.HostNetworkKey] = config.EnvHostNetworkDefaultValue
	}

	// set default CPU limits for driver container when it is not present in envMap
	if _, ok := envMap[config.DriverCPULimits]; !ok {
		logger.Info("Driver CPU limits is empty or incorrect.", "Defaulting CPU limits to", config.DriverCPULimitsDefaultValue)
		envMap[config.DriverCPULimits] = config.DriverCPULimitsDefaultValue
	}
	// set default Memory limits for driver container when it is not present in envMap
	if _, ok := envMap[config.DriverMemoryLimits]; !ok {
		logger.Info("Driver Memory limits is empty or incorrect.", "Defaulting Memory limits to", config.DriverMemoryLimitsDefaultValue)
		envMap[config.DriverMemoryLimits] = config.DriverMemoryLimitsDefaultValue
	}
	// set default CPU limits for sidecars container when it is not present in envMap
	if _, ok := envMap[config.SidecarCPULimits]; !ok {
		logger.Info("Sidecars CPU limits is empty or incorrect.", "Defaulting CPU limits to", config.SidecarCPULimitsDefaultValue)
		envMap[config.SidecarCPULimits] = config.SidecarCPULimitsDefaultValue
	}
	// set default Memory limits for sidecars container when it is not present in envMap
	if _, ok := envMap[config.SidecarMemoryLimits]; !ok {
		logger.Info("Sidecars Memory limits is empty or incorrect.", "Defaulting Memory limits to", config.SidecarMemoryLimitsDefaultValue)
		envMap[config.SidecarMemoryLimits] = config.SidecarMemoryLimitsDefaultValue
	}
	// set default EnvPrimaryFilesystemKey when it is not present in envMap
	if _, ok := envMap[config.EnvPrimaryFilesystemKey]; !ok {
		logger.Info("PRIMARY_FILESYSTEM is empty or incorrect.", "Defaulting PRIMARY_FILESYSTEM to", config.EnvPrimaryFilesystemDefaultValue)
		envMap[config.EnvPrimaryFilesystemKey] = config.EnvPrimaryFilesystemDefaultValue
	}
}

// getDiscoverCGFilesetDefaultValue returns default value for CG fileset discovery
func getDiscoverCGFilesetDefaultValue(ctx context.Context) string {
	logger := csiLog.FromContext(ctx).WithName("getDiscoverCGFilesetDefaultValue")
	logger.Info("Getting DISCOVER_CG_FILESET default value")

	if CNSAClusterPresence, ok := clusterTypeData[config.ENVClusterCNSAPresenceCheck]; ok {

		if CNSAClusterPresence == "True" {
			//CG fileset discovery is enabled by default with CNSA cluster required for RDR
			return "ENABLED"
		}
	}
	//CG fileset discovery is disabled by default on k8s cluster
	return "DISABLED"
}

// listGUIPasswdExpiredClusters returns a list having clusterIds whose password is expired
func (r *CSIScaleOperatorReconciler) listGUIPasswdExpiredClusters(ctx context.Context, instance *csiscaleoperator.CSIScaleOperator, clusters []csiv1.CSICluster) []string {
	logger := csiLog.FromContext(ctx).WithName("listPasswdExpiredClusters")
	logger.Info("Checking each cluster if its GUI password expired")

	expiredGui := []string{}
	for _, cls := range clusters {
		logger.Info("Checking each cluster ", "cluster", cls)
		connector, err := r.newConnector(ctx, instance, cls)
		if err != nil {
			logger.Error(err, "Failed to create scale connector")
		} else {
			_, err = connector.GetClusterId(context.TODO())
			if err != nil && isGUIUnauthorized(err) {
				expiredGui = append(expiredGui, cls.Id)
			}
		}
	}
	return expiredGui
}

func isGUIUnauthorized(err error) bool {
	return strings.Contains(err.Error(), config.ErrorUnauthorized)
}

// getClusterTypeAndCNSAOperatorPresence fetches CNSA operator deployment from the cluster
// and sets two parameters in the clusterTypeData map which is being set later in environment of the driver
func getClusterTypeAndCNSAOperatorPresence(ctx context.Context, namespace string) (err error) {

	// Checking the presence of CNSA operator into the cluster
	logger := csiLog.FromContext(ctx).WithName("getClusterTypeAndCNSAOperatorPresence")
	logger.Info("Reading resources from the cluster in the", "Namespace", namespace)

	// Default env/cluster type setup
	clusterTypeData[config.ENVClusterConfigurationType] = config.ENVClusterTypeKubernetes
	clusterTypeData[config.ENVClusterCNSAPresenceCheck] = "False"

	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Error(err, "Failed to set inclusterconfig instance")
		return err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(inClusterConfig)
	if err != nil {
		logger.Error(err, "Failed to set clientset with config")
		return err
	}

	// get deployments in a ibm-spectrum-scale-operator namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to list of all deployments from the ibm-spectrum-scale-operator namespace")
		return err
	}

	for _, deployment := range deployments.Items {
		if strings.Contains(deployment.GetName(), config.CNSAOperatorDeploymentName) {
			logger.Info("Cluster is having CNSA deployment")
			clusterTypeData[config.ENVClusterCNSAPresenceCheck] = "True"
			break
		}
	}

	// Checking if the platform is OpenShift
	apiList, err := clientset.ServerGroups()
	if err != nil {
		logger.Error(err, "Failed to get ServerGroups while getting openshift platform type")
		return err
	}

	apiGroups := apiList.Groups
	for i := 0; i < len(apiGroups); i++ {
		if apiGroups[i].Name == "route.openshift.io" {
			logger.Info("Found apiGroup having route.openshift.io :", "apiGroup", apiGroups[i])
			logger.Info("Platform is a Openshift")
			clusterTypeData[config.ENVClusterConfigurationType] = config.ENVClusterTypeOpenshift
			break
		}
	}
	return nil
}
