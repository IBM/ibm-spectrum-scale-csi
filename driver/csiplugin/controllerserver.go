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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	no                   = "no"
	yes                  = "yes"
	notFound             = "NOT_FOUND"
	filesystemTypeRemote = "remote"
	filesystemMounted    = "mounted"
	filesetUnlinkedPath  = "--"
)

type ScaleControllerServer struct {
	Driver *ScaleDriver
}

func (cs *ScaleControllerServer) IfSameVolReqInProcess(scVol *scaleVolume) (bool, error) {
	cap, volpresent := cs.Driver.reqmap[scVol.VolName]
	glog.Infof("reqmap: %v", cs.Driver.reqmap)
	if volpresent {
		if cap == int64(scVol.VolSize) {
			return true, nil
		} else {
			return false, status.Error(codes.Internal, fmt.Sprintf("Volume %v present in map but requested size %v does not match with size %v in map", scVol.VolName, scVol.VolSize, cap))
		}
	}
	return false, nil
}

func (cs *ScaleControllerServer) GetPriConnAndSLnkPath() (connectors.SpectrumScaleConnector, string, string, string, string, string, error) {
	primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]

	if isprimaryConnPresent {
		return primaryConn, cs.Driver.primary.SymlinkRelativePath, cs.Driver.primary.GetPrimaryFs(), cs.Driver.primary.PrimaryFSMount, cs.Driver.primary.SymlinkAbsolutePath, cs.Driver.primary.PrimaryCid, nil
	}

	return nil, "", "", "", "", "", status.Error(codes.Internal, "Primary connector not present in configMap")
}

func (cs *ScaleControllerServer) createLWVol(scVol *scaleVolume) (string, error) {
	glog.V(4).Infof("volume: [%v] - ControllerServer:createLWVol", scVol.VolName)
	var err error

	// check if directory exist
	dirExists, err := scVol.PrimaryConnector.CheckIfFileDirPresent(scVol.VolBackendFs, scVol.VolDirBasePath)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to check if DirBasePath %v is present in filesystem %v. Error : %v", scVol.VolName, scVol.VolDirBasePath, scVol.VolBackendFs, err)
		return "", status.Error(codes.Internal, fmt.Sprintf("unable to check if DirBasePath %v is present in filesystem %v. Error : %v", scVol.VolDirBasePath, scVol.VolBackendFs, err))
	}

	if !dirExists {
		glog.Errorf("volume:[%v] - directory base path %v not present in filesystem %v", scVol.VolName, scVol.VolDirBasePath, scVol.VolBackendFs)
		return "", status.Error(codes.Internal, fmt.Sprintf("directory base path %v not present in filesystem %v", scVol.VolDirBasePath, scVol.VolBackendFs))
	}

	// create directory in the filesystem specified in storageClass
	dirPath := fmt.Sprintf("%s/%s", scVol.VolDirBasePath, scVol.VolName)

	glog.V(4).Infof("volume: [%v] - creating directory %v", scVol.VolName, dirPath)
	err = cs.createDirectory(scVol, dirPath)
	if err != nil {
		glog.Errorf("volume:[%v] - failed to create directory %v. Error : %v", scVol.VolName, dirPath, err)
		return "", status.Error(codes.Internal, err.Error())
	}
	return dirPath, nil
}

func (cs *ScaleControllerServer) generateVolID(scVol *scaleVolume) (string, error) {
	glog.V(4).Infof("volume: [%v] - ControllerServer:generateVolId", scVol.VolName)
	var volId string

	/* We need to put FSUUID for localFS in volID */
	uid, err := scVol.PrimaryConnector.GetFsUid(scVol.LocalFS)
	if err != nil {
		return "", fmt.Errorf("unable to get FS UUID for FS [%v]. Error [%v]", scVol.LocalFS, err)
	}

	if scVol.IsFilesetBased {
		fSetuid, err := scVol.Connector.GetFileSetUid(scVol.VolBackendFs, scVol.VolName)

		if err != nil {
			return "", fmt.Errorf("unable to get Fset UID for [%v] in FS [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err)
		}

		/* <cluster_id>;<filesystem_uuid>;fileset=<fileset_id>; path=<symlink_path> */
		slink := fmt.Sprintf("%s/%s", scVol.PrimarySLnkPath, scVol.VolName)
		volId = fmt.Sprintf("%s;%s;fileset=%s;path=%s", scVol.ClusterId, uid, fSetuid, slink)
	} else {
		/* <cluster_id>;<filesystem_uuid>;path=<symlink_path> */
		slink := fmt.Sprintf("%s/%s", scVol.PrimarySLnkPath, scVol.VolName)
		volId = fmt.Sprintf("%s;%s;path=%s", scVol.ClusterId, uid, slink)
	}
	return volId, nil
}

func (cs *ScaleControllerServer) getTargetPath(fsetLinkPath, fsMountPoint, volumeName string) (string, error) {
	if fsetLinkPath == "" || fsMountPoint == "" {
		glog.Errorf("volume:[%v] - missing details to generate target path fileset junctionpath: [%v], filesystem mount point: [%v]", volumeName, fsetLinkPath, fsMountPoint)
		return "", fmt.Errorf("missing details to generate target path fileset junctionpath: [%v], filesystem mount point: [%v]", fsetLinkPath, fsMountPoint)
	}
	glog.V(4).Infof("volume: [%v] - ControllerServer:getTargetPath", volumeName)
	targetPath := strings.Replace(fsetLinkPath, fsMountPoint, "", 1)
	targetPath = strings.Trim(targetPath, "!/")
	targetPath = fmt.Sprintf("%s/%s-data", targetPath, volumeName)
	return targetPath, nil
}

func (cs *ScaleControllerServer) createDirectory(scVol *scaleVolume, targetPath string) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:createDirectory", scVol.VolName)
	dirExists, err := scVol.Connector.CheckIfFileDirPresent(scVol.VolBackendFs, targetPath)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", scVol.VolName, targetPath, scVol.VolBackendFs, err)
		return fmt.Errorf("unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
	}

	if !dirExists {
		err = scVol.Connector.MakeDirectory(scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid)
		if err != nil {
			// Directory creation failed, no cleanup will retry in next retry
			glog.Errorf("volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", scVol.VolName, targetPath, scVol.VolBackendFs, err)
			return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
		}
	}
	return nil
}

func (cs *ScaleControllerServer) createSoftlink(scVol *scaleVolume, target string) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:createSoftlink", scVol.VolName)
	volSlnkPath := fmt.Sprintf("%s/%s", scVol.PrimarySLnkRelPath, scVol.VolName)
	symLinkExists, err := scVol.PrimaryConnector.CheckIfFileDirPresent(scVol.PrimaryFS, volSlnkPath)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to check if symlink path [%v] exists in filesystem [%v]. Error: %v", scVol.VolName, scVol.PrimaryFS, err)
		return fmt.Errorf("unable to check if symlink path [%v] exists in filesystem [%v]. Error: %v", volSlnkPath, scVol.PrimaryFS, err)
	}

	if !symLinkExists {
		glog.Infof("symlink info filesystem [%v] TargetFS [%v]  target Path [%v] linkPath [%v]", scVol.PrimaryFS, scVol.LocalFS, target, volSlnkPath)
		err = scVol.PrimaryConnector.CreateSymLink(scVol.PrimaryFS, scVol.LocalFS, target, volSlnkPath)
		if err != nil {
			glog.Errorf("volume:[%v] - failed to create symlink [%v] in filesystem [%v], for target [%v] in filesystem [%v]. Error [%v]", scVol.VolName, volSlnkPath, scVol.PrimaryFS, target, scVol.LocalFS, err)
			return fmt.Errorf("failed to create symlink [%v] in filesystem [%v], for target [%v] in filesystem [%v]. Error [%v]", volSlnkPath, scVol.PrimaryFS, target, scVol.LocalFS, err)
		}
	}
	return nil
}

func (cs *ScaleControllerServer) setQuota(scVol *scaleVolume) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:setQuota", scVol.VolName)
	quota, err := scVol.Connector.ListFilesetQuota(scVol.VolBackendFs, scVol.VolName)
	if err != nil {
		return fmt.Errorf("unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err)
	}

	filesetQuotaBytes, err := ConvertToBytes(quota)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid number specified") {
			// Invalid number specified means quota is not set
			filesetQuotaBytes = 0
		} else {
			return fmt.Errorf("unable to convirt quota for fileset [%v] in filesystem [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err)
		}
	}
	if filesetQuotaBytes != scVol.VolSize && filesetQuotaBytes != 0 {
		// quota does not match and it is not 0 - It might not be fileset created by us
		return fmt.Errorf("Fileset %v present but quota %v does not match with requested size %v", scVol.VolName, filesetQuotaBytes, scVol.VolSize)
	}

	volsiz := strconv.FormatUint(scVol.VolSize, 10)
	err = scVol.Connector.SetFilesetQuota(scVol.VolBackendFs, scVol.VolName, volsiz)
	if err != nil {
		// failed to set quota, no cleanup, next retry might be able to set quota
		return fmt.Errorf("unable to set quota [%v] on fileset [%v] of FS [%v]", scVol.VolSize, scVol.VolName, scVol.VolBackendFs)
	}
	return nil
}

// CreateFilesetBasedVol : Create Fileset based volumes
func (cs *ScaleControllerServer) createFilesetBasedVol(scVol *scaleVolume) (string, error) { //nolint:gocyclo,funlen
	glog.V(4).Infof("volume: [%v] - ControllerServer:createFilesetBasedVol", scVol.VolName)
	opt := make(map[string]interface{})

	// fileset can not be created if filesystem is remote.
	glog.V(4).Infof("check if volumes filesystem [%v] is remote or local for cluster [%v]", scVol.VolBackendFs, scVol.ClusterId)
	fsDetails, err := scVol.Connector.GetFilesystemDetails(scVol.VolBackendFs)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in filesystemName") {
			glog.Errorf("volume:[%v] - filesystem %s in not known to cluster %v. Error: %v", scVol.VolName, scVol.VolBackendFs, scVol.ClusterId, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("Filesystem %s in not known to cluster %v. Error: %v", scVol.VolBackendFs, scVol.ClusterId, err))
		}
		glog.Errorf("volume:[%v] - unable to check type of filesystem [%v]. Error: %v", scVol.VolName, scVol.VolBackendFs, err)
		return "", status.Error(codes.Internal, fmt.Sprintf("unable to check type of filesystem [%v]. Error: %v", scVol.VolBackendFs, err))
	}

	if fsDetails.Type == filesystemTypeRemote {
		glog.Errorf("volume:[%v] - filesystem [%v] is not local to cluster [%v]", scVol.VolName, scVol.VolBackendFs, scVol.ClusterId)
		return "", status.Error(codes.Internal, fmt.Sprintf("filesystem [%v] is not local to cluster [%v]", scVol.VolBackendFs, scVol.ClusterId))
	}

	// if filesystem is remote, check it is mounted on remote GUI node.
	if cs.Driver.primary.PrimaryCid != scVol.ClusterId {
		glog.V(4).Infof("check if volumes filesystem [%v] is mounted on remote GUI of cluster [%v]", scVol.VolBackendFs, scVol.ClusterId)
		if fsDetails.Mount.Status != filesystemMounted {
			glog.Errorf("volume:[%v] -  filesystem [%v] is [%v] on remote GUI of cluster [%v]", scVol.VolName, scVol.VolBackendFs, fsDetails.Mount.Status, scVol.ClusterId)
			return "", status.Error(codes.Internal, fmt.Sprintf("Filesystem %v in cluster %v is not mounted", scVol.VolBackendFs, scVol.ClusterId))
		}
	}

	// check if quota is enabled on volume filesystem
	glog.V(4).Infof("check if quota is enabled on filesystem [%v] ", scVol.VolBackendFs)
	if scVol.VolSize != 0 {
		err = scVol.Connector.CheckIfFSQuotaEnabled(scVol.VolBackendFs)
		if err != nil {
			glog.Errorf("volume:[%v] - quota not enabled for filesystem %v of cluster %v. Error: %v", scVol.VolName, scVol.VolBackendFs, scVol.ClusterId, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("quota not enabled for filesystem %v of cluster %v", scVol.VolBackendFs, scVol.ClusterId))
		}
	}

	if scVol.VolUid != "" {
		opt[connectors.UserSpecifiedUid] = scVol.VolUid
	}
	if scVol.VolGid != "" {
		opt[connectors.UserSpecifiedGid] = scVol.VolGid
	}
	if scVol.FilesetType != "" {
		opt[connectors.UserSpecifiedFilesetType] = scVol.FilesetType
	}
	if scVol.InodeLimit != "" {
		opt[connectors.UserSpecifiedInodeLimit] = scVol.InodeLimit
	}
	if scVol.ParentFileset != "" {
		opt[connectors.UserSpecifiedParentFset] = scVol.ParentFileset
	}

	// Check if fileset exist
	filesetInfo, err := scVol.Connector.ListFileset(scVol.VolBackendFs, scVol.VolName)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in 'filesetName'") {
			// This means fileset is not present, create it
			fseterr := scVol.Connector.CreateFileset(scVol.VolBackendFs, scVol.VolName, opt)

			if fseterr != nil {
				// fileset creation failed return without cleanup
				glog.Errorf("volume:[%v] - unable to create fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, fseterr)
				return "", status.Error(codes.Internal, fmt.Sprintf("unable to create fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolBackendFs, fseterr))
			}
			// list fileset and update filesetInfo
			filesetInfo, err = scVol.Connector.ListFileset(scVol.VolBackendFs, scVol.VolName)
			if err != nil {
				// fileset got created but listing failed, return without cleanup
				glog.Errorf("volume:[%v] - unable to list newly created fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("unable to list newly created fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolBackendFs, err))
			}
		} else {
			glog.Errorf("volume:[%v] - unable to list fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolBackendFs, err))
		}
	}

	// fileset is present/created. Confirm if fileset is linked
	if (filesetInfo.Config.Path == "") || (filesetInfo.Config.Path == filesetUnlinkedPath) {
		// this means not linked, link it
		junctionPath := fmt.Sprintf("%s/%s", fsDetails.Mount.MountPoint, scVol.VolName)
		err := scVol.Connector.LinkFileset(scVol.VolBackendFs, scVol.VolName, junctionPath)
		if err != nil {
			glog.Errorf("volume:[%v] - linking fileset [%v] in filesystem [%v] at path [%v] failed. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, junctionPath, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("linking fileset [%v] in filesystem [%v] at path [%v] failed. Error: %v", scVol.VolName, scVol.VolBackendFs, junctionPath, err))
		}
		// update fileset details
		filesetInfo, err = scVol.Connector.ListFileset(scVol.VolBackendFs, scVol.VolName)
		if err != nil {
			glog.Errorf("volume:[%v] - unable to list fileset [%v] in filesystem [%v] after linking. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v] after linking. Error: %v", scVol.VolName, scVol.VolBackendFs, err))
		}
	}

	if scVol.VolSize != 0 {
		err = cs.setQuota(scVol)
		if err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}
	}

	targetBasePath, err := cs.getTargetPath(filesetInfo.Config.Path, fsDetails.Mount.MountPoint, scVol.VolName)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	err = cs.createDirectory(scVol, targetBasePath)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	return targetBasePath, nil
}

func (cs *ScaleControllerServer) getVolumeSizeInBytes(req *csi.CreateVolumeRequest) int64 {
	cap := req.GetCapacityRange()
	return cap.GetRequiredBytes()
}

func (cs *ScaleControllerServer) getConnFromClusterID(cid string) (connectors.SpectrumScaleConnector, error) {
	connector, isConnPresent := cs.Driver.connmap[cid]
	if isConnPresent {
		return connector, nil
	}

	return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get connector for ClusterID : %v", cid))
}

// CreateVolume - Create Volume
func (cs *ScaleControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) { //nolint:gocyclo,funlen
	glog.V(3).Infof("create volume req: %v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateVolume ValidateControllerServiceRequest failed: %v", err))
	}

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Request cannot be empty")
	}

	volName := req.GetName()
	if volName == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume Name is a required field")
	}

	/* Get volume size in bytes */
	volSize := cs.getVolumeSizeInBytes(req)

	reqCapabilities := req.GetVolumeCapabilities()
	if reqCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities is a required field")
	}

	for _, reqCap := range reqCapabilities {
		if reqCap.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented, "Block Volume is not supported")
		}
		if reqCap.GetAccessMode().GetMode() == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			return nil, status.Error(codes.Unimplemented, "Volume with Access Mode ReadOnlyMany is not supported")
		}
	}

	scaleVol, err := getScaleVolumeOptions(req.GetParameters())

	if err != nil {
		return nil, err
	}

	scaleVol.VolName = volName
	scaleVol.VolSize = uint64(volSize)

	/* Get details for Primary Cluster */
	pConn, PSLnkRelPath, PFS, PFSMount, PSLnkPath, PCid, err := cs.GetPriConnAndSLnkPath()

	if err != nil {
		return nil, err
	}

	scaleVol.PrimaryConnector = pConn
	scaleVol.PrimarySLnkRelPath = PSLnkRelPath
	scaleVol.PrimaryFS = PFS
	scaleVol.PrimaryFSMount = PFSMount
	scaleVol.PrimarySLnkPath = PSLnkPath

	// Check if Primary Fileset is linked
	primaryFileset := cs.Driver.primary.PrimaryFset
	glog.V(4).Infof("volume:[%v] - check if primary fileset [%v] is linked", scaleVol.VolName, primaryFileset)
	isPrimaryFilesetLinked, err := scaleVol.PrimaryConnector.IsFilesetLinked(scaleVol.PrimaryFS, primaryFileset)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to get details of Primary Fileset [%v]. Error : [%v]", scaleVol.VolName, primaryFileset, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get details of Primary Fileset [%v]. Error : [%v]", primaryFileset, err))
	}
	if !isPrimaryFilesetLinked {
		glog.Errorf("volume:[%v] - primary fileset [%v] is not linked", scaleVol.VolName, primaryFileset)
		return nil, status.Error(codes.Internal, fmt.Sprintf("primary fileset [%v] is not linked", primaryFileset))
	}

	// primary filesytem must be mounted on GUI node so that we can create the softlink
	glog.V(4).Infof("volume:[%v] - check if primary filesystem [%v] is mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.PrimaryFS)
	isPfsMounted, err := scaleVol.PrimaryConnector.IsFilesystemMountedOnGUINode(scaleVol.PrimaryFS)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to get filesystem mount details for %s on Primary cluster", scaleVol.VolName, scaleVol.PrimaryFS, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem mount details for %s on Primary cluster. Error: %v", scaleVol.PrimaryFS, err))
	}
	if !isPfsMounted {
		glog.Errorf("volume:[%v] - primary filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.PrimaryFS)
		return nil, status.Error(codes.Internal, fmt.Sprintf("primary filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.PrimaryFS))
	}

	// TODO - avoid additional check if primary filesystem is also volume
	glog.V(4).Infof("volume:[%v] - check if volume filesystem [%v] is mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.VolBackendFs)
	volFsInfo, err := scaleVol.PrimaryConnector.GetFilesystemDetails(scaleVol.VolBackendFs)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in filesystemName") {
			glog.Errorf("volume:[%v] - filesystem %s in not known to primary cluster. Error: %v", scaleVol.VolName, scaleVol.VolBackendFs, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("filesystem %s in not known to primary cluster. Error: %v", scaleVol.VolBackendFs, err))
		}
		glog.Errorf("volume:[%v] - unable to get details for filesystem [%v] in Primary cluster. Error: %v", scaleVol.VolName, scaleVol.VolBackendFs, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get details for filesystem [%v] in Primary cluster. Error: %v", scaleVol.VolBackendFs, err))
	}

	if volFsInfo.Mount.Status != filesystemMounted {
		glog.Errorf("volume:[%v] - volume filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.VolBackendFs)
		return nil, status.Error(codes.Internal, fmt.Sprintf("volume filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.VolBackendFs))
	}

	/* scaleVol.VolBackendFs will always be local cluster FS. So we need to find a
	   remote cluster FS in case local cluster FS is remotely mounted. We will find local FS RemoteDeviceName on local cluster, will use that as VolBackendFs and   create fileset on that FS. */

	if scaleVol.IsFilesetBased {
		remoteDeviceName := volFsInfo.Mount.RemoteDeviceName
		splitDevName := strings.Split(remoteDeviceName, ":")
		remDevFs := splitDevName[len(splitDevName)-1]

		scaleVol.LocalFS = scaleVol.VolBackendFs
		scaleVol.VolBackendFs = remDevFs
	} else {
		scaleVol.LocalFS = scaleVol.VolBackendFs
	}

	if scaleVol.IsFilesetBased {
		if scaleVol.ClusterId == "" {
			scaleVol.ClusterId = PCid
			glog.V(3).Infof("clusterID not provided in storage Class, using Primary ClusterID. Volume Name [%v]", scaleVol.VolName)
			if volFsInfo.Type == filesystemTypeRemote {
				glog.Errorf("volume filesystem %s is remotely mounted on Primary cluster, Specify owning cluster ID in storageClass", scaleVol.VolBackendFs)
				return nil, status.Error(codes.Internal, fmt.Sprintf("volume filesystem %s is remotely mounted on Primary cluster, Specify owning cluster ID in storageClass", scaleVol.VolBackendFs))
			}
		}
		conn, err := cs.getConnFromClusterID(scaleVol.ClusterId)
		if err != nil {
			return nil, err
		}

		scaleVol.Connector = conn
	} else {
		scaleVol.Connector = scaleVol.PrimaryConnector
		scaleVol.ClusterId = PCid
	}

	glog.Infof("volume:[%v] -  spectrum scale volume create params : %v\n", scaleVol.VolName, scaleVol)

	volReqInProcess, err := cs.IfSameVolReqInProcess(scaleVol)
	if err != nil {
		return nil, err
	}

	if volReqInProcess {
		glog.Errorf("volume:[%v] - volume creation already in process ", scaleVol.VolName)
		return nil, status.Error(codes.Aborted, fmt.Sprintf("volume creation already in process : %v", scaleVol.VolName))
	}

	/* Update driver map with new volume. Make sure to defer delete */

	cs.Driver.reqmap[scaleVol.VolName] = volSize
	defer delete(cs.Driver.reqmap, scaleVol.VolName)

	glog.V(4).Infof("reqmap After: %v", cs.Driver.reqmap)

	var targetPath string

	if scaleVol.IsFilesetBased {
		targetPath, err = cs.createFilesetBasedVol(scaleVol)
	} else {
		targetPath, err = cs.createLWVol(scaleVol)
	}

	if err != nil {
		return nil, err
	}

	// Create symbolic link if not present
	err = cs.createSoftlink(scaleVol, targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	volID, err := cs.generateVolID(scaleVol)
	if err != nil {
		glog.Errorf("volume:[%v] - failed to generate volume id. Error: %v", scaleVol.VolName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to generate volume id. Error: %v", err))
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volID,
			CapacityBytes: int64(scaleVol.VolSize),
			VolumeContext: req.GetParameters(),
		},
	}, nil
}

func (cs *ScaleControllerServer) GetVolIdMembers(vId string) (scaleVolId, error) {
	splitVid := strings.Split(vId, ";")
	var vIdMem scaleVolId

	if len(splitVid) == 3 {
		/* This is LW volume */
		/* <cluster_id>;<filesystem_uuid>;path=<symlink_path> */
		vIdMem.ClusterId = splitVid[0]
		vIdMem.FsUUID = splitVid[1]
		SlnkPart := splitVid[2]
		slnkSplit := strings.Split(SlnkPart, "=")
		if len(slnkSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vId))
		}
		vIdMem.SymLnkPath = slnkSplit[1]
		vIdMem.IsFilesetBased = false
		return vIdMem, nil
	}

	if len(splitVid) == 4 {
		/* This is fileset Based volume */
		/* <cluster_id>;<filesystem_uuid>;fileset=<fileset_id>; path=<symlink_path> */
		vIdMem.ClusterId = splitVid[0]
		vIdMem.FsUUID = splitVid[1]
		fileSetPart := splitVid[2]
		fileSetSplit := strings.Split(fileSetPart, "=")
		if len(fileSetSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vId))
		}
		vIdMem.FsetId = fileSetSplit[1]
		SlnkPart := splitVid[3]
		slnkSplit := strings.Split(SlnkPart, "=")
		if len(slnkSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vId))
		}
		vIdMem.SymLnkPath = slnkSplit[1]
		vIdMem.IsFilesetBased = true
		return vIdMem, nil
	}

	return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vId))
}

func (cs *ScaleControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Warningf("invalid delete volume req: %v", req)
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("invalid delete volume req (%v): %v", req, err))
	}
	// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeID := req.GetVolumeId()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "volume Id is missing")
	}

	volumeIdMembers, err := cs.GetVolIdMembers(volumeID)
	if err != nil {
		return &csi.DeleteVolumeResponse{}, nil
	}
	glog.V(4).Infof("volume Id Members [%v]", volumeIdMembers)

	conn, err := cs.getConnFromClusterID(volumeIdMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
	if !isprimaryConnPresent {
		return nil, status.Error(codes.Internal, "unable to get connector for Primary cluster")
	}

	/* FsUUID in volumeIdMembers will be of Primary cluster. So lets get Name of it
	   from Primary cluster */
	FilesystemName, err := primaryConn.GetFilesystemName(volumeIdMembers.FsUUID)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", volumeIdMembers.FsUUID, volumeIdMembers.ClusterId, err))
	}

	mountInfo, err := primaryConn.GetFilesystemMountDetails(FilesystemName)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get mount info for FS [%v] in primary cluster", FilesystemName))
	}

	remoteDeviceName := mountInfo.RemoteDeviceName
	splitDevName := strings.Split(remoteDeviceName, ":")
	remDevFs := splitDevName[len(splitDevName)-1]

	FilesystemName = remDevFs

	sLinkRelPath := strings.Replace(volumeIdMembers.SymLnkPath, cs.Driver.primary.PrimaryFSMount, "", 1)
	sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

	if volumeIdMembers.IsFilesetBased {
		FilesetName, err := conn.GetFileSetNameFromId(FilesystemName, volumeIdMembers.FsetId)

		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get Fileset Name for Id [%v] FS [%v] ClusterId [%v]", volumeIdMembers.FsetId, FilesystemName, volumeIdMembers.ClusterId))
		}

		if FilesetName != "" {
			/* Confirm it is same fileset which was created for this PV */
			pvName := filepath.Base(sLinkRelPath)
			if pvName == FilesetName {
				//Check if fileset exist has any snapshot
				snapshotList, err := conn.ListFilesetSnapshot(FilesystemName, FilesetName)
				if err != nil {
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshot for fileset [%v]. Error: [%v]", FilesetName, err))
				}

				if len(snapshotList) > 0 {
					return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume fileset [%v] contains one or more snapshot, delete snapshot/volumesnapshot", FilesetName))
				}
				glog.V(4).Infof("there is no snapshot present in the fileset [%v], continue DeleteVolume", FilesetName)

				err = conn.DeleteFileset(FilesystemName, FilesetName)
				if err != nil {
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Fileset [%v] for FS [%v] and clusterId [%v].Error : [%v]", FilesetName, FilesystemName, volumeIdMembers.ClusterId, err))
				}
			} else {
				glog.Infof("PV name from path [%v] does not match with filesetName [%v]. Skipping delete of fileset", pvName, FilesetName)
			}
		}
	} else {
		/* Delete Dir for Lw volume */
		err = primaryConn.DeleteDirectory(cs.Driver.primary.GetPrimaryFs(), sLinkRelPath)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Dir using FS [%v] Relative SymLink [%v]. Error [%v]", cs.Driver.primary.GetPrimaryFs(), sLinkRelPath, err))
		}
	}

	err = primaryConn.DeleteSymLnk(cs.Driver.primary.GetPrimaryFs(), sLinkRelPath)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete symlnk [%v:%v] Error [%v]", cs.Driver.primary.GetPrimaryFs(), sLinkRelPath, err))
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerGetCapabilities implements the default GRPC callout.
func (cs *ScaleControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	glog.V(4).Infof("ControllerGetCapabilities called with req: %#v", req)
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *ScaleControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeID := req.GetVolumeId()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "VolumeID not present")
	}

	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No volume capability specified")
	}

	for _, cap := range req.VolumeCapabilities {
		if cap.GetAccessMode().GetMode() != csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER {
			return &csi.ValidateVolumeCapabilitiesResponse{Message: ""}, nil
		}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.VolumeCapabilities,
		},
	}, nil
}

func (cs *ScaleControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	glog.V(3).Infof("controllerserver ControllerUnpublishVolume")
	glog.V(4).Infof("ControllerUnpublishVolume : req %#v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); err != nil {
		glog.V(3).Infof("invalid Unpublish volume request: %v", req)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerUnpublishVolume: ValidateControllerServiceRequest failed: %v", err))
	}

	volumeID := req.GetVolumeId()

	/* <cluster_id>;<filesystem_uuid>;path=<symlink_path> */
	splitVolID := strings.Split(volumeID, ";")
	if len(splitVolID) < 3 {
		return nil, status.Error(codes.InvalidArgument, "ControllerUnpublishVolume VolumeID is not in proper format")
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (cs *ScaleControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) { //nolint:gocyclo,funlen
	glog.V(3).Infof("controllerserver ControllerPublishVolume")
	glog.V(4).Infof("ControllerPublishVolume : req %#v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); err != nil {
		glog.V(3).Infof("invalid Publish volume request: %v", req)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume: ValidateControllerServiceRequest failed: %v", err))
	}

	//ControllerPublishVolumeRequest{VolumeId:"934225357755027944;09762E35:5D26932A;path=/ibm/gpfs0/volume1", NodeId:"node4", VolumeCapability:(*csi.VolumeCapability)(0xc00005d6c0), Readonly:false, Secrets:map[string]string(nil), VolumeContext:map[string]string(nil), XXX_NoUnkeyedLiteral:struct {}{}, XXX_unrecognized:[]uint8(nil), XXX_sizecache:0}

	nodeID := req.GetNodeId()

	if nodeID == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeID not present")
	}

	volumeID := req.GetVolumeId()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume : VolumeID is not present")
	}

	var isFsMounted bool

	/* VolumeID format : <cluster_id>;<filesystem_uuid>;path=<symlink_path> */
	//Assumption : filesystem_uuid is always from local/primary cluster.
	splitVolID := strings.Split(volumeID, ";")
	if len(splitVolID) < 3 {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume : VolumeID is not in proper format")
	}
	filesystemID := splitVolID[1]

	// if SKIP_MOUNT_UNMOUNT == "yes" then mount/unmount will not be invoked
	skipMountUnmount := utils.GetEnv("SKIP_MOUNT_UNMOUNT", yes)
	glog.V(4).Infof("ControllerPublishVolume : SKIP_MOUNT_UNMOUNT is set to %s", skipMountUnmount)

	//Get filesystem name from UUID
	fsName, err := cs.Driver.connmap["primary"].GetFilesystemName(filesystemID)
	if err != nil {
		glog.Errorf("ControllerPublishVolume : Error in getting filesystem Name for filesystem ID of %s.", filesystemID)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem Name for filesystem ID of %s. Error [%v]", filesystemID, err))
	}

	//Check if primary filesystem is mounted.
	primaryfsName := cs.Driver.primary.GetPrimaryFs()
	pfsMount, err := cs.Driver.connmap["primary"].GetFilesystemMountDetails(primaryfsName)
	if err != nil {
		glog.Errorf("ControllerPublishVolume : Error in getting filesystem mount details for %s", primaryfsName)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem mount details for %s. Error [%v]", primaryfsName, err))
	}

	// Node mapping check
	scalenodeID := utils.GetEnv(nodeID, notFound)
	// Additional node mapping check in case of k8s node id start with number.
	if scalenodeID == notFound {
		prefix := utils.GetEnv("SCALE_NODE_MAPPING_PREFIX", "K8sNodePrefix_")
		scalenodeID = utils.GetEnv(prefix+nodeID, notFound)
		if scalenodeID == notFound {
			glog.V(4).Infof("ControllerPublishVolume : scale node mapping not found for %s using %s", prefix+nodeID, nodeID)
			scalenodeID = nodeID
		}
	}

	glog.V(4).Infof("ControllerUnpublishVolume : scalenodeID:%s --known as-- k8snodeName: %s", scalenodeID, nodeID)
	ispFsMounted := utils.StringInSlice(scalenodeID, pfsMount.NodesMounted)

	glog.V(4).Infof("ControllerPublishVolume : Primary FS is mounted on %v", pfsMount.NodesMounted)
	glog.V(4).Infof("ControllerPublishVolume : Primary Fileystem is %s and Volume is from Filesystem %s", primaryfsName, fsName)
	// Skip if primary filesystem and volume filesystem is same
	if primaryfsName != fsName {
		//Check if filesystem is mounted
		fsMount, err := cs.Driver.connmap["primary"].GetFilesystemMountDetails(fsName)
		if err != nil {
			glog.Errorf("ControllerPublishVolume : Error in getting filesystem mount details for %s", fsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem mount details for %s. Error [%v]", fsName, err))
		}
		isFsMounted = utils.StringInSlice(scalenodeID, fsMount.NodesMounted)
		glog.V(4).Infof("ControllerPublishVolume : Volume Source FS is mounted on %v", fsMount.NodesMounted)
	} else {
		isFsMounted = ispFsMounted
	}

	glog.V(4).Infof("ControllerPublishVolume : Mount Status Primaryfs [ %t ], Sourcefs [ %t ]", ispFsMounted, isFsMounted)

	if isFsMounted && ispFsMounted {
		glog.V(4).Infof("ControllerPublishVolume : %s and %s are mounted on %s so returning success", fsName, primaryfsName, scalenodeID)
		return &csi.ControllerPublishVolumeResponse{}, nil
	}

	if skipMountUnmount == "yes" && (!isFsMounted || !ispFsMounted) {
		glog.Errorf("ControllerPublishVolume : SKIP_MOUNT_UNMOUNT == yes and either %s or %s is not mounted on node %s", primaryfsName, fsName, scalenodeID)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : SKIP_MOUNT_UNMOUNT == yes and either %s or %s is not mounted on node %s.", primaryfsName, fsName, scalenodeID))
	}

	//mount the primary filesystem if not mounted
	if !(ispFsMounted) && skipMountUnmount == no {
		glog.V(4).Infof("ControllerPublishVolume : mounting Filesystem %s on %s", primaryfsName, scalenodeID)
		err = cs.Driver.connmap["primary"].MountFilesystem(primaryfsName, scalenodeID)
		if err != nil {
			glog.Errorf("ControllerPublishVolume : Error in mounting filesystem %s on node %s", primaryfsName, scalenodeID)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume :  Error in mounting filesystem %s on node %s. Error [%v]", primaryfsName, scalenodeID, err))
		}
	}

	//mount the volume filesystem if mounted
	if !(isFsMounted) && skipMountUnmount == no && primaryfsName != fsName {
		glog.V(4).Infof("ControllerPublishVolume : mounting %s on %s", fsName, scalenodeID)
		err = cs.Driver.connmap["primary"].MountFilesystem(fsName, scalenodeID)
		if err != nil {
			glog.Errorf("ControllerPublishVolume : Error in mounting filesystem %s on node %s", fsName, scalenodeID)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in mounting filesystem %s on node %s. Error [%v]", fsName, scalenodeID, err))
		}
	}
	return &csi.ControllerPublishVolumeResponse{}, nil
}

//CreateSnapshot Create Snapshot
func (cs *ScaleControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	glog.V(3).Infof("CreateSnapshot - create snapshot req: %v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		glog.V(3).Infof("CreateSnapshot - invalid create snapshot req: %v", req)
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot ValidateControllerServiceRequest failed: %v", err))
	}

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "CreateSnapshot - Request cannot be empty")
	}

	volID := req.GetSourceVolumeId()
	if volID == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateSnapshot - Source Volume ID is a required field")
	}

	volumeIDMembers, err := cs.GetVolIdMembers(volID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - Error in source Volume ID %v: %v", volID, err))
	}

	if !volumeIDMembers.IsFilesetBased {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - Source Volume %v is not fileset based", volID))
	}

	conn, err := cs.getConnFromClusterID(volumeIDMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	filesystemName, err := conn.GetFilesystemName(volumeIDMembers.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
	}

	glog.V(5).Infof("CreateSnapshot - getting filesystem Name from Filesystem Uid [%s] ", volumeIDMembers.FsetId)
	filesetResp, err := conn.GetFileSetResponseFromId(filesystemName, volumeIDMembers.FsetId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get Fileset Name for Fileset Id [%v] FS [%v] ClusterId [%v]", volumeIDMembers.FsetId, filesystemName, volumeIDMembers.ClusterId))
	}

	if filesetResp.Config.ParentId > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - Source Volume %v is not independent fileset based", volID))
	}
	filesetName := filesetResp.FilesetName
	sLinkRelPath := strings.Replace(volumeIDMembers.SymLnkPath, cs.Driver.primary.PrimaryFSMount, "", 1)
	sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

	/* Confirm it is same fileset which was created for this PV */
	pvName := filepath.Base(sLinkRelPath)
	if pvName != filesetName {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - PV name from path [%v] does not match with filesetName [%v].", pvName, filesetName))
	}

	snapName := req.GetName()

	glog.V(5).Infof("CreateSnapshot - check if snapshot [%s] exist in fileset [%s] under filesystem [%s]", snapName, filesetName, filesystemName)
	snapExist, err := conn.CheckIfSnapshotExist(filesystemName, filesetName, snapName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get the snapshot details. Error [%v]", err))
	}

	if !snapExist {
		glog.V(5).Infof("CreateSnapshot - creating snapshot [%s] in fileset [%s] under filesystem [%s]", snapName, filesetName, filesystemName)
		snaperr := conn.CreateSnapshot(filesystemName, filesetName, snapName)
		if snaperr != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to create snapshot. Error [%v]", snaperr))
		}
	}

	//clusterId;FSUUID;filesetName;snapshotName
	snapID := fmt.Sprintf("%s;%s;%s;%s", volumeIDMembers.ClusterId, volumeIDMembers.FsUUID, filesetName, snapName)
	glog.V(5).Infof("CreateSnapshot - Snapshot ID: %s ", snapID)

	glog.V(5).Infof("CreateSnapshot - get timestamp of snapshot [%s] in fileset [%s] under filesystem [%s]", snapName, filesetName, filesystemName)
	createTS, err := conn.GetSnapshotCreateTimestamp(filesystemName, filesetName, snapName)
	if err != nil {
		return nil, err
	}

	const longForm = "2006-01-02 15:04:05,000" // This is the format in which REST API return creation timestamp
	//nolint::staticcheck
	t, _ := time.Parse(longForm, createTS)
	var timestamp timestamp.Timestamp
	timestamp.Seconds = t.Unix()
	timestamp.Nanos = 0

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     snapID,
			SourceVolumeId: volID,
			ReadyToUse:     true,
			CreationTime:   &timestamp,
		},
	}, nil
}

// DeleteSnapshot - Delete snapshot
func (cs *ScaleControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	glog.V(3).Infof("DeleteSnapshot - delete snapshot req: %v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		glog.V(3).Infof("DeleteSnapshot - invalid create snapshot req: %v", req)
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - ValidateControllerServiceRequest failed: %v", err))
	}

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "DeleteSnapshot - request cannot be empty")
	}
	snapID := req.GetSnapshotId()

	if snapID == "" {
		return nil, status.Error(codes.InvalidArgument, "DeleteSnapshot - snapshot Id is a required field")
	}

	splitSid := strings.Split(snapID, ";")
	if len(splitSid) != 4 {
		return nil, status.Error(codes.InvalidArgument, "DeleteSnapshot - snapshot Id is not in a valid format")
	}

	clusterID := splitSid[0]
	fsUUID := splitSid[1]
	filesetName := splitSid[2]
	snapshotName := splitSid[3]

	conn, err := cs.getConnFromClusterID(clusterID)
	if err != nil {
		return nil, err
	}

	glog.V(5).Infof("DeleteSnapshot - getting filesystem Name from Filesystem Uid [%s] ", fsUUID)
	filesystemName, err := conn.GetFilesystemName(fsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get filesystem Name for Filesystem UID [%v] and clusterId [%v]. Error [%v]", fsUUID, clusterID, err))
	}

	glog.V(5).Infof("DeleteSnapshot - check if fileset [%s] exist in filesystem [%s]", filesetName, filesystemName)
	filesetExist, err := conn.CheckIfFilesetExist(filesystemName, filesetName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the fileset %s details details. Error [%v]", filesetName, err))
	}

	//skip delete if fileset not exist, return success
	if filesetExist {
		glog.V(5).Infof("DeleteSnapshot - check if snapshot [%s] exist in fileset [%s] under filesystem [%s]", snapshotName, filesetName, filesystemName)
		snapExist, err := conn.CheckIfSnapshotExist(filesystemName, filesetName, snapshotName)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the snapshot details. Error [%v]", err))
		}

		// skip delete snapshot if not exist, return success
		if snapExist {
			glog.V(5).Infof("DeleteSnapshot - deleting snapshot [%s] from fileset [%s] under filesystem [%s]", snapshotName, filesetName, filesystemName)
			snaperr := conn.DeleteSnapshot(filesystemName, filesetName, snapshotName)
			if snaperr != nil {
				return nil, err
			}
		}
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

func (cs *ScaleControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ScaleControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *ScaleControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *ScaleControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
