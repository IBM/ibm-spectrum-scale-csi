package apis

import (
	"github.com/IBM/ibm-spectrum-scale-csi/operator/pkg/apis/ibm/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1.SchemeBuilder.AddToScheme)
}
