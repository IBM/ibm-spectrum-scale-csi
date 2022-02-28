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
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
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
	defaultSnapWindow           = "30"  // default snapWindow for Consistency Group snapshots is 30 minutes

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
	err = cs.createDirectory(scVol, scVol.VolName, dirPath)
	if err != nil {
		glog.Errorf("volume:[%v] - failed to create directory %v. Error : %v", scVol.VolName, dirPath, err)
		return "", status.Error(codes.Internal, err.Error())
	}
	return dirPath, nil
}

//generateVolID: Generate volume ID
//VolID format for all newly created volumes (from 2.5.0 onwards):

//<storageclass_type>;<volume_type>;<cluster_id>;<filesystem_uuid>;<consistency_group>;<fileset_name>;<path>
func (cs *ScaleControllerServer) generateVolID(scVol *scaleVolume, uid string, isNewVolumeType bool, targetPath string) (string, error) {
	glog.V(4).Infof("volume: [%v] - ControllerServer:generateVolId", scVol.VolName)
	var volID string
	var storageClassType string
	var volumeType string

	filesetName := scVol.VolName
	consistencyGroup := ""
	path := ""

	if isNewVolumeType {
		primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
		if !isprimaryConnPresent {
			glog.Errorf("unable to get connector for primary cluster")
			return "", status.Error(codes.Internal, "unable to find primary cluster details in custom resource")
		}
		fsMountPoint, err := primaryConn.GetFilesystemMountDetails(scVol.LocalFS)
		if err != nil {
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to get mount info for FS [%v] in cluster", scVol.LocalFS))
		}
		path = fmt.Sprintf("%s/%s", fsMountPoint.MountPoint, targetPath)
	} else {
		path = fmt.Sprintf("%s/%s", scVol.PrimarySLnkPath, scVol.VolName)
	}
	glog.V(4).Infof("volume: [%v] - ControllerServer:generateVolId: targetPath: [%v]", scVol.VolName, path)

	if isNewVolumeType {
		storageClassType = STORAGECLASS_ADVANCED
		volumeType = FILE_DEPENDENTFILESET_VOLUME
		consistencyGroup = scVol.ConsistencyGroup
	} else {
		storageClassType = STORAGECLASS_CLASSIC
		if scVol.IsFilesetBased {
			if scVol.FilesetType == independentFileset {
				volumeType = FILE_INDEPENDENTFILESET_VOLUME
			} else {
				volumeType = FILE_DEPENDENTFILESET_VOLUME
			}
		} else {
			volumeType = FILE_DIRECTORYBASED_VOLUME
			//filesetName for LW volume is empty
			filesetName = ""
		}
	}

	volID = fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s", storageClassType, volumeType, scVol.ClusterId, uid, consistencyGroup, filesetName, path)
	return volID, nil
}

//getTargetPath: retrun relative volume path from filesystem mount point
func (cs *ScaleControllerServer) getTargetPath(fsetLinkPath, fsMountPoint, volumeName string, createDataDir bool, isNewVolumeType bool) (string, error) {
	if fsetLinkPath == "" || fsMountPoint == "" {
		glog.Errorf("volume:[%v] - missing details to generate target path fileset junctionpath: [%v], filesystem mount point: [%v]", volumeName, fsetLinkPath, fsMountPoint)
		return "", fmt.Errorf("missing details to generate target path fileset junctionpath: [%v], filesystem mount point: [%v]", fsetLinkPath, fsMountPoint)
	}
	glog.V(4).Infof("volume: [%v] - ControllerServer:getTargetPath", volumeName)
	targetPath := strings.Replace(fsetLinkPath, fsMountPoint, "", 1)
	if createDataDir && !isNewVolumeType {
		targetPath = fmt.Sprintf("%s/%s-data", targetPath, volumeName)
	}
	targetPath = strings.Trim(targetPath, "!/")

	return targetPath, nil
}

//createDirectory: Create directory if not present
func (cs *ScaleControllerServer) createDirectory(scVol *scaleVolume, volName string, targetPath string) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:createDirectory", volName)
	dirExists, err := scVol.Connector.CheckIfFileDirPresent(scVol.VolBackendFs, targetPath)
	if err != nil {
		glog.Errorf("volume:[%v] - unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", volName, targetPath, scVol.VolBackendFs, err)
		return fmt.Errorf("unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
	}

	if !dirExists {
		if scVol.VolPermissions != "" {
			err = scVol.Connector.MakeDirectoryV2(scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid, scVol.VolPermissions)
			if err != nil {
				// Directory creation failed, no cleanup will retry in next retry
				glog.Errorf("volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", volName, targetPath, scVol.VolBackendFs, err)
				return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
			}
		} else {
			err = scVol.Connector.MakeDirectory(scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid)
			if err != nil {
				// Directory creation failed, no cleanup will retry in next retry
				glog.Errorf("volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", volName, targetPath, scVol.VolBackendFs, err)
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
func (cs *ScaleControllerServer) setQuota(scVol *scaleVolume, volName string) error {
	glog.V(4).Infof("volume: [%v] - ControllerServer:setQuota", volName)
	quota, err := scVol.Connector.ListFilesetQuota(scVol.VolBackendFs, volName)
	if err != nil {
		return fmt.Errorf("unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", volName, scVol.VolBackendFs, err)
	}

	filesetQuotaBytes, err := ConvertToBytes(quota)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid number specified") {
			// Invalid number specified means quota is not set
			filesetQuotaBytes = 0
		} else {
			return fmt.Errorf("unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", volName, scVol.VolBackendFs, err)
		}
	}

	if filesetQuotaBytes < scVol.VolSize && filesetQuotaBytes != 0 {
		// quota does not match and it is not 0 - It might not be fileset created by us
		return fmt.Errorf("fileset %v present but quota %v does not match with requested size %v", volName, filesetQuotaBytes, scVol.VolSize)
	}

	if filesetQuotaBytes == 0 {
		volsiz := strconv.FormatUint(scVol.VolSize, 10)
		err = scVol.Connector.SetFilesetQuota(scVol.VolBackendFs, volName, volsiz)
		if err != nil {
			// failed to set quota, no cleanup, next retry might be able to set quota
			return fmt.Errorf("unable to set quota [%v] on fileset [%v] of FS [%v]", scVol.VolSize, volName, scVol.VolBackendFs)
		}
	}
	return nil
}

//createFilesetBasedVol: Create fileset based volume  - return relative path of volume created
func (cs *ScaleControllerServer) createFilesetBasedVol(scVol *scaleVolume, isNewVolumeType bool) (string, error) { //nolint:gocyclo,funlen
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

	if isNewVolumeType {
		// For new storageClass first create independent fileset if not present
		indepFilesetName := scVol.ConsistencyGroup
		glog.V(4).Infof("creating independent fileset for new storageClass with fileset name: [%v]", indepFilesetName)
		opt[connectors.UserSpecifiedFilesetType] = independentFileset
		opt[connectors.UserSpecifiedParentFset] = ""
		//Set uid and gid as 0 for CG independent fileset
		opt[connectors.UserSpecifiedUid] = "0"
		opt[connectors.UserSpecifiedGid] = "0"
		if scVol.InodeLimit != "" {
			opt[connectors.UserSpecifiedInodeLimit] = scVol.InodeLimit
		} else {
			opt[connectors.UserSpecifiedInodeLimit] = "1M"
			// Assumption: On an average a consistency group contains 10 volumes
		}
		scVol.ParentFileset = ""
		createDataDir := false
		filesetPath, err := cs.createFilesetVol(scVol, indepFilesetName, fsDetails, opt, createDataDir, true, isNewVolumeType)
		if err != nil {
			glog.Errorf("volume:[%v] - failed to create independent fileset [%v] in filesystem [%v]. Error: %v", indepFilesetName, indepFilesetName, scVol.VolBackendFs, err)
			return "", err
		}
		glog.V(4).Infof("finished creation of independent fileset for new storageClass with fileset name: [%v]", indepFilesetName)

		// Now create dependent fileset
		glog.V(4).Infof("creating dependent fileset for new storageClass with fileset name: [%v]", scVol.VolName)
		opt[connectors.UserSpecifiedFilesetType] = dependentFileset
		opt[connectors.UserSpecifiedParentFset] = indepFilesetName
		delete(opt, connectors.UserSpecifiedUid)
		delete(opt, connectors.UserSpecifiedGid)
		if scVol.VolUid != "" {
			opt[connectors.UserSpecifiedUid] = scVol.VolUid
		}
		if scVol.VolGid != "" {
			opt[connectors.UserSpecifiedGid] = scVol.VolGid
		}
		scVol.ParentFileset = indepFilesetName
		createDataDir = true
		filesetPath, err = cs.createFilesetVol(scVol, scVol.VolName, fsDetails, opt, createDataDir, false, isNewVolumeType)
		if err != nil {
			glog.Errorf("volume:[%v] - failed to create dependent fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
			return "", err
		}
		glog.V(4).Infof("finished creation of dependent fileset for new storageClass with fileset name: [%v]", scVol.VolName)
		return filesetPath, nil
	} else {
		// Create volume for classic storageClass
		// Check if FileSetType not specified
		if scVol.FilesetType != "" {
			opt[connectors.UserSpecifiedFilesetType] = scVol.FilesetType
		}
		if scVol.ParentFileset != "" {
			opt[connectors.UserSpecifiedParentFset] = scVol.ParentFileset
		}

		// Create fileset
		glog.V(4).Infof("creating fileset for classic storageClass with fileset name: [%v]", scVol.VolName)
		createDataDir := true
		filesetPath, err := cs.createFilesetVol(scVol, scVol.VolName, fsDetails, opt, createDataDir, false, isNewVolumeType)
		if err != nil {
			glog.Errorf("volume:[%v] - failed to create fileset [%v] in filesystem [%v]. Error: %v", scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
			return "", err
		}
		glog.V(4).Infof("finished creation of fileset for classic storageClass with fileset name: [%v]", scVol.VolName)
		return filesetPath, nil
	}

}

func (cs *ScaleControllerServer) createFilesetVol(scVol *scaleVolume, volName string, fsDetails connectors.FileSystem_v2, opt map[string]interface{}, createDataDir bool, isCGIndependentFset bool, isNewVolumeType bool) (string, error) { //nolint:gocyclo,funlen
	// Check if fileset exist
	filesetInfo, err := scVol.Connector.ListFileset(scVol.VolBackendFs, volName)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in 'filesetName'") {
			// This means fileset is not present, create it
			fseterr := scVol.Connector.CreateFileset(scVol.VolBackendFs, volName, opt)

			if fseterr != nil {
				// fileset creation failed return without cleanup
				glog.Errorf("volume:[%v] - unable to create fileset [%v] in filesystem [%v]. Error: %v", volName, volName, scVol.VolBackendFs, fseterr)
				return "", status.Error(codes.Internal, fmt.Sprintf("unable to create fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, fseterr))
			}
			// list fileset and update filesetInfo
			filesetInfo, err = scVol.Connector.ListFileset(scVol.VolBackendFs, volName)
			if err != nil {
				// fileset got created but listing failed, return without cleanup
				glog.Errorf("volume:[%v] - unable to list newly created fileset [%v] in filesystem [%v]. Error: %v", volName, volName, scVol.VolBackendFs, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("unable to list newly created fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, err))
			}
		} else {
			glog.Errorf("volume:[%v] - unable to list fileset [%v] in filesystem [%v]. Error: %v", volName, volName, scVol.VolBackendFs, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, err))
		}
	} else {
		// fileset is present. Confirm if creator is IBM Spectrum Scale CSI driver and fileset type is correct.
		if filesetInfo.Config.Comment != connectors.FilesetComment {
			glog.Errorf("volume:[%v] - the fileset is not created by IBM Spectrum Scale CSI driver. Cannot use it.", volName)
			return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - the fileset is not created by IBM Spectrum Scale CSI driver. Cannot use it.", volName))
		}
		listFilesetType := ""
		if filesetInfo.Config.IsInodeSpaceOwner == true {
			listFilesetType = independentFileset
		} else {
			listFilesetType = dependentFileset
		}
		if opt[connectors.UserSpecifiedFilesetType] != listFilesetType {
			glog.Errorf("volume:[%v] - the fileset type is not as expected, got type: [%s], expected type: [%s]", volName, listFilesetType, opt[connectors.UserSpecifiedFilesetType])
			return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - the fileset type is not as expected, got type: [%s], expected type: [%s]", volName, listFilesetType, opt[connectors.UserSpecifiedFilesetType]))
		}
	}

	// fileset is present/created. Confirm if fileset is linked
	if (filesetInfo.Config.Path == "") || (filesetInfo.Config.Path == filesetUnlinkedPath) {
		// this means not linked, link it
		var junctionPath string
		junctionPath = fmt.Sprintf("%s/%s", fsDetails.Mount.MountPoint, volName)

		if scVol.ParentFileset != "" {
			parentfilesetInfo, err := scVol.Connector.ListFileset(scVol.VolBackendFs, scVol.ParentFileset)
			if err != nil {
				glog.Errorf("volume:[%v] - unable to get details of parent fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.ParentFileset, scVol.VolBackendFs, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - unable to get details of parent fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.ParentFileset, scVol.VolBackendFs, err))
			}
			if (parentfilesetInfo.Config.Path == "") || (parentfilesetInfo.Config.Path == filesetUnlinkedPath) {
				glog.Errorf("volume:[%v] - parent fileset [%v] is not linked", volName, scVol.ParentFileset)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - parent fileset [%v] is not linked", volName, scVol.ParentFileset))
			}
			junctionPath = fmt.Sprintf("%s/%s", parentfilesetInfo.Config.Path, volName)
		}

		err := scVol.Connector.LinkFileset(scVol.VolBackendFs, volName, junctionPath)
		if err != nil {
			glog.Errorf("volume:[%v] - linking fileset [%v] in filesystem [%v] at path [%v] failed. Error: %v", volName, volName, scVol.VolBackendFs, junctionPath, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("linking fileset [%v] in filesystem [%v] at path [%v] failed. Error: %v", volName, scVol.VolBackendFs, junctionPath, err))
		}
		// update fileset details
		filesetInfo, err = scVol.Connector.ListFileset(scVol.VolBackendFs, volName)
		if err != nil {
			glog.Errorf("volume:[%v] - unable to list fileset [%v] in filesystem [%v] after linking. Error: %v", volName, volName, scVol.VolBackendFs, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v] after linking. Error: %v", volName, scVol.VolBackendFs, err))
		}
	}
	targetBasePath := ""
	if !isCGIndependentFset {
		if scVol.VolSize != 0 {
			err = cs.setQuota(scVol, volName)
			if err != nil {
				return "", status.Error(codes.Internal, err.Error())
			}
		}

		targetBasePath, err = cs.getTargetPath(filesetInfo.Config.Path, fsDetails.Mount.MountPoint, volName, createDataDir, isNewVolumeType)
		if err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}

		err = cs.createDirectory(scVol, volName, targetBasePath)
		if err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}
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

// checkSCSupportedParams checks if given CreateVolume request parameter keys
// are supported by Spectrum Scale CSI and returns ("", true) if all parameter
// keys are supported, otherwise returns (<list of invalid keys seperated by
// comma>, false)
func checkSCSupportedParams(params map[string]string) (string, bool) {
	var invalidParams []string
	for k := range params {
		switch k {
		case "csi.storage.k8s.io/pv/name", "csi.storage.k8s.io/pvc/name",
			"csi.storage.k8s.io/pvc/namespace", "storage.kubernetes.io/csiProvisionerIdentity",
			"volBackendFs", "volDirBasePath", "uid", "gid", "permissions",
			"clusterId", "filesetType", "parentFileset", "inodeLimit",
			"nodeClass", "version", "tier", "compression":
			// These are valid parameters, do nothing here
		default:
			invalidParams = append(invalidParams, k)
		}
	}
	if len(invalidParams) == 0 {
		return "", true
	}
	return strings.Join(invalidParams[:], ", "), false
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

	invalidParams, allValid := checkSCSupportedParams(req.GetParameters())
	if !allValid {
		return nil, status.Error(codes.InvalidArgument, "The Parameter(s) not supported in storageClass: "+invalidParams)
	}
	scaleVol, err := getScaleVolumeOptions(req.GetParameters())
	if err != nil {
		return nil, err
	}

	isNewVolumeType := false
	if scaleVol.StorageClassType == storageClassAdvanced {
		isNewVolumeType = true
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
	isVolSource := false

	snapIdMembers := scaleSnapId{}
	srcVolumeIDMembers := scaleVolId{}

	if volSrc != nil {
		srcVolume := volSrc.GetVolume()
		if srcVolume != nil {
			srcVolumeID := srcVolume.GetVolumeId()
			srcVolumeIDMembers, err = getVolIDMembers(srcVolumeID)
			if err != nil {
				glog.Errorf("volume:[%v] - Invalid Volume ID %s [%v]", volName, srcVolumeID, err)
				return nil, err
			}
			isVolSource = true
		} else {

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

	var remoteClusterID string
	if scaleVol.ClusterId == "" && volFsInfo.Type == filesystemTypeRemote {
		glog.V(5).Infof("filesystem %s is remotely mounted, getting cluster ID information of the owning cluster.", volFsInfo.Name)
		clusterName := strings.Split(volFsInfo.Mount.RemoteDeviceName, ":")[0]
		if remoteClusterID, err = cs.getRemoteClusterID(clusterName); err != nil {
			return nil, err
		}
		glog.V(5).Infof("cluster ID for remote cluster %s is %s", clusterName, remoteClusterID)
	}

	if scaleVol.IsFilesetBased {
		if scaleVol.ClusterId == "" {
			if volFsInfo.Type == filesystemTypeRemote { // if fileset based and remotely mounted.
				glog.V(3).Infof("volume filesystem %s is remotely mounted on Primary cluster, using owning cluster ID %s.", scaleVol.LocalFS, remoteClusterID)
				scaleVol.ClusterId = remoteClusterID
			} else {
				glog.V(3).Infof("volume filesystem %s is locally mounted on Primary cluster, using primary cluster ID %s.", scaleVol.LocalFS, PCid)
				scaleVol.ClusterId = PCid
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

	if isNewVolumeType {
		if err := cs.checkCGSupport(scaleVol.Connector); err != nil {
			return nil, err
		}
	}
	if isVolSource {
		if !scaleVol.IsFilesetBased {
			if volFsInfo.Type == filesystemTypeRemote {
				return nil, status.Error(codes.Unimplemented, "Volume cloning for directories for remote file system is not supported")
			}
		}

		err = cs.validateSrcVolumeID(&srcVolumeIDMembers, scaleVol, PCid)
		if err != nil {
			glog.Errorf("volume:[%v] - Error in source volume validation [%v]", volName, err)
			return nil, err
		}
	}

	if isSnapSource {
		err = cs.validateSnapId(&snapIdMembers, scaleVol, PCid)
		if err != nil {
			glog.Errorf("volume:[%v] - Error in source snapshot validation [%v]", volName, err)
			return nil, err
		}
	}

	glog.Infof("volume:[%v] -  spectrum scale volume create params : %v\n", scaleVol.VolName, scaleVol)

	if scaleVol.IsFilesetBased && scaleVol.Compression != "" {
		glog.Infof("createvolume: compression is enabled: changing volume name")
		scaleVol.VolName = fmt.Sprintf("%s-COMPRESS%scsi", scaleVol.VolName, strings.ToUpper(scaleVol.Compression))
	}

	if scaleVol.IsFilesetBased && scaleVol.Tier != "" {
		if err := cs.checkVolTierSupport(volFsInfo.Version); err != nil {
			// TODO: Remove this secondary call to local gui when GUI refreshes remote cache immediately
			tempFsInfo, err := scaleVol.Connector.GetFilesystemDetails(scaleVol.VolBackendFs)
			if err != nil {
				return nil, err
			}
			if err := cs.checkVolTierSupport(tempFsInfo.Version); err != nil {
				return nil, err
			}
		}

		if err := scaleVol.Connector.DoesTierExist(scaleVol.Tier, scaleVol.VolBackendFs); err != nil {
			return nil, err
		}

		rule := "RULE 'csi-T%s' SET POOL '%s' WHERE FILESET_NAME LIKE 'pvc-%%-T%scsi%%'"
		policy := connectors.Policy{}

		policy.Policy = fmt.Sprintf(rule, scaleVol.Tier, scaleVol.Tier, scaleVol.Tier)
		policy.Priority = -5
		policy.Partition = fmt.Sprintf("csi-T%s", scaleVol.Tier)

		scaleVol.VolName = fmt.Sprintf("%s-T%scsi", scaleVol.VolName, scaleVol.Tier)
		err = scaleVol.Connector.SetFilesystemPolicy(&policy, scaleVol.VolBackendFs)
		if err != nil {
			glog.Errorf("volume:[%v] - setting policy failed [%v]", volName, err)
			return nil, err
		}

		// Since we are using a SET POOL rule, if there is not already a default rule in place in the policy partition
		// then all files that do not match our rules will have no defined place to go. This sets a default rule with
		// "lower" priority than the main policy as a catch all. If there is already a default rule in the main policy
		// file then that will take precedence
		defaultPartitionName := "csi-defaultRule"
		if !scaleVol.Connector.CheckIfDefaultPolicyPartitionExists(defaultPartitionName, scaleVol.VolBackendFs) {
			glog.Infof("createvolume: setting default policy partition rule")

			dataTierName, err := scaleVol.Connector.GetFirstDataTier(scaleVol.VolBackendFs)
			if err != nil {
				return nil, status.Error(codes.Unavailable, fmt.Sprintf("tier info request could not be completed: filesystemName %s", scaleVol.VolBackendFs))
			}
			defaultPolicy := connectors.Policy{}
			defaultPolicy.Policy = fmt.Sprintf("RULE 'csi-defaultRule' SET POOL '%s'", dataTierName)
			defaultPolicy.Priority = 5
			defaultPolicy.Partition = defaultPartitionName
			err = scaleVol.Connector.SetFilesystemPolicy(&defaultPolicy, scaleVol.VolBackendFs)
			if err != nil {
				glog.Errorf("volume:[%v] - setting default policy failed [%v]", volName, err)
				return nil, err
			}
		}
	}

	volReqInProcess, err := cs.IfSameVolReqInProcess(scaleVol)
	if err != nil {
		return nil, err
	}

	if volReqInProcess {
		glog.Errorf("volume:[%v] - volume creation already in process ", scaleVol.VolName)
		return nil, status.Error(codes.Aborted, fmt.Sprintf("volume creation already in process : %v", scaleVol.VolName))
	}

	if isVolSource {
		jobDetails, found := cs.Driver.volcopyjobstatusmap.Load(scaleVol.VolName)
		if found {
			jobStatus := jobDetails.(VolCopyJobDetails).jobStatus
			volID := jobDetails.(VolCopyJobDetails).volID
			glog.V(5).Infof("volume: [%v] found in volcopyjobstatusmap with volID: [%v], jobStatus: [%v]", scaleVol.VolName, volID, jobStatus)
			if jobStatus == VOLCOPY_JOB_RUNNING {
				glog.Errorf("volume:[%v] -  volume cloning request in progress.", scaleVol.VolName)
				return nil, status.Error(codes.Aborted, fmt.Sprintf("volume cloning request in progress for volume: %s", scaleVol.VolName))
			} else if jobStatus == VOLCOPY_JOB_FAILED {
				glog.Errorf("volume:[%v] -  volume cloning job had failed", scaleVol.VolName)
				return nil, status.Error(codes.Internal, fmt.Sprintf("volume cloning job had failed for volume:[%v]", scaleVol.VolName))
			} else if jobStatus == VOLCOPY_JOB_COMPLETED {
				glog.V(5).Infof("volume:[%v] -  volume cloning request has already completed successfully.", scaleVol.VolName)
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{
						VolumeId:      volID,
						CapacityBytes: int64(scaleVol.VolSize),
						VolumeContext: req.GetParameters(),
						ContentSource: volSrc,
					},
				}, nil
			}
		} else {
			glog.V(5).Infof("volume: [%v] not found in volcopyjobstatusmap", scaleVol.VolName)
		}
	}

	if isSnapSource {
		jobDetails, found := cs.Driver.snapjobstatusmap.Load(scaleVol.VolName)
		if found {
			jobStatus := jobDetails.(SnapCopyJobDetails).jobStatus
			volID := jobDetails.(SnapCopyJobDetails).volID
			glog.V(5).Infof("volume: [%v] found in snapjobstatusmap with volID: [%v], jobStatus: [%v]", scaleVol.VolName, volID, jobStatus)
			if jobStatus == SNAP_JOB_RUNNING {
				glog.Errorf("volume:[%v] -  snapshot copy request in progress for snapshot: %s.", scaleVol.VolName, snapIdMembers.SnapName)
				return nil, status.Error(codes.Aborted, fmt.Sprintf("snapshot copy request in progress for snapshot: %s", snapIdMembers.SnapName))
			} else if jobStatus == SNAP_JOB_FAILED {
				glog.Errorf("volume:[%v] -  snapshot copy job had failed for snapshot %s", scaleVol.VolName, snapIdMembers.SnapName)
				return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot copy job had failed for snapshot: %s", snapIdMembers.SnapName))
			} else if jobStatus == SNAP_JOB_COMPLETED {
				glog.V(5).Infof("volume:[%v] -  snapshot copy request has already completed successfully for snapshot: %s", scaleVol.VolName, snapIdMembers.SnapName)
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{
						VolumeId:      volID,
						CapacityBytes: int64(scaleVol.VolSize),
						VolumeContext: req.GetParameters(),
						ContentSource: volSrc,
					},
				}, nil
			}
		} else {
			glog.V(5).Infof("volume: [%v] not found in snapjobstatusmap", scaleVol.VolName)
		}
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
		targetPath, err = cs.createFilesetBasedVol(scaleVol, isNewVolumeType)
	} else {
		targetPath, err = cs.createLWVol(scaleVol)
	}

	if err != nil {
		return nil, err
	}

	if !isNewVolumeType {
		// Create symbolic link if not present
		err = cs.createSoftlink(scaleVol, targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	volID, volIDErr := cs.generateVolID(scaleVol, volFsInfo.UUID, isNewVolumeType, targetPath)
	if volIDErr != nil {
		return nil, volIDErr
	}

	if isVolSource {
		err = cs.copyVolumeContent(scaleVol, srcVolumeIDMembers, volFsInfo, targetPath, volID)
		if err != nil {
			glog.Errorf("CreateVolume [%s]: [%v]", volName, err)
			return nil, err
		}
	}

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

	snapIDPath := snapId.Path
	filesetForCopy := snapId.FsetName
	if snapId.StorageClassType == STORAGECLASS_ADVANCED {
		snapIDPath = fmt.Sprintf("/%s", snapId.FsetName)
		filesetForCopy = snapId.ConsistencyGroup
	}
	jobStatus, jobID, err := conn.CopyFsetSnapshotPath(snapId.FsName, filesetForCopy, snapId.SnapName, snapIDPath, targetPath, scVol.NodeClass)
	if err != nil {
		glog.Errorf("failed to create volume from snapshot %s: [%v]", snapId.SnapName, err)
		return status.Error(codes.Internal, fmt.Sprintf("failed to create volume from snapshot %s: [%v]", snapId.SnapName, err))

	}

	jobDetails := SnapCopyJobDetails{SNAP_JOB_RUNNING, volID}
	cs.Driver.snapjobstatusmap.Store(scVol.VolName, jobDetails)

	err = conn.WaitForJobCompletion(jobStatus, jobID)
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
		cs.Driver.snapjobstatusmap.Store(scVol.VolName, jobDetails)
		return err
	}

	glog.Infof("copy snapshot completed for snapId: [%v], scaleVolume: [%v]", snapId, scVol)
	jobDetails.jobStatus = SNAP_JOB_COMPLETED
	cs.Driver.snapjobstatusmap.Store(scVol.VolName, jobDetails)
	//delete(cs.Driver.snapjobmap, scVol.VolName)
	return nil
}

func (cs *ScaleControllerServer) copyVolumeContent(scVol *scaleVolume, vID scaleVolId, fsDetails connectors.FileSystem_v2, targetPath string, volID string) error {
	glog.V(3).Infof("copyVolContent volume ID: [%v], scaleVolume: [%v], volume name: [%v]", vID, scVol, scVol.VolName)
	conn, err := cs.getConnFromClusterID(vID.ClusterId)
	if err != nil {
		return err
	}

	// err = cs.validateRemoteFs(fsDetails, scVol)
	// if err != nil {
	// 	return err
	// }

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

	jobDetails := VolCopyJobDetails{VOLCOPY_JOB_NOT_STARTED, volID}
	if scVol.IsFilesetBased {
		path := ""
		if vID.StorageClassType == STORAGECLASS_ADVANCED {
			path = "/"
		} else {
			path = fmt.Sprintf("%s%s", vID.FsetName, "-data")
		}

		jobStatus, jobID, jobErr := conn.CopyFilesetPath(vID.FsName, vID.FsetName, path, targetPath, scVol.NodeClass)
		if jobErr != nil {
			glog.Errorf("failed to clone volume from volume. Error: [%v]", jobErr)
			return status.Error(codes.Internal, fmt.Sprintf("failed to clone volume from volume. Error: [%v]", jobErr))
		}

		jobDetails = VolCopyJobDetails{VOLCOPY_JOB_RUNNING, volID}
		cs.Driver.volcopyjobstatusmap.Store(scVol.VolName, jobDetails)
		err = conn.WaitForJobCompletion(jobStatus, jobID)
	} else {
		sLinkRelPath := strings.Replace(vID.Path, cs.Driver.primary.PrimaryFSMount, "", 1)
		sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

		jobStatus, jobID, jobErr := conn.CopyDirectoryPath(vID.FsName, sLinkRelPath, targetPath, scVol.NodeClass)

		if jobErr != nil {
			glog.Errorf("failed to clone volume from volume. Error: [%v]", jobErr)
			return status.Error(codes.Internal, fmt.Sprintf("failed to clone volume from volume. Error: [%v]", jobErr))
		}

		jobDetails = VolCopyJobDetails{VOLCOPY_JOB_RUNNING, volID}
		cs.Driver.volcopyjobstatusmap.Store(scVol.VolName, jobDetails)
		err = conn.WaitForJobCompletion(jobStatus, jobID)
	}

	if err != nil {
		glog.Errorf("unable to copy volume: %v.", err)
		if strings.Contains(err.Error(), "EFSSG0632C") {
			// EFSSG0632C = Command execution aborted
			// Store VOLCOPY_JOB_NOT_STARTED in volcopyjobstatusmap if error was due to same mmxcp in progress
			// or max no. of mmxcp already running. In these cases we want to retry again
			// in the next k8s rety cycle
			jobDetails.jobStatus = VOLCOPY_JOB_NOT_STARTED
		} else {
			jobDetails.jobStatus = VOLCOPY_JOB_FAILED
		}
		glog.Errorf("logging volume cloning error for VolName: [%v] Error: [%v] JobDetails: [%v]", scVol.VolName, err, jobDetails)
		cs.Driver.volcopyjobstatusmap.Store(scVol.VolName, jobDetails)
		return err
	}

	glog.Infof("volume copy completed for volumeID: [%v], scaleVolume: [%v]", vID, scVol)
	jobDetails.jobStatus = VOLCOPY_JOB_COMPLETED
	cs.Driver.volcopyjobstatusmap.Store(scVol.VolName, jobDetails)
	//delete(cs.Driver.volcopyjobstatusmap, scVol.VolName)
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

func (cs *ScaleControllerServer) checkMinFsVersion(fsVersion string, version string) bool {
	/* Assuming Filesystem version (fsVersion) in a format like 27.00 and version as 2700 */
	assembledFsVer := strings.ReplaceAll(fsVersion, ".", "")

	glog.Infof("fs version (%s) vs min required version (%s)", assembledFsVer, version)
	if assembledFsVer < version {
		return false
	}
	return true
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

func (cs *ScaleControllerServer) checkVolCloneSupport(conn connectors.SpectrumScaleConnector) error {
	/* Verify Spectrum Scale Version is not below 5.1.2-1 */
	versionCheck, err := cs.checkMinScaleVersion(conn, "5121")
	if err != nil {
		return err
	}

	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required Spectrum Scale version for volume cloning support with CSI is 5.1.2-1")
	}
	return nil
}

func (cs *ScaleControllerServer) checkVolTierSupport(version string) error {
	/* Verify Spectrum Scale Filesystem Version is not below 5.1.3-0 (27.00) */

	versionCheck := cs.checkMinFsVersion(version, "2700")

	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required Spectrum Scale Filesystem version for tiering support with CSI is 27.00 (5.1.3-0)")
	}
	return nil
}

func (cs *ScaleControllerServer) checkCGSupport(conn connectors.SpectrumScaleConnector) error {
	/* Verify Spectrum Scale Version is not below 5.1.3-0 */

	versionCheck, err := cs.checkMinScaleVersion(conn, "5130")
	if err != nil {
		return err
	}

	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required Spectrum Scale version for consistency group support with CSI is 5.1.3-0")
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

	filesetToCheck := sId.FsetName
	if sId.StorageClassType == STORAGECLASS_ADVANCED {
		// check if independent fileset is linked
		filesetToCheck = sId.ConsistencyGroup
		isFsetLinked, err := conn.IsFilesetLinked(sId.FsName, filesetToCheck)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("unable to get fileset link information for [%v]", filesetToCheck))
		}
		if !isFsetLinked {
			return status.Error(codes.Internal, fmt.Sprintf("fileset [%v] of source snapshot is not linked", filesetToCheck))
		}
	}

	isSnapExist, err := conn.CheckIfSnapshotExist(sId.FsName, filesetToCheck, sId.SnapName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get snapshot information for [%v]", sId.SnapName))
	}
	if !isSnapExist {
		return status.Error(codes.Internal, fmt.Sprintf("snapshot [%v] does not exist for fileset [%v]", sId.SnapName, filesetToCheck))
	}

	return nil
}

func (cs *ScaleControllerServer) validateSrcVolumeID(vID *scaleVolId, scVol *scaleVolume, pCid string) error {
	glog.V(3).Infof("validateVolId [%v]", vID)
	conn, err := cs.getConnFromClusterID(vID.ClusterId)
	if err != nil {
		return err
	}

	// This is kind of snapshot restore
	chkVolCloneErr := cs.checkVolCloneSupport(conn)
	if chkVolCloneErr != nil {
		return chkVolCloneErr
	}

	if scVol.ClusterId != "" && vID.ClusterId != scVol.ClusterId {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot create volume from a source volume from another cluster. Volume is being created in cluster %s, source volume is from cluster %s.", scVol.ClusterId, vID.ClusterId))
	}

	if scVol.ClusterId == "" && vID.ClusterId != pCid {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot create volume from a source volume from another cluster. Volume is being created in cluster %s, source volume is from cluster %s.", pCid, vID.ClusterId))
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

	vID.FsName, err = conn.GetFilesystemName(vID.FsUUID)

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", vID.FsUUID, vID.ClusterId, err))
	}

	if vID.FsName != scVol.VolBackendFs {
		isFsMounted, err := conn.IsFilesystemMountedOnGUINode(vID.FsName)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error in getting filesystem mount details for %s", vID.FsName))
		}
		if !isFsMounted {
			return status.Error(codes.Internal, fmt.Sprintf("filesystem %s is not mounted on GUI node", vID.FsName))
		}
	}

	if scVol.IsFilesetBased {
		if vID.FsetName == "" {
			vID.FsetName, err = conn.GetFileSetNameFromId(vID.FsName, vID.FsetId)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("error in getting fileset details for %s", vID.FsetId))
			}
		}

		isFsetLinked, err := conn.IsFilesetLinked(vID.FsName, vID.FsetName)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("unable to get fileset link information for [%v]", vID.FsetName))
		}
		if !isFsetLinked {
			return status.Error(codes.Internal, fmt.Sprintf("fileset [%v] of source volume is not linked", vID.FsetName))
		}
	}

	return nil
}

func (cs *ScaleControllerServer) GetSnapIdMembers(sId string) (scaleSnapId, error) {
	splitSid := strings.Split(sId, ";")
	var sIdMem scaleSnapId

	if len(splitSid) < 4 {
		return scaleSnapId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Snapshot Id : [%v]", sId))
	}

	if len(splitSid) >= 8 {
		/* storageclass_type;volumeType;clusterId;FSUUID;consistency_group;filesetName;snapshotName;path */
		sIdMem.StorageClassType = splitSid[0]
		sIdMem.VolType = splitSid[1]
		sIdMem.ClusterId = splitSid[2]
		sIdMem.FsUUID = splitSid[3]
		sIdMem.ConsistencyGroup = splitSid[4]
		sIdMem.FsetName = splitSid[5]
		sIdMem.SnapName = splitSid[6]
		sIdMem.MetaSnapName = splitSid[7]
		if len(splitSid) == 9 && splitSid[8] != "" {
			sIdMem.Path = splitSid[8]
		} else {
			sIdMem.Path = "/"
		}
	} else {
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
	}
	return sIdMem, nil
}

func (cs *ScaleControllerServer) DeleteFilesetVol(FilesystemName string, FilesetName string, volumeIdMembers scaleVolId, conn connectors.SpectrumScaleConnector) (bool, error) {
	//Check if fileset exist has any snapshot
	snapshotList, err := conn.ListFilesetSnapshots(FilesystemName, FilesetName)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0072C") ||
			strings.Contains(err.Error(), "400 Invalid value in 'filesetName'") { // fileset is already deleted
			glog.V(4).Infof("fileset seems already deleted - %v", err)
			return true, nil
		}
		return false, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshot for fileset [%v]. Error: [%v]", FilesetName, err))
	}

	if len(snapshotList) > 0 {
		return false, status.Error(codes.InvalidArgument, fmt.Sprintf("volume fileset [%v] contains one or more snapshot, delete snapshot/volumesnapshot", FilesetName))
	}
	glog.V(4).Infof("there is no snapshot present in the fileset [%v], continue DeleteFilesetVol", FilesetName)

	err = conn.DeleteFileset(FilesystemName, FilesetName)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0072C") ||
			strings.Contains(err.Error(), "400 Invalid value in 'filesetName'") { // fileset is already deleted
			glog.V(4).Infof("fileset seems already deleted - %v", err)
			return true, nil
		}
		return false, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Fileset [%v] for FS [%v] and clusterId [%v].Error : [%v]", FilesetName, FilesystemName, volumeIdMembers.ClusterId, err))
	}
	return false, nil
}

// This function deletes fileset for Consitency Group
func (cs *ScaleControllerServer) DeleteCGFileset(FilesystemName string, inodeSpace int, volumeIdMembers scaleVolId, conn connectors.SpectrumScaleConnector) error {
	glog.V(4).Infof("trying to delete independent fileset for consistency group [%v]", volumeIdMembers.ConsistencyGroup)
	filesets, err := conn.GetFilesetsInodeSpace(FilesystemName, inodeSpace)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("listing of filesets for filesystem: [%v] failed. Error: [%v]", FilesystemName, err))
	}

	if len(filesets) > 1 {
		glog.V(4).Infof("found atleast one dependent fileset for consistency group: [%v]", volumeIdMembers.ConsistencyGroup)
		return nil
	}

	// Check if fileset was created by IBM Spectrum Scale CSI Driver
	filesetDetails, err := conn.ListFileset(FilesystemName, volumeIdMembers.ConsistencyGroup)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0072C") ||
			strings.Contains(err.Error(), "400 Invalid value in 'filesetName'") { // fileset is already deleted
			glog.V(4).Infof("fileset seems already deleted - %v", err)
			return nil
		}
		return status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v]. Error: [%v]", volumeIdMembers.ConsistencyGroup, err))
	}

	if filesetDetails.Config.Comment == connectors.FilesetComment {
		// Delete independent fileset for consistency group
		_, err := cs.DeleteFilesetVol(FilesystemName, volumeIdMembers.ConsistencyGroup, volumeIdMembers, conn)
		if err != nil {
			return err
		}
		glog.V(4).Infof("deleted independent fileset for consistency group [%v]", volumeIdMembers.ConsistencyGroup)
	} else {
		glog.V(4).Infof("independent fileset for consistency group [%v] not created by IBM Spectrum Scale CSI Driver. Cannot delete it.", volumeIdMembers.ConsistencyGroup)
	}

	return nil
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

	volumeIdMembers, err := getVolIDMembers(volumeID)
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

	relPath := ""
	if volumeIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
		relPath = strings.Replace(volumeIdMembers.Path, mountInfo.MountPoint, "", 1)
	} else {
		relPath = strings.Replace(volumeIdMembers.Path, cs.Driver.primary.PrimaryFSMount, "", 1)
	}
	relPath = strings.Trim(relPath, "!/")

	if volumeIdMembers.IsFilesetBased {
		var FilesetName string

		FilesystemName = getRemoteFsName(mountInfo.RemoteDeviceName)
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
			pvName := filepath.Base(relPath)
			if pvName == FilesetName {
				// before deletion of fileset get its inodeSpace if storageClassType=STORAGECLASS_ADVANCED
				// this will help to identify if there are one or more dependent filesets for same inodeSpace
				// which is shared with independent fileset
				inodeSpace := 0
				if volumeIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
					// Save inodeSpace information
					filesetDetails, err := conn.ListFileset(FilesystemName, FilesetName)
					if err != nil {
						if strings.Contains(err.Error(), "EFSSG0072C") ||
							strings.Contains(err.Error(), "400 Invalid value in 'filesetName'") { // fileset is already deleted

							glog.V(4).Infof("fileset seems already deleted - %v", err)
							return &csi.DeleteVolumeResponse{}, nil
						}
						return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get the fileset details. Error [%v]", err))
					}
					inodeSpace = filesetDetails.Config.InodeSpace
				}

				isFilesetAlreadyDel, err := cs.DeleteFilesetVol(FilesystemName, FilesetName, volumeIdMembers, conn)
				if err != nil {
					return nil, err
				}

				// Delete fileset related symlink
				if !isFilesetAlreadyDel && volumeIdMembers.StorageClassType != STORAGECLASS_ADVANCED {
					err = primaryConn.DeleteSymLnk(cs.Driver.primary.GetPrimaryFs(), relPath)
					if err != nil {
						return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete symlnk [%v:%v] Error [%v]", cs.Driver.primary.GetPrimaryFs(), relPath, err))
					}
				}

				if volumeIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
					err := cs.DeleteCGFileset(FilesystemName, inodeSpace, volumeIdMembers, conn)
					if err != nil {
						return nil, err
					}
				}
				return &csi.DeleteVolumeResponse{}, nil
			} else {
				glog.Infof("pv name from path [%v] does not match with filesetName [%v]. Skipping delete of fileset", pvName, FilesetName)
			}
		}
	} else {
		/* Delete Dir for Lw volume */
		err = primaryConn.DeleteDirectory(cs.Driver.primary.GetPrimaryFs(), relPath, false)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Dir using FS [%v] Relative SymLink [%v]. Error [%v]", FilesystemName, relPath, err))
		}
	}

	if volumeIdMembers.StorageClassType != STORAGECLASS_ADVANCED {
		err = primaryConn.DeleteSymLnk(cs.Driver.primary.GetPrimaryFs(), relPath)

		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete symlnk [%v:%v] Error [%v]", cs.Driver.primary.GetPrimaryFs(), relPath, err))
		}
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
	_, err := getVolIDMembers(volumeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ControllerUnpublishVolume : VolumeID is not in proper format")
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

	//Assumption : filesystem_uuid is always from local/primary cluster.

	volumeIDMembers, err := getVolIDMembers(volumeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume : VolumeID is not in proper format")
	}

	filesystemID := volumeIDMembers.FsUUID
	volumePath := volumeIDMembers.Path

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
	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED || primaryfsName != fsName {
		//Check if filesystem is mounted
		fsMount, err := cs.Driver.connmap["primary"].GetFilesystemMountDetails(fsName)
		if err != nil {
			glog.Errorf("ControllerPublishVolume : Error in getting filesystem mount details for %s", fsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem mount details for %s. Error [%v]", fsName, err))
		}

		if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED &&
			!strings.HasPrefix(volumePath, fsMount.MountPoint) {
			glog.Errorf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s", volumePath, fsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s", volumePath, fsName))
		} else if !strings.HasPrefix(volumePath, fsMount.MountPoint) &&
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

func (cs *ScaleControllerServer) CheckNewSnapRequired(conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string, snapWindow string) (string, error) {
	latestSnapList, err := conn.GetLatestFilesetSnapshots(filesystemName, filesetName)
	if err != nil {
		glog.Errorf("CheckNewSnapRequired - getting latest snapshot list failed for fileset: [%s:%s]. Error: [%v]", filesystemName, filesetName, err)
		return "", err
	}

	if len(latestSnapList) == 0 {
		// No snapshot exists, so create new one
		return "", nil
	}

	timestamp, err := cs.getSnapshotCreateTimestamp(conn, filesystemName, filesetName, latestSnapList[0].SnapshotName)
	if err != nil {
		glog.Errorf("error getting create timestamp for snapshot %s:%s:%s", filesystemName, filesetName, latestSnapList[0].SnapshotName)
		return "", err
	}

	var timestampSecs int64
	timestampSecs = timestamp.GetSeconds()
	lastSnapTime := time.Unix(timestampSecs, 0)
	passedTime := time.Now().Sub(lastSnapTime).Seconds()
	glog.V(3).Infof("for fileset [%s:%s], last snapshot time: [%v], current time: [%v], passed time: %v seconds, snapWindow: %v minutes", filesystemName, filesetName, lastSnapTime, time.Now(), int64(passedTime), snapWindow)

	snapWindowInt, err := strconv.Atoi(snapWindow)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("CheckNewSnapRequired - invalid snapWindow value: [%v]", snapWindow))
	}
	snapWindowSeconds := snapWindowInt * 60

	if passedTime < float64(snapWindowSeconds) {
		// we don't need to take new snapshot
		glog.V(3).Infof("CheckNewSnapRequired - for fileset [%s:%s], using existing snapshot [%s]", filesystemName, filesetName, latestSnapList[0].SnapshotName)
		return latestSnapList[0].SnapshotName, nil
	}

	glog.V(3).Infof("CheckNewSnapRequired - for fileset [%s:%s] we need to create new snapshot", filesystemName, filesetName)
	return "", nil
}

func (cs *ScaleControllerServer) MakeSnapMetadataDir(conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string, indepFileset string, cgSnapName string, metaSnapName string) error {
	path := fmt.Sprintf("%s/%s/%s", indepFileset, cgSnapName, metaSnapName)
	glog.V(3).Infof("MakeSnapMetadataDir - creating directory [%s] for fileset: [%s:%s]", path, filesystemName, filesetName)
	err := conn.MakeDirectory(filesystemName, path, "0", "0")
	if err != nil {
		// Directory creation failed
		glog.Errorf("for volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", filesetName, path, filesystemName, err)
		return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", path, filesystemName, err)
	}
	return nil
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

	volumeIDMembers, err := getVolIDMembers(volID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - Error in source Volume ID %v: %v", volID, err))
	}

	if !volumeIDMembers.IsFilesetBased {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is fileset", volID))
	}

	if (volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED) && (volumeIDMembers.VolType != FILE_DEPENDENTFILESET_VOLUME) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is dependent fileset for new storageClass", volID))
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

	primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
	if !isprimaryConnPresent {
		glog.Errorf("CreateSnapshot - unable to get connector for primary cluster")
		return nil, status.Error(codes.Internal, "CreateSnapshot - unable to find primary cluster details in custom resource")
	}

	filesystemName, err := primaryConn.GetFilesystemName(volumeIDMembers.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
	}

	mountInfo, err := primaryConn.GetFilesystemMountDetails(filesystemName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - unable to get mount info for FS [%v] in primary cluster", filesystemName))
	}

	filesetResp := connectors.Fileset_v2{}
	filesystemName = getRemoteFsName(mountInfo.RemoteDeviceName)
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

	if volumeIDMembers.StorageClassType != STORAGECLASS_ADVANCED {
		if filesetResp.Config.ParentId > 0 {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is independent fileset", volID))
		}
	}

	filesetName := filesetResp.FilesetName
	relPath := ""
	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		glog.V(3).Infof("CreateSnapshot - creating snapshot for advanced storageClass")
		relPath = strings.Replace(volumeIDMembers.Path, mountInfo.MountPoint, "", 1)
	} else {
		glog.V(3).Infof("CreateSnapshot - creating snapshot for classic storageClass")
		relPath = strings.Replace(volumeIDMembers.Path, cs.Driver.primary.PrimaryFSMount, "", 1)
	}
	relPath = strings.Trim(relPath, "!/")

	/* Confirm it is same fileset which was created for this PV */
	pvName := filepath.Base(relPath)
	if pvName != filesetName {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - PV name from path [%v] does not match with filesetName [%v].", pvName, filesetName))
	}

	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		filesetName = volumeIDMembers.ConsistencyGroup
	}

	snapName := req.GetName()
	snapWindow := ""
	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		snapParams := req.GetParameters()
		snapWindowSpecified := false
		snapWindow, snapWindowSpecified = snapParams[connectors.UserSpecifiedSnapWindow]
		if !snapWindowSpecified {
			// use default snapshot window for consistency group
			snapWindow = defaultSnapWindow
			glog.V(3).Infof("snapWindow not specified. Using default snapWindow: [%s] for for fileset[%s:%s]", snapWindow, filesetResp.FilesetName, filesystemName)
		}
	}

	snapExist, err := conn.CheckIfSnapshotExist(filesystemName, filesetName, snapName)
	if err != nil {
		glog.Errorf("CreateSnapshot [%s] - Unable to get the snapshot details. Error [%v]", snapName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get the snapshot details for [%s]. Error [%v]", snapName, err))
	}

	if !snapExist {
		/* For new storageClass check last snapshot creation time, if time passed is less than
		 * snapWindow then return existing snapshot */
		createNewSnap := true
		if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
			// check if snapshot creation already in progress
			snapIdentifier := fmt.Sprintf("%s:%s", filesystemName, filesetName)
			snapReadyStatus, found := cs.Driver.cgsnapreadystatusmap.Load(snapIdentifier)
			if found && !snapReadyStatus.(CGSnapReadyDetails).snapStatus {
				// snapshot not yet ready
				return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot [%s] not yet ready for [%s:%s]", snapReadyStatus.(CGSnapReadyDetails).snapName, filesystemName, filesetName))	
			}
			cgSnapName, err := cs.CheckNewSnapRequired(conn, filesystemName, filesetName, snapWindow)
			if err != nil {
				glog.Errorf("CreateSnapshot [%s] - unable to check if snapshot is required for new storageClass for fileset [%s:%s]. Error: [%v]", snapName, filesystemName, filesetName, err)
				return nil, err
			}
			if cgSnapName != "" {
				usable, err := cs.isExistingSnapUseableForVol(conn, filesystemName, filesetName, filesetResp.FilesetName, cgSnapName)
				if !usable {
					return nil, err
				}
				createNewSnap = false
				snapName = cgSnapName
			} else {
				cs.Driver.cgsnapmutex.Lock()
				// re-check if snapshot creation was started by another thread or not
				snapReadyStatus, found := cs.Driver.cgsnapreadystatusmap.Load(snapIdentifier)
				if found && !snapReadyStatus.(CGSnapReadyDetails).snapStatus {
					// another thread already started creating snapshot
					cs.Driver.cgsnapmutex.Unlock()
					return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot [%s] not yet ready for [%s:%s]", snapReadyStatus.(CGSnapReadyDetails).snapName, filesystemName, filesetName))
				}
				snapReadyDetails := CGSnapReadyDetails{false, snapName}
				cs.Driver.cgsnapreadystatusmap.Store(snapIdentifier, snapReadyDetails)
				cs.Driver.cgsnapmutex.Unlock()

				glog.V(3).Infof("CreateSnapshot - creating new snapshot for consistency group for fileset: [%s:%s]", filesystemName, filesetName)
			}
		}

		if createNewSnap {
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
			if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
				snapIdentifier := fmt.Sprintf("%s:%s", filesystemName, filesetName)
				snapReadyDetails := CGSnapReadyDetails{true, snapName}
				cs.Driver.cgsnapreadystatusmap.Store(snapIdentifier, snapReadyDetails)
			}
		}
	}

	snapID := ""
	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		// storageclass_type;volumeType;clusterId;FSUUID;consistency_group;filesetName;snapshotName;metaSnapshotName
		snapID = fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s;%s", volumeIDMembers.StorageClassType, volumeIDMembers.VolType, volumeIDMembers.ClusterId, volumeIDMembers.FsUUID, filesetName, filesetResp.FilesetName, snapName, req.GetName())
	} else {
		if filesetResp.Config.Comment == connectors.FilesetComment &&
			(cs.Driver.primary.PrimaryFset != filesetName || cs.Driver.primary.PrimaryFs != filesystemName) {
			// Dynamically created PVC, here path is the xxx-data directory within the fileset where all volume data resides
			// storageclass_type;volumeType;clusterId;FSUUID;consistency_group;filesetName;snapshotName;metaSnapshotName;path
			snapID = fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s;%s;%s-data", volumeIDMembers.StorageClassType, volumeIDMembers.VolType, volumeIDMembers.ClusterId, volumeIDMembers.FsUUID, "", filesetName, snapName, "", filesetName)
		} else {
			// This is statically created PVC from an independent fileset, here path is the root of fileset
			// storageclass_type;volumeType;clusterId;FSUUID;consistency_group;filesetName;snapshotName;metaSnapshotName;/
			snapID = fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s;%s;/", volumeIDMembers.StorageClassType, volumeIDMembers.VolType, volumeIDMembers.ClusterId, volumeIDMembers.FsUUID, "", filesetName, snapName, "")
		}
	}

	timestamp, err := cs.getSnapshotCreateTimestamp(conn, filesystemName, filesetName, snapName)
	if err != nil {
		glog.Errorf("error getting create timestamp for snapshot %s:%s:%s", filesystemName, filesetName, snapName)
		return nil, err
	}

	restoreSize, err := cs.getSnapRestoreSize(conn, filesystemName, filesetResp.FilesetName)
	if err != nil {
		glog.Errorf("error getting the snapshot restore size for snapshot %s:%s:%s", filesystemName, filesetResp.FilesetName, snapName)
		return nil, err
	}

	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		err := cs.MakeSnapMetadataDir(conn, filesystemName, filesetResp.FilesetName, filesetName, snapName, req.GetName())
		if err != nil {
			glog.Errorf("error in creating directory for storing metadata information for advanced storageClass. Error: [%v]", err)
			return nil, err
		}
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

	createTSTZ := strings.Replace(createTS, ",000", timezoneOffset, 1)
	t, err := time.Parse(longForm, createTSTZ)
	if err != nil {
		glog.Errorf("snapshot - for fileset [%s:%s] error in parsing timestamp: [%v]. Error: [%v]", fs, fset, createTS, err)
		return timestamp, err
	}
	timestamp.Seconds = t.Unix()
	timestamp.Nanos = 0

	glog.V(3).Infof("getSnapshotCreateTimestamp: for fileset [%s:%s] snapshot creation timestamp: [%v]", fs, fset, createTSTZ)
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

func (cs *ScaleControllerServer) isExistingSnapUseableForVol(conn connectors.SpectrumScaleConnector, filesystemName string, consistencyGroup string, filesetName string, cgSnapName string) (bool, error) {
	pathDir := fmt.Sprintf("%s/.snapshots/%s/%s", consistencyGroup, cgSnapName, filesetName)
	_, err := conn.StatDirectory(filesystemName, pathDir)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist") { // directory does not exist
			return false, status.Error(codes.Internal, fmt.Sprintf("snapshot for volume [%v] in filesystem [%v] is not taken. Wait till current snapWindow expires.", filesetName, filesystemName))
		} else {
			return false, err
		}
	}
	return true, nil
}

func (cs *ScaleControllerServer) DelSnapMetadataDir(conn connectors.SpectrumScaleConnector, filesystemName string, consistencyGroup string, filesetName string, cgSnapName string, metaSnapName string) (bool, error) {
	pathDir := fmt.Sprintf("%s/%s/%s", consistencyGroup, cgSnapName, metaSnapName)
	err := conn.DeleteDirectory(filesystemName, pathDir, false)
	if err != nil {
		if !(strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist")) { // directory is already deleted
			return false, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Dir using FS [%v] at path [%v]. Error [%v]", filesystemName, pathDir, err))
		}
	}

	// Now check if consistency group snapshot metadata directory can be deleted
	pathDir = fmt.Sprintf("%s/%s", consistencyGroup, cgSnapName)
	statInfo, err := conn.StatDirectory(filesystemName, pathDir)
	if err != nil {
		if !(strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist")) { // directory is already deleted
			return false, status.Error(codes.Internal, fmt.Sprintf("unable to stat directory using FS [%v] at path [%v]. Error [%v]", filesystemName, pathDir, err))
		}
		return true, nil
	}

	statSplit := strings.Split(statInfo, "\n")
	thirdLineSplit := strings.Split(statSplit[2], " ")
	lenSplit := len(thirdLineSplit)
	linkStr := strings.TrimRight(thirdLineSplit[lenSplit-1], "\n")
	nlink, err := strconv.Atoi(linkStr)
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("invalid number of links [%v] returned in stat output for FS [%v] at path [%v]. Error [%v]", linkStr, filesystemName, pathDir, err))
	}

	glog.V(3).Infof("DelSnapMetadataDir - number of links for directory in FS [%v] at path [%v] is [%v]", filesystemName, pathDir, nlink)

	if nlink == 2 {
		// directory can be deleted
		err := conn.DeleteDirectory(filesystemName, pathDir, true)
		if err != nil {
			if !(strings.Contains(err.Error(), "EFSSG0264C") ||
				strings.Contains(err.Error(), "does not exist")) {
				return false, status.Error(codes.Internal, fmt.Sprintf("unable to delete directory for FS [%v] at path [%v]. Error: [%v]", filesystemName, pathDir, err))
			}
		}
		return true, nil
	}
	return false, nil
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

	snapIdMembers, err := cs.GetSnapIdMembers(snapID)
	if err != nil {
		glog.Errorf("invalid snapshot ID %s [%v]", snapID, err)
		return nil, err
	}

	conn, err := cs.getConnFromClusterID(snapIdMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	filesystemName, err := conn.GetFilesystemName(snapIdMembers.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get filesystem Name for Filesystem UID [%v] and clusterId [%v]. Error [%v]", snapIdMembers.FsUUID, snapIdMembers.ClusterId, err))
	}

	filesetExist := false
	if snapIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
		filesetExist, err = conn.CheckIfFilesetExist(filesystemName, snapIdMembers.ConsistencyGroup)
	} else {
		filesetExist, err = conn.CheckIfFilesetExist(filesystemName, snapIdMembers.FsetName)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the fileset %s details details. Error [%v]", snapIdMembers.FsetName, err))
	}

	//skip delete if snapshot not exist, return success
	if filesetExist {
		snapExist := false
		if snapIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
			glog.V(5).Infof("DeleteSnapshot - for advanced storageClass check if snapshot [%s] exist in fileset [%s] under filesystem [%s]", snapIdMembers.SnapName, snapIdMembers.ConsistencyGroup, filesystemName)
			chkSnapExist, err := conn.CheckIfSnapshotExist(filesystemName, snapIdMembers.ConsistencyGroup, snapIdMembers.SnapName)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the snapshot details. Error [%v]", err))
			}
			snapExist = chkSnapExist
		} else {
			glog.V(5).Infof("DeleteSnapshot - for classic storageClass check if snapshot [%s] exist in fileset [%s] under filesystem [%s]", snapIdMembers.SnapName, snapIdMembers.FsetName, filesystemName)
			chkSnapExist, err := conn.CheckIfSnapshotExist(filesystemName, snapIdMembers.FsetName, snapIdMembers.SnapName)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the snapshot details. Error [%v]", err))
			}
			snapExist = chkSnapExist
		}

		// skip delete snapshot if not exist, return success
		if snapExist {
			deleteSnapshot := true
			filesetName := snapIdMembers.FsetName
			if snapIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
				delSnap, snaperr := cs.DelSnapMetadataDir(conn, filesystemName, snapIdMembers.ConsistencyGroup, snapIdMembers.FsetName, snapIdMembers.SnapName, snapIdMembers.MetaSnapName)
				if snaperr != nil {
					glog.Errorf("DeleteSnapshot - error while deleting snapshot %v: Error: %v", snapIdMembers.SnapName, snaperr)
					return nil, snaperr
				}
				if delSnap {
					filesetName = snapIdMembers.ConsistencyGroup
					glog.V(5).Infof("DeleteSnapshot - for advanced storageClass we can delete snapshot [%s] from fileset [%s] under filesystem [%s]", snapIdMembers.SnapName, filesetName, filesystemName)
				} else {
					deleteSnapshot = false
				}
			}

			if deleteSnapshot {
				glog.V(5).Infof("DeleteSnapshot - deleting snapshot [%s] from fileset [%s] under filesystem [%s]", snapIdMembers.SnapName, filesetName, filesystemName)
				snaperr := conn.DeleteSnapshot(filesystemName, filesetName, snapIdMembers.SnapName)
				if snaperr != nil {
					glog.Errorf("DeleteSnapshot - error deleting snapshot %v: %v", snapIdMembers.SnapName, snaperr)
					return nil, snaperr
				}
				glog.V(5).Infof("DeleteSnapshot - successfully deleted snapshot [%s] from fileset [%s] under filesystem [%s]", snapIdMembers.SnapName, filesetName, filesystemName)
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
		return nil, status.Error(codes.InvalidArgument, "volume ID missing in request")
	}

	capRange := req.GetCapacityRange()
	if capRange == nil {
		return nil, status.Error(codes.InvalidArgument, "capacity range not provided")
	}

	capacity := uint64(capRange.GetRequiredBytes())

	volumeIDMembers, err := getVolIDMembers(volID)

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
		glog.Errorf("ControllerExpandVolume - unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerExpandVolume - unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
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
			glog.Errorf("unable to update the quota. Error [%v]", err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to expand the volume. Error [%v]", err))
		}
	}

	fsetDetails, err := conn.ListFileset(filesystemName, filesetName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get the fileset details. Error [%v]", err))
	}
	//check if fileset is dependent of independent\
	maxInodesCombination := []int{100096, 100352, 102400, 106496, 114688, 131072}

	if fsetDetails.Config.ParentId == 0 {
		if capacity > 10*oneGB {
			if numberInSlice(fsetDetails.Config.MaxNumInodes, maxInodesCombination) {
				opt := make(map[string]interface{})
				opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)
				fseterr := conn.UpdateFileset(filesystemName, filesetName, opt)
				if fseterr != nil {
					glog.Errorf("volume:[%v] - unable to update fileset [%v] in filesystem [%v]. Error: %v", filesetName, filesetName, filesystemName, fseterr)
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to update fileset [%v] in filesystem [%v]. Error: %v", filesetName, filesystemName, fseterr))
				}
			}
		}
	}
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(capacity),
		NodeExpansionRequired: false,
	}, nil
}

func (cs *ScaleControllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// getRemoteClusterID returns the cluster ID for the passed cluster name.
func (cs *ScaleControllerServer) getRemoteClusterID(clusterName string) (string, error) {

	glog.V(5).Infof("fetching cluster details from cache map for cluster %s", clusterName)
	clusterDetails, found := cs.Driver.clusterMap.Load(ClusterName{clusterName})
	if found {
		glog.V(5).Infof("checking if cluster details found from cache map for cluster %s has expired.", clusterName)
		if expired := checkExpiry(clusterDetails); !expired { // cluster details are not expired.
			glog.V(5).Infof("cluster details found from cache map for cluster %s are valid.", clusterName)
			return clusterDetails.(ClusterDetails).id, nil
		} else { // cluster details are expired
			glog.V(5).Infof("cluster details found from cache map for cluster %s are expired.", clusterName)
			cID := clusterDetails.(ClusterDetails).id
			conn, err := cs.getConnFromClusterID(cID)
			if err != nil {
				return "", err
			}
			clusterSummary, err := conn.GetClusterSummary()
			if err != nil {
				return "", err
			}
			cName := clusterSummary.ClusterName
			if cName == clusterName {
				glog.V(5).Infof("updating cluster details in cache map for cluster %s.", clusterName)
				cs.Driver.clusterMap.Store(ClusterName{cName}, ClusterDetails{cID, cName, time.Now(), 24})
				cs.Driver.clusterMap.Store(ClusterID{cID}, ClusterDetails{cID, cName, time.Now(), 24})
				glog.V(5).Infof("ClusterMap updated, [%s : %s]", cID, cName)
				return cID, nil
			} else {
				found = false
			}
		}
	}

	if !found {
		glog.V(5).Infof("cluster details are either expired or not found in cache map for cluster %s. Updating the cache map.", clusterName)
		scaleconfig := settings.LoadScaleConfigSettings()

		for i := range scaleconfig.Clusters {

			cID := scaleconfig.Clusters[i].ID
			glog.V(5).Infof("fetching cluster details from cache map for cluster %s", scaleconfig.Clusters[i].ID)
			clusterDetails, found := cs.Driver.clusterMap.Load(ClusterID{cID})
			if found {
				glog.V(5).Infof("checking if cluster details found from cache map for cluster %s has expired.", scaleconfig.Clusters[i].ID)
				if expired := checkExpiry(clusterDetails); !expired {
					glog.V(5).Infof("cluster details found from cache map for cluster %s are valid.", scaleconfig.Clusters[i].ID)
					cName := clusterDetails.(ClusterDetails).name
					if cName == clusterName {
						return cID, nil
					}
				} else {
					glog.V(5).Infof("cluster details found from cache map for cluster %s are expired.", scaleconfig.Clusters[i].ID)
					glog.V(5).Infof("updating cluster details in cache map for cluster %s.", scaleconfig.Clusters[i].ID)
					cName, updated := cs.updateClusterMap(cID)
					if !updated {
						continue
					}
					if cName == clusterName {
						return cID, nil
					}
				}
			} else { // if !found
				glog.V(5).Infof("cluster details not found in cache map for cluster %s.", scaleconfig.Clusters[i].ID)
				glog.V(5).Infof("adding cluster details in cache map for cluster %s.", scaleconfig.Clusters[i].ID)
				cName, updated := cs.updateClusterMap(cID)
				if !updated {
					continue
				}
				if cName == clusterName {
					return cID, nil
				}
			}
		}
	}

	return "", status.Error(codes.Internal, fmt.Sprintf("unable to get cluster ID for cluster %s", clusterName))
}

// checkExpiry returns false if cluster detials are valid.
// It returns true if cluster details have expired.
func checkExpiry(clusterDetails interface{}) bool {
	updateTime := clusterDetails.(ClusterDetails).lastupdated
	expiryDuration := clusterDetails.(ClusterDetails).expiryDuration
	if time.Since(updateTime).Hours() < float64(expiryDuration) {
		return false
	} else {
		return true
	}
}

// updateClusterMap updates the clusterMap with cluster details.
// It returns true if cache map is updated else it returns false.
func (cs *ScaleControllerServer) updateClusterMap(cID string) (string, bool) {
	glog.V(5).Infof("creating new connector for the cluster %s", cID)
	clusterConnector, err := cs.getConnFromClusterID(cID)
	// clusterConnector, err := connectors.NewSpectrumRestV2(cluster)
	if err != nil {
		glog.V(5).Infof("unable to create new connector for the cluster %s", cID)
		return "", false
	}

	clusterSummary, err := clusterConnector.GetClusterSummary()
	if err != nil {
		glog.V(5).Infof("unable to get cluster summary for cluster %s", cID)
		return "", false
	}

	cName := clusterSummary.ClusterName
	// cID = fmt.Sprint(clusterSummary.ClusterID)
	cs.Driver.clusterMap.Store(ClusterName{cName}, ClusterDetails{cID, cName, time.Now(), 24})
	cs.Driver.clusterMap.Store(ClusterID{cID}, ClusterDetails{cID, cName, time.Now(), 24})
	glog.V(5).Infof("ClusterMap updated: [%s : %s]", cID, cName)
	return cName, true
}
