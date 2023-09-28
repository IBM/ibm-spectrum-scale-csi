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
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const (
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

	defaultPrimaryFileset = "spectrum-scale-csi-volume-store"
	symlinkDir            = ".volumes"
	volumeStatsCapability = "VOLUME_STATS_CAPABILITY"
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
	// id of the IBM Storage Scale cluster
	id string
	// name of the IBM Storage Scale cluster
	name string
	// time when the object was last updated.
	lastupdated time.Time
	// expiry duration in hours.
	expiryDuration float64
}

// ClusterName stores the name of the cluster.
type ClusterName struct {
	// name of the IBM Storage Scale cluster
	name string
}

// ClusterID stores the id of the cluster.
type ClusterID struct {
	// id of the IBM Storage Scale cluster
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

func GetScaleDriver(ctx context.Context) *ScaleDriver {
	klog.V(4).Infof("[%s] IBM Storage Scale GetScaleDriver", utils.GetLoggerId(ctx))
	return &ScaleDriver{}
}

func NewIdentityServer(ctx context.Context, d *ScaleDriver) *ScaleIdentityServer {
	klog.V(4).Infof("[%s] Starting IdentityServer", utils.GetLoggerId(ctx))
	return &ScaleIdentityServer{
		Driver: d,
	}
}

func NewControllerServer(ctx context.Context, d *ScaleDriver, connMap map[string]connectors.SpectrumScaleConnector, cmap settings.ScaleSettingsConfigMap, primary settings.Primary) *ScaleControllerServer {
	klog.V(4).Infof("[%s] Starting ControllerServer", utils.GetLoggerId(ctx))
	d.connmap = connMap
	d.cmap = cmap
	d.primary = primary
	d.reqmap = make(map[string]int64)
	return &ScaleControllerServer{
		Driver: d,
	}
}

func NewNodeServer(ctx context.Context, d *ScaleDriver) *ScaleNodeServer {
	klog.V(4).Infof("[%s] Starting NewNodeServer", utils.GetLoggerId(ctx))
	return &ScaleNodeServer{
		Driver: d,
	}
}

func (driver *ScaleDriver) AddVolumeCapabilityAccessModes(ctx context.Context, vc []csi.VolumeCapability_AccessMode_Mode) error {
	klog.V(4).Infof("[%s] AddVolumeCapabilityAccessModes", utils.GetLoggerId(ctx))
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		klog.Infof("[%s] Enabling volume access mode: %v", utils.GetLoggerId(ctx), c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	driver.vcap = vca
	return nil
}

func (driver *ScaleDriver) AddControllerServiceCapabilities(ctx context.Context, cl []csi.ControllerServiceCapability_RPC_Type) error {
	klog.V(4).Infof("[%s] AddControllerServiceCapabilities", utils.GetLoggerId(ctx))
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		klog.Infof("[%s] Enabling controller service capability: %v", utils.GetLoggerId(ctx), c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	driver.cscap = csc
	return nil
}

func (driver *ScaleDriver) AddNodeServiceCapabilities(ctx context.Context, nl []csi.NodeServiceCapability_RPC_Type) error {
	klog.V(4).Infof("[%s] AddNodeServiceCapabilities", utils.GetLoggerId(ctx))
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		klog.V(4).Infof("[%s] Enabling node service capability: %v", utils.GetLoggerId(ctx), n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	driver.nscap = nsc
	return nil
}

func (driver *ScaleDriver) ValidateControllerServiceRequest(ctx context.Context, c csi.ControllerServiceCapability_RPC_Type) error {
	klog.Infof("[%s] ValidateControllerServiceRequest", utils.GetLoggerId(ctx))
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

func (driver *ScaleDriver) SetupScaleDriver(ctx context.Context, name, vendorVersion, nodeID string) error {
	klog.Infof("[%s] SetupScaleDriver. name: %s, version: %v, nodeID: %s", utils.GetLoggerId(ctx), name, vendorVersion, nodeID)
	if name == "" {
		return fmt.Errorf("driver name missing")
	}

	scmap, cmap, primary, err := driver.PluginInitialize(ctx)
	if err != nil {
		klog.Errorf("[%s] Error in plugin initialization: %s", utils.GetLoggerId(ctx), err)
		return err
	}

	driver.name = name
	driver.vendorVersion = vendorVersion
	driver.nodeID = nodeID

	// Adding Capabilities
	vcam := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}
	_ = driver.AddVolumeCapabilityAccessModes(ctx, vcam)

	csc := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	}
	_ = driver.AddControllerServiceCapabilities(ctx, csc)

	ns := []csi.NodeServiceCapability_RPC_Type{}
	statsCapability := os.Getenv(volumeStatsCapability)
	if strings.ToUpper(statsCapability) != "DISABLED" {
		klog.Infof("[%s] volume stats capability is enabled", utils.GetLoggerId(ctx))
		ns = append(ns, csi.NodeServiceCapability_RPC_GET_VOLUME_STATS)
	} else {
		klog.Infof("[%s] volume stats capability is disabled", utils.GetLoggerId(ctx))
	}
	_ = driver.AddNodeServiceCapabilities(ctx, ns)

	driver.ids = NewIdentityServer(ctx, driver)
	driver.ns = NewNodeServer(ctx, driver)
	driver.cs = NewControllerServer(ctx, driver, scmap, cmap, primary)
	return nil
}

func (driver *ScaleDriver) PluginInitialize(ctx context.Context) (map[string]connectors.SpectrumScaleConnector, settings.ScaleSettingsConfigMap, settings.Primary, error) { //nolint:funlen
	klog.Infof("[%s] Initialize IBM Storage Scale CSI driver", utils.GetLoggerId(ctx))
	scaleConfig := settings.LoadScaleConfigSettings(ctx)

	scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
	primaryInfo := settings.Primary{}

	for i := 0; i < len(scaleConfig.Clusters); i++ {
		cluster := scaleConfig.Clusters[i]

		sc, err := connectors.GetSpectrumScaleConnector(ctx, cluster)
		if err != nil {
			klog.Errorf("[%s] Unable to initialize IBM Storage Scale connector for cluster %s", utils.GetLoggerId(ctx), cluster.ID)
			return nil, scaleConfig, primaryInfo, err
		}

		scaleConnMap[cluster.ID] = sc

		if cluster.Primary != (settings.Primary{}) {

			// Check if GUI is reachable - only for primary cluster
			clusterId, err := sc.GetClusterId(ctx)
			if err != nil {
				klog.Errorf("[%s] Error getting cluster ID: %v", utils.GetLoggerId(ctx), err)
				return nil, scaleConfig, primaryInfo, err
			}

			scaleConnMap["primary"] = sc
			scaleConfig.Clusters[i].Primary.PrimaryCid = clusterId

			//If primary fileset value is not specified then use the default one
			if scaleConfig.Clusters[i].Primary.PrimaryFset == "" {
				scaleConfig.Clusters[i].Primary.PrimaryFset = defaultPrimaryFileset
			}
			primaryInfo = scaleConfig.Clusters[i].Primary
		}
	}

	klog.Infof("[%s] IBM Storage Scale CSI driver initialized", utils.GetLoggerId(ctx))
	return scaleConnMap, scaleConfig, primaryInfo, nil
}

func (driver *ScaleDriver) Run(ctx context.Context, endpoint string) {
	s := NewNonBlockingGRPCServer()
	s.Start(endpoint, driver.ids, driver.cs, driver.ns)
	s.Wait()
}
