/**
 * Copyright 2019 IBM Corp.
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

package scale

import (
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors/rest/v2"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

type Driver struct {
	IdentityService
	NodeService
	ControllerService

	gRPC *grpc.Server
}

/*NewDriver creates new CSI plugin driver from ConfigMap
 */
func NewDriver(
	driverName string,
	vendorVersion string,
	nodeID string,
	config *settings.ConfigMap,
) *Driver {
	klog.Infof(`Driver: %v Version: %v`, driverName, vendorVersion)

	fab := rest.NewSpectrumV2(config)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(LogGRPC),
	}
	d := &Driver{
		IdentityService:   newIdentityService(driverName, vendorVersion),
		NodeService:       newNodeService(nodeID, fab),
		ControllerService: newControllerService(config, fab),
		gRPC:              grpc.NewServer(opts...),
	}

	d.AddVolumeCapabilityAccessModes(
		[]csi.VolumeCapability_AccessMode_Mode{
			csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	)

	d.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		},
	)

	d.AddNodeServiceCapabilities(
		[]csi.NodeServiceCapability_RPC_Type{},
	)

	csi.RegisterIdentityServer(d.gRPC, d)
	csi.RegisterControllerServer(d.gRPC, d)
	csi.RegisterNodeServer(d.gRPC, d)

	return d
}

/*Run the CSI plugin at the endpoint
Note: blocks until gRPC Server stops
*/
func (d *Driver) Run(endpoint string) error {
	scheme, addr, err := ParseEndpoint(endpoint)
	if err != nil {
		return err
	}

	listener, err := net.Listen(scheme, addr)
	if err != nil {
		return err
	}

	klog.Infof(`Listening for connections on address: %#v`, listener.Addr())
	return d.gRPC.Serve(listener)
}

/*Stop the CSI plugin server
 */
func (d *Driver) Stop() {
	klog.Infof(`Stopping server...`)
	d.gRPC.GracefulStop()
}

/*ForceStop the CSI plugin server
 */
func (d *Driver) ForceStop() {
	klog.Infof(`Killing server...`)
	d.gRPC.Stop()
}

//TODO
func (driver *Driver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) {
	glog.V(3).Infof("gpfs AddVolumeCapabilityAccessModes")
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.V(3).Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	//driver.vcap = vca
}

func (driver *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	glog.V(3).Infof("gpfs AddControllerServiceCapabilities")
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		glog.V(3).Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	driver.ControllerService.cscap = csc
}

func (driver *Driver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	glog.V(3).Infof("gpfs AddNodeServiceCapabilities")
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		glog.V(3).Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	driver.NodeService.nscap = nsc
}

func (driver *Driver) PluginInitialize(config *settings.ConfigMap) error {
	klog.V(3).Infof(
		"gpfs PluginInitialize. driverName: %s, vendorVersion: %v, nodeID: %s",
		driver.IdentityService.driverName,
		driver.IdentityService.vendorVersion,
		driver.NodeService.id,
	)
	if driver.IdentityService.driverName == "" {
		return fmt.Errorf("Driver name missing")
	}

	for _, cluster := range config.Clusters {
		remote := driver.ControllerService.fab.NewConnector(&cluster)

		// validate cluster ID
		clusterId, err := remote.GetClusterId()
		if err != nil {
			glog.Errorf("Error getting cluster ID: %v", err)
			return err
		}
		if cluster.ID != clusterId {
			glog.Errorf("Cluster ID %s from scale config doesnt match the ID from cluster %s.", cluster.ID, clusterId)
			return fmt.Errorf("Cluster ID doesnt match the cluster")
		}
	}

	primary := driver.ControllerService.fab.NewConnector(config.Primary)
	remote := primary

	// check if primary filesystem exists and mounted on atleast one node
	_fsMount, err := primary.GetFilesystemMountDetails(config.Primary.GetPrimaryFs())
	if err != nil {
		glog.Errorf("Error in getting filesystem details for %s", config.Primary.GetPrimaryFs())
		return err
	}
	if _fsMount.NodesMounted == nil || len(_fsMount.NodesMounted) == 0 {
		return fmt.Errorf("Primary filesystem not mounted on any node")
	}

	config.Primary.PrimaryFSMount = _fsMount.MountPoint
	fsMount := _fsMount.MountPoint

	if config.Primary.RemoteCluster != "" {
		if fs := config.Primary.GetRemoteFs(); fs != "" {
			remote = driver.ControllerService.fab.NewConnector(
				&settings.Cluster{ID: config.Primary.RemoteCluster},
			)

			// check if primary filesystem exists on remote cluster and mounted on atleast one node
			_fsMount, err = remote.GetFilesystemMountDetails(fs)
			if err != nil {
				glog.Errorf("Error in getting filesystem details for %s from cluster %s", fs, config.Primary.RemoteCluster)
				return err
			}
			glog.Infof("remote fsMount = %v", _fsMount)
			if _fsMount.NodesMounted == nil || len(_fsMount.NodesMounted) == 0 {
				return fmt.Errorf("Primary filesystem not mounted on any node on cluster %s", config.Primary.RemoteCluster)
			}
			fsMount = _fsMount.MountPoint
		}
	}

	fsetlinkpath, err := driver.CreatePrimaryFileset(
		remote,
		config.Primary.GetPrimaryFs(),
		config.Primary.PrimaryFSMount,
		config.Primary.PrimaryFset,
		config.Primary.GetInodeLimit(),
	)
	if err != nil {
		glog.Errorf("Error in creating primary fileset")
		return err
	}

	if fsMount != config.Primary.PrimaryFSMount {
		fsetlinkpath = strings.Replace(fsetlinkpath, fsMount, config.Primary.PrimaryFSMount, 1)
	}

	// Validate hostpath from daemonset is valid
	err = driver.ValidateHostpath(config.Primary.PrimaryFSMount, fsetlinkpath)
	if err != nil {
		glog.Errorf("Hostpath validation failed")
		return err
	}

	// Create directory where volume symlinks will reside
	symlinkPath, relativePath, err := driver.CreateSymlinkPath(primary, config.Primary.GetPrimaryFs(), config.Primary.PrimaryFSMount, fsetlinkpath)
	if err != nil {
		glog.Errorf("Error in creating volumes directory")
		return err
	}
	config.Primary.SymlinkAbsolutePath = symlinkPath
	config.Primary.SymlinkRelativePath = relativePath
	config.Primary.PrimaryFsetLink = fsetlinkpath

	glog.Infof("IBM Spectrum Scale: Plugin initialized")
	return nil
}

func (driver *Driver) CreatePrimaryFileset(primary connectors.Connector, primaryFS string, fsmount string, filesetName string, inodeLimit string) (string, error) {
	glog.V(4).Infof("gpfs CreatePrimaryFileset. primaryFS: %s, mountpoint: %s, filesetName: %s", primaryFS, fsmount, filesetName)

	// create primary fileset if not already created
	fsetResponse, err := primary.ListFileset(primaryFS, filesetName)
	linkpath := fsetResponse.Config.Path
	newlinkpath := path.Join(fsmount, filesetName)

	if err != nil {
		glog.Infof("Primary fileset %s not found. Creating it.", filesetName)
		opts := make(map[string]interface{})
		if inodeLimit != "" {
			opts[settings.InodeLimit] = inodeLimit
		}
		err = primary.CreateFileset(primaryFS, filesetName, opts)
		if err != nil {
			glog.Errorf("Unable to create primary fileset %s", filesetName)
			return "", err
		}
		linkpath = newlinkpath
	} else if linkpath == "" || linkpath == "--" {
		glog.Infof("Primary fileset %s not linked. Linking it.", filesetName)
		err = primary.LinkFileset(primaryFS, filesetName, newlinkpath)
		if err != nil {
			glog.Errorf("Unable to link primary fileset %s", filesetName)
			return "", err
		} else {
			glog.Infof("Linked primary fileset %s. Linkpath: %s", newlinkpath, filesetName)
		}
		linkpath = newlinkpath
	} else {
		glog.Infof("Primary fileset %s exists and linked at %s", filesetName, linkpath)
	}

	return linkpath, nil
}

func (driver *Driver) CreateSymlinkPath(conn connectors.Connector, fs string, fsmount string, fsetlinkpath string) (string, string, error) {
	glog.V(4).Infof("gpfs CreateSymlinkPath. filesystem: %s, mountpoint: %s, filesetlinkpath: %s", fs, fsmount, fsetlinkpath)

	dirpath := strings.Replace(fsetlinkpath, fsmount, "", 1)
	dirpath = strings.Trim(dirpath, "!/")
	fsetlinkpath = strings.TrimSuffix(fsetlinkpath, "/")

	dirpath = fmt.Sprintf("%s/.volumes", dirpath)
	symlinkpath := fmt.Sprintf("%s/.volumes", fsetlinkpath)

	err := conn.MakeDirectory(fs, dirpath, "0", "0")
	if err != nil {
		glog.Errorf("Make directory failed on filesystem %s, path = %s", fs, dirpath)
		return symlinkpath, dirpath, err
	}

	return symlinkpath, dirpath, nil
}

func (driver *Driver) ValidateHostpath(mountpath string, linkpath string) error {
	glog.V(4).Infof("gpfs ValidateHostpath. mountpath: %s, linkpath: %s", mountpath, linkpath)

	hostpath := utils.GetEnv("SCALE_HOSTPATH", "")
	if hostpath == "" {
		return fmt.Errorf("SCALE_HOSTPATH not defined in daemonset")
	}

	if !strings.HasSuffix(hostpath, "/") {
		hostpathslice := []string{hostpath}
		hostpathslice = append(hostpathslice, "/")
		hostpath = strings.Join(hostpathslice, "")
	}

	if !strings.HasSuffix(linkpath, "/") {
		linkpathslice := []string{linkpath}
		linkpathslice = append(linkpathslice, "/")
		linkpath = strings.Join(linkpathslice, "")
	}

	if !strings.HasSuffix(mountpath, "/") {
		mountpathslice := []string{mountpath}
		mountpathslice = append(mountpathslice, "/")
		mountpath = strings.Join(mountpathslice, "")
	}

	if !strings.HasPrefix(hostpath, linkpath) &&
		!strings.HasPrefix(hostpath, mountpath) &&
		!strings.HasPrefix(linkpath, hostpath) &&
		!strings.HasPrefix(mountpath, hostpath) {
		return fmt.Errorf("Invalid SCALE_HOSTPATH")
	}

	return nil
}
