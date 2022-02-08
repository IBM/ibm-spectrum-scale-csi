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

import "fmt"

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

func (rn ResourceName) String() string {
	return string(rn)
}

const (
	CSIController             ResourceName = "csi-controller"
	CSIControllerAttacher     ResourceName = "csi-controller-attacher"
	CSIControllerProvisioner  ResourceName = "csi-controller-provisioner"
	CSIControllerSnapshotter  ResourceName = "csi-controller-snapshotter"
	CSIControllerResizer      ResourceName = "csi-controller-resizer"
	CSINode                   ResourceName = "csi"
	NodeAgent                 ResourceName = "ibm-node-agent"
	CSIAttacherServiceAccount ResourceName = "csi-attacher-sa"
	CSIControllerServiceAccount  ResourceName = "csi-controller-sa"
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
	Sidecar		ResourceName = "controller"
)

// GetNameForResource returns the name of a resource for a CSI driver
func GetNameForResource(name ResourceName, driverName string) string {
	switch name {
	case CSIController:
		return fmt.Sprintf("%s-controller", driverName)
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
		return fmt.Sprintf("%s-controller", driverName)
	case CSINodeServiceAccount:
		return fmt.Sprintf("%s-node", driverName)
	case CSIProvisionerServiceAccount:
		return fmt.Sprintf("%s-provisioner", driverName)
	case CSISnapshotterServiceAccount:
		return fmt.Sprintf("%s-snapshotter", driverName)
	case CSIResizerServiceAccount:
		return fmt.Sprintf("%s-resizer", driverName)
	//case CSISidecarServiceAccount:
	//	return fmt.Sprintf("%s-controller", driverName)
	default:
		return fmt.Sprintf("%s-%s", driverName, name)
	}
}
