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
	LabelArchitecture = "kubernetes.io/arch"
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

	ENVEndpoint    = "ENDPOINT"
	ENVNodeName    = "NODE_NAME"
	ENVKubeVersion = "KUBE_VERSION"
	ENVIsOpenShift = "IS_OpenShift"
	ENVCGPrefix    = "CSI_CG_PREFIX"
	ENVSymDirPath  = "SYMLINK_DIR_PATH"

	DriverVersion   = "2.10.0"
	OperatorVersion = "2.10.0"

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
	CSIDriverPluginImage = "quay.io/ibm-spectrum-scale/ibm-spectrum-scale-csi-driver:v2.10.0"
	//  registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.7.0
	CSINodeDriverRegistrarImage = "registry.k8s.io/sig-storage/csi-node-driver-registrar@sha256:4a4cae5118c4404e35d66059346b7fa0835d7e6319ff45ed73f4bba335cf5183"
	//  registry.k8s.io/sig-storage/livenessprobe:v2.10.0
	LivenessProbeImage = "registry.k8s.io/sig-storage/livenessprobe@sha256:4dc0b87ccd69f9865b89234d8555d3a614ab0a16ed94a3016ffd27f8106132ce"
	//  registry.k8s.io/sig-storage/csi-attacher:v4.2.0
	CSIAttacherImage = "registry.k8s.io/sig-storage/csi-attacher@sha256:34cf9b32736c6624fc9787fb149ea6e0fbeb45415707ac2f6440ac960f1116e6"
	//  registry.k8s.io/sig-storage/csi-provisioner:v3.4.1
	CSIProvisionerImage = "registry.k8s.io/sig-storage/csi-provisioner@sha256:893e37a388e7a7463d6c3523311b28cfbc5ae536dbef35430eed272cdc6850dc"
	//  registry.k8s.io/sig-storage/csi-snapshotter:v6.2.2
	CSISnapshotterImage = "registry.k8s.io/sig-storage/csi-snapshotter@sha256:becc53e25b96573f61f7469923a92fb3e9d3a3781732159954ce0d9da07233a2"
	//  registry.k8s.io/sig-storage/csi-resizer:v1.7.0
	CSIResizerImage = "registry.k8s.io/sig-storage/csi-resizer@sha256:3a7bdf5d105783d05d0962fa06ca53032b01694556e633f27366201c2881e01d"

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
	DefaultLogLevel           = "DEBUG"

	// Constants for Optional ConfigMap
	CSIEnvVarConfigMap                   = "ibm-spectrum-scale-csi-config"
	CSIEnvVarPrefix                      = "VAR_DRIVER_"
	CSIDaemonSetUpgradeMaxUnavailable    = "DRIVER_UPGRADE_MAXUNAVAILABLE"
	CSIDaemonSetUpgradeUpdateStrateyType = "RollingUpdate"

	//Default imagePullSecrets
	ImagePullSecretRegistryKey    = "ibm-spectrum-scale-csi-registrykey" // #nosec G101 false positive
	ImagePullSecretEntitlementKey = "ibm-entitlement-key"                // #nosec G101 false positive

)

const (
	StatusConditionReady   = "Ready"
	StatusConditionSuccess = "Success"
	StatusConditionEnabled = "Enabled"

	SecretUsername    = "username" // #nosec G101 false positive
	SecretPassword    = "password" // #nosec G101 false positive
	Primary           = "primary"
	HTTPClientTimeout = 60

	DefaultPrimaryFileset = "spectrum-scale-csi-volume-store"
	SymlinkDir            = ".volumes"
	DefaultUID            = "0"
	DefaultGID            = "0"
	RunAsUser             = 10001
	RunAsGroup            = 10001

	ErrorForbidden    = "403: Forbidden"
	ErrorUnauthorized = "401: Unauthorized"
)
