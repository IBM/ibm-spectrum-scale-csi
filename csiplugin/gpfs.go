/*
Copyright 2018 The Kubernetes Authors.

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

package gpfs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/golang/glog"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
        "k8s.io/kubernetes/pkg/util/mount"

        "github.ibm.com/FSaaS/scale-image/pkg/gpfs"
)

// PluginFolder defines the location of gpfsplugin
const (
	PluginFolder      = "/var/lib/kubelet/plugins/csi-gpfs"
)

type GPFSDriver struct {
	name          string
	vendorVersion string
	nodeID        string

	ids *GPFSIdentityServer
	ns  *GPFSNodeServer
	cs  *GPFSControllerServer

	vcap  []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
	nscap []*csi.NodeServiceCapability
}

var gpfsVolumes map[string]*gpfsVolume

// GPFS Operations
var ops = gpfs.NewGpfsOps()


// Init checks for the persistent volume file and loads all found volumes
// into a memory structure
func init() {
	gpfsVolumes = map[string]*gpfsVolume{}
	if _, err := os.Stat(path.Join(PluginFolder, "controller")); os.IsNotExist(err) {
		glog.Infof("gpfs: folder %s not found. Creating... \n", path.Join(PluginFolder, "controller"))
		if err := os.Mkdir(path.Join(PluginFolder, "controller"), 0755); err != nil {
			glog.Fatalf("Failed to create a controller's volumes folder with error: %v\n", err)
		}
	} else {
		// Since "controller" folder exists, it means the plugin has already been running, it means
		// there might be some volumes left, they must be re-inserted into gpfsVolumes map
		loadExVolumes()
	}
}


// loadExVolumes check for any *.json files in the  PluginFolder/controller folder
// and loads then into gpfsVolumes map
func loadExVolumes() {
	gpfsVol := gpfsVolume{}
	files, err := ioutil.ReadDir(path.Join(PluginFolder, "controller"))
	if err != nil {
		glog.Infof("gpfs: failed to read controller's volumes folder: %s error:%v", path.Join(PluginFolder, "controller"), err)
		return
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		fp, err := os.Open(path.Join(PluginFolder, "controller", f.Name()))
		if err != nil {
			glog.Infof("gpfs: open file: %s err %%v", f.Name(), err)
			continue
		}
		decoder := json.NewDecoder(fp)
		if err = decoder.Decode(&gpfsVol); err != nil {
			glog.Infof("gpfs: decode file: %s err: %v", f.Name(), err)
			fp.Close()
			continue
		}
		gpfsVolumes[gpfsVol.VolID] = &gpfsVol
	}
	glog.Infof("gpfs: Loaded %d volumes from %s", len(gpfsVolumes), path.Join(PluginFolder, "controller"))
}

func GetGpfsDriver() *GPFSDriver {
	return &GPFSDriver{}
}

func NewIdentityServer(d *GPFSDriver) *GPFSIdentityServer {
	return &GPFSIdentityServer{
		Driver: d,
	}
}

func NewControllerServer(d *GPFSDriver) *GPFSControllerServer {
	return &GPFSControllerServer{
		Driver: d,
	}
}

func NewNodeServer(d *GPFSDriver, mounter *mount.SafeFormatAndMount) *GPFSNodeServer {
	return &GPFSNodeServer{
		Driver: d,
		Mounter: mounter,
	}
}

func (driver *GPFSDriver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) error {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.V(4).Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	driver.vcap = vca
	return nil
}

func (driver *GPFSDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) error {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		glog.V(4).Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	driver.cscap = csc
	return nil
}

func (driver *GPFSDriver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) error {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		glog.V(4).Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	driver.nscap = nsc
	return nil
}

func (driver *GPFSDriver) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}
	for _, cap := range driver.cscap {
		if c == cap.GetRpc().Type {
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, "Invalid controller service request")
}

func (driver *GPFSDriver) SetupGPFSDriver(name, vendorVersion, nodeID string, mounter *mount.SafeFormatAndMount) error {
	if name == "" {
		return fmt.Errorf("Driver name missing")
	}

	driver.name = name
	driver.vendorVersion = vendorVersion
	driver.nodeID = nodeID

	// Adding Capabilities
	vcam := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}
	driver.AddVolumeCapabilityAccessModes(vcam)

	csc := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	}
	driver.AddControllerServiceCapabilities(csc)

	ns := []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
	}
	driver.AddNodeServiceCapabilities(ns)

	// Set up RPC Servers
	driver.ids = NewIdentityServer(driver)
	driver.ns = NewNodeServer(driver, mounter) 
	driver.cs = NewControllerServer(driver)

	return nil
}

func (driver *GPFSDriver) Run(endpoint string) {
	glog.Infof("Driver: %v version: %v", driver.name, driver.vendorVersion)
	s := NewNonBlockingGRPCServer()
	s.Start(endpoint, driver.ids, driver.cs, driver.ns)
	s.Wait()
}
