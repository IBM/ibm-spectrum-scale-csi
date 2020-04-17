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
	"path"
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PluginFolder defines the location of scaleplugin
const (
	PluginFolder          = "/var/lib/kubelet/plugins/ibm-spectrum-scale-csi"
	DefaultPrimaryFileset = "spectrum-scale-csi-volume-store"
)

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
	}
	_ = driver.AddControllerServiceCapabilities(csc)

	ns := []csi.NodeServiceCapability_RPC_Type{}
	_ = driver.AddNodeServiceCapabilities(ns)

	driver.ids = NewIdentityServer(driver)
	driver.ns = NewNodeServer(driver)
	driver.cs = NewControllerServer(driver, scmap, cmap, primary)
	return nil
}

func (driver *ScaleDriver) PluginInitialize() (map[string]connectors.SpectrumScaleConnector, settings.ScaleSettingsConfigMap, settings.Primary, error) { //nolint:funlen
	glog.V(3).Infof("gpfs PluginInitialize")
	scaleConfig := settings.LoadScaleConfigSettings()

	isValid, err := driver.ValidateScaleConfigParameters(scaleConfig)
	if !isValid {
		glog.Errorf("Parameter validation failure")
		return nil, settings.ScaleSettingsConfigMap{}, settings.Primary{}, err
	}

	scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
	primaryInfo := settings.Primary{}
	remoteFilesystemName := ""

	for i := 0; i < len(scaleConfig.Clusters); i++ {
		cluster := scaleConfig.Clusters[i]

		sc, err := connectors.GetSpectrumScaleConnector(cluster)
		if err != nil {
			glog.Errorf("Unable to initialize Spectrum Scale connector for cluster %s", cluster.ID)
			return nil, scaleConfig, primaryInfo, err
		}

		// validate cluster ID
		clusterId, err := sc.GetClusterId()
		if err != nil {
			glog.Errorf("Error getting cluster ID: %v", err)
			return nil, scaleConfig, primaryInfo, err
		}
		if cluster.ID != clusterId {
			glog.Errorf("Cluster ID %s from scale config doesnt match the ID from cluster %s.", cluster.ID, clusterId)
			return nil, scaleConfig, primaryInfo, fmt.Errorf("Cluster ID doesnt match the cluster")
		}

		scaleConnMap[clusterId] = sc

		if cluster.Primary != (settings.Primary{}) {
			scaleConnMap["primary"] = sc

			// check if primary filesystem exists and mounted on atleast one node
			fsMount, err := sc.GetFilesystemMountDetails(cluster.Primary.GetPrimaryFs())
			if err != nil {
				glog.Errorf("Error in getting filesystem details for %s", cluster.Primary.GetPrimaryFs())
				return nil, scaleConfig, cluster.Primary, err
			}
			if fsMount.NodesMounted == nil || len(fsMount.NodesMounted) == 0 {
				return nil, scaleConfig, cluster.Primary, fmt.Errorf("Primary filesystem not mounted on any node")
			}
			// In case primary fset value is not specified in configuation then use default
			if scaleConfig.Clusters[i].Primary.PrimaryFset == "" {
				scaleConfig.Clusters[i].Primary.PrimaryFset = DefaultPrimaryFileset
				glog.Infof("primaryFset is not specified in configuration using default %s", DefaultPrimaryFileset)
			}
			scaleConfig.Clusters[i].Primary.PrimaryFSMount = fsMount.MountPoint
			scaleConfig.Clusters[i].Primary.PrimaryCid = clusterId

			primaryInfo = scaleConfig.Clusters[i].Primary

			// RemoteFS name from Local Filesystem details
			remoteDeviceName := strings.Split(fsMount.RemoteDeviceName, ":")
			remoteFilesystemName = remoteDeviceName[len(remoteDeviceName)-1]
		}
	}

	fs := primaryInfo.GetPrimaryFs()
	sconn := scaleConnMap["primary"]
	fsmount := primaryInfo.PrimaryFSMount
	if primaryInfo.RemoteCluster != "" {
		sconn = scaleConnMap[primaryInfo.RemoteCluster]
		if remoteFilesystemName == "" {
			return scaleConnMap, scaleConfig, primaryInfo, fmt.Errorf("Failed to get the name of remote Filesystem")
		}
		fs = remoteFilesystemName
		// check if primary filesystem exists on remote cluster and mounted on atleast one node
		fsMount, err := sconn.GetFilesystemMountDetails(fs)
		if err != nil {
			glog.Errorf("Error in getting filesystem details for %s from cluster %s", fs, primaryInfo.RemoteCluster)
			return scaleConnMap, scaleConfig, primaryInfo, err
		}

		glog.Infof("remote fsMount = %v", fsMount)
		if fsMount.NodesMounted == nil || len(fsMount.NodesMounted) == 0 {
			return scaleConnMap, scaleConfig, primaryInfo, fmt.Errorf("Primary filesystem not mounted on any node on cluster %s", primaryInfo.RemoteCluster)
		}
		fsmount = fsMount.MountPoint
	}

	fsetlinkpath, err := driver.CreatePrimaryFileset(sconn, fs, fsmount, primaryInfo.PrimaryFset, primaryInfo.GetInodeLimit())
	if err != nil {
		glog.Errorf("Error in creating primary fileset")
		return scaleConnMap, scaleConfig, primaryInfo, err
	}

	if fsmount != primaryInfo.PrimaryFSMount {
		fsetlinkpath = strings.Replace(fsetlinkpath, fsmount, primaryInfo.PrimaryFSMount, 1)
	}

	// Validate hostpath from daemonset is valid
	err = driver.ValidateHostpath(primaryInfo.PrimaryFSMount, fsetlinkpath)
	if err != nil {
		glog.Errorf("Hostpath validation failed")
		return scaleConnMap, scaleConfig, primaryInfo, err
	}

	// Create directory where volume symlinks will reside
	symlinkPath, relativePath, err := driver.CreateSymlinkPath(scaleConnMap["primary"], primaryInfo.GetPrimaryFs(), primaryInfo.PrimaryFSMount, fsetlinkpath)
	if err != nil {
		glog.Errorf("Error in creating volumes directory")
		return scaleConnMap, scaleConfig, primaryInfo, err
	}
	primaryInfo.SymlinkAbsolutePath = symlinkPath
	primaryInfo.SymlinkRelativePath = relativePath
	primaryInfo.PrimaryFsetLink = fsetlinkpath

	glog.Infof("IBM Spectrum Scale: Plugin initialized")
	return scaleConnMap, scaleConfig, primaryInfo, nil
}

func (driver *ScaleDriver) CreatePrimaryFileset(sc connectors.SpectrumScaleConnector, primaryFS string, fsmount string, filesetName string, inodeLimit string) (string, error) {
	glog.V(4).Infof("gpfs CreatePrimaryFileset. primaryFS: %s, mountpoint: %s, filesetName: %s", primaryFS, fsmount, filesetName)

	// create primary fileset if not already created
	fsetResponse, err := sc.ListFileset(primaryFS, filesetName)
	linkpath := fsetResponse.Config.Path
	newlinkpath := path.Join(fsmount, filesetName)

	if err != nil {
		glog.Infof("Primary fileset %s not found. Creating it.", filesetName)
		opts := make(map[string]interface{})
		if inodeLimit != "" {
			opts[connectors.UserSpecifiedInodeLimit] = inodeLimit
		}
		err = sc.CreateFileset(primaryFS, filesetName, opts)
		if err != nil {
			glog.Errorf("Unable to create primary fileset %s", filesetName)
			return "", err
		}
		linkpath = newlinkpath
	} else if linkpath == "" || linkpath == "--" {
		glog.Infof("Primary fileset %s not linked. Linking it.", filesetName)
		err = sc.LinkFileset(primaryFS, filesetName, newlinkpath)
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

func (driver *ScaleDriver) CreateSymlinkPath(sc connectors.SpectrumScaleConnector, fs string, fsmount string, fsetlinkpath string) (string, string, error) {
	glog.V(4).Infof("gpfs CreateSymlinkPath. filesystem: %s, mountpoint: %s, filesetlinkpath: %s", fs, fsmount, fsetlinkpath)

	dirpath := strings.Replace(fsetlinkpath, fsmount, "", 1)
	dirpath = strings.Trim(dirpath, "!/")
	fsetlinkpath = strings.TrimSuffix(fsetlinkpath, "/")

	dirpath = fmt.Sprintf("%s/.volumes", dirpath)
	symlinkpath := fmt.Sprintf("%s/.volumes", fsetlinkpath)

	err := sc.MakeDirectory(fs, dirpath, "0", "0")
	if err != nil {
		glog.Errorf("Make directory failed on filesystem %s, path = %s", fs, dirpath)
		return symlinkpath, dirpath, err
	}

	return symlinkpath, dirpath, nil
}

func (driver *ScaleDriver) ValidateHostpath(mountpath string, linkpath string) error {
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

// ValidateScaleConfigParameters : Validating the Configuration provided for Spectrum Scale CSI Driver
func (driver *ScaleDriver) ValidateScaleConfigParameters(scaleConfig settings.ScaleSettingsConfigMap) (bool, error) {
	glog.V(4).Infof("gpfs ValidateScaleConfigParameters.")
	if len(scaleConfig.Clusters) == 0 {
		return false, fmt.Errorf("Missing cluster information in Spectrum Scale configuration")
	}

	primaryClusterFound := false
	rClusterForPrimaryFS := ""
	var cl = make([]string, len(scaleConfig.Clusters))
	issueFound := false

	for i := 0; i < len(scaleConfig.Clusters); i++ {
		cluster := scaleConfig.Clusters[i]

		if cluster.ID == "" {
			issueFound = true
			glog.Errorf("Mandatory parameter 'id' is not specified")
		}
		if len(cluster.RestAPI) == 0 {
			issueFound = true
			glog.Errorf("Mandatory section 'restApi' is not specified for cluster %v", cluster.ID)
		}
		if len(cluster.RestAPI) != 0 && cluster.RestAPI[0].GuiHost == "" {
			issueFound = true
			glog.Errorf("Mandatory parameter 'guiHost' is not specified for cluster %v", cluster.ID)
		}

		if cluster.Primary != (settings.Primary{}) {
			if primaryClusterFound {
				issueFound = true
				glog.Errorf("More than one primary clusters specified")
			}

			primaryClusterFound = true

			if cluster.Primary.GetPrimaryFs() == "" {
				issueFound = true
				glog.Errorf("Mandatory parameter 'primaryFs' is not specified for primary cluster %v", cluster.ID)
			}

			rClusterForPrimaryFS = cluster.Primary.RemoteCluster
		} else {
			cl[i] = cluster.ID
		}

		if cluster.Secrets == "" {
			issueFound = true
			glog.Errorf("Mandatory parameter 'secrets' is not specified for cluster %v", cluster.ID)
		}

		if cluster.SecureSslMode && cluster.CacertValue == nil {
			issueFound = true
			glog.Errorf("CA certificate not specified in secure SSL mode for cluster %v", cluster.ID)
		}
	}

	if !primaryClusterFound {
		issueFound = true
		glog.Errorf("No primary clusters specified")
	}

	if rClusterForPrimaryFS != "" && !utils.StringInSlice(rClusterForPrimaryFS, cl) {
		issueFound = true
		glog.Errorf("Remote cluster specified for primary filesystem: %s, but no definition found for it in config", rClusterForPrimaryFS)
	}

	if issueFound {
		return false, fmt.Errorf("one or more issue found in Spectrum scale csi driver configuration, check Spectrum Scale csi driver logs")
	}

	return true, nil
}

func (driver *ScaleDriver) Run(endpoint string) {
	glog.Infof("Driver: %v version: %v", driver.name, driver.vendorVersion)
	s := NewNonBlockingGRPCServer()
	s.Start(endpoint, driver.ids, driver.cs, driver.ns)
	s.Wait()
}
