package config

// Kubernetes built-in well-known constants
const (
	// LabelAppName is the name of the component application.
	LabelAppName = "app.kubernetes.io/name"
	// LabelAppInstance identifies resources related to a unique Cluster deployment.
	LabelAppInstance = "app.kubernetes.io/instance"
	// LabelAppManagedBy is the controller/user who created the resource.
	LabelAppManagedBy = "app.kubernetes.io/managed-by"
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
	APIGroup    = "csi.ibm.com"
	APIVersion  = "v1"
	ID          = "ibm-spectrum-scale-csi-operator"
	DriverName  = "spectrumscale.csi.ibm.com"
	Kind        = "CSIScaleOperator"

	Masterlabel = "node-role.kubernetes.io/master"
	ProductName = "IBM Spectrum Scale CSI Operator"

	NodeAgentRepository = "ibmcom/ibm-node-agent"

	ENVEndpoint    = "ENDPOINT"
	ENVNodeName    = "NODE_NAME"
	ENVKubeVersion = "KUBE_VERSION"
	ENVIsOpenShift = "IS_OpenShift"

	DriverVersion = "2.5.0"

	//  Default images for containers
	CSIDriverPluginImage        = "quay.io/ibm-spectrum-scale/ibm-spectrum-scale-csi-driver:v2.4.0"
	CSINodeDriverRegistrarImage = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-node-driver-registrar:v2.3.0"
	LivenessProbeImage          = "us.gcr.io/k8s-artifacts-prod/sig-storage/livenessprobe:v2.4.0"
	CSIAttacherImage            = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-attacher:v3.3.0"
	CSIProvisionerImage         = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-provisioner:v3.0.0"
	CSISnapshotterImage         = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-snapshotter:v4.2.1"
	CSIResizerImage             = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-resizer:v1.3.0"

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
	SecretsMountPath          = "/var/lib/ibm/"
	ConfigMapPath             = "/var/lib/ibm/config"
	CAcertMountPath           = "/var/lib/ibm/ssl/public/"
	CSIFinalizer              = "finalizer.csiscaleoperators.csi.ibm.com"
)

const (
	StatusConditionReady   = "Ready"
	StatusConditionSuccess = "Success"
	StatusConditionEnabled = "Enabled"
)
