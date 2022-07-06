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

	DriverVersion   = "2.7.0"
	OperatorVersion = "2.7.0"

	// Number of replica pods for CSI Sidecar deployment
	ReplicaCount = int32(2)
	// Tolerations seconds for the CSI Sidecar deployment
	TolerationsSeconds = int64(300)
	// ContainerPort for /healthz/leader-election endpoint
	LeaderLivenessPort = int32(8080)
	// 64-Bit machine architecture supported by Spectrum Scale CSI.
	AMD64 = "amd64"
	// Power PC machine architecture supported by Spectrum Scale CSI.
	PPC = "ppc64le"
	// IBM zSystems machine architecture supported by Spectrum Scale CSI.
	IBMSystem390 = "s390x"

	//  Default images for containers
	CSIDriverPluginImage = "quay.io/ibm-spectrum-scale/ibm-spectrum-scale-csi-driver:v2.7.0"
	//  us.gcr.io/k8s-artifacts-prod/sig-storage/csi-node-driver-registrar:v2.5.0
	CSINodeDriverRegistrarImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-node-driver-registrar@sha256:4fd21f36075b44d1a423dfb262ad79202ce54e95f5cbc4622a6c1c38ab287ad6"
	//  us.gcr.io/k8s-artifacts-prod/sig-storage/livenessprobe:v2.7.0
	LivenessProbeImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/livenessprobe@sha256:933940f13b3ea0abc62e656c1aa5c5b47c04b15d71250413a6b821bd0c58b94e"
	//  us.gcr.io/k8s-artifacts-prod/sig-storage/csi-attacher:v3.4.0
	CSIAttacherImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-attacher@sha256:8b9c313c05f54fb04f8d430896f5f5904b6cb157df261501b29adc04d2b2dc7b"
	//  us.gcr.io/k8s-artifacts-prod/sig-storage/csi-provisioner:v3.2.1
	CSIProvisionerImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-provisioner@sha256:4ad5fcdbe7e9147b541a863d74e4d1d519bf435ecda4c7bde5abe237a43f7029"
	//  us.gcr.io/k8s-artifacts-prod/sig-storage/csi-snapshotter:v6.0.1
	CSISnapshotterImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-snapshotter@sha256:ad16874e2140256a809cada2b4ac3d931d5b73b0bee23ed0f8d60bdd778cfec2"
	//  us.gcr.io/k8s-artifacts-prod/sig-storage/csi-resizer:v1.4.0
	CSIResizerImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-resizer@sha256:9ebbf9f023e7b41ccee3d52afe39a89e3ddacdbb69269d583abfc25847cfd9e4"

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

	//Default imagePullSecrets
	ImagePullSecretRegistryKey    = "ibm-spectrum-scale-csi-registrykey" // #nosec G101 false positive
	ImagePullSecretEntitlementKey = "ibm-entitlement-key"                // #nosec G101 false positive

)

const (
	StatusConditionReady   = "Ready"
	StatusConditionSuccess = "Success"
	StatusConditionEnabled = "Enabled"
)
