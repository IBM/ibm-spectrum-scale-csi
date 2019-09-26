package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CSIPrimarySpec struct {
  primaryFs   string `json:"primaryFS"`
  primaryFset string `json:"primaryFset"`
}

type CSIRestApiSpec struct {
  GuiHost string `json:"guiHost"`
  GuiPort string `json:"guiPort"`
}


type CSIClusterSpec struct {
    Id             string  `json:"id"`
    SecureSslMode  bool    `json:"secureSslMode"`

    //TODO make a secret ref?
    Secrets        string  `json:"secrets"`
    Cacert         string  `json:"cacert"`

    Primary   CSIPrimarySpec `json:"primary"`
    RestApi []CSIRestApiSpec `json:"restApi"`

}

// CSIScaleOperatorSpec defines the desired state of CSIScaleOperator
// +k8s:openapi-gen=true
type CSIScaleOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

  Attacher        string `json:"csi_attacher"`
  Provisioner     string `json:"csi_provisioner"`
  DriverRegistrar string `json:"csi_driver_registrar"`
  SpectrumScale   string `json:"csi_spectrum_scale"`
  ScaleHostpath   string `json:"csi_scale_hostpath"`
  Clusters []CSIClusterSpec `json:"csi_clusters"`

}

// CSIScaleOperatorStatus defines the observed state of CSIScaleOperator
// +k8s:openapi-gen=true
type CSIScaleOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
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
