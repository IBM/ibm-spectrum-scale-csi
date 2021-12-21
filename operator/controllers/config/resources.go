package config

import "fmt"

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

func (rn ResourceName) String() string {
	return string(rn)
}

const (
	CSIController                ResourceName = "csi-controller"
	CSIControllerAttacher        ResourceName = "csi-controller-attacher"
	CSIControllerProvisioner     ResourceName = "csi-controller-provisioner"
	CSIControllerSnapshotter     ResourceName = "csi-controller-snapshotter"
	CSIControllerResizer         ResourceName = "csi-controller-resizer"
	CSINode                      ResourceName = "csi"
	NodeAgent                    ResourceName = "ibm-node-agent"
	CSIAttacherServiceAccount    ResourceName = "csi-attacher-sa"
	CSIControllerServiceAccount  ResourceName = "csi-operator"
	CSINodeServiceAccount        ResourceName = "csi-node-sa"
	CSIProvisionerServiceAccount ResourceName = "csi-provisioner-sa"
	CSISnapshotterServiceAccount ResourceName = "csi-snapshotter-sa"
	CSIResizerServiceAccount     ResourceName = "csi-resizer-sa"

	// Suffixes for ClusterRole and ClusterRoleBinding names.
	Provisioner ResourceName = "provisioner"
	NodePlugin  ResourceName = "node"
	Attacher    ResourceName = "attacher"
	Snapshotter ResourceName = "snapshotter"
	Resizer     ResourceName = "resizer"
)

// GetNameForResource returns the name of a resource for a CSI driver
func GetNameForResource(name ResourceName, driverName string) string {
	switch name {
	case CSIController:
		return fmt.Sprintf("%s-operator", driverName)
	case CSIControllerAttacher:
		return fmt.Sprintf("%s-attacher", driverName)
	case CSIControllerProvisioner:
		return fmt.Sprintf("%s-provisioner", driverName)
	case CSIControllerSnapshotter:
		return fmt.Sprintf("%s-snapshotter", driverName)
	case CSIControllerResizer:
		return fmt.Sprintf("%s-resizer", driverName)
	case CSINode:
		return driverName
	case CSIAttacherServiceAccount:
		return fmt.Sprintf("%s-attacher", driverName)
	case CSIControllerServiceAccount:
		return fmt.Sprintf("%s-operator", driverName)
	case CSINodeServiceAccount:
		return fmt.Sprintf("%s-node", driverName)
	case CSIProvisionerServiceAccount:
		return fmt.Sprintf("%s-provisioner", driverName)
	case CSISnapshotterServiceAccount:
		return fmt.Sprintf("%s-snapshotter", driverName)
	case CSIResizerServiceAccount:
		return fmt.Sprintf("%s-resizer", driverName)
	default:
		return fmt.Sprintf("%s-%s", driverName, name)
	}
}

