/*
Copyright 2022, 2024 IBM Corp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package csiscaleoperator

import (
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/log"

	csiv1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
)

var csiLog = log.Log.WithName("csiscaleoperator")

// CSIScaleOperator is the wrapper for csiv1.CSIScaleOperator type
type CSIScaleOperator struct {
	*csiv1.CSIScaleOperator
	//	ServerVersion string
}

// New returns a wrapper for csiv1.CSIScaleOperator
func New(c *csiv1.CSIScaleOperator) *CSIScaleOperator {
	return &CSIScaleOperator{
		CSIScaleOperator: c,
	}
}

// Unwrap returns the csiv1.CSIScaleOperator object
func (c *CSIScaleOperator) Unwrap() *csiv1.CSIScaleOperator {
	return c.CSIScaleOperator
}

// GetLabels returns all the labels to be set on all resources
// func (c *CSIScaleOperator) GetLabels() labels.Set {
func (c *CSIScaleOperator) GetLabels() map[string]string {
	labels := labels.Set{
		config.LabelAppName:      config.ResourceAppName,
		config.LabelAppInstance:  config.ResourceInstance,
		config.LabelAppManagedBy: config.ResourceManagedBy,
		config.LabelProduct:      config.Product,
		config.LabelRelease:      config.Release,
	}

	if c.Labels != nil {
		for k, v := range c.Labels {
			if !labels.Has(k) {
				labels[k] = v
			}
		}
	}

	return labels
}

// GetAnnotations returns all the annotations to be set on all resources
// func (c *CSIScaleOperator) GetAnnotations(daemonSetRestartedKey string, daemonSetRestartedValue string) labels.Set {
func (c *CSIScaleOperator) GetAnnotations(daemonSetRestartedKey string, daemonSetRestartedValue string) map[string]string {
	//func (c *CSIScaleOperator) GetAnnotations() map[string]string {
	labels := labels.Set{
		"productID":      config.ID,
		"productName":    config.ProductName,
		"productVersion": config.DriverVersion,
	}

	if c.Annotations != nil {
		for k, v := range c.Annotations {
			if !labels.Has(k) {
				labels[k] = v
			}
		}
	}

	// removing the annotations that are not required in the daemonset/statefulset and their pods.
	_, ok := c.Annotations["kubectl.kubernetes.io/last-applied-configuration"]
	if ok {
		delete(c.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	}

	if !labels.Has(daemonSetRestartedKey) && daemonSetRestartedKey != "" {
		labels[daemonSetRestartedKey] = daemonSetRestartedValue
	}

	return labels
}

// GetSelectorLabels returns labels used in label selectors
func (c *CSIScaleOperator) GetSelectorLabels(appName string) labels.Set {
	return labels.Set{
		// "app.kubernetes.io/component": component,
		config.LabelProduct: config.Product,
		config.LabelApp:     appName,
	}
}

func (c *CSIScaleOperator) GetCSIControllerSelectorLabels(appName string) labels.Set {
	return c.GetSelectorLabels(appName)
}

func (c *CSIScaleOperator) GetCSINodeSelectorLabels(appName string) labels.Set {
	return c.GetSelectorLabels(appName)
}

func (c *CSIScaleOperator) GetCSIControllerPodLabels(appName string) labels.Set {
	return labels.Merge(c.GetLabels(), c.GetCSIControllerSelectorLabels(appName))
}

func (c *CSIScaleOperator) GetCSINodePodLabels(appName string) labels.Set {
	return labels.Merge(c.GetLabels(), c.GetCSINodeSelectorLabels(appName))
}

func (c *CSIScaleOperator) GetDefaultImage(name string) string {

	logger := csiLog.WithName("GetDefaultImage")
	logger.Info("Getting default image for ", "container:", name)

	image := ""
	switch name {
	case config.CSINodeDriverRegistrar:
		image = config.CSINodeDriverRegistrarImage
	case config.CSINodeDriverPlugin:
		image = config.CSIDriverPluginImage
	case config.LivenessProbe:
		image = config.LivenessProbeImage
	case config.CSIProvisioner:
		image = config.CSIProvisionerImage
	case config.CSIAttacher:
		image = config.CSIAttacherImage
	case config.CSISnapshotter:
		image = config.CSISnapshotterImage
	case config.CSIResizer:
		image = config.CSIResizerImage
	}
	logger.Info("Got default image for ", "container:", name, ", image:", image)
	return image
}

// GetKubeletPodsDir returns the kubelet pods directory
// TODO: Unexport these functions to fix lint warnings
func (c *CSIScaleOperator) GetKubeletPodsDir() string {
	logger := csiLog.WithName("GetKubeletPodsDir")
	kubeletPodsDir := c.GetKubeletRootDirPath() + "/pods"
	logger.Info("GetKubeletPodsDir", "kubeletPodsDir: ", kubeletPodsDir)
	return kubeletPodsDir
}

func (c *CSIScaleOperator) GetKubeletRootDirPath() string {
	logger := csiLog.WithName("GetKubeletRootDirPath")

	if c.Spec.KubeletRootDirPath == "" {
		logger.Info("In GetKubeletRootDirPath", "using default kubeletRootDirPath: ", config.CSIKubeletRootDirPath)
		return config.CSIKubeletRootDirPath
	}
	logger.Info("In GetKubeletRootDirPath", "using kubeletRootDirPath: ", c.Spec.KubeletRootDirPath)
	return c.Spec.KubeletRootDirPath
}

func (c *CSIScaleOperator) GetSocketPath() string {
	logger := csiLog.WithName("GetSocketPath")

	socketPath := c.GetKubeletRootDirPath() + config.SocketPath
	logger.Info("In GetSocketPath", "socketPath", socketPath)
	return socketPath
}

func (c *CSIScaleOperator) GetSocketDir() string {
	logger := csiLog.WithName("GetSocketDir")

	socketDir := c.GetKubeletRootDirPath() + config.SocketDir
	logger.Info("In GetSocketDir", "socketDir", socketDir)
	return socketDir
}

func (c *CSIScaleOperator) GetCSIEndpoint() string {
	logger := csiLog.WithName("GetCSIEndpoint")

	CSIEndpoint := "unix://" + c.GetKubeletRootDirPath() + config.SocketPath
	logger.Info("In GetCSIEndpoint", "CSIEndpoint", CSIEndpoint)
	return CSIEndpoint
}
