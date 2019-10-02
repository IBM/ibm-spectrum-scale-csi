package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"github.com/operator-framework/operator-sdk/pkg/ansible/proxy/controllermap"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, cMap controllermap.ControllerMap) error {
	return nil
}
