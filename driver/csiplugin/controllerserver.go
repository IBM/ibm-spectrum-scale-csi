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
	no                          = "no"
	yes                         = "yes"
	notFound                    = "NOT_FOUND"
	filesystemTypeRemote        = "remote"
	filesystemMounted           = "mounted"
	filesetUnlinkedPath         = "--"
	oneGB                uint64 = 1024 * 1024 * 1024
	smallestVolSize      uint64 = oneGB // 1GB

)

type ScaleControllerServer struct {
	Driver *ScaleDriver
}

func (cs *ScaleControllerServer) IfSameVolReqInProcess(scVol *scaleVolume) (bool, error) {
	cap, volpresent := cs.Driver.reqmap[scVol.VolName]
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

//createLWVol: Create lightweight volume - return relative path of directory created
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

//generateVolID: Generate volume ID
func (cs *ScaleControllerServer) generateVolID(scVol *scaleVolume, uid string) string {
	glog.V(4).Infof("volume: [%v] - ControllerServer:generateVolId", scVol.VolName)
	var volID string

	if scVol.IsFilesetBased {
		/* <cluster_id>;<filesystem_uuid>;fileset=<fileset_id>; path=<symlink_path> */
		slink := fmt.Sprintf("%s/%s", scVol.PrimarySLnkPath, scVol.VolName)
		volID = fmt.Sprintf("%s;%s;filesetName=%s;path=%s", scVol.ClusterId, uid, scVol.VolName, slink)
	} else {
		/* <cluster_id>;<filesystem_uuid>;path=<symlink_path> */
		slink := fmt.Sprintf("%s/%s", scVol.PrimarySLnkPath, scVol.VolName)
		volID = fmt.Sprintf("%s;%s;path=%s", scVol.ClusterId, uid, slink)
	}
	return volID
}

//getTargetPath: retrun relative volume path from filesystem mount point
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

//createDirectory: Create directory if not present
func (cs *ScaleControllerServer) createDirectory(scVol *scaleVolume, targetPath string) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:createDirectory", scVol.VolName)
	dirExists, err := scVol.Connector.CheckIfFileDirPresent(scVol.VolBackendFs, targetPath)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", scVol.VolName, targetPath, scVol.VolBackendFs, err)
		return fmt.Errorf("unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
	}

	if !dirExists {
		if scVol.VolPermissions != "" {
			err = scVol.Connector.MakeDirectoryV2(scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid, scVol.VolPermissions)
			if err != nil {
				// Directory creation failed, no cleanup will retry in next retry
				glog.Errorf("volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", scVol.VolName, targetPath, scVol.VolBackendFs, err)
				return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
			}
		} else {
			err = scVol.Connector.MakeDirectory(scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid)
			if err != nil {
				// Directory creation failed, no cleanup will retry in next retry
				glog.Errorf("volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", scVol.VolName, targetPath, scVol.VolBackendFs, err)
				return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
			}
		}
	}
	return nil
}

//createSoftlink: Create soft link if not present
func (cs *ScaleControllerServer) createSoftlink(scVol *scaleVolume, target string) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:createSoftlink", scVol.VolName)
	volSlnkPath := fmt.Sprintf("%s/%s", scVol.PrimarySLnkRelPath, scVol.VolName)
	symLinkExists, err := scVol.PrimaryConnector.CheckIfFileDirPresent(scVol.PrimaryFS, volSlnkPath)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to check if symlink path [%v] exists in filesystem [%v]. Error: %v", scVol.VolName, volSlnkPath, scVol.PrimaryFS, err)
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

//setQuota: Set quota if not set
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
			return fmt.Errorf("unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err)
		}
	}

	if filesetQuotaBytes < scVol.VolSize && filesetQuotaBytes != 0 {
		// quota does not match and it is not 0 - It might not be fileset created by us
		return fmt.Errorf("fileset %v present but quota %v does not match with requested size %v", scVol.VolName, filesetQuotaBytes, scVol.VolSize)
	}

	if filesetQuotaBytes == 0 {
		volsiz := strconv.FormatUint(scVol.VolSize, 10)
		err = scVol.Connector.SetFilesetQuota(scVol.VolBackendFs, scVol.VolName, volsiz)
		if err != nil {
			// failed to set quota, no cleanup, next retry might be able to set quota
			return fmt.Errorf("unable to set quota [%v] on fileset [%v] of FS [%v]", scVol.VolSize, scVol.VolName, scVol.VolBackendFs)
		}
	}
	return nil
}

//createFilesetBasedVol: Create fileset based volume  - return relative path of volume created
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
		if fsDetails.Mount.Status != filesystemMounted {
			glog.Errorf("volume:[%v] -  filesystem [%v] is [%v] on remote GUI of cluster [%v]", scVol.VolName, scVol.VolBackendFs, fsDetails.Mount.Status, scVol.ClusterId)
			return "", status.Error(codes.Internal, fmt.Sprintf("Filesystem %v in cluster %v is not mounted", scVol.VolBackendFs, scVol.ClusterId))
		}
		glog.V(4).Infof("volume:[%v] - mount point of volume filesystem [%v] on owning cluster is %v", scVol.VolName, scVol.VolBackendFs, fsDetails.Mount.MountPoint)
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
	} else {
		var inodeLimit uint64
		if scVol.VolSize > 10*oneGB {
			inodeLimit = 200000
		} else {
			inodeLimit = 100000
		}
		opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(inodeLimit, 10)
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
		var junctionPath string
		junctionPath = fmt.Sprintf("%s/%s", fsDetails.Mount.MountPoint, scVol.VolName)

		if scVol.ParentFileset != "" {
			parentfilesetInfo, err := scVol.Connector.ListFileset(scVol.VolBackendFs, scVol.ParentFileset)
			if err != nil {
				glog.Errorf("volume:[%v] - unable to get details of parent fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.ParentFileset, scVol.VolBackendFs, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - unable to get details of parent fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.ParentFileset, scVol.VolBackendFs, err))
			}
			if (parentfilesetInfo.Config.Path == "") || (parentfilesetInfo.Config.Path == filesetUnlinkedPath) {
				glog.Errorf("volume:[%v] - parent fileset [%v] is not linked", scVol.VolName, scVol.ParentFileset)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - parent fileset [%v] is not linked", scVol.VolName, scVol.ParentFileset))
			}
			junctionPath = fmt.Sprintf("%s/%s", parentfilesetInfo.Config.Path, scVol.VolName)
		}

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
	glog.Errorf("unable to get connector for cluster ID %v", cid)
	return nil, status.Error(codes.Internal, fmt.Sprintf("unable to find cluster [%v] details in custom resource", cid))
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
	if scaleVol.IsFilesetBased && uint64(volSize) < smallestVolSize {
		scaleVol.VolSize = smallestVolSize
	} else {
		scaleVol.VolSize = uint64(volSize)
	}

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

	volSrc := req.GetVolumeContentSource()
	isSnapSource := false
	snapIdMembers := scaleSnapId{}
	if volSrc != nil {
		if volSrc.GetVolume() != nil {
			return nil, status.Error(codes.Unimplemented, "Volume as volume content source is not supported")
		}

		srcSnap := volSrc.GetSnapshot()
		if srcSnap != nil {
			snapId := srcSnap.GetSnapshotId()
			snapIdMembers, err = cs.GetSnapIdMembers(snapId)
			if err != nil {
				glog.Errorf("volume:[%v] - Invalid snapshot ID %s [%v]", volName, snapId, err)
				return nil, err
			}
			isSnapSource = true
		}
	}

	// Check if Primary Fileset is linked
	primaryFileset := cs.Driver.primary.PrimaryFset
	glog.V(5).Infof("volume:[%v] - check if primary fileset [%v] is linked", scaleVol.VolName, primaryFileset)
	isPrimaryFilesetLinked, err := scaleVol.PrimaryConnector.IsFilesetLinked(scaleVol.PrimaryFS, primaryFileset)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to get details of Primary Fileset [%v]. Error : [%v]", scaleVol.VolName, primaryFileset, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get details of Primary Fileset [%v]. Error : [%v]", primaryFileset, err))
	}
	if !isPrimaryFilesetLinked {
		glog.Errorf("volume:[%v] - primary fileset [%v] is not linked", scaleVol.VolName, primaryFileset)
		return nil, status.Error(codes.Internal, fmt.Sprintf("primary fileset [%v] is not linked", primaryFileset))
	}

	if scaleVol.PrimaryFS != scaleVol.VolBackendFs {
		// primary filesytem must be mounted on GUI node so that we can create the softlink
		// skip if primary and volume filesystem is same
		glog.V(4).Infof("volume:[%v] - check if primary filesystem [%v] is mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.PrimaryFS)
		isPfsMounted, err := scaleVol.PrimaryConnector.IsFilesystemMountedOnGUINode(scaleVol.PrimaryFS)
		if err != nil {
			glog.Errorf("volume:[%v] - unable to get filesystem mount details for %s on Primary cluster. Error: %v", scaleVol.VolName, scaleVol.PrimaryFS, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem mount details for %s on Primary cluster. Error: %v", scaleVol.PrimaryFS, err))
		}
		if !isPfsMounted {
			glog.Errorf("volume:[%v] - primary filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.PrimaryFS)
			return nil, status.Error(codes.Internal, fmt.Sprintf("primary filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.PrimaryFS))
		}
	}

	glog.V(5).Infof("volume:[%v] - check if volume filesystem [%v] is mounted on GUI node of Primary cluster", scaleVol.VolName, scaleVol.VolBackendFs)
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

	glog.V(5).Infof("volume:[%v] - mount point of volume filesystem [%v] is on Primary cluster is %v", scaleVol.VolName, scaleVol.VolBackendFs, volFsInfo.Mount.MountPoint)

	/* scaleVol.VolBackendFs will always be local cluster FS. So we need to find a
	   remote cluster FS in case local cluster FS is remotely mounted. We will find local FS RemoteDeviceName on local cluster, will use that as VolBackendFs and	create fileset on that FS. */

	if scaleVol.IsFilesetBased {
		remoteDeviceName := volFsInfo.Mount.RemoteDeviceName
		scaleVol.LocalFS = scaleVol.VolBackendFs
		scaleVol.VolBackendFs = getRemoteFsName(remoteDeviceName)
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

	if isSnapSource {
		err = cs.validateSnapId(&snapIdMembers, scaleVol, PCid)
		if err != nil {
			glog.Errorf("volume:[%v] - Error in source snapshot validation [%v]", volName, err)
			return nil, err
		}
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

	jobDetails := cs.Driver.snapjobstatusmap[scaleVol.VolName]
	if jobDetails.jobStatus == SNAP_JOB_RUNNING {
		glog.Errorf("volume:[%v] -  snapshot copy request in progress for snapshot: %s.", scaleVol.VolName, snapIdMembers.SnapName)
		return nil, status.Error(codes.Aborted, fmt.Sprintf("snapshot copy request in progress for snapshot: %s", snapIdMembers.SnapName))
	} else if jobDetails.jobStatus == SNAP_JOB_FAILED {
		glog.Errorf("volume:[%v] -  snapshot copy job had failed for snapshot %s", scaleVol.VolName, snapIdMembers.SnapName)
		return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot copy job had failed for snapshot: %s", snapIdMembers.SnapName))
	} else if jobDetails.jobStatus == SNAP_JOB_COMPLETED {
		glog.Infof("volume:[%v] -  snapshot copy request has already completed successfully for snapshot: %s", scaleVol.VolName, snapIdMembers.SnapName)
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				VolumeId:      jobDetails.volID,
				CapacityBytes: int64(scaleVol.VolSize),
				VolumeContext: req.GetParameters(),
				ContentSource: volSrc,
			},
		}, nil
	}

	if scaleVol.VolPermissions != "" {
		versionCheck, err := cs.checkMinScaleVersion(scaleVol.Connector, "5112")
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("the minimum Spectrum Scale version check for permissions failed with error %s", err))
		}
		if !versionCheck {
			return nil, status.Error(codes.Internal, "the minimum required Spectrum Scale version for permissions support with CSI is 5.1.1-2")
		}
	}

	/* Update driver map with new volume. Make sure to defer delete */

	cs.Driver.reqmap[scaleVol.VolName] = int64(scaleVol.VolSize)
	defer delete(cs.Driver.reqmap, scaleVol.VolName)

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

	volID := cs.generateVolID(scaleVol, volFsInfo.UUID)

	if isSnapSource {
		err = cs.copySnapContent(scaleVol, snapIdMembers, volFsInfo, targetPath, volID)
		if err != nil {
			glog.Errorf("createVolume failed while copying snapshot content [%s]: [%v]", volName, err)
			return nil, err
		}
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volID,
			CapacityBytes: int64(scaleVol.VolSize),
			VolumeContext: req.GetParameters(),
			ContentSource: volSrc,
		},
	}, nil
}

func (cs *ScaleControllerServer) copySnapContent(scVol *scaleVolume, snapId scaleSnapId, fsDetails connectors.FileSystem_v2, targetPath string, volID string) error {
	glog.V(3).Infof("copySnapContent snapId: [%v], scaleVolume: [%v]", snapId, scVol)
	conn, err := cs.getConnFromClusterID(snapId.ClusterId)
	if err != nil {
		return err
	}

	//err = cs.validateRemoteFs(fsDetails, scVol)
	//if err != nil {
	//	return err
	//}

	targetFsName, err := conn.GetFilesystemName(fsDetails.UUID)
	if err != nil {
		return err
	}

	targetFsDetails, err := conn.GetFilesystemDetails(targetFsName)
	if err != nil {
		return err
	}

	fsMntPt := targetFsDetails.Mount.MountPoint
	targetPath = fmt.Sprintf("%s/%s", fsMntPt, targetPath)

	jobStatus, jobID, err := conn.CopyFsetSnapshotPath(snapId.FsName, snapId.FsetName, snapId.SnapName, snapId.Path, targetPath, scVol.NodeClass)
	if err != nil {
		glog.Errorf("failed to create volume from snapshot %s: [%v]", snapId.SnapName, err)
		return status.Error(codes.Internal, fmt.Sprintf("failed to create volume from snapshot %s: [%v]", snapId.SnapName, err))

	}

	jobDetails := SnapCopyJobDetails{SNAP_JOB_RUNNING, volID}
	cs.Driver.snapjobstatusmap[scVol.VolName] = jobDetails

	err = conn.WaitForSnapshotCopy(jobStatus, jobID)
	if err != nil {
		glog.Errorf("unable to copy snapshot %s: %v.", snapId.SnapName, err)
		if strings.Contains(err.Error(), "EFSSG0632C") {
			// EFSSG0632C = Command execution aborted
			// Store SNAP_JOB_NOT_STARTED in snapjobstatusmap if error was due to same mmxcp in progress
			// or max no. of mmxcp already running. In these cases we want to retry again
			// in the next k8s rety cycle
			jobDetails.jobStatus = SNAP_JOB_NOT_STARTED
		} else {
			jobDetails.jobStatus = SNAP_JOB_FAILED
		}
		cs.Driver.snapjobstatusmap[scVol.VolName] = jobDetails
		return err
	}

	glog.Infof("copy snapshot completed for snapId: [%v], scaleVolume: [%v]", snapId, scVol)
	jobDetails.jobStatus = SNAP_JOB_COMPLETED
	cs.Driver.snapjobstatusmap[scVol.VolName] = jobDetails
	//delete(cs.Driver.snapjobmap, scVol.VolName)
	return nil
}

func (cs *ScaleControllerServer) checkMinScaleVersion(conn connectors.SpectrumScaleConnector, version string) (bool, error) {
	scaleVersion, err := conn.GetScaleVersion()
	if err != nil {
		return false, err
	}
	/* Assuming Spectrum Scale version is in a format like 5.0.0-0_170818.165000 */
	// "serverVersion" : "5.1.1.1-developer build",
	splitScaleVer := strings.Split(scaleVersion, ".")
	if len(splitScaleVer) < 3 {
		return false, status.Error(codes.Internal, fmt.Sprintf("invalid Spectrum Scale version - %s", scaleVersion))
	}
	splitMinorVer := strings.Split(splitScaleVer[2], "-")
	assembledScaleVer := splitScaleVer[0] + splitScaleVer[1] + splitMinorVer[0] + splitMinorVer[1][0:1]

	if assembledScaleVer < version {
		return false, nil
	}
	return true, nil
}

func (cs *ScaleControllerServer) checkSnapshotSupport(conn connectors.SpectrumScaleConnector) error {
	/* Verify Spectrum Scale Version is not below 5.1.1-0 */
	versionCheck, err := cs.checkMinScaleVersion(conn, "5110")
	if err != nil {
		return err
	}

	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required Spectrum Scale version for snapshot support with CSI is 5.1.1-0")
	}
	return nil
}

func (cs *ScaleControllerServer) validateSnapId(sId *scaleSnapId, scVol *scaleVolume, pCid string) error {
	glog.V(3).Infof("validateSnapId [%v]", sId)
	conn, err := cs.getConnFromClusterID(sId.ClusterId)
	if err != nil {
		return err
	}

	/* Check if Spectrum Scale supports Snapshot */
	chkSnapshotErr := cs.checkSnapshotSupport(conn)
	if chkSnapshotErr != nil {
		return chkSnapshotErr
	}

	if scVol.IsFilesetBased {
		if scVol.ClusterId != "" && sId.ClusterId != scVol.ClusterId {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot create volume from a source snapshot from another cluster. Volume is being created in cluster %s, source snapshot is from cluster %s.", scVol.ClusterId, sId.ClusterId))
		}

		if scVol.ClusterId == "" && sId.ClusterId != pCid {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot create volume from a source snapshot from another cluster. Volume is being created in cluster %s, source snapshot is from cluster %s.", pCid, sId.ClusterId))
		}
	}

	if scVol.NodeClass != "" {
		isValidNodeclass, err := conn.IsValidNodeclass(scVol.NodeClass)
		if err != nil {
			return err
		}

		if !isValidNodeclass {
			return status.Error(codes.NotFound, fmt.Sprintf("nodeclass [%s] not found on cluster [%v]", scVol.NodeClass, scVol.ClusterId))
		}
	}

	sId.FsName, err = conn.GetFilesystemName(sId.FsUUID)

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", sId.FsUUID, sId.ClusterId, err))
	}

	if sId.FsName != scVol.VolBackendFs {
		isFsMounted, err := conn.IsFilesystemMountedOnGUINode(sId.FsName)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error in getting filesystem mount details for %s", sId.FsName))
		}
		if !isFsMounted {
			return status.Error(codes.Internal, fmt.Sprintf("filesystem %s is not mounted on GUI node", sId.FsName))
		}
	}

	isFsetLinked, err := conn.IsFilesetLinked(sId.FsName, sId.FsetName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get fileset link information for [%v]", sId.FsetName))
	}
	if !isFsetLinked {
		return status.Error(codes.Internal, fmt.Sprintf("fileset [%v] of source snapshot is not linked", sId.FsetName))
	}

	isSnapExist, err := conn.CheckIfSnapshotExist(sId.FsName, sId.FsetName, sId.SnapName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get snapshot information for [%v]", sId.SnapName))
	}
	if !isSnapExist {
		return status.Error(codes.Internal, fmt.Sprintf("snapshot [%v] does not exist for fileset [%v]", sId.SnapName, sId.FsetName))
	}

	return nil
}

func (cs *ScaleControllerServer) GetSnapIdMembers(sId string) (scaleSnapId, error) {
	splitSid := strings.Split(sId, ";")
	var sIdMem scaleSnapId

	if len(splitSid) < 4 {
		return scaleSnapId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Snapshot Id : [%v]", sId))
	}

	/* clusterId;FSUUID;filesetName;snapshotName;path */
	sIdMem.ClusterId = splitSid[0]
	sIdMem.FsUUID = splitSid[1]
	sIdMem.FsetName = splitSid[2]
	sIdMem.SnapName = splitSid[3]
	if len(splitSid) == 5 && splitSid[4] != "" {
		sIdMem.Path = splitSid[4]
	} else {
		sIdMem.Path = "/"
	}

	return sIdMem, nil

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
		/* <cluster_id>;<filesystem_uuid>;fileset=<fileset_id>;path=<symlink_path> */
		vIdMem.ClusterId = splitVid[0]
		vIdMem.FsUUID = splitVid[1]
		fileSetPart := splitVid[2]
		fileSetSplit := strings.Split(fileSetPart, "=")
		if len(fileSetSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vId))
		}

		if fileSetSplit[0] == "filesetName" {
			vIdMem.FsetName = fileSetSplit[1]
		} else {
			vIdMem.FsetId = fileSetSplit[1]
		}

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
	glog.V(3).Infof("DeleteVolume [%v]", req)

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
		return &csi.DeleteVolumeResponse{}, err
	}

	glog.Infof("Volume Id Members [%v]", volumeIdMembers)

	conn, err := cs.getConnFromClusterID(volumeIdMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
	if !isprimaryConnPresent {
		glog.Errorf("unable to get connector for primary cluster")
		return nil, status.Error(codes.Internal, "unable to find primary cluster details in custom resource")
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

	FilesystemName = getRemoteFsName(mountInfo.RemoteDeviceName)

	sLinkRelPath := strings.Replace(volumeIdMembers.SymLnkPath, cs.Driver.primary.PrimaryFSMount, "", 1)
	sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

	if volumeIdMembers.IsFilesetBased {
		var FilesetName string

		if volumeIdMembers.FsetName != "" {
			FilesetName = volumeIdMembers.FsetName
		} else {
			FilesetName, err = conn.GetFileSetNameFromId(FilesystemName, volumeIdMembers.FsetId)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get Fileset Name for Id [%v] FS [%v] ClusterId [%v]", volumeIdMembers.FsetId, FilesystemName, volumeIdMembers.ClusterId))
			}
		}

		if FilesetName != "" {
			/* Confirm it is same fileset which was created for this PV */
			pvName := filepath.Base(sLinkRelPath)
			if pvName == FilesetName {
				//Check if fileset exist has any snapshot
				snapshotList, err := conn.ListFilesetSnapshots(FilesystemName, FilesetName)
				if err != nil {
					if strings.Contains(err.Error(), "EFSSG0072C") ||
						strings.Contains(err.Error(), "400 Invalid value in 'filesetName'") { // fileset is already deleted
						glog.V(4).Infof("fileset seems already deleted - %v", err)
						return &csi.DeleteVolumeResponse{}, nil
					}
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshot for fileset [%v]. Error: [%v]", FilesetName, err))
				}

				if len(snapshotList) > 0 {
					return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume fileset [%v] contains one or more snapshot, delete snapshot/volumesnapshot", FilesetName))
				}
				glog.V(4).Infof("there is no snapshot present in the fileset [%v], continue DeleteVolume", FilesetName)

				err = conn.DeleteFileset(FilesystemName, FilesetName)
				if err != nil {
					if strings.Contains(err.Error(), "EFSSG0072C") ||
						strings.Contains(err.Error(), "400 Invalid value in 'filesetName'") { // fileset is already deleted
						glog.V(4).Infof("fileset seems already deleted - %v", err)
						return &csi.DeleteVolumeResponse{}, nil
					}
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Fileset [%v] for FS [%v] and clusterId [%v].Error : [%v]", FilesetName, FilesystemName, volumeIdMembers.ClusterId, err))
				}
			} else {
				glog.Infof("pv name from path [%v] does not match with filesetName [%v]. Skipping delete of fileset", pvName, FilesetName)
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

	volumePath := splitVolID[2]
	if len(splitVolID) == 4 {
		volumePath = splitVolID[3]
	}
	volumePath = strings.TrimPrefix(volumePath, "path=")

	// if SKIP_MOUNT_UNMOUNT == "yes" then mount/unmount will not be invoked
	skipMountUnmount := utils.GetEnv(SKIP_MOUNT_UNMOUNT, yes)
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
	scalenodeID := getNodeMapping(nodeID)
	glog.V(4).Infof("ControllerUnpublishVolume : scalenodeID:%s --known as-- k8snodeName: %s", scalenodeID, nodeID)

	shortnameNodeMapping := utils.GetEnv(SHORTNAME_NODE_MAPPING, no)
	if shortnameNodeMapping == yes {
		glog.V(4).Infof("ControllerPublishVolume : SHORTNAME_NODE_MAPPING is set to %s", shortnameNodeMapping)
	}

	var ispFsMounted bool
	// NodesMounted has admin node names
	// This means node mapping must be to admin names.
	// Unless shortnameNodeMapping=="yes", then we should check shortname portion matches.
	if shortnameNodeMapping == yes {
		ispFsMounted = shortnameInSlice(scalenodeID, pfsMount.NodesMounted)
	} else {
		ispFsMounted = utils.StringInSlice(scalenodeID, pfsMount.NodesMounted)
	}

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

		if !strings.HasPrefix(volumePath, fsMount.MountPoint) &&
			!strings.HasPrefix(volumePath, pfsMount.MountPoint) {
			glog.Errorf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s or %s", volumePath, primaryfsName, fsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s or %s", volumePath, primaryfsName, fsName))
		}

		// NodesMounted has admin node names
		// This means node mapping must be to admin names.
		// Unless shortnameNodeMapping=="yes", then we should check shortname portion matches.
		if shortnameNodeMapping == yes {
			isFsMounted = shortnameInSlice(scalenodeID, pfsMount.NodesMounted)
		} else {
			isFsMounted = utils.StringInSlice(scalenodeID, pfsMount.NodesMounted)
		}

		glog.V(4).Infof("ControllerPublishVolume : Volume Source FS is mounted on %v", fsMount.NodesMounted)
	} else {
		if !strings.HasPrefix(volumePath, pfsMount.MountPoint) {
			glog.Errorf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s", volumePath, primaryfsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s", volumePath, primaryfsName))
		}

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
func (cs *ScaleControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) { //nolint:gocyclo,funlen
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
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is independent fileset", volID))
	}

	conn, err := cs.getConnFromClusterID(volumeIDMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	/* Check if Spectrum Scale supports Snapshot */
	chkSnapshotErr := cs.checkSnapshotSupport(conn)
	if chkSnapshotErr != nil {
		return nil, chkSnapshotErr
	}

	filesystemName, err := conn.GetFilesystemName(volumeIDMembers.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
	}

	filesetResp := connectors.Fileset_v2{}
	if volumeIDMembers.FsetName != "" {
		filesetResp, err = conn.GetFileSetResponseFromName(filesystemName, volumeIDMembers.FsetName)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get Fileset response for Fileset [%v] FS [%v] ClusterId [%v]", volumeIDMembers.FsetName, filesystemName, volumeIDMembers.ClusterId))
		}
	} else {
		filesetResp, err = conn.GetFileSetResponseFromId(filesystemName, volumeIDMembers.FsetId)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get Fileset response for Fileset Id [%v] FS [%v] ClusterId [%v]", volumeIDMembers.FsetId, filesystemName, volumeIDMembers.ClusterId))
		}
	}

	if filesetResp.Config.ParentId > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is independent fileset", volID))
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

	snapExist, err := conn.CheckIfSnapshotExist(filesystemName, filesetName, snapName)
	if err != nil {
		glog.Errorf("CreateSnapshot [%s] - Unable to get the snapshot details. Error [%v]", snapName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get the snapshot details for [%s]. Error [%v]", snapName, err))
	}

	if !snapExist {
		snapshotList, err := conn.ListFilesetSnapshots(filesystemName, filesetName)
		if err != nil {
			glog.Errorf("CreateSnapshot [%s] - unable to list snapshots for fileset [%s:%s]. Error: [%v]", snapName, filesystemName, filesetName, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshots for fileset [%s:%s]. Error: [%v]", filesystemName, filesetName, err))
		}

		if len(snapshotList) >= 256 {
			glog.Errorf("CreateSnapshot [%s] - max limit of snapshots reached for fileset [%s:%s]. No more snapshots can be created for this fileset.", snapName, filesystemName, filesetName)
			return nil, status.Error(codes.OutOfRange, fmt.Sprintf("max limit of snapshots reached for fileset [%s:%s]. No more snapshots can be created for this fileset.", filesystemName, filesetName))
		}

		snaperr := conn.CreateSnapshot(filesystemName, filesetName, snapName)
		if snaperr != nil {
			glog.Errorf("snapshot [%s] - Unable to create snapshot. Error [%v]", snapName, snaperr)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create snapshot [%s]. Error [%v]", snapName, snaperr))
		}
	}

	snapID := ""
	if filesetResp.Config.Comment == connectors.FilesetComment &&
		(cs.Driver.primary.PrimaryFset != filesetName || cs.Driver.primary.PrimaryFs != filesystemName) {
		// Dynamically created PVC, here path is the xxx-data directory within the fileset where all volume data resides
		//clusterId;FSUUID;filesetName;snapshotName;path
		snapID = fmt.Sprintf("%s;%s;%s;%s;%s-data", volumeIDMembers.ClusterId, volumeIDMembers.FsUUID, filesetName, snapName, filesetName)
	} else {
		// This is statically created PVC from an independent fileset, here path is the root of fileset
		//clusterId;FSUUID;filesetName;snapshotName;/
		snapID = fmt.Sprintf("%s;%s;%s;%s;/", volumeIDMembers.ClusterId, volumeIDMembers.FsUUID, filesetName, snapName)
	}

	timestamp, err := cs.getSnapshotCreateTimestamp(conn, filesystemName, filesetName, snapName)
	if err != nil {
		glog.Errorf("error getting create timestamp for snapshot %s:%s:%s", filesystemName, filesetName, snapName)
		return nil, err
	}

	restoreSize, err := cs.getSnapRestoreSize(conn, filesystemName, filesetName)
	if err != nil {
		glog.Errorf("error getting the snapshot restore size for snapshot %s:%s:%s", filesystemName, filesetName, snapName)
		return nil, err
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     snapID,
			SourceVolumeId: volID,
			ReadyToUse:     true,
			CreationTime:   &timestamp,
			SizeBytes:      restoreSize,
		},
	}, nil
}

func (cs *ScaleControllerServer) getSnapshotCreateTimestamp(conn connectors.SpectrumScaleConnector, fs string, fset string, snap string) (timestamp.Timestamp, error) {
	var timestamp timestamp.Timestamp

	createTS, err := conn.GetSnapshotCreateTimestamp(fs, fset, snap)
	if err != nil {
		glog.Errorf("snapshot [%s] - Unable to get snapshot create timestamp", snap)
		return timestamp, err
	}

	timezoneOffset, err := conn.GetTimeZoneOffset()
	if err != nil {
		glog.Errorf("snapshot [%s] - Unable to get cluster timezone", snap)
		return timestamp, err
	}

	// Rest API returns create timestamp in the format 2006-01-02 15:04:05,000
	// irrespective of the cluster timezone. We replace the last part of this date
	// with the timezone offset returned by cluster config REST API and then parse
	// the timestamp with correct zone info
	const longForm = "2006-01-02 15:04:05-07:00"
	//nolint::staticcheck

	strings.Replace(createTS, ",000", timezoneOffset, 1)
	t, _ := time.Parse(longForm, createTS)
	timestamp.Seconds = t.Unix()
	timestamp.Nanos = 0

	return timestamp, nil
}

func (cs *ScaleControllerServer) getSnapRestoreSize(conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string) (int64, error) {
	quotaResp, err := conn.GetFilesetQuotaDetails(filesystemName, filesetName)

	if err != nil {
		return 0, err
	}

	if quotaResp.BlockLimit < 0 {
		glog.Errorf("getSnapRestoreSize: Invalid block limit [%v] for fileset [%s:%s] found", quotaResp.BlockLimit, filesystemName, filesetName)
		return 0, status.Error(codes.Internal, fmt.Sprintf("invalid block limit [%v] for fileset [%s:%s] found", quotaResp.BlockLimit, filesystemName, filesetName))
	}

	// REST API returns block limit in kb, convert it to bytes and return
	return int64(quotaResp.BlockLimit * 1024), nil
}

// DeleteSnapshot - Delete snapshot
func (cs *ScaleControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	glog.V(3).Infof("DeleteSnapshot - delete snapshot req: %v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		glog.Errorf("DeleteSnapshot - invalid delete snapshot req %v: %v", req, err)
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
	if len(splitSid) < 4 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("DeleteSnapshot - snapshot Id [%s] is not in a valid format", snapID))
	}

	clusterID := splitSid[0]
	fsUUID := splitSid[1]
	filesetName := splitSid[2]
	snapshotName := splitSid[3]

	conn, err := cs.getConnFromClusterID(clusterID)
	if err != nil {
		return nil, err
	}

	filesystemName, err := conn.GetFilesystemName(fsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get filesystem Name for Filesystem UID [%v] and clusterId [%v]. Error [%v]", fsUUID, clusterID, err))
	}

	filesetExist, err := conn.CheckIfFilesetExist(filesystemName, filesetName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the fileset %s details details. Error [%v]", filesetName, err))
	}

	//skip delete if snapshot not exist, return success
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
				glog.Errorf("DeleteSnapshot - error deleting snapshot %v: %v", snapshotName, snaperr)
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
	glog.V(3).Infof("ControllerExpandVolume - Volume expand req: %v", req)

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_EXPAND_VOLUME); err != nil {
		glog.V(3).Infof("ControllerExpandVolume - invalid expand volume req: %v", req)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerExpandVolume ValidateControllerServiceRequest failed: %v", err))
	}

	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	capRange := req.GetCapacityRange()
	if capRange == nil {
		return nil, status.Error(codes.InvalidArgument, "Capacity range not provided")
	}

	capacity := uint64(capRange.GetRequiredBytes())

	volumeIDMembers, err := cs.GetVolIdMembers(volID)
	if err != nil {
		glog.Errorf("ControllerExpandVolume - Error in source Volume ID %v: %v", volID, err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("ControllerExpandVolume - Error in source Volume ID %v: %v", volID, err))
	}

	// For lightweight return volume expanded as no action is required
	if !volumeIDMembers.IsFilesetBased {
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(capacity),
			NodeExpansionRequired: false,
		}, nil
	}

	conn, err := cs.getConnFromClusterID(volumeIDMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	filesystemName, err := conn.GetFilesystemName(volumeIDMembers.FsUUID)
	if err != nil {
		glog.Errorf("ControllerExpandVolume - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerExpandVolume - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
	}

	filesetName := volumeIDMembers.FsetName

	fsetExist, err := conn.CheckIfFilesetExist(filesystemName, filesetName)
	if err != nil {
		glog.Errorf("unable to check fileset [%v] existance in filesystem [%v]. Error [%v]", filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to check fileset [%v] existance in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	if !fsetExist {
		glog.Errorf("fileset [%v] does not exist in filesystem [%v]. Error [%v]", filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("fileset [%v] does not exist in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	quota, err := conn.ListFilesetQuota(filesystemName, filesetName)
	if err != nil {
		glog.Errorf("unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	filesetQuotaBytes, err := ConvertToBytes(quota)
	if err != nil {
		glog.Errorf("unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	if filesetQuotaBytes < capacity {
		volsize := strconv.FormatUint(capacity, 10)
		err = conn.SetFilesetQuota(filesystemName, filesetName, volsize)
		if err != nil {
			glog.Errorf("unable to expand the volume. Error [%v]", err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to expand the volume. Error [%v]", err))
		}
	} else {
		capacity = filesetQuotaBytes
	}

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(capacity),
		NodeExpansionRequired: false,
	}, nil
}

func (cs *ScaleControllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
