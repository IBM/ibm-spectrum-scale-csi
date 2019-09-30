package controller

import (
	"github.ibm.com/jdunham/ibm-spectrum-scale-csi-operator/pkg/controller/csiscaleoperator"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, csiscaleoperator.Add)
}
