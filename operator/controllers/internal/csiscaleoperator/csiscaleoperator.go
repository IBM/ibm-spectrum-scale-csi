package csiscaleoperator

import (
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/log"

	csiv1 "github.com/IBM/ibm-spectrum-scale-csi/api/v1"
	"github.com/IBM/ibm-spectrum-scale-csi/controllers/config"
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
//func (c *CSIScaleOperator) GetLabels() labels.Set {
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
//func (c *CSIScaleOperator) GetAnnotations(daemonSetRestartedKey string, daemonSetRestartedValue string) labels.Set {
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

func (c *CSIScaleOperator) GetKubeletRootDirPath() string {
	logger := csiLog.WithName("GetKubeletRootDirPath")

	if c.Spec.KubeletRootDirPath == "" {
		logger.Info("in GetKubeletRootDirPath", "using default kubeletRootDirPath: ", config.CSIKubeletRootDirPath)
		return config.CSIKubeletRootDirPath
	}
	logger.Info("in GetKubeletRootDirPath", "using kubeletRootDirPath: ", c.Spec.KubeletRootDirPath)
	return c.Spec.KubeletRootDirPath
}

func (c *CSIScaleOperator) GetSocketPath() string {
	logger := csiLog.WithName("GetSocketPath")

	socketPath := c.GetKubeletRootDirPath() + config.SocketPath
	logger.Info("in GetSocketPath", "socketPath", socketPath)
	return socketPath
}

func (c *CSIScaleOperator) GetSocketDir() string {
	logger := csiLog.WithName("GetSocketDir")

	socketDir := c.GetKubeletRootDirPath() + config.SocketDir
	logger.Info("in GetSocketDir", "socketDir", socketDir)
	return socketDir
}

func (c *CSIScaleOperator) GetCSIEndpoint() string {
	logger := csiLog.WithName("GetCSIEndpoint")

	CSIEndpoint := "unix://" + c.GetKubeletRootDirPath() + config.SocketPath
	logger.Info("in GetCSIEndpoint", "CSIEndpoint", CSIEndpoint)
	return CSIEndpoint
}
