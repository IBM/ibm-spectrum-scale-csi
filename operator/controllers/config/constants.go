/**
 * Copyright 2022, 2024 IBM Corp.
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
	DriverVersion   = "3.1.0"
	OperatorVersion = "3.1.0"
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
	// 64-Bit ARM architecture supported by IBM Storage Scale CSI.
	ARM64 = "arm64"

	//  Default images for containers

	CSIDriverPluginImage = "quay.io/ibm-spectrum-scale/ibm-spectrum-scale-csi-driver:v3.1.0"
	//  registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.15.0
	CSINodeDriverRegistrarImage = "registry.k8s.io/sig-storage/csi-node-driver-registrar@sha256:11f199f6bec47403b03cb49c79a41f445884b213b382582a60710b8c6fdc316a" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/livenessprobe:v2.17.0
	LivenessProbeImage = "registry.k8s.io/sig-storage/livenessprobe@sha256:9b75b9ade162136291d5e8f13a1dfc3dec71ee61419b1bfc112e0796ff8a6aa9" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-attacher:v4.10.0
	CSIAttacherImage = "registry.k8s.io/sig-storage/csi-attacher@sha256:be59d0556508d3dd419cc34c53062f170a28902ef36e832b947c7796458a083d" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-provisioner:v6.0.0
	CSIProvisionerImage = "registry.k8s.io/sig-storage/csi-provisioner@sha256:0c537015fe9cf9d53d79d0181e17a78ee784303cd46bf957016605488f212327" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-snapshotter:v8.4.0
	CSISnapshotterImage = "registry.k8s.io/sig-storage/csi-snapshotter@sha256:c7e0a3718832b6197ce8b29fefb3fed3d84f4fbcdf08f4606140dbec2566501d" // #nosec G101 false positive
	//  registry.k8s.io/sig-storage/csi-resizer:v2.0.0
	CSIResizerImage = "registry.k8s.io/sig-storage/csi-resizer@sha256:4a95d94e57ad82f6977cd8d4fdcfcfc0b83f02d990e4e7715b688c20970a906d" // #nosec G101 false positive

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
	ImagePullSecretEntitlementKey = "ibm-entitlement-key" // #nosec G101 false positive

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
	EnvVolNamePrefixKey               = "VOLUME_NAME_PREFIX"

	// Optional ConfigMap keys with prefix
	EnvLogLevelKeyPrefixed              = EnvVarPrefix + EnvLogLevelKey
	EnvPersistentLogKeyPrefixed         = EnvVarPrefix + EnvPersistentLogKey
	EnvNodePublishMethodKeyPrefixed     = EnvVarPrefix + EnvNodePublishMethodKey
	EnvVolumeStatsCapabilityKeyPrefixed = EnvVarPrefix + EnvVolumeStatsCapabilityKey
	EnvDiscoverCGFilesetKeyPrefixed     = EnvVarPrefix + EnvDiscoverCGFilesetKey
	EnvVolNamePrefixKeyPrefixed         = EnvVarPrefix + EnvVolNamePrefixKey

	// Optional ConfigMap default values if not provided in the cm
	DriverCPULimitsDefaultValue          = "600m"
	DriverMemoryLimitsDefaultValue       = "600Mi"
	SidecarCPULimitsDefaultValue         = "300m"
	SidecarMemoryLimitsDefaultValue      = "800Mi"
	EnvLogLevelDefaultValue              = "INFO"
	EnvPersistentLogDefaultValue         = "DISABLED"
	EnvNodePublishMethodDefaultValue     = "BINDMOUNT"
	EnvVolumeStatsCapabilityDefaultValue = "ENABLED"
	EnvHostNetworkDefaultValue           = "ENABLED"
	EnvVolNamePrefixDefaultValue         = "pvc"

	// Driver and Sidecar Containers Resources limits
	PodsCPULimitsLowerValue    = "20m"
	PodsMemoryLimitsLowerValue = "20Mi"

	// // For CNSA Dev setup, if the GUI host is set to localroute env
	// To run local in cnsa dev env
	IBMSpectrumScaleGUI string = "ibm-spectrum-scale-gui"
	ScaleProduct        string = "ibm-spectrum-scale"
	ScaleGUIRoute       string = IBMSpectrumScaleGUI
	ScaleGUIService     string = IBMSpectrumScaleGUI
)

// allowed keys of the optional cm variables
var CSIOptionalConfigMapKeys = []string{
	EnvLogLevelKeyPrefixed,
	EnvPersistentLogKeyPrefixed,
	EnvNodePublishMethodKeyPrefixed,
	EnvVolumeStatsCapabilityKeyPrefixed,
	DaemonSetUpgradeMaxUnavailableKey,
	EnvDiscoverCGFilesetKeyPrefixed,
	EnvVolNamePrefixKeyPrefixed,
	HostNetworkKey,
	DriverCPULimits,
	DriverMemoryLimits,
	SidecarCPULimits,
	SidecarMemoryLimits}

// allowed values of the optional cm variables
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
	CNSAScaleNamespace          = "ibm-spectrum-scale"
)
