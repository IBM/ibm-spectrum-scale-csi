/**
 * Copyright 2022 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

// Kubernetes built-in well-known constants
// For more information: https://kubernetes.io/docs/reference/labels-annotations-taints/
const (
	// LabelAppName is the name of the component application.
	LabelAppName = "app.kubernetes.io/name"
	// LabelAppInstance identifies resources related to a unique Cluster deployment.
	LabelAppInstance = "app.kubernetes.io/instance"
	// LabelAppManagedBy is the controller/user who created the resource.
	LabelAppManagedBy = "app.kubernetes.io/managed-by"
	// LabelArchitecture is the label applied on node, used to identify the architecture of node.
	LabelArchitecture     = "kubernetes.io/arch"
	LabelNodeMaster       = "node-role.kubernetes.io/master"
	LabelNodeInfra        = "node-role.kubernetes.io/infra"
	LabelNodeControlPlane = "node-role.kubernetes.io/control-plane"
)

// CSI resource labels
const (
	// LabelName is the name of the product.
	LabelProduct = "product"
	// LabelRelease is the current version of the application.
	LabelRelease = "release"
	// LabelApp is the name of the CSI application.
	LabelApp = "app"
)

// CSI resource label values
const (
	// Product is the name of the application.
	Product = "ibm-spectrum-scale-csi"
	// Release is the current version of the application. // TODO: Update description with relevant information.
	Release = "ibm-spectrum-scale-csi-operator"
	// ResourceName is the name of the operator
	ResourceAppName = "ibm-spectrum-scale-csi-operator"
	// ResourceInstance is unique name identifying the instance of an application.
	ResourceInstance = "ibm-spectrum-scale-csi-operator"
	// ResourceManagedBy is the controller/user who created this resource.
	ResourceManagedBy = "ibm-spectrum-scale-csi-operator"
)

// Add a field here if it never changes, if it changes over time, put it to settings.go
const (
	APIGroup   = "csi.ibm.com"
	APIVersion = "v1"
	ID         = "ibm-spectrum-scale-csi-operator"
	DriverName = "spectrumscale.csi.ibm.com"
	Kind       = "CSIScaleOperator"

	ProductName = "IBM Spectrum Scale CSI Operator"

	ENVEndpoint     = "ENDPOINT"
	ENVNodeName     = "NODE_NAME"
	ENVKubeVersion  = "KUBE_VERSION"
	ENVCGPrefix     = "CSI_CG_PREFIX"
	ENVSymDirPath   = "SYMLINK_DIR_PATH"
	DriverVersion   = "2.10.6"
	OperatorVersion = "2.10.6"

	// Number of replica pods for CSI Sidecar deployment
	ReplicaCount = int32(2)
	// Tolerations seconds for the CSI Sidecar deployment
	TolerationsSeconds = int64(300)
	// ContainerPort for /healthz/leader-election endpoint
	LeaderLivenessPort = int32(8080)
	// 64-Bit machine architecture supported by IBM Storage Scale CSI.
	AMD64 = "amd64"
	// Power PC machine architecture supported by IBM Storage Scale CSI.
	PPC = "ppc64le"
	// IBM zSystems machine architecture supported by IBM Storage Scale CSI.
	IBMSystem390 = "s390x"

	//  Default images for containers
	CSIDriverPluginImage = "quay.io/ibm-spectrum-scale/ibm-spectrum-scale-csi-driver:v2.10.5"
	//  registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.8.0
	CSINodeDriverRegistrarImage = "registry.k8s.io/sig-storage/csi-node-driver-registrar@sha256:f6717ce72a2615c7fbc746b4068f788e78579c54c43b8716e5ce650d97af2df1" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/livenessprobe:v2.10.0
	LivenessProbeImage = "registry.k8s.io/sig-storage/livenessprobe@sha256:4dc0b87ccd69f9865b89234d8555d3a614ab0a16ed94a3016ffd27f8106132ce" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-attacher:v4.3.0
	CSIAttacherImage = "registry.k8s.io/sig-storage/csi-attacher@sha256:4eb73137b66381b7b5dfd4d21d460f4b4095347ab6ed4626e0199c29d8d021af" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-provisioner:v3.5.0
	CSIProvisionerImage = "registry.k8s.io/sig-storage/csi-provisioner@sha256:d078dc174323407e8cc6f0f9abd4efaac5db27838f1564d0253d5e3233e3f17f" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-snapshotter:v6.2.2
	CSISnapshotterImage = "registry.k8s.io/sig-storage/csi-snapshotter@sha256:becc53e25b96573f61f7469923a92fb3e9d3a3781732159954ce0d9da07233a2" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-resizer:v1.8.0
	CSIResizerImage = "registry.k8s.io/sig-storage/csi-resizer@sha256:2e2b44393539d744a55b9370b346e8ebd95a77573064f3f9a8caf18c22f4d0d0" // #nosec G101 false positive

	//ImagePullPolicies for containers
	CSIDriverImagePullPolicy              = "IfNotPresent"
	CSINodeDriverRegistrarImagePullPolicy = "IfNotPresent"
	LivenessProbeImagePullPolicy          = "IfNotPresent"
	CSIAttacherImagePullPolicy            = "IfNotPresent"
	CSIProvisionerImagePullPolicy         = "IfNotPresent"
	CSISnapshotterImagePullPolicy         = "IfNotPresent"
	CSIResizerImagePullPolicy             = "IfNotPresent"

	CSINodeDriverPlugin       = "csi-spectrum-scale"
	CSINodeDriverRegistrar    = "csi-node-driver-registrar"
	CSIProvisioner            = "csi-provisioner"
	CSIAttacher               = "ibm-spectrum-scale-csi-attacher"
	CSISnapshotter            = "csi-snapshotter"
	CSIResizer                = "csi-resizer"
	LivenessProbe             = "livenessprobe"
	CSIConfigMap              = "spectrum-scale-config"
	ClusterFirstWithHostNet   = "ClusterFirstWithHostNet"
	NodeSocketVolumeMountPath = "/var/lib/ibm/config"
	SocketDir                 = "/plugins/spectrumscale.csi.ibm.com"
	SocketPath                = "/plugins/spectrumscale.csi.ibm.com/csi.sock"
	PluginsRegistry           = "/plugins_registry"
	Plugins                   = "/plugins"
	CSIKubeletRootDirPath     = "/var/lib/kubelet"
	CSISCC                    = "spectrum-scale-csiaccess"
	SecretsMountPath          = "/var/lib/ibm/" // #nosec G101 false positive
	ConfigMapPath             = "/var/lib/ibm/config"
	CAcertMountPath           = "/var/lib/ibm/ssl/public/"
	CSIFinalizer              = "finalizer.csiscaleoperators.csi.ibm.com"

	//Default imagePullSecrets
	ImagePullSecretRegistryKey    = "ibm-spectrum-scale-csi-registrykey" // #nosec G101 false positive
	ImagePullSecretEntitlementKey = "ibm-entitlement-key"                // #nosec G101 false positive

	DaemonSetUpgradeUpdateStrategyType = "RollingUpdate"

	// Optional ConfigMap constants for CSI driver environment variables
	EnvVarConfigMap = "ibm-spectrum-scale-csi-config"
	EnvVarPrefix    = "VAR_DRIVER_"

	// Optional ConfigMap keys
	DaemonSetUpgradeMaxUnavailableKey = "DRIVER_UPGRADE_MAXUNAVAILABLE"
	DriverCPULimits                   = "DRIVER_CPU_LIMITS"
	DriverMemoryLimits                = "DRIVER_MEMORY_LIMITS"
	SidecarCPULimits                  = "SIDECAR_CPU_LIMITS"
	SidecarMemoryLimits               = "SIDECAR_MEMORY_LIMITS"
	EnvLogLevelKey                    = "LOGLEVEL"
	EnvPersistentLogKey               = "PERSISTENT_LOG"
	EnvNodePublishMethodKey           = "NODEPUBLISH_METHOD"
	EnvVolumeStatsCapabilityKey       = "VOLUME_STATS_CAPABILITY"
	EnvDiscoverCGFilesetKey           = "DISCOVER_CG_FILESET"
	HostNetworkKey                    = "HOST_NETWORK"

	// Optional ConfigMap keys with prefix
	EnvLogLevelKeyPrefixed              = EnvVarPrefix + EnvLogLevelKey
	EnvPersistentLogKeyPrefixed         = EnvVarPrefix + EnvPersistentLogKey
	EnvNodePublishMethodKeyPrefixed     = EnvVarPrefix + EnvNodePublishMethodKey
	EnvVolumeStatsCapabilityKeyPrefixed = EnvVarPrefix + EnvVolumeStatsCapabilityKey
	EnvDiscoverCGFilesetKeyPrefixed     = EnvVarPrefix + EnvDiscoverCGFilesetKey

	// Optional ConfigMap default values
	DriverCPULimitsDefaultValue          = "600m"
	DriverMemoryLimitsDefaultValue       = "600Mi"
	SidecarCPULimitsDefaultValue         = "300m"
	SidecarMemoryLimitsDefaultValue      = "800Mi"
	EnvLogLevelDefaultValue              = "INFO"
	EnvPersistentLogDefaultValue         = "DISABLED"
	EnvNodePublishMethodDefaultValue     = "BINDMOUNT"
	EnvVolumeStatsCapabilityDefaultValue = "ENABLED"
	EnvHostNetworkDefaultValue           = "ENABLED"

	// Driver and Sidecar Containers Resources limits
	PodsCPULimitsLowerValue    = "20m"
	PodsMemoryLimitsLowerValue = "20Mi"
)

var CSIOptionalConfigMapKeys = []string{
	EnvLogLevelKeyPrefixed,
	EnvPersistentLogKeyPrefixed,
	EnvNodePublishMethodKeyPrefixed,
	EnvVolumeStatsCapabilityKeyPrefixed,
	DaemonSetUpgradeMaxUnavailableKey,
	EnvDiscoverCGFilesetKeyPrefixed,
	HostNetworkKey,
	DriverCPULimits,
	DriverMemoryLimits,
	SidecarCPULimits,
	SidecarMemoryLimits}
var EnvLogLevelValues = []string{"TRACE", "DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}
var EnvNodePublishMethodValues = []string{"SYMLINK", "BINDMOUNT"}
var EnvPersistentLogValues = []string{"ENABLED", "DISABLED"}
var EnvVolumeStatsCapabilityValues = []string{"ENABLED", "DISABLED"}
var EnvDiscoverCGFilesetValues = []string{"ENABLED", "DISABLED"}
var EnvHostNetworkValues = []string{"ENABLED", "DISABLED"}

const (
	StatusConditionReady   = "Ready"
	StatusConditionSuccess = "Success"
	StatusConditionEnabled = "Enabled"

	SecretUsername     = "username" // #nosec G101 false positive
	SecretPassword     = "password" // #nosec G101 false positive
	SecretVolumeSuffix = "-secret"  // #nosec G101 false positive
	CacertVolumeSuffix = "-cacert"
	Primary            = "primary"
	HTTPClientTimeout  = 60

	DefaultPrimaryFileset = "spectrum-scale-csi-volume-store"
	SymlinkDir            = ".volumes"
	DefaultUID            = "0"
	DefaultGID            = "0"
	RunAsUser             = 10001
	RunAsGroup            = 10001

	ErrorForbidden    = "403: Forbidden"
	ErrorUnauthorized = "401: Unauthorized"

	ENVClusterConfigurationType = "ClusterConfigurationType"
	ENVClusterTypeOpenshift     = "OpenShiftPlatform"
	ENVClusterTypeKubernetes    = "KubernetesPlatform"
	ENVClusterCNSAPresenceCheck = "CNSADeployment"
	CNSAOperatorNamespace       = "ibm-spectrum-scale-operator"
	CNSAOperatorDeploymentName  = "ibm-spectrum-scale-controller-manager"
)
