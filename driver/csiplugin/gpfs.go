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
	"strings"
	"sync"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DefaultPrimaryFileset = "spectrum-scale-csi-volume-store"

	SNAP_JOB_NOT_STARTED    = 0
	SNAP_JOB_RUNNING        = 1
	SNAP_JOB_COMPLETED      = 2
	SNAP_JOB_FAILED         = 3
	VOLCOPY_JOB_FAILED      = 4
	VOLCOPY_JOB_RUNNING     = 5
	VOLCOPY_JOB_COMPLETED   = 6
	VOLCOPY_JOB_NOT_STARTED = 7
	JOB_STATUS_UNKNOWN      = 8

	STORAGECLASS_CLASSIC  = "0"
	STORAGECLASS_ADVANCED = "1"

	// Volume types
	FILE_DIRECTORYBASED_VOLUME     = "0"
	FILE_DEPENDENTFILESET_VOLUME   = "1"
	FILE_INDEPENDENTFILESET_VOLUME = "2"

	//	BLOCK_FILESET_VOLUME = 3

	ENVSymDirPath = "SYMLINK_DIR_PATH"
)

type SnapCopyJobDetails struct {
	jobStatus int
	volID     string
}

type VolCopyJobDetails struct {
	jobStatus int
	volID     string
}

// ClusterDetails stores information of the cluster.
type ClusterDetails struct {
	// id of the Spectrum Scale cluster
	id string
	// name of the Spectrum Scale cluster
	name string
	// time when the object was last updated.
	lastupdated time.Time
	// expiry duration in hours.
	expiryDuration float64
}

// ClusterName stores the name of the cluster.
type ClusterName struct {
	// name of the Spectrum Scale cluster
	name string
}

// ClusterID stores the id of the cluster.
type ClusterID struct {
	// id of the Spectrum Scale cluster
	id string
}

type ScaleDriver struct {
	name          string
	vendorVersion string
	nodeID        string

	ids *ScaleIdentityServer
	ns  *ScaleNodeServer
	cs  *ScaleControllerServer

	connmap map[string]connectors.SpectrumScaleConnector
	cmap    settings.ScaleSettingsConfigMap
	primary settings.Primary
	reqmap  map[string]int64

	snapjobstatusmap    sync.Map
	volcopyjobstatusmap sync.Map

	// clusterMap map stores the cluster name as key and cluster details as value.
	clusterMap sync.Map

	vcap  []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
	nscap []*csi.NodeServiceCapability
}

func GetScaleDriver() *ScaleDriver {
	glog.V(3).Infof("gpfs GetScaleDriver")
	return &ScaleDriver{}
}

func NewIdentityServer(d *ScaleDriver) *ScaleIdentityServer {
	glog.V(3).Infof("gpfs NewIdentityServer")
	return &ScaleIdentityServer{
		Driver: d,
	}
}

func NewControllerServer(d *ScaleDriver, connMap map[string]connectors.SpectrumScaleConnector, cmap settings.ScaleSettingsConfigMap, primary settings.Primary) *ScaleControllerServer {
	glog.V(3).Infof("gpfs NewControllerServer")
	d.connmap = connMap
	d.cmap = cmap
	d.primary = primary
	d.reqmap = make(map[string]int64)
	return &ScaleControllerServer{
		Driver: d,
	}
}

func NewNodeServer(d *ScaleDriver) *ScaleNodeServer {
	glog.V(3).Infof("gpfs NewNodeServer")
	return &ScaleNodeServer{
		Driver: d,
	}
}

func (driver *ScaleDriver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) error {
	glog.V(3).Infof("gpfs AddVolumeCapabilityAccessModes")
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.V(3).Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	driver.vcap = vca
	return nil
}

func (driver *ScaleDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) error {
	glog.V(3).Infof("gpfs AddControllerServiceCapabilities")
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		glog.V(3).Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	driver.cscap = csc
	return nil
}

func (driver *ScaleDriver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) error {
	glog.V(3).Infof("gpfs AddNodeServiceCapabilities")
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		glog.V(3).Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	driver.nscap = nsc
	return nil
}

func (driver *ScaleDriver) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	glog.V(3).Infof("gpfs ValidateControllerServiceRequest")
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

func (driver *ScaleDriver) SetupScaleDriver(name, vendorVersion, nodeID string) error {
	glog.V(3).Infof("gpfs SetupScaleDriver. name: %s, version: %v, nodeID: %s", name, vendorVersion, nodeID)
	if name == "" {
		return fmt.Errorf("Driver name missing")
	}

	scmap, cmap, primary, err := driver.PluginInitialize()
	if err != nil {
		glog.Errorf("Error in plugin initialization: %s", err)
		return err
	}

	driver.name = name
	driver.vendorVersion = vendorVersion
	driver.nodeID = nodeID

	// Adding Capabilities
	vcam := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}
	_ = driver.AddVolumeCapabilityAccessModes(vcam)

	csc := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	}
	_ = driver.AddControllerServiceCapabilities(csc)

	ns := []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
	}
	_ = driver.AddNodeServiceCapabilities(ns)

	driver.ids = NewIdentityServer(driver)
	driver.ns = NewNodeServer(driver)
	driver.cs = NewControllerServer(driver, scmap, cmap, primary)
	return nil
}

func (driver *ScaleDriver) PluginInitialize() (map[string]connectors.SpectrumScaleConnector, settings.ScaleSettingsConfigMap, settings.Primary, error) { //nolint:funlen
	glog.V(3).Infof("gpfs PluginInitialize")
	scaleConfig := settings.LoadScaleConfigSettings()

	scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
	primaryInfo := settings.Primary{}

	for i := 0; i < len(scaleConfig.Clusters); i++ {
		cluster := scaleConfig.Clusters[i]

		sc, err := connectors.GetSpectrumScaleConnector(cluster)
		if err != nil {
			glog.Errorf("Unable to initialize Spectrum Scale connector for cluster %s", cluster.ID)
			return nil, scaleConfig, primaryInfo, err
		}

		scaleConnMap[cluster.ID] = sc

		if cluster.Primary != (settings.Primary{}) {

			// Check if GUI is reachable - only for primary cluster
			_, err := sc.GetClusterId()
			if err != nil {
				glog.Errorf("Error getting cluster ID: %v", err)
				return nil, scaleConfig, primaryInfo, err
			}

			scaleConnMap["primary"] = sc
			primaryInfo = scaleConfig.Clusters[i].Primary
		}
	}

	symlinkDirPath := utils.GetEnv(ENVSymDirPath, notFound)
	if symlinkDirPath == notFound {
		message := fmt.Sprintf("Unable to get environmental variable %s", ENVSymDirPath)
		glog.Errorf(message)
		return nil, scaleConfig, primaryInfo, fmt.Errorf(message)
	}

	primaryInfo.SymlinkAbsolutePath = symlinkDirPath

	//get relative path
	pathTokens := strings.Split(symlinkDirPath, "/")
	len := len(pathTokens)
	symlinkDirRelPath := pathTokens[len-2] + "/" + pathTokens[len-1]
	primaryInfo.SymlinkRelativePath = symlinkDirRelPath

	glog.V(3).Infof("Symlink directory path", "absolute:", symlinkDirPath, "relative", symlinkDirRelPath)
	glog.Infof("IBM Spectrum Scale: Plugin initialized")
	return scaleConnMap, scaleConfig, primaryInfo, nil
}

func (driver *ScaleDriver) Run(endpoint string) {
	glog.Infof("Driver: %v version: %v", driver.name, driver.vendorVersion)
	s := NewNonBlockingGRPCServer()
	s.Start(endpoint, driver.ids, driver.cs, driver.ns)
	s.Wait()
}
