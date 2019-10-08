package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Defines the primary filesystem.
// +k8s:openapi-gen=true
type CSIPrimarySpec struct {
	// The name of the primary filesystem.
	PrimaryFS string `json:"primaryFS,omitempty"`

	// The name of the primary fileset, created in primaryFS.
	PrimaryFset string `json:"primaryFset,omitempty"`
}

// Defines the desired REST API access info.
// +k8s:openapi-gen=true
type CSIRestApiSpec struct {
	// The hostname of the REST server.
	GuiHost string `json:"guiHost,omitempty"`

	// The port number running the REST server.
	GuiPort string `json:"guiPort,omitempty"`
}

//  CSIClusterSpec  defines the desired state of CSIi Scale Cluster
// +k8s:openapi-gen=true
type CSIClusterSpec struct {
	// The cluster id of the gpfs cluster specified (mandatory).
	Id string `json:"id"`

	// Require a secure SSL connection to connect to GPFS.
	SecureSslMode bool `json:"secureSslMode,omitempty"`

	// A string specifying a secret resource name.
	Secrets string `json:"secrets,omitempty"`

	// A string specifying a cacert resource name.
	Cacert string `json:"cacert,omitempty"`

	// The primary file system for the GPFS cluster.
	Primary CSIPrimarySpec `json:"primary,omitempty"`

	// A collection of targets for REST calls.
	RestApi []CSIRestApiSpec `json:"restApi,omitempty"`
}

// CSIScaleOperatorSpec defines the desired state of CSIScaleOperator
// +k8s:openapi-gen=true
type CSIScaleOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Attacher image for csi (actually attaches to the storage).
	Attacher string `json:"attacher,omitempty"`

	// Provisioner image for csi (actually issues provision requests).
	Provisioner string `json:"provisioner,omitempty"`

	// Sidecar container image for the csi spectrum scale plugin pods.
	DriverRegistrar string `json:"driverRegistrar,omitempty"`

	// Image name for the csi spectrum scale plugin container.
	SpectrumScale string `json:"spectrumScale,omitempty"`

	// The path to the gpfs file system mounted on the host machine.
	ScaleHostpath string `json:"scaleHostpath,omitempty"`

	// A collection of gpfs cluster properties for the csi driver to mount.
	Clusters []CSIClusterSpec `json:"clusters,omitempty"`

	// Trigger used by the operator for secret changes.
	SecretCounter int `json:"secretCounter,omitempty"`
}

// CSIScaleOperatorCondition defines the observed Condition of CSIScaleOperator
// +k8s:openapi-gen=true
type CSIScaleOperatorCondition struct {
	IsRunning bool `json:"isRunning,omitempty"`
}

// CSIScaleOperatorStatus defines the observed state of CSIScaleOperator
// +k8s:openapi-gen=true
type CSIScaleOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Conditions []CSIScaleOperatorCondition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CSIScaleOperator is the Schema for the csiscaleoperators API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type CSIScaleOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIScaleOperatorSpec   `json:"spec,omitempty"`
	Status CSIScaleOperatorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CSIScaleOperatorList contains a list of CSIScaleOperator
type CSIScaleOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIScaleOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSIScaleOperator{}, &CSIScaleOperatorList{})
}
