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
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/klog/v2"
)

const (
	no                           = "no"
	yes                          = "yes"
	notFound                     = "NOT_FOUND"
	filesystemTypeRemote         = "remote"
	filesystemMounted            = "mounted"
	filesetUnlinkedPath          = "--"
	ResponseStatusUnknown        = "UNKNOWN"
	oneGB                 uint64 = 1024 * 1024 * 1024
	smallestVolSize       uint64 = oneGB // 1GB
	defaultSnapWindow            = "30"  // default snapWindow for Consistency Group snapshots is 30 minutes
	cgPrefixLen                  = 37
	softQuotaPercent             = 70 // This value is % of the hardQuotaLimit e.g. 70%

	discoverCGFileset         = "DISCOVER_CG_FILESET"
	discoverCGFilesetDisabled = "DISABLED"

	fsetNotFoundErrCode = "EFSSG0072C"
	fsetNotFoundErrMsg  = "400 Invalid value in 'filesetName'"

	pvcNameKey      = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"

	defaultS3Port = "443"
)

var bucketLock = make(map[string]bool)
var bucketMutex sync.Mutex

type ScaleControllerServer struct {
	Driver *ScaleDriver
}

func (cs *ScaleControllerServer) IfSameVolReqInProcess(scVol *scaleVolume) (bool, error) {
	capacity, volpresent := cs.Driver.reqmap[scVol.VolName]
	if volpresent {
		/*  #nosec G115 -- false positive  */
		if capacity == int64(scVol.VolSize) {
			return true, nil
		} else {
			return false, status.Error(codes.Internal, fmt.Sprintf("Volume %v present in map but requested size %v does not match with size %v in map", scVol.VolName, scVol.VolSize, capacity))
		}
	}
	return false, nil
}

// createLWVol: Create lightweight volume - return relative path of directory created
func (cs *ScaleControllerServer) createLWVol(ctx context.Context, scVol *scaleVolume) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] volume: [%v] - ControllerServer:createLWVol", loggerId, scVol.VolName)
	var err error

	// check if directory exist
	dirExists, err := scVol.PrimaryConnector.CheckIfFileDirPresent(ctx, scVol.VolBackendFs, scVol.VolDirBasePath)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - unable to check if DirBasePath %v is present in filesystem %v. Error : %v", loggerId, scVol.VolName, scVol.VolDirBasePath, scVol.VolBackendFs, err)
		return "", status.Error(codes.Internal, fmt.Sprintf("unable to check if DirBasePath %v is present in filesystem %v. Error : %v", scVol.VolDirBasePath, scVol.VolBackendFs, err))
	}

	if !dirExists {
		klog.Errorf("[%s] volume:[%v] - directory base path %v not present in filesystem %v", loggerId, scVol.VolName, scVol.VolDirBasePath, scVol.VolBackendFs)
		return "", status.Error(codes.Internal, fmt.Sprintf("directory base path %v not present in filesystem %v", scVol.VolDirBasePath, scVol.VolBackendFs))
	}

	// create directory in the filesystem specified in storageClass
	dirPath := fmt.Sprintf("%s/%s", scVol.VolDirBasePath, scVol.VolName)

	klog.V(4).Infof("[%s] volume: [%v] - creating directory %v", loggerId, scVol.VolName, dirPath)
	err = cs.createDirectory(ctx, scVol, scVol.VolName, dirPath)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - failed to create directory %v. Error : %v", loggerId, scVol.VolName, dirPath, err)
		return "", status.Error(codes.Internal, err.Error())
	}
	return dirPath, nil
}

//generateVolID: Generate volume ID
//VolID format for all newly created volumes (from 2.5.0 onwards):

// <storageclass_type>;<volume_type>;<cluster_id>;<filesystem_uuid>;<consistency_group>;<fileset_name>;<path>
func (cs *ScaleControllerServer) generateVolID(ctx context.Context, scVol *scaleVolume, uid string, isCGVolume, isShallowCopyVolume bool, targetPath string) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] volume: [%v] - ControllerServer:generateVolId", loggerId, scVol.VolName)
	var volID string
	var storageClassType string
	var volumeType string

	filesetName := scVol.VolName
	consistencyGroup := ""
	path := ""

	if !isShallowCopyVolume {
		if isCGVolume || scVol.VolumeType == cacheVolume {
			primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
			if !isprimaryConnPresent {
				klog.Errorf("[%s] unable to get connector for primary cluster", loggerId)
				return "", status.Error(codes.Internal, "unable to find primary cluster details in custom resource")
			}
			fsMountPoint, err := primaryConn.GetFilesystemMountDetails(ctx, scVol.LocalFS)
			if err != nil {
				return "", status.Error(codes.Internal, fmt.Sprintf("unable to get mount info for FS [%v] in cluster", scVol.LocalFS))
			}
			path = fmt.Sprintf("%s/%s", fsMountPoint.MountPoint, targetPath)
		} else {
			path = fmt.Sprintf("%s/%s", scVol.PrimarySLnkPath, scVol.VolName)
		}
	} else {
		path = targetPath
	}
	klog.V(4).Infof("[%s] volume: [%v] - ControllerServer:generateVolId: targetPath: [%v]", loggerId, scVol.VolName, path)

	if isCGVolume {
		storageClassType = STORAGECLASS_ADVANCED
		volumeType = FILE_DEPENDENTFILESET_VOLUME
		consistencyGroup = scVol.ConsistencyGroup
	} else if scVol.VolumeType == cacheVolume {
		storageClassType = STORAGECLASS_CACHE
		volumeType = FILE_INDEPENDENTFILESET_VOLUME
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

	if isShallowCopyVolume {
		volumeType = FILE_SHALLOWCOPY_VOLUME
	}

	volID = fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s", storageClassType, volumeType, scVol.ClusterId, uid, consistencyGroup, filesetName, path)
	return volID, nil
}

// getTargetPath: retrun relative volume path from filesystem mount point
func (cs *ScaleControllerServer) getTargetPath(ctx context.Context, fsetLinkPath, fsMountPoint, volumeName string, createDataDir bool, isCGVolume bool, isCacheVolume bool) (string, error) {
	if fsetLinkPath == "" || fsMountPoint == "" {
		klog.Errorf("[%s] volume:[%v] - missing details to generate target path fileset junctionpath: [%v], filesystem mount point: [%v]", utils.GetLoggerId(ctx), volumeName, fsetLinkPath, fsMountPoint)
		return "", fmt.Errorf("missing details to generate target path fileset junctionpath: [%v], filesystem mount point: [%v]", fsetLinkPath, fsMountPoint)
	}
	klog.V(4).Infof("[%s] volume: [%v] - ControllerServer:getTargetPath", utils.GetLoggerId(ctx), volumeName)
	targetPath := strings.Replace(fsetLinkPath, fsMountPoint, "", 1)
	if createDataDir && !isCGVolume && !isCacheVolume {
		targetPath = fmt.Sprintf("%s/%s-data", targetPath, volumeName)
	}
	targetPath = strings.Trim(targetPath, "!/")
	klog.V(4).Infof("[%s] ControllerServer:getTargetPath volumeName : [%s],fsetLinkPath : [%s],fsMountPoint : [%s],targetPath : [%s]", utils.GetLoggerId(ctx), volumeName, fsetLinkPath, fsMountPoint, targetPath)
	return targetPath, nil
}

// createDirectory: Create directory if not present
func (cs *ScaleControllerServer) createDirectory(ctx context.Context, scVol *scaleVolume, volName string, targetPath string) error {
	klog.Infof("[%s] volume: [%v] - ControllerServer:createDirectory", utils.GetLoggerId(ctx), volName)
	dirExists, err := scVol.Connector.CheckIfFileDirPresent(ctx, scVol.VolBackendFs, targetPath)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", utils.GetLoggerId(ctx), volName, targetPath, scVol.VolBackendFs, err)
		return fmt.Errorf("unable to check if directory path [%v] exists in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
	}

	if !dirExists {
		if scVol.VolPermissions != "" {
			err = scVol.Connector.MakeDirectoryV2(ctx, scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid, scVol.VolPermissions)
			if err != nil {
				// Directory creation failed, no cleanup will eetry in next retry
				klog.Errorf("[%s] volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", utils.GetLoggerId(ctx), volName, targetPath, scVol.VolBackendFs, err)
				return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
			}
		} else {
			err = scVol.Connector.MakeDirectory(ctx, scVol.VolBackendFs, targetPath, scVol.VolUid, scVol.VolGid)
			if err != nil {
				// Directory creation failed, no cleanup will retry in next retry
				klog.Errorf("[%s] volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", utils.GetLoggerId(ctx), volName, targetPath, scVol.VolBackendFs, err)
				return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", targetPath, scVol.VolBackendFs, err)
			}
		}
	}
	return nil
}

// createSoftlink: Create soft link if not present
func (cs *ScaleControllerServer) createSoftlink(ctx context.Context, scVol *scaleVolume, target string) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] volume: [%v] - ControllerServer:createSoftlink", loggerId, scVol.VolName)
	volSlnkPath := fmt.Sprintf("%s/%s", scVol.PrimarySLnkRelPath, scVol.VolName)
	symLinkExists, err := scVol.PrimaryConnector.CheckIfFileDirPresent(ctx, scVol.PrimaryFS, volSlnkPath)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - unable to check if symlink path [%v] exists in filesystem [%v]. Error: %v", loggerId, scVol.VolName, volSlnkPath, scVol.PrimaryFS, err)
		return fmt.Errorf("unable to check if symlink path [%v] exists in filesystem [%v]. Error: %v", volSlnkPath, scVol.PrimaryFS, err)
	}

	if !symLinkExists {
		klog.Infof("[%s] symlink info filesystem [%v] TargetFS [%v]  target Path [%v] linkPath [%v]", loggerId, scVol.PrimaryFS, scVol.LocalFS, target, volSlnkPath)
		err = scVol.PrimaryConnector.CreateSymLink(ctx, scVol.PrimaryFS, scVol.LocalFS, target, volSlnkPath)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - failed to create symlink [%v] in filesystem [%v], for target [%v] in filesystem [%v]. Error [%v]", loggerId, scVol.VolName, volSlnkPath, scVol.PrimaryFS, target, scVol.LocalFS, err)
			return fmt.Errorf("failed to create symlink [%v] in filesystem [%v], for target [%v] in filesystem [%v]. Error [%v]", volSlnkPath, scVol.PrimaryFS, target, scVol.LocalFS, err)
		}
	}
	return nil
}

// setQuota: Set quota if not set
func (cs *ScaleControllerServer) setQuota(ctx context.Context, scVol *scaleVolume, volName string) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] volume: [%v] - ControllerServer:setQuota", loggerId, volName)
	quota, err := scVol.Connector.ListFilesetQuota(ctx, scVol.VolBackendFs, volName)
	if err != nil {
		return fmt.Errorf("unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", volName, scVol.VolBackendFs, err)
	}

	filesetQuotaBytes, err := ConvertToBytes(quota)
	if err != nil {
		if strings.Contains(err.Error(), "invalid number specified") {
			// Invalid number specified means quota is not set
			filesetQuotaBytes = 0
		} else {
			return fmt.Errorf("unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", volName, scVol.VolBackendFs, err)
		}
	}

	if filesetQuotaBytes != scVol.VolSize {
		var hardLimit, softLimit string
		hardLimit = strconv.FormatUint(scVol.VolSize, 10)
		if scVol.VolumeType == cacheVolume {
			softLimit = strconv.FormatUint(uint64(math.Round(softQuotaPercent/float64(100)*float64(scVol.VolSize))), 10)
		} else {
			softLimit = hardLimit
		}

		err = scVol.Connector.SetFilesetQuota(ctx, scVol.VolBackendFs, volName, hardLimit, softLimit)
		if err != nil {
			// failed to set quota, no cleanup, next retry might be able to set quota
			return fmt.Errorf("unable to set quota [%v] on fileset [%v] of FS [%v]", scVol.VolSize, volName, scVol.VolBackendFs)
		}
	}
	return nil
}

func (cs *ScaleControllerServer) validateCG(ctx context.Context, scVol *scaleVolume) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] Validate CG for volume [%v]", loggerId, scVol)

	fsetlist, err := scVol.Connector.ListCSIIndependentFilesets(ctx, scVol.VolBackendFs)
	if err != nil {
		return "", err
	}

	var flist []string
	pvcns := scVol.ConsistencyGroup[cgPrefixLen:]

	for _, fset := range fsetlist {
		if len(fset.FilesetName) > cgPrefixLen {
			if fset.FilesetName[cgPrefixLen:] == pvcns {
				flist = append(flist, fset.FilesetName)
			}
		}
	}

	klog.Infof("[%s] Filesets with namespace [%s] as suffix: [%v]", loggerId, pvcns, flist)

	// no fileset with this namespace found
	if len(flist) == 0 {
		return scVol.ConsistencyGroup, nil
	}

	// multiple filesets with this namespace found
	if len(flist) > 1 {
		return "", status.Error(codes.Internal, fmt.Sprintf("conflicting filesets found %+v", flist))
	}

	// this is either local CG or Remote CG
	return flist[0], nil
}

// createFilesetBasedVol: Create fileset based volume  - return relative path of volume created
func (cs *ScaleControllerServer) createFilesetBasedVol(ctx context.Context, scVol *scaleVolume, isCGVolume bool, fsType string, bucketInfo map[string]string) (string, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] volume: [%v] - ControllerServer:createFilesetBasedVol", loggerId, scVol.VolName)
	opt := make(map[string]interface{})

	// fileset can not be created if filesystem is remote.
	klog.Infof("[%s] check if volumes filesystem [%v] is remote or local for cluster [%v]", loggerId, scVol.VolBackendFs, scVol.ClusterId)
	fsDetails, err := scVol.Connector.GetFilesystemDetails(ctx, scVol.VolBackendFs)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in filesystemName") {
			klog.Errorf("[%s] volume:[%v] - filesystem %s in not known to cluster %v. Error: %v", loggerId, scVol.VolName, scVol.VolBackendFs, scVol.ClusterId, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("Filesystem %s in not known to cluster %v. Error: %v", scVol.VolBackendFs, scVol.ClusterId, err))
		}
		klog.Errorf("[%s] volume:[%v] - unable to check type of filesystem [%v]. Error: %v", loggerId, scVol.VolName, scVol.VolBackendFs, err)
		return "", status.Error(codes.Internal, fmt.Sprintf("unable to check type of filesystem [%v]. Error: %v", scVol.VolBackendFs, err))
	}

	if fsDetails.Type == filesystemTypeRemote {
		klog.Errorf("[%s] volume:[%v] - filesystem [%v] is not local to cluster [%v]", loggerId, scVol.VolName, scVol.VolBackendFs, scVol.ClusterId)
		return "", status.Error(codes.Internal, fmt.Sprintf("filesystem [%v] is not local to cluster [%v]", scVol.VolBackendFs, scVol.ClusterId))
	}

	// if filesystem is remote, check it is mounted on remote GUI node.
	if cs.Driver.primary.PrimaryCid != scVol.ClusterId {
		if fsDetails.Mount.Status != filesystemMounted {
			klog.Errorf("[%s] volume:[%v] -  filesystem [%v] is [%v] on remote GUI of cluster [%v]", loggerId, scVol.VolName, scVol.VolBackendFs, fsDetails.Mount.Status, scVol.ClusterId)
			return "", status.Error(codes.Internal, fmt.Sprintf("Filesystem %v in cluster %v is not mounted", scVol.VolBackendFs, scVol.ClusterId))
		}
		klog.V(4).Infof("[%s] volume:[%v] - mount point of volume filesystem [%v] on owning cluster is %v", loggerId, scVol.VolName, scVol.VolBackendFs, fsDetails.Mount.MountPoint)
	}

	// check if quota is enabled on volume filesystem
	klog.Infof("[%s] check if quota is enabled on filesystem [%v] ", loggerId, scVol.VolBackendFs)
	if scVol.VolSize != 0 {
		klog.Infof("[%s] quota status on filesystem [%v] is [%t]", loggerId, scVol.VolBackendFs, fsDetails.Quota.FilesetdfEnabled)
		if !fsDetails.Quota.FilesetdfEnabled {
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

	if isCGVolume {
		// For new storageClass first create independent fileset if not present

		discoverCGFileset := strings.ToUpper(os.Getenv(discoverCGFileset))
		klog.Infof("[%s] discoverCGFileset is : %s", loggerId, discoverCGFileset)

		if discoverCGFileset != discoverCGFilesetDisabled && len(scVol.ConsistencyGroup) > cgPrefixLen {
			// Check for consistencyGroup
			if fsType != filesystemTypeRemote {
				newcg, err := cs.validateCG(ctx, scVol)
				if err != nil {
					klog.Errorf("ValidateCG failed. Error: %v", err)
					return "", err
				}
				scVol.ConsistencyGroup = newcg
			}
		}
		indepFilesetName := scVol.ConsistencyGroup
		klog.Infof("[%s] creating independent fileset for new storageClass with fileset name: [%v]", loggerId, indepFilesetName)
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
		_, err = cs.createFilesetVol(ctx, scVol, indepFilesetName, fsDetails, opt, createDataDir, true, isCGVolume, nil)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - failed to create independent fileset [%v] in filesystem [%v]. Error: %v", loggerId, indepFilesetName, indepFilesetName, scVol.VolBackendFs, err)
			return "", err
		}
		klog.Infof("[%s] finished creation of independent fileset for new storageClass with fileset name: [%v]", loggerId, indepFilesetName)

		// Now create dependent fileset
		klog.Infof("[%s] creating dependent fileset for new storageClass with fileset name: [%v]", loggerId, scVol.VolName)
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
		if scVol.VolPermissions != "" {
			opt[connectors.UserSpecifiedPermissions] = scVol.VolPermissions
		}

		scVol.ParentFileset = indepFilesetName
		createDataDir = true
		filesetPath, err := cs.createFilesetVol(ctx, scVol, scVol.VolName, fsDetails, opt, createDataDir, false, isCGVolume, nil)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - failed to create dependent fileset [%v] in filesystem [%v]. Error: %v", loggerId, scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
			return "", err
		}
		klog.Infof("[%s] finished creation of dependent fileset for new storageClass with fileset name: [%v]", loggerId, scVol.VolName)
		return filesetPath, nil
	} else if scVol.VolumeType == cacheVolume {
		createDataDir := false
		klog.Infof("[%s] creating a fileset for a cache volume, fileset name: [%s] in filesystem [%s]", loggerId, scVol.VolName, scVol.VolBackendFs)
		filesetPath, err := cs.createFilesetVol(ctx, scVol, scVol.VolName, fsDetails, opt, createDataDir, false, isCGVolume, bucketInfo)
		if err != nil {
			klog.Errorf("[%s] failed to create a cache fileset [%s] in filesystem [%s]. Error: %v", loggerId, scVol.VolName, scVol.VolBackendFs, err)
			return "", err
		}
		klog.Infof("[%s] finished creation of a fileset for a cache volume, fileset [%s] in filesystem [%s]", loggerId, scVol.VolName, scVol.VolBackendFs)
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
		klog.Infof("[%s] creating fileset for classic storageClass with fileset name: [%v]", loggerId, scVol.VolName)
		createDataDir := true
		filesetPath, err := cs.createFilesetVol(ctx, scVol, scVol.VolName, fsDetails, opt, createDataDir, false, isCGVolume, nil)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - failed to create fileset [%v] in filesystem [%v]. Error: %v", loggerId, scVol.VolName, scVol.VolName, scVol.VolBackendFs, err)
			return "", err
		}
		klog.Infof("[%s] finished creation of fileset for classic storageClass with fileset name: [%v]", loggerId, scVol.VolName)
		return filesetPath, nil
	}
}

func (cs *ScaleControllerServer) createFilesetVol(ctx context.Context, scVol *scaleVolume, volName string, fsDetails connectors.FileSystem_v2, opt map[string]interface{}, createDataDir bool, isCGIndependentFset bool, isCGVolume bool, bucketInfo map[string]string) (string, error) { //nolint:gocyclo,funlen
	// Check if fileset exist
	filesetInfo, err := scVol.Connector.ListFileset(ctx, scVol.VolBackendFs, volName)
	loggerId := utils.GetLoggerId(ctx)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - unable to list fileset [%v] in filesystem [%v]. Error: %v", loggerId, volName, volName, scVol.VolBackendFs, err)
		return "", status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, err))
	} else if reflect.ValueOf(filesetInfo).IsZero() {
		// This means fileset is not present, create it
		klog.V(4).Infof("[%s] createFilesetVol fileset: %s is not present in the filesystem, creating", utils.GetLoggerId(ctx), volName)

		var fseterr error
		if scVol.VolumeType == cacheVolume {
			endpoint := bucketInfo[connectors.BucketEndpoint]
			parsedURL, err := url.Parse(endpoint)
			if err != nil {
				klog.Errorf("[%s] volume:[%v] - failed to parse the endpoint URL [%s]. Error: [%v]", loggerId, volName, endpoint, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - failed to parse the endpoint URL [%s]. Error: [%v]", volName, endpoint, err))
			}
			if parsedURL.Port() == "" {
				endpoint += ":" + string(defaultS3Port)
			}
			afmTarget := endpoint + "/" + bucketInfo[connectors.BucketName]

			scheme := parsedURL.Scheme
			lockSuccess := lockBucket(loggerId, volName, afmTarget)
			if !lockSuccess {
				klog.Errorf("[%s] volume:[%v] - the bucket [%s] is already locked for another volume creation", loggerId, volName, afmTarget)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - the bucket [%s] is already locked for another volume creation", volName, afmTarget))
			} else {
				defer unlockBucket(loggerId, volName, afmTarget)
			}

			// Before creating a cache fileset, check if there is any other cache
			// fileset pointing to the same bucket, if such fileset is found then
			// disallow creation of another cache fileset.
			filesetWitAFMTarget, err := scVol.Connector.CheckFilesetWithAFMTarget(ctx, scVol.VolBackendFs, afmTarget)
			if err != nil {
				klog.Errorf("[%s] volume:[%v] - failed to get a cache fileset with bucket [%v] in filesystem [%v]. Error: [%v]", loggerId, volName, afmTarget, scVol.VolBackendFs, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("failed to get a cache fileset with bucket [%v] in filesystem [%v]. Error: [%v]", afmTarget, scVol.VolBackendFs, err))
			}
			if filesetWitAFMTarget != "" {
				klog.Errorf("[%s] volume:[%v] - failed to create an AFM cache fileset [%v] in filesystem [%v] as another fileset [%v] with the same bucket [%v] exists already", loggerId, volName, volName, scVol.VolBackendFs, filesetWitAFMTarget, afmTarget)
				return "", status.Error(codes.Internal, fmt.Sprintf("failed to create an AFM cache fileset [%v] in filesystem [%v] as another fileset [%v] with the same bucket [%v] exists already", volName, scVol.VolBackendFs, filesetWitAFMTarget, afmTarget))
			}

			// Set bucket keys for a cache volume
			keyerr := scVol.Connector.SetBucketKeys(ctx, bucketInfo)
			if keyerr != nil {
				klog.Errorf("[%s] failed to set bucket keys for volume %s", loggerId, volName)
				return "", status.Error(codes.Internal, fmt.Sprintf("failed to set bucket keys for volume %s, error: %v", volName, keyerr))
			}

			// Create a cache fileset
			if cacheFsetErr := scVol.Connector.CreateS3CacheFileset(ctx, scVol.VolBackendFs, volName, scVol.CacheMode, opt, bucketInfo, scheme); cacheFsetErr != nil {
				klog.Errorf("[%s] volume:[%v] - failed to create cache fileset [%v] in filesystem [%v]. Error: %v", loggerId, volName, volName, scVol.VolBackendFs, cacheFsetErr.Error())
				return "", status.Error(codes.Internal, fmt.Sprintf("failed to create cache fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, cacheFsetErr.Error()))
			}

			// For cache fileset, add a comment as the create COS fileset
			// interface doesn't allow setting the fileset comment.
			if err := handleUpdateComment(ctx, scVol); err != nil {
				return "", err
			}
		} else {
			// This means fileset is not present, create it
			fseterr = scVol.Connector.CreateFileset(ctx, scVol.VolBackendFs, volName, opt)
		}

		if fseterr != nil {
			// fileset creation failed return without cleanup
			klog.Errorf("[%s] volume:[%v] - unable to create fileset [%v] in filesystem [%v]. Error: %v", loggerId, volName, volName, scVol.VolBackendFs, fseterr)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to create fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, fseterr))
		}
		// list fileset and update filesetInfo
		filesetInfo, err = scVol.Connector.ListFileset(ctx, scVol.VolBackendFs, volName)
		if err != nil {
			// fileset got created but listing failed, return without cleanup
			klog.Errorf("[%s] volume:[%v] - unable to list newly created fileset [%v] in filesystem [%v]. Error: %v", loggerId, volName, volName, scVol.VolBackendFs, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to list newly created fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.VolBackendFs, err))
		}

	} else {
		// fileset is present. Confirm if creator is IBM Storage Scale CSI driver and fileset type is correct.
		if filesetInfo.Config.Comment != connectors.FilesetComment {
			if scVol.VolumeType == cacheVolume {
				if err := handleUpdateComment(ctx, scVol); err != nil {
					return "", err
				}
			} else {
				klog.Errorf("[%s] volume:[%v] - the fileset is not created by IBM Storage Scale CSI driver. Cannot use it.", loggerId, volName)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - the fileset is not created by IBM Storage Scale CSI driver. Cannot use it.", volName))
			}
		}
		if scVol.VolumeType != cacheVolume {
			listFilesetType := ""
			if filesetInfo.Config.IsInodeSpaceOwner {
				listFilesetType = independentFileset
			} else {
				listFilesetType = dependentFileset
			}
			if opt[connectors.UserSpecifiedFilesetType] != listFilesetType {
				klog.Errorf("[%s] volume:[%v] - the fileset type is not as expected, got type: [%s], expected type: [%s]", loggerId, volName, listFilesetType, opt[connectors.UserSpecifiedFilesetType])
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - the fileset type is not as expected, got type: [%s], expected type: [%s]", volName, listFilesetType, opt[connectors.UserSpecifiedFilesetType]))
			}
		}
	}

	// fileset is present/created. Confirm if fileset is linked
	if (filesetInfo.Config.Path == "") || (filesetInfo.Config.Path == filesetUnlinkedPath) {
		// this means not linked, link it
		var junctionPath string
		junctionPath = fmt.Sprintf("%s/%s", fsDetails.Mount.MountPoint, volName)

		if scVol.ParentFileset != "" {
			parentfilesetInfo, err := scVol.Connector.ListFileset(ctx, scVol.VolBackendFs, scVol.ParentFileset)
			if err != nil {
				klog.Errorf("[%s] volume:[%v] - unable to get details of parent fileset [%v] in filesystem [%v]. Error: %v", loggerId, volName, scVol.ParentFileset, scVol.VolBackendFs, err)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - unable to get details of parent fileset [%v] in filesystem [%v]. Error: %v", volName, scVol.ParentFileset, scVol.VolBackendFs, err))
			}
			if (parentfilesetInfo.Config.Path == "") || (parentfilesetInfo.Config.Path == filesetUnlinkedPath) {
				klog.Errorf("[%s] volume:[%v] - parent fileset [%v] is not linked", loggerId, volName, scVol.ParentFileset)
				return "", status.Error(codes.Internal, fmt.Sprintf("volume:[%v] - parent fileset [%v] is not linked", volName, scVol.ParentFileset))
			}
			junctionPath = fmt.Sprintf("%s/%s", parentfilesetInfo.Config.Path, volName)
		}

		err := scVol.Connector.LinkFileset(ctx, scVol.VolBackendFs, volName, junctionPath)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - linking fileset [%v] in filesystem [%v] at path [%v] failed. Error: %v", loggerId, volName, volName, scVol.VolBackendFs, junctionPath, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("linking fileset [%v] in filesystem [%v] at path [%v] failed. Error: %v", volName, scVol.VolBackendFs, junctionPath, err))
		}
		// update fileset details
		filesetInfo, err = scVol.Connector.ListFileset(ctx, scVol.VolBackendFs, volName)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - unable to list fileset [%v] in filesystem [%v] after linking. Error: %v", loggerId, volName, volName, scVol.VolBackendFs, err)
			return "", status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v] after linking. Error: %v", volName, scVol.VolBackendFs, err))
		}
	}
	targetBasePath := ""
	if !isCGIndependentFset {
		if scVol.VolSize != 0 {
			err = cs.setQuota(ctx, scVol, volName)
			if err != nil {
				return "", status.Error(codes.Internal, err.Error())
			}
		}

		isCacheVolume := false
		if scVol.VolumeType == cacheVolume {
			isCacheVolume = true
		}

		targetBasePath, err = cs.getTargetPath(ctx, filesetInfo.Config.Path, fsDetails.Mount.MountPoint, volName, createDataDir, isCGVolume, isCacheVolume)
		if err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}

		err = cs.createDirectory(ctx, scVol, volName, targetBasePath)
		if err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}

		// Create a cacheTempDir inside the fileset for all the cacheModes except ro mode.
		if scVol.VolumeType == cacheVolume && scVol.CacheMode != afmModeRO {
			err = cs.createDirectory(ctx, scVol, volName, fmt.Sprintf("%s/%s", targetBasePath, connectors.CacheTempDirName))
			if err != nil {
				return "", status.Error(codes.Internal, err.Error())
			}
		}
	}
	return targetBasePath, nil
}

func handleUpdateComment(ctx context.Context, scVol *scaleVolume) error {
	loggerId := utils.GetLoggerId(ctx)
	volName := scVol.VolName

	if updateerr := updateComment(ctx, scVol); updateerr != nil {
		if strings.Contains(updateerr.Error(), fsetNotFoundErrCode) ||
			strings.Contains(updateerr.Error(), fsetNotFoundErrMsg) {
			// Filset is not found, refresh filesets
			if err := scVol.Connector.FilesetRefreshTask(ctx); err != nil {
				klog.Errorf("[%s] failed to refresh fileset. Error: %v", loggerId, err)
				return status.Error(codes.Internal, fmt.Sprintf("failed to refresh fileset. Error: %v", err))
			}

			// Try update again after fileset refresh
			if updateerr := updateComment(ctx, scVol); updateerr != nil {
				klog.Errorf("[%s] failed to update comment for fileset [%s] in filesystem [%s] even after fileset refresh. Error: %v", loggerId, volName, scVol.VolBackendFs, updateerr)
				return status.Error(codes.Internal, fmt.Sprintf("failed to update comment for fileset [%s] in filesystem [%s] even after fileset refresh. Error: %v", volName, scVol.VolBackendFs, updateerr))
			}
		} else {
			klog.Errorf("[%s] failed to update comment for fileset [%s] in filesystem [%s]. Error: %v", loggerId, volName, scVol.VolBackendFs, updateerr)
			return status.Error(codes.Internal, fmt.Sprintf("failed to update comment for fileset [%s] in filesystem [%s]. Error: %v", volName, scVol.VolBackendFs, updateerr))
		}
	}
	return nil
}

func (cs *ScaleControllerServer) getVolumeSizeInBytes(req *csi.CreateVolumeRequest) int64 {
	capacity := req.GetCapacityRange()
	return capacity.GetRequiredBytes()
}

func updateComment(ctx context.Context, scVol *scaleVolume) error {
	updateOpts := make(map[string]interface{})
	updateOpts[connectors.FilesetComment] = connectors.FilesetComment
	return scVol.Connector.UpdateFileset(ctx, scVol.VolBackendFs, scVol.VolName, updateOpts)
}

func (cs *ScaleControllerServer) getConnFromClusterID(ctx context.Context, cid string) (connectors.SpectrumScaleConnector, error) {
	loggerId := utils.GetLoggerId(ctx)
	connector, isConnPresent := cs.Driver.connmap[cid]
	if isConnPresent {
		return connector, nil
	}
	klog.Errorf("[%s] unable to get connector for cluster ID %v", loggerId, cid)
	return nil, status.Error(codes.Internal, fmt.Sprintf("unable to find cluster [%v] details in custom resource", cid))
}

// checkSCSupportedParams checks if given CreateVolume request parameter keys
// are supported by IBM Storage Scale CSI and returns ("", true) if all parameter
// keys are supported, otherwise returns (<list of invalid keys seperated by
// comma>, false)
func checkSCSupportedParams(params map[string]string) (string, bool) {
	var invalidParams []string
	for k := range params {
		switch k {
		case "csi.storage.k8s.io/pv/name", "csi.storage.k8s.io/pvc/name",
			"csi.storage.k8s.io/pvc/namespace", "storage.kubernetes.io/csiProvisionerIdentity",
			"volBackendFs", "volDirBasePath", "uid", "gid", "permissions",
			"clusterId", "filesetType", "parentFileset", "inodeLimit", "nodeClass",
			"version", "tier", "compression", "consistencyGroup", "shared",
			"volumeType", "cacheMode", "volNamePrefix":
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

func (cs *ScaleControllerServer) getPrimaryClusterDetails(ctx context.Context) (connectors.SpectrumScaleConnector, string, string, string, string, string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] getPrimaryClusterDetails", loggerId)

	symlinkDirAbsolutePath := ""
	symlinkDirRelativePath := ""

	primaryConn := cs.Driver.connmap["primary"]
	primaryFS := cs.Driver.primary.GetPrimaryFs()
	primaryFset := cs.Driver.primary.PrimaryFset

	// check if primary filesystem exists
	fsMountInfo, err := primaryConn.GetFilesystemMountDetails(ctx, primaryFS)
	if err != nil {
		klog.Errorf("[%s] Failed to get details of primary filesystem %s", loggerId, primaryFS)
		return nil, "", "", "", "", "", err
	}

	primaryFSMount := fsMountInfo.MountPoint
	// If primary fset is not specified, then use default
	if primaryFset == "" {
		primaryFset = defaultPrimaryFileset
	}

	symlinkDirRelativePath = primaryFset + "/" + symlinkDir
	symlinkDirAbsolutePath = fsMountInfo.MountPoint + "/" + symlinkDirRelativePath
	klog.Infof("[%s] symlinkDirPath [%s], symlinkDirRelPath [%s]", loggerId, symlinkDirAbsolutePath, symlinkDirRelativePath)

	return primaryConn, symlinkDirRelativePath, primaryFS, primaryFSMount, symlinkDirAbsolutePath, cs.Driver.primary.PrimaryCid, err
}

func (cs *ScaleControllerServer) getPrimaryFSMountPoint(ctx context.Context) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] getPrimaryFSMountPoint", loggerId)

	primaryConn := cs.Driver.connmap["primary"]
	primaryFS := cs.Driver.primary.GetPrimaryFs()
	fsMountInfo, err := primaryConn.GetFilesystemMountDetails(ctx, primaryFS)
	if err != nil {
		klog.Errorf("[%s] Failed to get details of primary filesystem %s:Error: %v", loggerId, primaryFS, err)
		return "", status.Error(codes.NotFound, fmt.Sprintf("Failed to get details of primary filesystem %s. Error: %v", primaryFS, err))

	}
	return fsMountInfo.MountPoint, nil
}

// CreateVolume - Create Volume
func (cs *ScaleControllerServer) CreateVolume(newctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(newctx)
	ctx := utils.SetModuleName(newctx, createVolume)

	// Mask the secrets from request before logging
	reqToLog := *req
	reqToLog.Secrets = nil
	klog.Infof("[%s] CreateVolume req: %v", loggerId, &reqToLog)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		klog.Errorf("[%s] invalid create volume req: %v", loggerId, req)
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

		if reqCap.GetMount().GetMountFlags() != nil {
			return nil, status.Error(codes.Unimplemented, "mountOptions are not supported")
		}
	}

	invalidParams, allValid := checkSCSupportedParams(req.GetParameters())
	if !allValid {
		return nil, status.Error(codes.InvalidArgument, "The Parameter(s) not supported in storageClass: "+invalidParams)
	}

	scaleVol, isCGVolume, primaryClusterID, err := cs.setScaleVolume(ctx, req, volName, volSize)
	if err != nil {
		return nil, err
	}

	isSnapSource, isVolSource, volSrc, snapIdMembers, srcVolumeIDMembers, err := cs.getVolORSnapMembers(ctx, req, volName)
	if err != nil {
		return nil, err
	}

	// Block creating a cache volume from another volume (clone) or
	// from a snapshot (restore)
	if scaleVol.VolumeType == cacheVolume && (isSnapSource || isVolSource) {
		return nil, status.Error(codes.InvalidArgument, "Creating a cache volume from another volume or snapshot is not supported")
	}

	isShallowCopyVolume := false
	for _, reqCap := range reqCapabilities {
		if reqCap.GetAccessMode().GetMode() == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			if isSnapSource {
				klog.Infof("[%s] Requested pvc is a shallow copy volume", loggerId)
				isShallowCopyVolume = true
			} else if scaleVol.VolumeType == cacheVolume {
				if scaleVol.CacheMode == "" {
					// cacheMode is not specified, use AFM mode RO by default for volume access mode ROX
					scaleVol.CacheMode = afmModeRO
				} else if scaleVol.CacheMode != afmModeRO {
					return nil, status.Error(codes.InvalidArgument, "The volume access mode ReadOnlyMany is only supported with the cacheMode readonly")
				}
			} else {
				return nil, status.Error(codes.Unimplemented, "Volume source with Access Mode ReadOnlyMany is not supported")
			}
		} else {
			if scaleVol.VolumeType == cacheVolume {
				if scaleVol.CacheMode == "" {
					// cacheMode is not specified, use AFM mode IW by default for other volume access modes
					scaleVol.CacheMode = afmModeIW
				} else if scaleVol.CacheMode == afmModeRO {
					return nil, status.Error(codes.InvalidArgument, "The cacheMode readonly is only supported with the volume access mode ReadOnlyMany")
				}
			}
		}
	}

	if srcVolumeIDMembers.VolType != FILE_SHALLOWCOPY_VOLUME {
		err = cs.checkFileSetLink(ctx, scaleVol.PrimaryConnector, scaleVol, scaleVol.PrimaryFS, cs.Driver.primary.PrimaryFset, "primary")
		if err != nil {
			return nil, err
		}
	}

	volFsInfo, err := checkVolumeFilesystemMountOnPrimary(ctx, scaleVol)
	if err != nil {
		return nil, err
	}
	err = cs.setScaleVolumeWithRemoteCluster(ctx, scaleVol, volFsInfo, primaryClusterID)
	if err != nil {
		return nil, err
	}
	assembledScaleversion, err := cs.assembledScaleVersion(ctx, scaleVol.Connector)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("IBM Storage Scale version check for permissions failed with error %s", err))
	}
	if isCGVolume {
		if err := cs.checkCGSupport(assembledScaleversion); err != nil {
			return nil, err
		}
	}

	if isVolSource {
		err = cs.validateCloneRequest(ctx, scaleVol, &srcVolumeIDMembers, scaleVol, volFsInfo, assembledScaleversion)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - Error in source volume validation [%v]", loggerId, volName, err)
			return nil, err
		}

	}

	if isSnapSource {
		err = cs.validateSnapId(ctx, scaleVol, &snapIdMembers, scaleVol, assembledScaleversion)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - Error in source snapshot validation [%v]", loggerId, volName, err)
			return nil, err
		}

		if isShallowCopyVolume {
			err = cs.validateShallowCopyVolume(ctx, &snapIdMembers, scaleVol)
			if err != nil {
				klog.Errorf("[%s] volume:[%v] - Error in validating shallow copy volume", loggerId, volName)
				return nil, status.Error(codes.Internal, fmt.Sprintf("CreateVolume ValidateShallowCopyVolume failed: %v", err))
			}
		}

	}

	var shallowCopyTargetPath string
	if isShallowCopyVolume {
		err = cs.createSnapshotDir(ctx, &snapIdMembers, scaleVol, isCGVolume)
		if err != nil {
			return nil, err
		}

		if isCGVolume {
			shallowCopyTargetPath = fmt.Sprintf("%s/%s/.snapshots/%s/%s", volFsInfo.Mount.MountPoint, snapIdMembers.ConsistencyGroup, snapIdMembers.SnapName, snapIdMembers.FsetName)
		} else {
			shallowCopyTargetPath = fmt.Sprintf("%s/%s/.snapshots/%s/%s", volFsInfo.Mount.MountPoint, snapIdMembers.FsetName, snapIdMembers.SnapName, snapIdMembers.Path)
		}

		volID, volIDErr := cs.generateVolID(ctx, scaleVol, volFsInfo.UUID, isCGVolume, isShallowCopyVolume, shallowCopyTargetPath)
		if volIDErr != nil {
			return nil, volIDErr
		}

		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				VolumeId:      volID,
				CapacityBytes: int64(scaleVol.VolSize), // #nosec G115 -- false positive
				VolumeContext: req.GetParameters(),
				ContentSource: volSrc,
			},
		}, nil
	}

	klog.Infof("[%s] volume:[%v] -  IBM Storage Scale volume create params : %v\n", loggerId, scaleVol.VolName, scaleVol)

	if scaleVol.IsFilesetBased && scaleVol.Compression != "" {
		klog.Infof("[%s] createvolume: compression is enabled: changing volume name", loggerId)
		scaleVol.VolName = fmt.Sprintf("%s-COMPRESS%scsi", scaleVol.VolName, strings.ToUpper(scaleVol.Compression))
	}

	if scaleVol.IsFilesetBased && scaleVol.Tier != "" {
		err = cs.checkVolTierAndSetFilesystemPolicy(ctx, scaleVol, volFsInfo, primaryClusterID)
		if err != nil {
			return nil, err
		}

	}

	volReqInProcess, err := cs.IfSameVolReqInProcess(scaleVol)
	if err != nil {
		return nil, err
	}

	if volReqInProcess {
		klog.Errorf("[%s] volume:[%v] - volume creation already in process ", loggerId, scaleVol.VolName)
		return nil, status.Error(codes.Aborted, fmt.Sprintf("volume creation already in process : %v", scaleVol.VolName))
	}

	volResponse, err := cs.getCopyJobStatus(ctx, req, volSrc, scaleVol, isVolSource, isSnapSource, snapIdMembers)
	if err != nil {
		return nil, err
	} else if volResponse != nil {
		return volResponse, nil
	}

	if scaleVol.VolPermissions != "" {
		versionCheck := checkMinScaleVersionValid(assembledScaleversion, "5112")
		if !versionCheck {
			return nil, status.Error(codes.Internal, "the minimum required IBM Storage Scale version for permissions support with CSI is 5.1.1-2")
		}
	}

	/* Update driver map with new volume. Make sure to defer delete */

	cs.Driver.reqmap[scaleVol.VolName] = int64(scaleVol.VolSize) // #nosec G115 -- false positive
	defer delete(cs.Driver.reqmap, scaleVol.VolName)

	var targetPath string

	if scaleVol.VolumeType == cacheVolume {
		// Validate the secret data in case of cache volumes
		missingKeys := validateCacheSecret(req.Secrets)
		if len(missingKeys) != 0 {
			reqParams := req.GetParameters()
			return nil, status.Error(codes.Aborted, fmt.Sprintf("The secret %s/%s-secret does not have required parameter(s): %v", reqParams[pvcNamespaceKey], reqParams[pvcNameKey], missingKeys))
		}

		// A gateway node is must for cache fileset, return error if no gateway found
		gatewayPresent, err := scaleVol.Connector.CheckIfGatewayNodePresent(ctx)
		if err != nil {
			return nil, err
		}
		if !gatewayPresent {
			return nil, status.Error(codes.Aborted, fmt.Sprintf("Failed to the create a cache volume as there in no gateway node in the cluster"))
		}
	}

	if scaleVol.IsFilesetBased {
		targetPath, err = cs.createFilesetBasedVol(ctx, scaleVol, isCGVolume, volFsInfo.Type, req.Secrets)
	} else {
		targetPath, err = cs.createLWVol(ctx, scaleVol)
	}

	if err != nil {
		return nil, err
	}

	if !isCGVolume && scaleVol.VolumeType != cacheVolume {
		// Create symbolic link if not present
		err = cs.createSoftlink(ctx, scaleVol, targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	volID, volIDErr := cs.generateVolID(ctx, scaleVol, volFsInfo.UUID, isCGVolume, isShallowCopyVolume, targetPath)
	if volIDErr != nil {
		return nil, volIDErr
	}

	if isVolSource {
		if srcVolumeIDMembers.VolType == FILE_SHALLOWCOPY_VOLUME {
			err = cs.copyShallowVolumeContent(ctx, scaleVol, srcVolumeIDMembers, volFsInfo, targetPath, volID)
			if err != nil {
				klog.Errorf("[%s] CreateVolume [%s]: [%v]", loggerId, volName, err)
				return nil, err
			}
		} else {
			err = cs.copyVolumeContent(ctx, scaleVol, srcVolumeIDMembers, volFsInfo, targetPath, volID)
			if err != nil {
				klog.Errorf("[%s] CreateVolume [%s]: [%v]", loggerId, volName, err)
				return nil, err
			}
		}
	}

	if isSnapSource {
		err = cs.copySnapContent(ctx, scaleVol, snapIdMembers, volFsInfo, targetPath, volID)
		if err != nil {
			klog.Errorf("[%s] createVolume failed while copying snapshot content [%s]: [%v]", loggerId, volName, err)
			return nil, err
		}
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volID,
			CapacityBytes: int64(scaleVol.VolSize), // #nosec G115 -- false positive
			VolumeContext: req.GetParameters(),
			ContentSource: volSrc,
		},
	}, nil
}

func validateCacheSecret(secretData map[string]string) []string {
	requiredKeys := []string{"endpoint", "bucket", "accesskey", "secretkey"}
	missingKeys := []string{}
	for _, key := range requiredKeys {
		if _, exists := secretData[key]; !exists {
			missingKeys = append(missingKeys, key)
		}
	}
	return missingKeys
}

func (cs *ScaleControllerServer) setScaleVolume(ctx context.Context, req *csi.CreateVolumeRequest, volName string, volSize int64) (*scaleVolume, bool, string, error) {
	scaleVol, err := getScaleVolumeOptions(ctx, req.GetParameters())
	if err != nil {
		return nil, false, "", err
	}

	isCGVolume := false
	if scaleVol.StorageClassType == STORAGECLASS_ADVANCED {
		isCGVolume = true
	}

	scaleVol.VolName = volName
	// #nosec G115 -- false positive
	if scaleVol.IsFilesetBased && uint64(volSize) < smallestVolSize {
		scaleVol.VolSize = smallestVolSize
	} else {
		scaleVol.VolSize = uint64(volSize) // #nosec G115 -- false positive
	}

	/* Get details for Primary Cluster */
	primaryConn, symlinkDirRelativePath, primaryFS, primaryFSMount, symlinkDirAbsolutePath, primaryClusterID, err := cs.getPrimaryClusterDetails(ctx)
	if err != nil {
		return nil, isCGVolume, "", err
	}

	scaleVol.PrimaryConnector = primaryConn
	scaleVol.PrimarySLnkRelPath = symlinkDirRelativePath
	scaleVol.PrimaryFS = primaryFS
	scaleVol.PrimaryFSMount = primaryFSMount
	scaleVol.PrimarySLnkPath = symlinkDirAbsolutePath
	return scaleVol, isCGVolume, primaryClusterID, nil
}

func (cs *ScaleControllerServer) getVolORSnapMembers(ctx context.Context, req *csi.CreateVolumeRequest, volName string) (bool, bool, *csi.VolumeContentSource, scaleSnapId, scaleVolId, error) {
	loggerId := utils.GetLoggerId(ctx)
	volSrc := req.GetVolumeContentSource()
	isSnapSource := false
	isVolSource := false

	snapIdMembers := scaleSnapId{}
	srcVolumeIDMembers := scaleVolId{}
	var err error
	if volSrc != nil {
		srcVolume := volSrc.GetVolume()
		if srcVolume != nil {
			srcVolumeID := srcVolume.GetVolumeId()
			srcVolumeIDMembers, err = getVolIDMembers(srcVolumeID)
			if err != nil {
				klog.Errorf("[%s] volume:[%v] - Invalid Volume ID %s [%v]", loggerId, volName, srcVolumeID, err)
				return isSnapSource, isVolSource, volSrc, snapIdMembers, srcVolumeIDMembers, status.Error(codes.NotFound, fmt.Sprintf("volume source volume is not found: %v", err))
			}
			isVolSource = true
		} else {

			srcSnap := volSrc.GetSnapshot()
			if srcSnap != nil {
				snapId := srcSnap.GetSnapshotId()
				snapIdMembers, err = cs.GetSnapIdMembers(snapId)
				if err != nil {
					klog.Errorf("[%s] volume:[%v] - Invalid snapshot ID %s [%v]", loggerId, volName, snapId, err)
					return isSnapSource, isVolSource, volSrc, snapIdMembers, srcVolumeIDMembers, status.Error(codes.NotFound, fmt.Sprintf("volume source snapshot is not found: %v", err))
				}
				isSnapSource = true
			}
		}
	}
	return isSnapSource, isVolSource, volSrc, snapIdMembers, srcVolumeIDMembers, nil
}

func (cs *ScaleControllerServer) checkFileSetLink(ctx context.Context, connector connectors.SpectrumScaleConnector, scaleVol *scaleVolume, filesystem string, fileset string, primaryOrSource string) error {
	loggerId := utils.GetLoggerId(ctx)
	// Check if Primary Fileset is linked
	//primaryFileset := cs.Driver.primary.PrimaryFset
	klog.Infof("[%s] volume:[%v] - check if %s fileset [%v] is linked", loggerId, scaleVol.VolName, primaryOrSource, fileset)
	isFilesetLinked, err := connector.IsFilesetLinked(ctx, filesystem, fileset)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - unable to get details of %s fileset [%v]. Error : [%v]", loggerId, scaleVol.VolName, primaryOrSource, fileset, err)
		return status.Error(codes.Internal, fmt.Sprintf("unable to get details of %s fileset link information for [%v]. Error : [%v]", primaryOrSource, fileset, err))
	}
	if !isFilesetLinked {
		klog.Errorf("[%s] volume:[%s] - [%s] is not linked", loggerId, scaleVol.VolName, fileset)
		return status.Error(codes.Internal, fmt.Sprintf("%s  fileset [%v] is not linked", primaryOrSource, fileset))
	}
	return nil
}

func checkVolumeFilesystemMountOnPrimary(ctx context.Context, scaleVol *scaleVolume) (connectors.FileSystem_v2, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(6).Infof("[%s] volume:[%v] - check if volume filesystem [%v] is mounted on GUI node of Primary cluster", loggerId, scaleVol.VolName, scaleVol.VolBackendFs)
	volFsInfo, err := scaleVol.PrimaryConnector.GetFilesystemDetails(ctx, scaleVol.VolBackendFs)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in filesystemName") {
			klog.Errorf("[%s] volume:[%v] - filesystem %s in not known to primary cluster. Error: %v", loggerId, scaleVol.VolName, scaleVol.VolBackendFs, err)
			return volFsInfo, status.Error(codes.Internal, fmt.Sprintf("filesystem %s in not known to primary cluster. Error: %v", scaleVol.VolBackendFs, err))
		}
		klog.Errorf("[%s] volume:[%v] - unable to get details for filesystem [%v] in Primary cluster. Error: %v", loggerId, scaleVol.VolName, scaleVol.VolBackendFs, err)
		return volFsInfo, status.Error(codes.Internal, fmt.Sprintf("unable to get details for filesystem [%v] in Primary cluster. Error: %v", scaleVol.VolBackendFs, err))
	}

	if volFsInfo.Mount.Status != filesystemMounted {
		klog.Errorf("[%s] volume:[%v] - volume filesystem %s is not mounted on GUI node of Primary cluster", loggerId, scaleVol.VolName, scaleVol.VolBackendFs)
		return volFsInfo, status.Error(codes.Internal, fmt.Sprintf("volume filesystem %s is not mounted on GUI node of Primary cluster", scaleVol.VolBackendFs))
	}

	klog.V(6).Infof("[%s] volume:[%v] - mount point of volume filesystem [%v] is on Primary cluster is %v", loggerId, scaleVol.VolName, scaleVol.VolBackendFs, volFsInfo.Mount.MountPoint)
	return volFsInfo, nil
}

func (cs *ScaleControllerServer) setScaleVolumeWithRemoteCluster(ctx context.Context, scaleVol *scaleVolume, volFsInfo connectors.FileSystem_v2, primaryClusterID string) error {
	loggerId := utils.GetLoggerId(ctx)
	/* scaleVol.VolBackendFs will always be local cluster FS. So we need to find a
	remote cluster FS in case local cluster FS is remotely mounted. We will find local FS RemoteDeviceName on local cluster, will use that as VolBackendFs and	create fileset on that FS. */
	if scaleVol.IsFilesetBased {
		remoteDeviceName := volFsInfo.Mount.RemoteDeviceName
		scaleVol.LocalFS = scaleVol.VolBackendFs
		scaleVol.VolBackendFs = getRemoteFsName(remoteDeviceName)
	} else {
		scaleVol.LocalFS = scaleVol.VolBackendFs
	}

	// LocalFs is name of filesystem on K8s cluster
	// VolBackendFs is changed to name on remote cluster in case of fileset based provisioning

	var remoteClusterID string
	var err error
	if scaleVol.ClusterId == "" && volFsInfo.Type == filesystemTypeRemote {
		klog.Infof("[%s] filesystem %s is remotely mounted, getting cluster ID information of the owning cluster.", loggerId, volFsInfo.Name)
		clusterName := strings.Split(volFsInfo.Mount.RemoteDeviceName, ":")[0]
		if remoteClusterID, err = cs.getRemoteClusterID(ctx, clusterName); err != nil {
			klog.Errorf("[%s] error in getting remote cluster ID for cluster [%s], error [%v]", loggerId, clusterName, err)
			return err
		}
		klog.V(6).Infof("[%s] cluster ID for remote cluster %s is %s", loggerId, clusterName, remoteClusterID)
	}

	if scaleVol.IsFilesetBased {
		if scaleVol.ClusterId == "" {
			if volFsInfo.Type == filesystemTypeRemote { // if fileset based and remotely mounted.
				klog.Infof("[%s] volume filesystem %s is remotely mounted on Primary cluster, using owning cluster ID %s.", loggerId, scaleVol.LocalFS, remoteClusterID)
				scaleVol.ClusterId = remoteClusterID
			} else {
				klog.Infof("[%s] volume filesystem %s is locally mounted on Primary cluster, using primary cluster ID %s.", loggerId, scaleVol.LocalFS, primaryClusterID)
				scaleVol.ClusterId = primaryClusterID
			}
		}
		conn, err := cs.getConnFromClusterID(ctx, scaleVol.ClusterId)
		if err != nil {
			return err
		}
		scaleVol.Connector = conn
	} else {
		scaleVol.Connector = scaleVol.PrimaryConnector
		scaleVol.ClusterId = primaryClusterID
	}
	return nil
}

func (cs *ScaleControllerServer) checkVolTierAndSetFilesystemPolicy(ctx context.Context, scaleVol *scaleVolume, volFsInfo connectors.FileSystem_v2, volName string) error {
	loggerId := utils.GetLoggerId(ctx)
	if err := cs.checkVolTierSupport(volFsInfo.Version); err != nil {
		// TODO: Remove this secondary call to local gui when GUI refreshes remote cache immediately
		tempFsInfo, err := scaleVol.Connector.GetFilesystemDetails(ctx, scaleVol.VolBackendFs)
		if err != nil {
			return err
		}
		if err := cs.checkVolTierSupport(tempFsInfo.Version); err != nil {
			return err
		}
	}

	if err := scaleVol.Connector.DoesTierExist(ctx, scaleVol.Tier, scaleVol.VolBackendFs); err != nil {
		return err
	}

	rule := "RULE 'csi-T%s' SET POOL '%s' WHERE FILESET_NAME LIKE '%s-%%-T%scsi%%'"
	policy := connectors.Policy{}

	policy.Policy = fmt.Sprintf(rule, scaleVol.Tier, scaleVol.Tier, scaleVol.VolNamePrefix, scaleVol.Tier)
	klog.Infof("[%s] checkVolTierAndSetFilesystemPolicy: setting policy:[%v]", loggerId, policy.Policy)
	policy.Priority = -5
	policy.Partition = fmt.Sprintf("csi-T%s", scaleVol.Tier)

	scaleVol.VolName = fmt.Sprintf("%s-T%scsi", scaleVol.VolName, scaleVol.Tier)
	err := scaleVol.Connector.SetFilesystemPolicy(ctx, &policy, scaleVol.VolBackendFs)
	if err != nil {
		klog.Errorf("[%s] volume:[%v] - setting policy failed [%v]", loggerId, volName, err)
		return err
	}

	// Since we are using a SET POOL rule, if there is not already a default rule in place in the policy partition
	// then all files that do not match our rules will have no defined place to go. This sets a default rule with
	// "lower" priority than the main policy as a catch all. If there is already a default rule in the main policy
	// file then that will take precedence
	defaultPartitionName := "csi-defaultRule"
	if !scaleVol.Connector.CheckIfDefaultPolicyPartitionExists(ctx, defaultPartitionName, scaleVol.VolBackendFs) {
		klog.Infof("[%s] createvolume: setting default policy partition rule", loggerId)

		dataTierName, err := scaleVol.Connector.GetFirstDataTier(ctx, scaleVol.VolBackendFs)
		if err != nil {
			return status.Error(codes.Unavailable, fmt.Sprintf("tier info request could not be completed: filesystemName %s", scaleVol.VolBackendFs))
		}
		defaultPolicy := connectors.Policy{}
		defaultPolicy.Policy = fmt.Sprintf("RULE 'csi-defaultRule' SET POOL '%s'", dataTierName)
		defaultPolicy.Priority = 5
		defaultPolicy.Partition = defaultPartitionName
		err = scaleVol.Connector.SetFilesystemPolicy(ctx, &defaultPolicy, scaleVol.VolBackendFs)
		if err != nil {
			klog.Errorf("[%s] volume:[%v] - setting default policy failed [%v]", loggerId, volName, err)
			return err
		}
	}
	return nil
}

func (cs *ScaleControllerServer) getCopyJobStatus(ctx context.Context, req *csi.CreateVolumeRequest, volSrc *csi.VolumeContentSource, scaleVol *scaleVolume, isVolSource bool, isSnapSource bool, snapIdMembers scaleSnapId) (*csi.CreateVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	if isVolSource {
		jobDetails, found := cs.Driver.volcopyjobstatusmap.Load(scaleVol.VolName)
		if found {
			jobStatus := jobDetails.(VolCopyJobDetails).jobStatus
			volID := jobDetails.(VolCopyJobDetails).volID
			klog.V(6).Infof("[%s] volume: [%v] found in volcopyjobstatusmap with volID: [%v], jobStatus: [%v]", loggerId, scaleVol.VolName, volID, jobStatus)
			switch jobStatus {
			case VOLCOPY_JOB_RUNNING:
				klog.Errorf("[%s] volume:[%v] -  volume cloning request in progress.", loggerId, scaleVol.VolName)
				return nil, status.Error(codes.Aborted, fmt.Sprintf("volume cloning request in progress for volume: %s", scaleVol.VolName))
			case VOLCOPY_JOB_FAILED:
				//Delete the entry from map, so that it is retried
				klog.Errorf("[%s] volume:[%v] -  volume cloning job had failed and it will be retried", loggerId, scaleVol.VolName)
				cs.Driver.volcopyjobstatusmap.Delete(scaleVol.VolName)
				return nil, status.Error(codes.Internal, fmt.Sprintf("volume cloning job had failed for volume:[%v] and it will be retried", scaleVol.VolName))
			case VOLCOPY_JOB_COMPLETED:
				klog.Infof("[%s] volume:[%v] -  volume cloning request has already completed successfully.", loggerId, scaleVol.VolName)
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{
						VolumeId:      volID,
						CapacityBytes: int64(scaleVol.VolSize), // #nosec G115 --  false positive
						VolumeContext: req.GetParameters(),
						ContentSource: volSrc,
					},
				}, nil
			case JOB_STATUS_UNKNOWN:
				//Remove the entry from map, so that it can be retried
				klog.Infof("[%s] volume:[%v] -  the status of volume cloning job is unknown.", loggerId, scaleVol.VolName)
				cs.Driver.volcopyjobstatusmap.Delete(scaleVol.VolName)
			}
		} else {
			klog.Infof("[%s] volume: [%v] not found in volcopyjobstatusmap", loggerId, scaleVol.VolName)
		}
	}

	if isSnapSource {
		jobDetails, found := cs.Driver.snapjobstatusmap.Load(scaleVol.VolName)
		if found {
			jobStatus := jobDetails.(SnapCopyJobDetails).jobStatus
			volID := jobDetails.(SnapCopyJobDetails).volID
			klog.V(6).Infof("[%s] volume: [%v] found in snapjobstatusmap with volID: [%v], jobStatus: [%v]", loggerId, scaleVol.VolName, volID, jobStatus)
			switch jobStatus {
			case SNAP_JOB_RUNNING:
				klog.Errorf("[%s] volume:[%v] -  snapshot copy request in progress for snapshot: %s.", loggerId, scaleVol.VolName, snapIdMembers.SnapName)
				return nil, status.Error(codes.Aborted, fmt.Sprintf("snapshot copy request in progress for snapshot: %s", snapIdMembers.SnapName))
			case SNAP_JOB_FAILED:
				klog.Errorf("[%s] volume:[%v] -  snapshot copy job had failed for snapshot %s and it will be retried", loggerId, scaleVol.VolName, snapIdMembers.SnapName)
				//Delete the entry from map, so that it is retried
				cs.Driver.snapjobstatusmap.Delete(scaleVol.VolName)
				return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot copy job had failed for snapshot: %s and it will be retried", snapIdMembers.SnapName))
			case SNAP_JOB_COMPLETED:
				klog.V(6).Infof("[%s] volume:[%v] -  snapshot copy request has already completed successfully for snapshot: %s", loggerId, scaleVol.VolName, snapIdMembers.SnapName)
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{
						VolumeId:      volID,
						CapacityBytes: int64(scaleVol.VolSize), // #nosec G115 --  false positive
						VolumeContext: req.GetParameters(),
						ContentSource: volSrc,
					},
				}, nil
			case JOB_STATUS_UNKNOWN:
				//Remove the entry from map, so that it can be retried
				klog.V(6).Infof("[%s] volume:[%v] -  the status of snapshot copy job for snapshot [%s] is unknown", loggerId, scaleVol.VolName, snapIdMembers.SnapName)
				cs.Driver.snapjobstatusmap.Delete(scaleVol.VolName)
			}
		} else {
			klog.V(6).Infof("[%s] volume: [%v] not found in snapjobstatusmap", loggerId, scaleVol.VolName)
		}
	}
	return nil, nil
}
func (cs *ScaleControllerServer) copySnapContent(ctx context.Context, scVol *scaleVolume, snapId scaleSnapId, fsDetails connectors.FileSystem_v2, targetPath string, volID string) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] copySnapContent snapId: [%v], scaleVolume: [%v]", loggerId, snapId, scVol)
	conn, err := cs.getConnFromClusterID(ctx, snapId.ClusterId)
	if err != nil {
		return err
	}

	//err = cs.validateRemoteFs(fsDetails, scVol)
	//if err != nil {
	//	return err
	//}

	targetFsName, err := conn.GetFilesystemName(ctx, fsDetails.UUID)
	if err != nil {
		return err
	}

	targetFsDetails, err := conn.GetFilesystemDetails(ctx, targetFsName)
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
	jobStatus, jobID, err := conn.CopyFsetSnapshotPath(ctx, snapId.FsName, filesetForCopy, snapId.SnapName, snapIDPath, targetPath, scVol.NodeClass)
	if err != nil {
		klog.Errorf("[%s] failed to create volume from snapshot %s: [%v]", loggerId, snapId.SnapName, err)
		return status.Error(codes.Internal, fmt.Sprintf("failed to create volume from snapshot %s: [%v]", snapId.SnapName, err))

	}

	jobDetails := SnapCopyJobDetails{SNAP_JOB_RUNNING, volID}
	cs.Driver.snapjobstatusmap.Store(scVol.VolName, jobDetails)

	isResponseStatusUnknown := false
	response, err := conn.WaitForJobCompletionWithResp(ctx, jobStatus, jobID)
	if len(response.Jobs) != 0 {
		if response.Jobs[0].Status == ResponseStatusUnknown {
			isResponseStatusUnknown = true
		}
	}
	if err != nil || isResponseStatusUnknown {
		klog.Errorf("[%s] unable to copy snapshot %s: %v.", loggerId, snapId.SnapName, err)
		if err != nil && strings.Contains(err.Error(), "EFSSG0632C") {
			//TODO: When the GUI issue https://jazz07.rchland.ibm.com:21443/jazz/web/projects/GPFS#action=com.ibm.team.workitem.viewWorkItem&id=300263
			// is fixed, check whether the err.Error() says mmxcp is already running for the same
			// source and destination and then set the job status as SNAP_JOB_RUNNING, so that
			// mmxcp is not run again for the same source and destination.

			// EFSSG0632C = Command execution aborted
			// Store SNAP_JOB_NOT_STARTED in snapjobstatusmap if error was due to same mmxcp in progress
			// or max no. of mmxcp already running. In these cases we want to retry again
			// in the next k8s rety cycle
			jobDetails.jobStatus = SNAP_JOB_NOT_STARTED
		} else if isResponseStatusUnknown {
			jobDetails.jobStatus = JOB_STATUS_UNKNOWN
		} else {
			jobDetails.jobStatus = SNAP_JOB_FAILED
		}
		cs.Driver.snapjobstatusmap.Store(scVol.VolName, jobDetails)
		return err
	}

	klog.Infof("[%s] copy snapshot completed for snapId: [%v], scaleVolume: [%v]", loggerId, snapId, scVol)
	jobDetails.jobStatus = SNAP_JOB_COMPLETED
	cs.Driver.snapjobstatusmap.Store(scVol.VolName, jobDetails)
	//delete(cs.Driver.snapjobmap, scVol.VolName)
	return nil
}

func (cs *ScaleControllerServer) copyShallowVolumeContent(ctx context.Context, newvolume *scaleVolume, sourcevolume scaleVolId, fsDetails connectors.FileSystem_v2, targetPath string, volID string) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] copyShallowVolContent volume ID: [%v], scaleVolume: [%v], volume name: [%v]", loggerId, sourcevolume, newvolume, newvolume.VolName)
	conn, err := cs.getConnFromClusterID(ctx, sourcevolume.ClusterId)
	if err != nil {
		return err
	}

	fsMntPt := fsDetails.Mount.MountPoint
	targetPath = fmt.Sprintf("%s/%s", fsMntPt, targetPath)

	jobDetails := VolCopyJobDetails{VOLCOPY_JOB_NOT_STARTED, volID}
	response := connectors.GenericResponse{}

	sLinkRelPath := strings.Replace(sourcevolume.Path, fsMntPt, "", 1)
	sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

	if fsDetails.Type == filesystemTypeRemote {
		remotefsDetails, err := conn.GetFilesystemDetails(ctx, newvolume.VolBackendFs)
		if err != nil {
			if strings.Contains(err.Error(), "Invalid value in filesystemName") {
				klog.Errorf("[%s] filesystem %s in not known to cluster %s. Error: %v", loggerId, newvolume.VolBackendFs, newvolume.ClusterId, err)
				return status.Error(codes.Internal, fmt.Sprintf("Filesystem %s in not known to cluster %s. Error: %v", newvolume.VolBackendFs, newvolume.ClusterId, err))
			}
			klog.Errorf("[%s] unable to check type of filesystem [%s]. Error: %v", loggerId, newvolume.VolBackendFs, err)
			return status.Error(codes.Internal, fmt.Sprintf("unable to check type of filesystem [%s]. Error: %v", newvolume.VolBackendFs, err))
		}
		remoteMntPt := remotefsDetails.Mount.MountPoint
		targetPath = strings.Replace(targetPath, fsMntPt, remoteMntPt, 1)
	}

	jobStatus, jobID, jobErr := conn.CopyDirectoryPath(ctx, sourcevolume.FsName, sLinkRelPath, targetPath, newvolume.NodeClass)

	if jobErr != nil {
		klog.Errorf("[%s] failed to clone volume from volume. Error: [%v]", loggerId, jobErr)
		return status.Error(codes.Internal, fmt.Sprintf("failed to clone volume from shallow copy volume. Error: [%v]", jobErr))
	}

	jobDetails = VolCopyJobDetails{VOLCOPY_JOB_RUNNING, volID}
	cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
	response, err = conn.WaitForJobCompletionWithResp(ctx, jobStatus, jobID)
	if err != nil {
		klog.Errorf("[%s] failed while calling WaitForJobCompletionWithResp: %v.", loggerId, err)
	}

	isResponseStatusUnknown := false
	if len(response.Jobs) != 0 {
		if response.Jobs[0].Status == ResponseStatusUnknown {
			isResponseStatusUnknown = true
		}
	}

	if err != nil || isResponseStatusUnknown {
		klog.Errorf("[%s] unable to clone shallow copy volume: %v.", loggerId, err)
		if err != nil && strings.Contains(err.Error(), "EFSSG0632C") {
			jobDetails.jobStatus = VOLCOPY_JOB_NOT_STARTED
		} else if isResponseStatusUnknown {
			jobDetails.jobStatus = JOB_STATUS_UNKNOWN
		} else {
			jobDetails.jobStatus = VOLCOPY_JOB_FAILED
		}
		klog.Errorf("[%s] logging volume cloning error for VolName: [%s] Error: [%v] JobDetails: [%v]", loggerId, newvolume.VolName, err, jobDetails)
		cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
		return err
	}

	klog.Infof("[%s] volume copy completed for volumeID: [%v], scaleVolume: [%v]", loggerId, sourcevolume, newvolume)
	jobDetails.jobStatus = VOLCOPY_JOB_COMPLETED
	cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
	return nil

}

func (cs *ScaleControllerServer) copyVolumeContent(ctx context.Context, newvolume *scaleVolume, sourcevolume scaleVolId, fsDetails connectors.FileSystem_v2, targetPath string, volID string) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] copyVolContent volume ID: [%v], scaleVolume: [%v], volume name: [%v]", loggerId, sourcevolume, newvolume, newvolume.VolName)
	conn, err := cs.getConnFromClusterID(ctx, sourcevolume.ClusterId)
	if err != nil {
		return err
	}

	// err = cs.validateRemoteFs(fsDetails, scVol)
	// if err != nil {
	// 	return err
	// }

	targetFsName, err := conn.GetFilesystemName(ctx, fsDetails.UUID)
	if err != nil {
		return err
	}

	targetFsDetails, err := conn.GetFilesystemDetails(ctx, targetFsName)
	if err != nil {
		return err
	}

	fsMntPt := targetFsDetails.Mount.MountPoint
	targetPath = fmt.Sprintf("%s/%s", fsMntPt, targetPath)

	jobDetails := VolCopyJobDetails{VOLCOPY_JOB_NOT_STARTED, volID}
	response := connectors.GenericResponse{}
	if newvolume.IsFilesetBased {
		path := ""
		if sourcevolume.StorageClassType == STORAGECLASS_ADVANCED {
			path = "/"
		} else {
			path = fmt.Sprintf("%s%s", sourcevolume.FsetName, "-data")
		}

		jobStatus, jobID, jobErr := conn.CopyFilesetPath(ctx, sourcevolume.FsName, sourcevolume.FsetName, path, targetPath, newvolume.NodeClass)
		if jobErr != nil {
			klog.Errorf("[%s] failed to clone volume from volume. Error: [%v]", loggerId, jobErr)
			return status.Error(codes.Internal, fmt.Sprintf("failed to clone volume from volume. Error: [%v]", jobErr))
		}

		jobDetails = VolCopyJobDetails{VOLCOPY_JOB_RUNNING, volID}
		cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
		response, err = conn.WaitForJobCompletionWithResp(ctx, jobStatus, jobID)
	} else {
		primaryFSMountPoint, err := cs.getPrimaryFSMountPoint(ctx)
		if err != nil {
			return err
		}
		sLinkRelPath := strings.Replace(sourcevolume.Path, primaryFSMountPoint, "", 1)
		sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

		jobStatus, jobID, jobErr := conn.CopyDirectoryPath(ctx, sourcevolume.FsName, sLinkRelPath, targetPath, newvolume.NodeClass)

		if jobErr != nil {
			klog.Errorf("[%s] failed to clone volume from volume. Error: [%v]", loggerId, jobErr)
			return status.Error(codes.Internal, fmt.Sprintf("failed to clone volume from volume. Error: [%v]", jobErr))
		}

		jobDetails = VolCopyJobDetails{VOLCOPY_JOB_RUNNING, volID}
		cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
		response, err = conn.WaitForJobCompletionWithResp(ctx, jobStatus, jobID)
		if err != nil {
			klog.Errorf("[%s] failed while calling WaitForJobCompletionWithResp: %v.", loggerId, err)
		}
	}
	isResponseStatusUnknown := false
	if len(response.Jobs) != 0 {
		if response.Jobs[0].Status == ResponseStatusUnknown {
			isResponseStatusUnknown = true
		}
	}
	if err != nil || isResponseStatusUnknown {
		klog.Errorf("[%s] unable to copy volume: %v.", loggerId, err)
		if err != nil && strings.Contains(err.Error(), "EFSSG0632C") {
			//TODO: When the GUI issue https://jazz07.rchland.ibm.com:21443/jazz/web/projects/GPFS#action=com.ibm.team.workitem.viewWorkItem&id=300263
			// is fixed, check whether the err.Error() says mmxcp is already running for the same
			// source and destination and then set the job status as VOLCOPY_JOB_RUNNING, so that
			// mmxcp is not run again for the same source and destination.

			// EFSSG0632C = Command execution aborted
			// Store VOLCOPY_JOB_NOT_STARTED in volcopyjobstatusmap if error was due to same mmxcp in progress
			// or max no. of mmxcp already running. In these cases we want to retry again
			// in the next k8s rety cycle
			jobDetails.jobStatus = VOLCOPY_JOB_NOT_STARTED
		} else if isResponseStatusUnknown {
			jobDetails.jobStatus = JOB_STATUS_UNKNOWN
		} else {
			jobDetails.jobStatus = VOLCOPY_JOB_FAILED
		}
		klog.Errorf("[%s] logging volume cloning error for VolName: [%v] Error: [%v] JobDetails: [%v]", loggerId, newvolume.VolName, err, jobDetails)
		cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
		return err
	}

	klog.Infof("[%s] volume copy completed for volumeID: [%v], scaleVolume: [%v]", loggerId, sourcevolume, newvolume)
	jobDetails.jobStatus = VOLCOPY_JOB_COMPLETED
	cs.Driver.volcopyjobstatusmap.Store(newvolume.VolName, jobDetails)
	//delete(cs.Driver.volcopyjobstatusmap, scVol.VolName)
	return nil
}

func (cs *ScaleControllerServer) assembledScaleVersion(ctx context.Context, conn connectors.SpectrumScaleConnector) (string, error) {
	assembledScaleVer := ""
	scaleVersion, err := conn.GetScaleVersion(ctx)
	if err != nil {
		return assembledScaleVer, err
	}
	/* Assuming IBM Storage Scale version is in a format like 5.0.0-0_170818.165000 */
	// "serverVersion" : "5.1.1.1-developer build",
	splitScaleVer := strings.Split(scaleVersion, ".")
	if len(splitScaleVer) < 3 {
		return assembledScaleVer, status.Error(codes.Internal, fmt.Sprintf("invalid IBM Storage Scale version - %s", scaleVersion))
	}
	var splitMinorVer []string
	if len(splitScaleVer) == 4 {
		//dev build e.g. "5.1.5.0-developer build"
		splitMinorVer = strings.Split(splitScaleVer[3], "-")
		assembledScaleVer = splitScaleVer[0] + splitScaleVer[1] + splitScaleVer[2] + splitMinorVer[0]
	} else {
		//GA build e.g. "5.1.5-0"
		splitMinorVer = strings.Split(splitScaleVer[2], "-")
		assembledScaleVer = splitScaleVer[0] + splitScaleVer[1] + splitMinorVer[0] + splitMinorVer[1][0:1]
	}
	return assembledScaleVer, nil
}

func checkMinScaleVersionValid(assembledScaleVer string, version string) bool {
	return assembledScaleVer >= version
}

func (cs *ScaleControllerServer) checkMinFsVersion(fsVersion string, version string) bool {
	/* Assuming Filesystem version (fsVersion) in a format like 27.00 and version as 2700 */
	assembledFsVer := strings.ReplaceAll(fsVersion, ".", "")

	klog.Infof("fs version (%s) vs min required version (%s)", assembledFsVer, version)
	return assembledFsVer >= version
}

func (cs *ScaleControllerServer) checkSnapshotSupport(assembledScaleversion string) error {
	/* Verify IBM Storage Scale Version is not below 5.1.1-0 */
	versionCheck := checkMinScaleVersionValid(assembledScaleversion, "5110")
	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required IBM Storage Scale version for snapshot support with CSI is 5.1.1-0")
	}
	return nil
}

func (cs *ScaleControllerServer) checkVolCloneSupport(assembledScaleversion string) error {
	/* Verify IBM Storage Scale Version is not below 5.1.2-1 */
	versionCheck := checkMinScaleVersionValid(assembledScaleversion, "5121")
	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required IBM Storage Scale version for volume cloning support with CSI is 5.1.2-1")
	}
	return nil
}

func (cs *ScaleControllerServer) checkVolTierSupport(version string) error {
	/* Verify IBM Storage Scale Filesystem Version is not below 5.1.3-0 (27.00) */

	versionCheck := cs.checkMinFsVersion(version, "2700")

	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required IBM Storage Scale Filesystem version for tiering support with CSI is 27.00 (5.1.3-0)")
	}
	return nil
}

func (cs *ScaleControllerServer) checkCGSupport(assembledScaleversion string) error {
	/* Verify IBM Storage Scale Version is not below 5.1.3-0 */
	versionCheck := checkMinScaleVersionValid(assembledScaleversion, "5130")
	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required IBM Storage Scale version for consistency group support with CSI is 5.1.3-0")
	}
	return nil
}

/*func (cs *ScaleControllerServer) checkGuiHASupport(ctx context.Context, conn connectors.SpectrumScaleConnector) error {
	  // Verify IBM Storage Scale Version is not below 5.1.5-0

	  versionCheck, err := cs.checkMinScaleVersion(ctx, conn, "5150")
	  if err != nil {
		  return err
	  }

	  if !versionCheck {
		  return status.Error(codes.FailedPrecondition, "the minimum required IBM Storage Scale version for GUI HA support with CSI is 5.1.5-0")
	  }
	  return nil
  }*/

func (cs *ScaleControllerServer) validateSnapId(ctx context.Context, scaleVol *scaleVolume, sourcesnapshot *scaleSnapId, newvolume *scaleVolume, assembledScaleversion string) error {

	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] validateSnapId [%v]", loggerId, sourcesnapshot)
	conn, err := cs.getConnFromClusterID(ctx, sourcesnapshot.ClusterId)
	if err != nil {
		return err
	}

	// Restrict cross cluster cloning
	if newvolume.ClusterId != sourcesnapshot.ClusterId {
		return status.Error(codes.Unimplemented, "creating volume from snapshot across clusters is not supported")
	}

	// Restrict cross storage class version volume from snapshot
	// if len(newvolume.StorageClassType) != 0 || len(sourcesnapshot.StorageClassType) != 0 {
	// 	if newvolume.StorageClassType != sourcesnapshot.StorageClassType {
	// 		return status.Error(codes.Unimplemented, "creating volume from snapshot between different version of storageClass is not supported")
	// 	}
	// }

	// Restrict creating LW volume from snapshot
	// if !newvolume.IsFilesetBased {
	// 	return status.Error(codes.Unimplemented, "creating lightweight volume from snapshot is not supported")
	// }

	// // Restrict creating dependent fileset based volume from snapshot
	// if newvolume.StorageClassType == STORAGECLASS_CLASSIC && newvolume.FilesetType == dependentFileset {
	// 	return status.Error(codes.Unimplemented, "creating dependent fileset based volume from snapshot is not supported")
	// }

	/* Check if IBM Storage Scale supports Snapshot */
	chkSnapshotErr := cs.checkSnapshotSupport(assembledScaleversion)
	if chkSnapshotErr != nil {
		return chkSnapshotErr
	}

	if newvolume.NodeClass != "" {
		isValidNodeclass, err := conn.IsValidNodeclass(ctx, newvolume.NodeClass)
		if err != nil {
			return err
		}

		if !isValidNodeclass {
			return status.Error(codes.NotFound, fmt.Sprintf("nodeclass [%s] not found on cluster [%v]", newvolume.NodeClass, newvolume.ClusterId))
		}
	}

	sourcesnapshot.FsName, err = conn.GetFilesystemName(ctx, sourcesnapshot.FsUUID)

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", sourcesnapshot.FsUUID, sourcesnapshot.ClusterId, err))
	}

	filesetToCheck := sourcesnapshot.FsetName
	if sourcesnapshot.StorageClassType == STORAGECLASS_ADVANCED {
		filesetToCheck = sourcesnapshot.ConsistencyGroup
	}
	/*isFsetLinked, err := conn.IsFilesetLinked(ctx, sourcesnapshot.FsName, filesetToCheck)
	  if err != nil {
		  return status.Error(codes.Internal, fmt.Sprintf("unable to get fileset link information for [%v]", filesetToCheck))
	  }
	  if !isFsetLinked {
		  return status.Error(codes.Internal, fmt.Sprintf("fileset [%v] of source snapshot is not linked", filesetToCheck))
	  }*/

	err = cs.checkFileSetLink(ctx, conn, scaleVol, sourcesnapshot.FsName, filesetToCheck, "source snapshot")
	if err != nil {
		return err
	}

	isSnapExist, err := conn.CheckIfSnapshotExist(ctx, sourcesnapshot.FsName, filesetToCheck, sourcesnapshot.SnapName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get snapshot information for [%v]", sourcesnapshot.SnapName))
	}
	if !isSnapExist {
		return status.Error(codes.Internal, fmt.Sprintf("snapshot [%v] does not exist for fileset [%v]", sourcesnapshot.SnapName, filesetToCheck))
	}

	return nil
}

func (cs *ScaleControllerServer) validateShallowCopyVolume(ctx context.Context, sourcesnapshot *scaleSnapId, newvolume *scaleVolume) error {
	loggerId := utils.GetLoggerId(ctx)

	if sourcesnapshot.VolType == "" && sourcesnapshot.ConsistencyGroup == "" {
		klog.Errorf("[%s] creating shallow copy volume is not supported for static volume or old snapshot handle", loggerId)
		return status.Error(codes.Internal, "creating shallow copy volume is not supported for static volume or old snapshot handle")
	}

	if !newvolume.IsFilesetBased {
		klog.Errorf("[%s] creating shallow copy volume as directory based volume is not supported", loggerId)
		return status.Error(codes.Internal, "creating shallow copy volume as directory based volume is not supported")
	}

	if newvolume.ClusterId != sourcesnapshot.ClusterId {
		klog.Errorf("[%s] shallow copy volume across clusters is not supported", loggerId)
		return status.Error(codes.Internal, "shallow copy volume across clusters is not supported")
	}

	if len(newvolume.StorageClassType) != 0 || len(sourcesnapshot.StorageClassType) != 0 {
		if newvolume.StorageClassType != sourcesnapshot.StorageClassType {
			klog.Errorf("[%s] validation of shallow copy volume [%s] failed as storage class type is different from source pvc [%s]", loggerId, newvolume.VolName, sourcesnapshot.SnapName)
			return status.Error(codes.Internal, fmt.Sprintf("validation of shallow copy volume [%s] failed as storage class type is different from source pvc [%s]", newvolume.VolName, sourcesnapshot.SnapName))
		} else {
			if newvolume.VolBackendFs != sourcesnapshot.FsName {
				klog.Errorf("[%s] validation of shallow copy volume [%s] failed as filesystem [%s] is different from source pvc [%s] failed ", loggerId, newvolume.VolName,
					newvolume.VolBackendFs, sourcesnapshot.SnapName)
				return status.Error(codes.Internal, fmt.Sprintf("validation of shallow copy volume [%s] failed as filesystem [%s] is different from source pvc [%s] failed", newvolume.VolName, newvolume.VolBackendFs, sourcesnapshot.SnapName))
			} else {
				if sourcesnapshot.StorageClassType == STORAGECLASS_CLASSIC {
					if !((newvolume.FilesetType == independentFileset && sourcesnapshot.VolType == FILE_INDEPENDENTFILESET_VOLUME) || (newvolume.FilesetType == dependentFileset && sourcesnapshot.VolType == FILE_DEPENDENTFILESET_VOLUME)) {
						klog.Errorf("[%s] Filesettype is not same for both source snapshot and new volume", loggerId)
						return status.Error(codes.Internal, "Filesettype is not same for both source snapshot and new volume")
					}

				}
			}
		}
	}
	return nil
}

func (cs *ScaleControllerServer) createSnapshotDir(ctx context.Context, sourcesnapshot *scaleSnapId, newvolume *scaleVolume, isCGVolume bool) error {
	loggerId := utils.GetLoggerId(ctx)
	var snapshotPath string

	if isCGVolume {
		snapshotPath = fmt.Sprintf("%s/%s", sourcesnapshot.ConsistencyGroup, sourcesnapshot.SnapName)
	} else {
		snapshotPath = fmt.Sprintf("%s/%s", sourcesnapshot.FsetName, sourcesnapshot.SnapName)
	}
	shallowCopyPath := fmt.Sprintf("%s/%s", snapshotPath, newvolume.VolName)

	if isCGVolume {
		klog.Infof("[%s] Target path in createSnapshotDir:[%s]", loggerId, snapshotPath)
		lockSuccess := CgSnapshotLock(ctx, snapshotPath, false)
		if !lockSuccess {
			message := fmt.Sprintf("create snapshot failed to acquire the lock as another operation is in progress for the same targetPath: [%s]", snapshotPath)
			klog.Errorf("[%s] %s", loggerId, message)
			return status.Error(codes.Internal, message)
		} else {
			defer CgSnapshotUnlock(ctx, snapshotPath)
		}
	}

	klog.Infof("[%s] createSnapshotDir reference path [%s] for shallow copy volume: [%s]", loggerId, shallowCopyPath, newvolume.VolName)
	err := cs.createDirectory(ctx, newvolume, newvolume.VolName, shallowCopyPath)
	if err != nil {
		klog.Errorf("[%s] Failed to create snapshot reference directory", loggerId)
		return err
	}
	return nil
}

func (cs *ScaleControllerServer) validateCloneRequest(ctx context.Context, scaleVol *scaleVolume, sourcevolume *scaleVolId, newvolume *scaleVolume, volFsInfo connectors.FileSystem_v2, assembledScaleversion string) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] validateVolId [%v]", loggerId, sourcevolume)

	conn, err := cs.getConnFromClusterID(ctx, sourcevolume.ClusterId)
	if err != nil {
		return err
	}

	// This is kind of snapshot restore
	chkVolCloneErr := cs.checkVolCloneSupport(assembledScaleversion)
	if chkVolCloneErr != nil {
		return chkVolCloneErr
	}

	// Block cloning for cache volume
	if sourcevolume.StorageClassType == STORAGECLASS_CACHE {
		return status.Error(codes.Unimplemented, "cloning of cache volume is not supported")
	}

	// Restrict cross cluster cloning
	if newvolume.ClusterId != sourcevolume.ClusterId {
		return status.Error(codes.Unimplemented, "cloning of volume across clusters is not supported")
	}

	// Restrict cross storage class version
	if len(newvolume.StorageClassType) != 0 || len(sourcevolume.StorageClassType) != 0 {
		if newvolume.StorageClassType != sourcevolume.StorageClassType {
			return status.Error(codes.Unimplemented, "cloning of volumes between different version of storageClass is not supported")
		}
	}

	// Restrict cloning LW to Fileset based or vise a versa
	if newvolume.IsFilesetBased != sourcevolume.IsFilesetBased {
		return status.Error(codes.Unimplemented, "cloning of directory based volume to fileset based volume or vice a versa is not supported")
	}

	// Restrict if new volune is lw and is from remote
	if !newvolume.IsFilesetBased {
		if volFsInfo.Type == filesystemTypeRemote {
			return status.Error(codes.Unimplemented, "Volume cloning for directories for remote file system is not supported")
		}
	}

	sourcevolume.FsName, err = conn.GetFilesystemName(ctx, sourcevolume.FsUUID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", sourcevolume.FsUUID, sourcevolume.ClusterId, err))
	}

	sourceFsDetails, err := conn.GetFilesystemDetails(ctx, sourcevolume.FsName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("error in getting filesystem mount details for %s", sourcevolume.FsName))
	}

	// restrict remote lw to local lw cloning
	if !sourcevolume.IsFilesetBased && sourceFsDetails.Type == filesystemTypeRemote {
		return status.Error(codes.Unimplemented, "cloning of directory based volume belonging to remote cluster is not supported")
	}

	if sourcevolume.FsName != newvolume.VolBackendFs {
		if sourceFsDetails.Mount.Status != "mounted" {
			return status.Error(codes.Internal, fmt.Sprintf("filesystem %s is not mounted on GUI node", sourcevolume.FsName))
		}
	}

	if sourcevolume.IsFilesetBased {
		if sourcevolume.FsetName == "" {
			sourcevolume.FsetName, err = conn.GetFileSetNameFromId(ctx, sourcevolume.FsName, sourcevolume.FsetId)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("error in getting fileset details for %s", sourcevolume.FsetId))
			}
		}

		if sourcevolume.VolType != FILE_SHALLOWCOPY_VOLUME {
			err = cs.checkFileSetLink(ctx, conn, scaleVol, sourcevolume.FsName, sourcevolume.FsetName, "source")
			if err != nil {
				return err
			}
		}
	}

	if newvolume.NodeClass != "" {
		isValidNodeclass, err := conn.IsValidNodeclass(ctx, newvolume.NodeClass)
		if err != nil {
			return err
		}

		if !isValidNodeclass {
			return status.Error(codes.NotFound, fmt.Sprintf("nodeclass [%s] not found on cluster [%v]", newvolume.NodeClass, newvolume.ClusterId))
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
		sIdMem.StorageClassType = STORAGECLASS_CLASSIC
	}
	return sIdMem, nil
}

func (cs *ScaleControllerServer) DeleteFilesetVol(ctx context.Context, FilesystemName string, FilesetName string, volumeIdMembers scaleVolId, conn connectors.SpectrumScaleConnector, checkForSnapshots bool) (bool, error) {
	//Check if fileset exist has any snapshot
	loggerId := utils.GetLoggerId(ctx)
	if checkForSnapshots {
		klog.Infof("[%s] Checking if there is any snapshot present in the fileset [%v]", loggerId, FilesetName)
		snapshotList, err := conn.ListFilesetSnapshots(ctx, FilesystemName, FilesetName)

		if err != nil {
			if strings.Contains(err.Error(), fsetNotFoundErrCode) ||
				strings.Contains(err.Error(), fsetNotFoundErrMsg) { // fileset is already deleted
				klog.V(4).Infof("[%s] fileset seems already deleted - %v", loggerId, err)
				return true, nil
			}
			return false, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshot for fileset [%v]. Error: [%v]", FilesetName, err))
		}

		if len(snapshotList) > 0 {
			return false, status.Error(codes.Internal, fmt.Sprintf("volume fileset [%v] contains one or more snapshot, delete snapshot/volumesnapshot", FilesetName))
		}
		klog.Infof("[%s] there is no snapshot present in the fileset [%v], continue DeleteFilesetVol", loggerId, FilesetName)
	}

	err := conn.DeleteFileset(ctx, FilesystemName, FilesetName)
	if err != nil {
		if strings.Contains(err.Error(), fsetNotFoundErrCode) ||
			strings.Contains(err.Error(), fsetNotFoundErrMsg) { // fileset is already deleted
			klog.V(4).Infof("[%s] fileset seems already deleted - %v", loggerId, err)
			return true, nil
		}
		return false, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Fileset [%v] for FS [%v] and clusterId [%v].Error : [%v]", FilesetName, FilesystemName, volumeIdMembers.ClusterId, err))
	}
	return false, nil
}

// GetAFMMode returns the AFM mode of the fileset and also the error
// if there is any (including the fileset not found error) while getting
// the fileset info
func (cs *ScaleControllerServer) GetAFMMode(ctx context.Context, filesystemName string, filesetName string, conn connectors.SpectrumScaleConnector) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	filesetDetails, err := conn.ListFileset(ctx, filesystemName, filesetName)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("failed to get fileset info, filesystem: [%v], fileset: [%v], error: [%v]", filesystemName, filesetName, err))
	}

	klog.V(4).Infof("[%s] AFM mode of the fileset [%v] is [%v]", loggerId, filesetName, filesetDetails.AFM.AFMMode)
	return filesetDetails.AFM.AFMMode, nil
}

// This function deletes fileset for Consitency Group
func (cs *ScaleControllerServer) DeleteCGFileset(ctx context.Context, FilesystemName string, volumeIdMembers scaleVolId, conn connectors.SpectrumScaleConnector) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] Trying to delete independent fileset for consistency group [%v]", loggerId, volumeIdMembers.ConsistencyGroup)

	filesetDetails, err := conn.ListFileset(ctx, FilesystemName, volumeIdMembers.ConsistencyGroup)
	if err != nil {
		if strings.Contains(err.Error(), fsetNotFoundErrCode) ||
			strings.Contains(err.Error(), fsetNotFoundErrMsg) { // fileset is already deleted
			klog.V(4).Infof("[%s] Fileset seems already deleted - %v", loggerId, err)
			return nil
		}
		return status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v]. Error: [%v]", volumeIdMembers.ConsistencyGroup, err))
	}

	// Check if fileset was created by IBM Storage Scale CSI Driver
	if filesetDetails.Config.Comment == connectors.FilesetComment {
		// before deletion of fileset get its inodeSpace.
		// this will help to identify if there are one or more dependent filesets for same inodeSpace
		// which is shared with independent fileset
		inodeSpace := filesetDetails.Config.InodeSpace
		filesets, err := conn.GetFilesetsInodeSpace(ctx, FilesystemName, inodeSpace)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("listing of filesets for filesystem: [%v] failed. Error: [%v]", FilesystemName, err))
		}

		if len(filesets) > 1 {
			klog.V(4).Infof("[%s] Found atleast one dependent fileset for consistency group: [%v]", loggerId, volumeIdMembers.ConsistencyGroup)
			return nil
		}

		// Delete independent fileset for consistency group
		_, err = cs.DeleteFilesetVol(ctx, FilesystemName, volumeIdMembers.ConsistencyGroup, volumeIdMembers, conn, true)
		if err != nil {
			return err
		}
		klog.Infof("[%s] Deleted independent fileset for consistency group [%v]", loggerId, volumeIdMembers.ConsistencyGroup)
	} else {
		klog.Infof("[%s] Independent fileset for consistency group [%v] not created by IBM Storage Scale CSI Driver. Cannot delete it.", loggerId, volumeIdMembers.ConsistencyGroup)
	}

	return nil
}

func (cs *ScaleControllerServer) DeleteVolume(newctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	loggerId := utils.GetLoggerId(newctx)
	ctx := utils.SetModuleName(newctx, deleteVolume)

	// Mask the secrets from request before logging
	reqToLog := *req
	reqToLog.Secrets = nil
	klog.Infof("[%s] DeleteVolume req: %v", loggerId, &reqToLog)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		klog.Errorf("[%s] Invalid delete volume req: %v", loggerId, req)
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("Invalid delete volume req (%v): %v", req, err))
	}
	// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeID := req.GetVolumeId()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "volume Id is missing")
	}

	volumeIdMembers, err := getVolIDMembers(volumeID)
	if err != nil {
		return &csi.DeleteVolumeResponse{}, nil
	}

	klog.V(4).Infof("[%s] Volume Id Members [%v]", loggerId, volumeIdMembers)

	conn, err := cs.getConnFromClusterID(ctx, volumeIdMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
	if !isprimaryConnPresent {
		klog.Errorf("[%s] unable to get connector for primary cluster", loggerId)
		return nil, status.Error(codes.Internal, "unable to find primary cluster details in custom resource")
	}

	/* FsUUID in volumeIdMembers will be of Primary cluster. So lets get Name of it
	from Primary cluster */
	FilesystemName, err := primaryConn.GetFilesystemName(ctx, volumeIdMembers.FsUUID)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", volumeIdMembers.FsUUID, volumeIdMembers.ClusterId, err))
	}

	mountInfo, err := primaryConn.GetFilesystemMountDetails(ctx, FilesystemName)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get mount info for FS [%v] in primary cluster", FilesystemName))
	}

	relPath := ""
	if volumeIdMembers.StorageClassType != STORAGECLASS_CLASSIC || volumeIdMembers.VolType == FILE_SHALLOWCOPY_VOLUME {
		relPath = strings.Replace(volumeIdMembers.Path, mountInfo.MountPoint, "", 1)
	} else {
		primaryFSMountPoint, err := cs.getPrimaryFSMountPoint(ctx)
		if err != nil {
			return nil, err
		}
		relPath = strings.Replace(volumeIdMembers.Path, primaryFSMountPoint, "", 1)
	}
	relPath = strings.Trim(relPath, "!/")
	isPvcFromSnapshot := false
	var shallowCopyRefPath string
	var snapshotName string
	var independentFileset string
	if volumeIdMembers.VolType == FILE_SHALLOWCOPY_VOLUME {
		if relPath != "" && strings.Contains(relPath, ".snapshots") {
			volPath := strings.Split(relPath, "/")
			if len(volPath) > 2 {
				if volPath[1] == ".snapshots" {
					isPvcFromSnapshot = true
					snapshotName = volPath[2]
					independentFileset = volPath[0]
					shallowCopyRefPath = fmt.Sprintf("%s/%s", volPath[0], volPath[2])
				}
			} else {
				klog.Errorf("[%s] Invalid volume path to delete shallow copy volume reference", loggerId)
			}
		}
	}

	if volumeIdMembers.IsFilesetBased {
		var FilesetName string

		FilesystemName = getRemoteFsName(mountInfo.RemoteDeviceName)
		if volumeIdMembers.FsetName != "" {
			FilesetName = volumeIdMembers.FsetName
		} else {
			FilesetName, err = conn.GetFileSetNameFromId(ctx, FilesystemName, volumeIdMembers.FsetId)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get Fileset Name for Id [%v] FS [%v] ClusterId [%v]", volumeIdMembers.FsetId, FilesystemName, volumeIdMembers.ClusterId))
			}
		}

		// Check if fileset exists and the creator is IBM Storage Scale CSI driver
		filesetInfo, err := conn.ListFileset(ctx, FilesystemName, FilesetName)
		loggerId := utils.GetLoggerId(ctx)
		if err != nil {
			klog.Errorf("[%s]  unable to list fileset [%v] in filesystem [%v]. Error: %v", loggerId, FilesetName, FilesystemName, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list fileset [%v] in filesystem [%v]. Error: %v", FilesetName, FilesystemName, err))
		} else if !reflect.ValueOf(filesetInfo).IsZero() && filesetInfo.Config.Comment != connectors.FilesetComment {
			klog.Infof("Fileset [%v] is not created by IBM Container Storage Interface driver, skipping the fileset delete", FilesetName)
			return &csi.DeleteVolumeResponse{}, nil
		}

		if FilesetName != "" && isPvcFromSnapshot {
			err := cs.DeleteShallowCopyRefPath(ctx, FilesystemName, FilesetName, shallowCopyRefPath, volumeIdMembers.StorageClassType, independentFileset, snapshotName, conn)
			if err != nil {
				return nil, err
			}
		}
		klog.Infof("[%s] Delete Volume FilesetName:[%s] and creator is IBM Storage Scale CSI driver", loggerId, FilesetName)

		// Additional check for RDR fileset in secondary mode
		if volumeIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
			AFMMode, err := cs.GetAFMMode(ctx, FilesystemName, volumeIdMembers.ConsistencyGroup, conn)
			if err != nil {
				if strings.Contains(err.Error(), fsetNotFoundErrCode) ||
					strings.Contains(err.Error(), fsetNotFoundErrMsg) { // fileset is already deleted
					klog.V(4).Infof("[%s] the ConsistencyGroup fileset [%v] is deleted already", loggerId, FilesetName)
					return &csi.DeleteVolumeResponse{}, nil
				}
				return nil, err
			}
			if AFMMode == connectors.AFMModeSecondary {
				// AFM will take care of deletion on secondary
				klog.Infof("[%s] skipping the deletion of fileset [%v] because ConsistencyGroup fileset [%v] is in AFM Secondary mode", loggerId, FilesetName, volumeIdMembers.ConsistencyGroup)
				return &csi.DeleteVolumeResponse{}, nil
			}
		}
		if FilesetName != "" {
			/* Confirm it is same fileset which was created for this PV */
			pvName := filepath.Base(relPath)

			if pvName == FilesetName {
				checkForSnapshots := false
				if volumeIdMembers.VolType == FILE_INDEPENDENTFILESET_VOLUME {
					checkForSnapshots = true
				}
				_, err := cs.DeleteFilesetVol(ctx, FilesystemName, FilesetName, volumeIdMembers, conn, checkForSnapshots)
				if err != nil {
					return nil, err
				}

				// Delete fileset related symlink
				if volumeIdMembers.StorageClassType == STORAGECLASS_CLASSIC {
					err = primaryConn.DeleteSymLnk(ctx, cs.Driver.primary.GetPrimaryFs(), relPath)
					if err != nil {
						return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete symlnk [%v:%v] Error [%v]", cs.Driver.primary.GetPrimaryFs(), relPath, err))
					}
				}

				if volumeIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
					err := cs.DeleteCGFileset(ctx, FilesystemName, volumeIdMembers, conn)
					// DeleteCGFileset calls DeleteFilesetVol function with checkForSnapshots=true to
					// check for snapshots before deleting the CG independent fileset
					if err != nil {
						return nil, err
					}
				}

				// Delete bucket keys for a cache volume
				if volumeIdMembers.StorageClassType == STORAGECLASS_CACHE {
					bucketName := req.Secrets[connectors.BucketName]
					endpoint := req.Secrets[connectors.BucketEndpoint]
					parsedURL, err := url.Parse(endpoint)
					if err != nil {
						return nil, fmt.Errorf("failed to parse endpoint URL %s, error %v", endpoint, err)
					}
					server := parsedURL.Hostname()

					err = conn.DeleteBucketKeys(ctx, bucketName+":"+server)
					if err != nil {
						volumeName := volumeIdMembers.FsetName
						klog.Errorf("[%s] failed to delete bucket keys for volume %s", loggerId, volumeName)
						return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete bucket keys for volume %s, error: %v", volumeName, err))
					}
				}

				return &csi.DeleteVolumeResponse{}, nil
			} else {
				klog.Infof("[%s] pv name from path [%v] does not match with filesetName [%v]. Skipping delete of fileset", loggerId, pvName, FilesetName)
			}
		}
	} else {
		/* Delete Dir for Lw volume */
		err = primaryConn.DeleteDirectory(ctx, cs.Driver.primary.GetPrimaryFs(), relPath, false)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Dir using FS [%v] Relative SymLink [%v]. Error [%v]", FilesystemName, relPath, err))
		}
	}

	if volumeIdMembers.StorageClassType == STORAGECLASS_CLASSIC && !isPvcFromSnapshot {
		err = primaryConn.DeleteSymLnk(ctx, cs.Driver.primary.GetPrimaryFs(), relPath)

		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete symlnk [%v:%v] Error [%v]", cs.Driver.primary.GetPrimaryFs(), relPath, err))
		}
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ScaleControllerServer) DeleteShallowCopyRefPath(ctx context.Context, FilesystemName, FilesetName, ShallowCopyRefPath, storageClassType, independentFileset, snapshotName string, conn connectors.SpectrumScaleConnector) error {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] Deleting shallow copy reference path [%s]", loggerId, ShallowCopyRefPath)

	if storageClassType == STORAGECLASS_ADVANCED {
		klog.Infof("[%s] Target path in DeleteShallowCopyRefPath:[%s]", loggerId, ShallowCopyRefPath)
		lockSuccess := CgSnapshotLock(ctx, ShallowCopyRefPath, false)
		if !lockSuccess {
			message := fmt.Sprintf("Delete shallow copy failed to acquire lock as another operation is in progress for the same targetPath: [%s]", ShallowCopyRefPath)
			klog.Errorf("[%s] %s", loggerId, message)
			return status.Error(codes.Internal, message)
		} else {
			defer CgSnapshotUnlock(ctx, ShallowCopyRefPath)
		}
	}
	shallowCopyRefCompletePath := fmt.Sprintf("%s/%s", ShallowCopyRefPath, FilesetName)

	isShallowCopyRefPathDeleted := false
	err := conn.DeleteDirectory(ctx, FilesystemName, shallowCopyRefCompletePath, false)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist") { // directory is already deleted
			isShallowCopyRefPathDeleted = true
		} else {
			return status.Error(codes.Internal, fmt.Sprintf("unable to Delete shallow copy reference Dir using FS [%s] Error [%v]", FilesystemName, err))
		}
	} else {
		isShallowCopyRefPathDeleted = true
	}

	if isShallowCopyRefPathDeleted {
		statInfo, err := conn.StatDirectory(ctx, FilesystemName, ShallowCopyRefPath)
		if err != nil {
			if strings.Contains(err.Error(), "EFSSG0264C") ||
				strings.Contains(err.Error(), "does not exist") {
				klog.Infof("[%s] snapshot path [%s] is already deleted", loggerId, ShallowCopyRefPath)
				return nil
			} else {
				klog.Errorf("[%s] unable to stat directory using FS [%s] at path [%s]. Error [%v]", loggerId, FilesystemName, ShallowCopyRefPath, err)
				return err
			}
		} else {
			nlink, err := parseStatDirInfo(statInfo)
			if err != nil {
				klog.Errorf("[%s] invalid number of links [%d] returned in stat output for FS [%s] at path [%s]", loggerId, nlink, FilesystemName, ShallowCopyRefPath)
				return err
			}

			if nlink == 2 {
				err = conn.DeleteDirectory(ctx, FilesystemName, ShallowCopyRefPath, false)
				if err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("unable to Delete shallow copy reference parent dir using FS [%s] Error [%v]", FilesystemName, err))
				}

				if storageClassType == STORAGECLASS_ADVANCED {
					snaperr := conn.DeleteSnapshot(ctx, FilesystemName, independentFileset, snapshotName)
					if snaperr != nil {
						return status.Error(codes.Internal, fmt.Sprintf("unable to delete snapshot dir [%s] Error [%v]", snapshotName, err))
					} else {
						klog.Infof("[%s] delete snapshot reference directory [%s] successfully", loggerId, snapshotName)
					}
				}
			}

		}

	}
	return nil
}

// ControllerGetCapabilities implements the default GRPC callout.
func (cs *ScaleControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] ControllerGetCapabilities called with req: %#v", loggerId, req)
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *ScaleControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	volumeID := req.GetVolumeId()
	klog.V(4).Infof("[%s] ValidateVolumeCapabilities called with req: %#v", loggerId, req)
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
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] controllerserver ControllerUnpublishVolume", loggerId)
	klog.V(4).Infof("[%s] ControllerUnpublishVolume : req %#v", loggerId, req)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); err != nil {
		klog.Errorf("[%s] invalid Unpublish volume request: %v", loggerId, req)
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
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] Controllerserver ControllerPublishVolume", loggerId)
	klog.V(4).Infof("[%s] ControllerPublishVolume : req %#v", loggerId, req)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); err != nil {
		klog.Errorf("[%s] Invalid Publish volume request: %v", loggerId, req)
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

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume :volume capabilities are empty")
	}

	volumeIDMembers, err := getVolIDMembers(volumeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume : VolumeID is not in proper format")
	}

	filesystemID := volumeIDMembers.FsUUID
	volumePath := volumeIDMembers.Path

	// if SKIP_MOUNT_UNMOUNT == "yes" then mount/unmount will not be invoked
	skipMountUnmount := utils.GetEnv(SKIP_MOUNT_UNMOUNT, yes)
	klog.Infof("[%s] ControllerPublishVolume : SKIP_MOUNT_UNMOUNT is set to %s", loggerId, skipMountUnmount)

	//Get filesystem name from UUID
	fsName, err := cs.Driver.connmap["primary"].GetFilesystemName(ctx, filesystemID)
	if err != nil {
		klog.Errorf("[%s] ControllerPublishVolume : Error in getting filesystem Name for filesystem ID of %s.", loggerId, filesystemID)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem Name for filesystem ID of %s. Error [%v]", filesystemID, err))
	}

	//Check if primary filesystem is mounted.
	primaryfsName := cs.Driver.primary.GetPrimaryFs()
	pfsMount, err := cs.Driver.connmap["primary"].GetFilesystemMountDetails(ctx, primaryfsName)
	if err != nil {
		klog.Errorf("[%s] ControllerPublishVolume : Error in getting filesystem mount details for %s", loggerId, primaryfsName)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem mount details for %s. Error [%v]", primaryfsName, err))
	}

	// Node mapping check
	scalenodeID := getNodeMapping(nodeID)
	klog.Infof("[%s] ControllerUnpublishVolume : scalenodeID:%s --known as-- k8snodeName: %s", loggerId, scalenodeID, nodeID)

	shortnameNodeMapping := utils.GetEnv(SHORTNAME_NODE_MAPPING, no)
	if shortnameNodeMapping == yes {
		klog.V(4).Infof("[%s] ControllerPublishVolume : SHORTNAME_NODE_MAPPING is set to %s", loggerId, shortnameNodeMapping)
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

	klog.Infof("[%s] ControllerPublishVolume : Primary FS is mounted on %v", loggerId, pfsMount.NodesMounted)
	klog.V(4).Infof("[%s] ControllerPublishVolume : Primary Fileystem is %s and Volume is from Filesystem %s", loggerId, primaryfsName, fsName)
	// Skip if primary filesystem and volume filesystem is same
	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED || primaryfsName != fsName {
		//Check if filesystem is mounted
		fsMount, err := cs.Driver.connmap["primary"].GetFilesystemMountDetails(ctx, fsName)
		if err != nil {
			klog.Errorf("[%s] ControllerPublishVolume : Error in getting filesystem mount details for %s", loggerId, fsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem mount details for %s. Error [%v]", fsName, err))
		}

		if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED &&
			!strings.HasPrefix(volumePath, fsMount.MountPoint) {
			klog.Errorf("[%s] ControllerPublishVolume : Volume path %s is not part of the filesystem %s", loggerId, volumePath, fsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s", volumePath, fsName))
		} else if !strings.HasPrefix(volumePath, fsMount.MountPoint) &&
			!strings.HasPrefix(volumePath, pfsMount.MountPoint) {
			klog.Errorf("[%s] ControllerPublishVolume : Volume path %s is not part of the filesystem %s or %s", loggerId, volumePath, primaryfsName, fsName)
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

		klog.Infof("[%s] ControllerPublishVolume : Volume Source FS is mounted on %v", loggerId, fsMount.NodesMounted)
	} else {
		if !strings.HasPrefix(volumePath, pfsMount.MountPoint) {
			klog.Errorf("[%s] ControllerPublishVolume : Volume path %s is not part of the filesystem %s", loggerId, volumePath, primaryfsName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Volume path %s is not part of the filesystem %s", volumePath, primaryfsName))
		}

		isFsMounted = ispFsMounted
	}

	klog.Infof("[%s] ControllerPublishVolume : Mount Status Primaryfs [ %t ], Sourcefs [ %t ]", loggerId, ispFsMounted, isFsMounted)

	if isFsMounted && ispFsMounted {
		klog.V(4).Infof("[%s] ControllerPublishVolume : %s and %s are mounted on %s so returning success", loggerId, fsName, primaryfsName, scalenodeID)
		return &csi.ControllerPublishVolumeResponse{}, nil
	}

	if skipMountUnmount == "yes" && (!isFsMounted || !ispFsMounted) {
		klog.Errorf("[%s] ControllerPublishVolume : SKIP_MOUNT_UNMOUNT == yes and either %s or %s is not mounted on node %s", loggerId, primaryfsName, fsName, scalenodeID)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : SKIP_MOUNT_UNMOUNT == yes and either %s or %s is not mounted on node %s.", primaryfsName, fsName, scalenodeID))
	}

	//mount the primary filesystem if not mounted
	if !(ispFsMounted) && skipMountUnmount == no {
		klog.V(4).Infof("[%s] ControllerPublishVolume : mounting Filesystem %s on %s", loggerId, primaryfsName, scalenodeID)
		err = cs.Driver.connmap["primary"].MountFilesystem(ctx, primaryfsName, scalenodeID)
		if err != nil {
			klog.Errorf("[%s] ControllerPublishVolume : Error in mounting filesystem %s on node %s", loggerId, primaryfsName, scalenodeID)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume :  Error in mounting filesystem %s on node %s. Error [%v]", primaryfsName, scalenodeID, err))
		}
	}

	//mount the volume filesystem if mounted
	if !(isFsMounted) && skipMountUnmount == no && primaryfsName != fsName {
		klog.V(4).Infof("[%s] ControllerPublishVolume : mounting %s on %s", loggerId, fsName, scalenodeID)
		err = cs.Driver.connmap["primary"].MountFilesystem(ctx, fsName, scalenodeID)
		if err != nil {
			klog.Errorf("[%s] ControllerPublishVolume : Error in mounting filesystem %s on node %s", loggerId, fsName, scalenodeID)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in mounting filesystem %s on node %s. Error [%v]", fsName, scalenodeID, err))
		}
	}
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (cs *ScaleControllerServer) CheckNewSnapRequired(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string, snapWindow int) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	latestSnapList, err := conn.GetLatestFilesetSnapshots(ctx, filesystemName, filesetName)
	if err != nil {
		klog.Errorf("[%s] CheckNewSnapRequired - getting latest snapshot list failed for fileset: [%s:%s]. Error: [%v]", loggerId, filesystemName, filesetName, err)
		return "", err
	}

	if len(latestSnapList) == 0 {
		// No snapshot exists, so create new one
		return "", nil
	}

	timestamp, err := cs.getSnapshotCreateTimestamp(ctx, conn, filesystemName, filesetName, latestSnapList[0].SnapshotName)
	if err != nil {
		klog.Errorf("[%s] Error getting create timestamp for snapshot %s:%s:%s", loggerId, filesystemName, filesetName, latestSnapList[0].SnapshotName)
		return "", err
	}

	klog.Infof("[%s] latestSnapList[0].SnapshotName:%s", loggerId, latestSnapList[0].SnapshotName)
	var timestampSecs int64 = timestamp.GetSeconds()
	lastSnapTime := time.Unix(timestampSecs, 0)
	passedTime := time.Since(lastSnapTime).Seconds()
	klog.Infof("[%s] Fileset [%s:%s], last snapshot time: [%v], current time: [%v], passed time: %v seconds, snapWindow: %v minutes", loggerId, filesystemName, filesetName, lastSnapTime, time.Now(), int64(passedTime), snapWindow)

	snapWindowSeconds := snapWindow * 60

	if passedTime < float64(snapWindowSeconds) {
		// we don't need to take new snapshot
		klog.Infof("[%s] CheckNewSnapRequired - for fileset [%s:%s], using existing snapshot [%s]", loggerId, filesystemName, filesetName, latestSnapList[0].SnapshotName)
		return latestSnapList[0].SnapshotName, nil
	}

	klog.Infof("[%s] CheckNewSnapRequired - for fileset [%s:%s] we need to create new snapshot", loggerId, filesystemName, filesetName)
	return "", nil
}

func (cs *ScaleControllerServer) MakeSnapMetadataDir(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string, indepFileset string, cgSnapName string, metaSnapName string) error {
	loggerId := utils.GetLoggerId(ctx)
	cgpath := fmt.Sprintf("%s/%s", indepFileset, cgSnapName)
	path := fmt.Sprintf("%s/%s", cgpath, metaSnapName)

	klog.Infof("[%s] Target path in MakeSnapMetadataDir:[%s]", loggerId, cgpath)
	lockSuccess := CgSnapshotLock(ctx, cgpath, true)
	if !lockSuccess {
		message := fmt.Sprintf("create snapshot failed to acquire the lock as another operation is in progress for the targetPath: [%s]", cgpath)
		klog.Errorf("[%s] %s", loggerId, message)
		return status.Error(codes.Internal, message)
	} else {
		defer CgSnapshotUnlock(ctx, cgpath)
	}
	klog.Infof("[%s] MakeSnapMetadataDir - creating directory [%s] for fileset: [%s:%s]", loggerId, path, filesystemName, filesetName)
	err := conn.MakeDirectory(ctx, filesystemName, path, "0", "0")
	if err != nil {
		// Directory creation failed
		klog.Errorf("[%s] Volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", loggerId, filesetName, path, filesystemName, err)
		return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", path, filesystemName, err)
	}
	return nil
}

// CreateSnapshot Create Snapshot
func (cs *ScaleControllerServer) CreateSnapshot(newctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(newctx)
	ctx := utils.SetModuleName(newctx, createSnapshot)
	klog.Infof("[%s] CreateSnapshot - create snapshot req: %v", loggerId, req)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		klog.Errorf("[%s] CreateSnapshot - invalid create snapshot req: %v", loggerId, req)
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

	// Block snapshot for cache volume
	if volumeIDMembers.StorageClassType == STORAGECLASS_CACHE {
		return nil, status.Error(codes.InvalidArgument, "CreateSnapshot - taking snapshot of cache volume is not supported")
	}

	if !volumeIDMembers.IsFilesetBased {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is fileset", volID))
	}

	if (volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED) && (volumeIDMembers.VolType != FILE_DEPENDENTFILESET_VOLUME) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateSnapshot - volume [%s] - Volume snapshot can only be created when source volume is dependent fileset for new storageClass", volID))
	}

	conn, err := cs.getConnFromClusterID(ctx, volumeIDMembers.ClusterId)
	if err != nil {
		return nil, err
	}
	assembledScaleversion, err := cs.assembledScaleVersion(ctx, conn)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("the  IBM Storage Scale version check for permissions failed with error %s", err))
	}
	/* Check if IBM Storage Scale supports Snapshot */
	chkSnapshotErr := cs.checkSnapshotSupport(assembledScaleversion)
	if chkSnapshotErr != nil {
		return nil, chkSnapshotErr
	}

	primaryConn, isprimaryConnPresent := cs.Driver.connmap["primary"]
	if !isprimaryConnPresent {
		klog.Errorf("[%s] CreateSnapshot - unable to get connector for primary cluster", loggerId)
		return nil, status.Error(codes.Internal, "CreateSnapshot - unable to find primary cluster details in custom resource")
	}

	filesystemName, err := primaryConn.GetFilesystemName(ctx, volumeIDMembers.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
	}

	mountInfo, err := primaryConn.GetFilesystemMountDetails(ctx, filesystemName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - unable to get mount info for FS [%v] in primary cluster", filesystemName))
	}

	filesetResp := connectors.Fileset_v2{}
	filesystemName = getRemoteFsName(mountInfo.RemoteDeviceName)
	if volumeIDMembers.FsetName != "" {
		filesetResp, err = conn.GetFileSetResponseFromName(ctx, filesystemName, volumeIDMembers.FsetName)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get Fileset response for Fileset [%v] FS [%v] ClusterId [%v]", volumeIDMembers.FsetName, filesystemName, volumeIDMembers.ClusterId))
		}
	} else {
		filesetResp, err = conn.GetFileSetResponseFromId(ctx, filesystemName, volumeIDMembers.FsetId)
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
		klog.V(4).Infof("[%s] CreateSnapshot - creating snapshot for advanced storageClass", loggerId)
		relPath = strings.Replace(volumeIDMembers.Path, mountInfo.MountPoint, "", 1)
	} else {
		klog.V(4).Infof("[%s] CreateSnapshot - creating snapshot for classic storageClass", loggerId)
		primaryFSMountPoint, err := cs.getPrimaryFSMountPoint(ctx)
		if err != nil {
			return nil, err
		}
		relPath = strings.Replace(volumeIDMembers.Path, primaryFSMountPoint, "", 1)
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
	snapWindowInt := 0
	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		snapParams := req.GetParameters()
		snapWindow, snapWindowSpecified := snapParams[connectors.UserSpecifiedSnapWindow]
		if !snapWindowSpecified {
			// use default snapshot window for consistency group
			snapWindow = defaultSnapWindow
			klog.Infof("[%s] SnapWindow not specified. Using default snapWindow: [%s] for for fileset[%s:%s]", loggerId, snapWindow, filesetResp.FilesetName, filesystemName)
		}
		snapWindowInt, err = strconv.Atoi(snapWindow)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot [%s] - invalid snapWindow value: [%v]", snapName, snapWindow))
		}

		// Additional check for RDR fileset in secondary mode
		AFMMode, err := cs.GetAFMMode(ctx, filesystemName, filesetName, conn)
		if err != nil {
			return nil, err
		}
		if AFMMode == connectors.AFMModeSecondary {
			klog.Errorf("[%s] snapshot is not supported for AFM Secondary mode of ConsistencyGroup fileset [%v]", loggerId, filesetName)
			return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot is not supported for AFM Secondary mode of ConsistencyGroup fileset [%v]", filesetName))
		}
	}

	snapExist, err := conn.CheckIfSnapshotExist(ctx, filesystemName, filesetName, snapName)
	if err != nil {
		klog.Errorf("[%s] CreateSnapshot [%s] - Unable to get the snapshot details. Error [%v]", loggerId, snapName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get the snapshot details for [%s]. Error [%v]", snapName, err))
	}

	snapName, createNewSnap, err := cs.CheckIfNewSnapshotIsRequired(ctx, conn, filesystemName, filesetName, filesetResp.FilesetName, snapName, volumeIDMembers.StorageClassType, snapWindowInt, snapExist)
	if err != nil {
		return nil, err
	}

	if createNewSnap {
		snapName, err = cs.CreateNewSnapshot(ctx, conn, filesystemName, filesetName, snapName, volumeIDMembers.StorageClassType, snapExist)
		if err != nil {
			klog.Errorf("[%s] CreateSnapshot [%s] unable to create new snapshot for fileset [%s:%s]. Error: [%v]", loggerId, snapName, filesystemName, filesetName, err)
			return nil, err
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

	timestamp, err := cs.getSnapshotCreateTimestamp(ctx, conn, filesystemName, filesetName, snapName)
	if err != nil {
		klog.Errorf("[%s] Error getting create timestamp for snapshot %s:%s:%s", loggerId, filesystemName, filesetName, snapName)
		return nil, err
	}

	restoreSize, err := cs.getSnapRestoreSize(ctx, conn, filesystemName, filesetResp.FilesetName)
	if err != nil {
		klog.Errorf("[%s] Error getting the snapshot restore size for snapshot %s:%s:%s", loggerId, filesystemName, filesetResp.FilesetName, snapName)
		return nil, err
	}

	if volumeIDMembers.StorageClassType == STORAGECLASS_ADVANCED {
		err := cs.MakeSnapMetadataDir(ctx, conn, filesystemName, filesetResp.FilesetName, filesetName, snapName, req.GetName())
		if err != nil {
			klog.Errorf("[%s] Error in creating directory for storing metadata information for advanced storageClass. Error: [%v]", loggerId, err)
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

func (cs *ScaleControllerServer) CheckIfNewSnapshotIsRequired(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName, filesetName, fsetName, snapName, storageClassType string, snapWindowInt int, snapExist bool) (string, bool, error) {

	loggerId := utils.GetLoggerId(ctx)
	createNewSnap := false

	if !snapExist {
		createNewSnap = true
		if storageClassType == STORAGECLASS_ADVANCED && createNewSnap {
			lockSuccess := CgSnapshotLock(ctx, filesetName, snapExist)
			if !lockSuccess {
				cs.retryToCreateNewSnap(ctx, conn, filesystemName, filesetName, snapName)
			} else {
				defer CgSnapshotUnlock(ctx, filesetName)
			}
		}

		//  For new storageClass check last snapshot creation time, if time passed is less than
		//  snapWindow then return existing snapshot

		if storageClassType == STORAGECLASS_ADVANCED {
			cgSnapName, err := cs.CheckNewSnapRequired(ctx, conn, filesystemName, filesetName, snapWindowInt)
			if err != nil {
				klog.Errorf("[%s] CreateSnapshot [%s] - unable to check if snapshot is required for new storageClass for fileset [%s:%s]. Error: [%v]", loggerId, snapName, filesystemName, filesetName, err)
				return snapName, createNewSnap, err
			}
			if cgSnapName != "" {
				usable, err := cs.isExistingSnapUseableForVol(ctx, conn, filesystemName, filesetName, fsetName, cgSnapName)
				if !usable {
					return snapName, createNewSnap, err
				}
				createNewSnap = false
				snapName = cgSnapName
			} else {
				klog.Infof("[%s] CreateSnapshot - creating new snapshot for consistency group for fileset: [%s:%s]", loggerId, filesystemName, filesetName)
			}
		}
	}
	return snapName, createNewSnap, nil
}

func (cs *ScaleControllerServer) CreateNewSnapshot(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName, filesetName, snapName, storageClassType string, snapExist bool) (string, error) {
	loggerId := utils.GetLoggerId(ctx)

	if storageClassType == STORAGECLASS_ADVANCED {
		lockSuccess := CgSnapshotLock(ctx, filesetName, snapExist)
		if !lockSuccess {
			klog.Errorf("[%s] CreateNewSnapshot [%s]: Failed to acquire the lock", loggerId, snapName)
			return snapName, status.Error(codes.Internal, fmt.Sprintf("CreateNewSnapshot [%s]: Failed to acquire the lock", snapName))
		} else {
			defer CgSnapshotUnlock(ctx, filesetName)
		}
	}

	snapshotList, err := conn.ListFilesetSnapshots(ctx, filesystemName, filesetName)
	if err != nil {
		klog.Errorf("[%s] CreateSnapshot [%s] - unable to list snapshots for fileset [%s:%s]. Error: [%v]", loggerId, snapName, filesystemName, filesetName, err)
		return snapName, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshots for fileset [%s:%s]. Error: [%v]", filesystemName, filesetName, err))
	}

	if len(snapshotList) >= 256 {
		klog.Errorf("[%s] CreateSnapshot [%s] - max limit of snapshots reached for fileset [%s:%s]. No more snapshots can be created for this fileset.", loggerId, snapName, filesystemName, filesetName)
		return snapName, status.Error(codes.OutOfRange, fmt.Sprintf("max limit of snapshots reached for fileset [%s:%s]. No more snapshots can be created for this fileset.", filesystemName, filesetName))
	}

	snaperr := conn.CreateSnapshot(ctx, filesystemName, filesetName, snapName)
	if snaperr != nil {
		klog.Errorf("[%s] Snapshot [%s] - Unable to create snapshot. Error [%v]", loggerId, snapName, snaperr)
		return snapName, status.Error(codes.Internal, fmt.Sprintf("unable to create snapshot [%s]. Error [%v]", snapName, snaperr))
	}

	return snapName, nil
}

func (cs *ScaleControllerServer) retryToCreateNewSnap(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName, filesetName, snapName string) {

	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] retrying to check whether new snapshot is required", loggerId)
	for i := 0; i < retryCount; i++ {
		time.Sleep(retryInterval * time.Second)
	}

}

func (cs *ScaleControllerServer) getSnapshotCreateTimestamp(ctx context.Context, conn connectors.SpectrumScaleConnector, fs string, fset string, snap string) (timestamppb.Timestamp, error) {
	var timestamp timestamppb.Timestamp

	createTS, err := conn.GetSnapshotCreateTimestamp(ctx, fs, fset, snap)
	if err != nil {
		klog.Errorf("[%s]snapshot [%s] - Unable to get snapshot create timestamp", utils.GetLoggerId(ctx), snap)
		return timestamp, err
	}

	timezoneOffset, err := conn.GetTimeZoneOffset(ctx)
	if err != nil {
		klog.Errorf("[%s] snapshot [%s] - Unable to get cluster timezone", utils.GetLoggerId(ctx), snap)
		return timestamp, err
	}

	// for GMT, REST API returns Z instead of 00:00
	if timezoneOffset == "Z" {
		timezoneOffset = "+00:00"
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
		klog.Errorf("[%s] snapshot - for fileset [%s:%s] error in parsing timestamp: [%v]. Error: [%v]", utils.GetLoggerId(ctx), fs, fset, createTS, err)
		return timestamp, err
	}
	timestamp.Seconds = t.Unix()
	timestamp.Nanos = 0

	klog.Infof("[%s] getSnapshotCreateTimestamp: for fileset [%s:%s] snapshot creation timestamp: [%v]", utils.GetLoggerId(ctx), fs, fset, createTSTZ)
	return timestamp, nil
}

func (cs *ScaleControllerServer) getSnapRestoreSize(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string) (int64, error) {
	quotaResp, err := conn.GetFilesetQuotaDetails(ctx, filesystemName, filesetName)

	if err != nil {
		return 0, err
	}

	if quotaResp.BlockLimit < 0 {
		klog.Errorf("[%s] getSnapRestoreSize: Invalid block limit [%v] for fileset [%s:%s] found", utils.GetLoggerId(ctx), quotaResp.BlockLimit, filesystemName, filesetName)
		return 0, status.Error(codes.Internal, fmt.Sprintf("invalid block limit [%v] for fileset [%s:%s] found", quotaResp.BlockLimit, filesystemName, filesetName))
	}

	// REST API returns block limit in kb, convert it to bytes and return
	return int64(quotaResp.BlockLimit * 1024), nil
}

func (cs *ScaleControllerServer) isExistingSnapUseableForVol(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, consistencyGroup string, filesetName string, cgSnapName string) (bool, error) {
	pathDir := fmt.Sprintf("%s/.snapshots/%s/%s", consistencyGroup, cgSnapName, filesetName)
	_, err := conn.StatDirectory(ctx, filesystemName, pathDir)
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

func (cs *ScaleControllerServer) DelSnapMetadataDir(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, consistencyGroup string, filesetName string, cgSnapName string, metaSnapName string) (bool, error) {
	loggerId := utils.GetLoggerId(ctx)
	cgpath := fmt.Sprintf("%s/%s", consistencyGroup, cgSnapName)
	pathDir := fmt.Sprintf("%s/%s", cgpath, metaSnapName)

	klog.Infof("[%s] Target path in DelSnapMetadataDir:[%s]", loggerId, cgpath)
	lockSuccess := CgSnapshotLock(ctx, cgpath, true)
	if !lockSuccess {
		message := fmt.Sprintf("Delete snapshot failed to acquire the lock as another operation is in progress for the targetPath: [%s]", cgpath)
		klog.Errorf("[%s] %s", loggerId, message)
		return false, status.Error(codes.Internal, message)
	} else {
		defer CgSnapshotUnlock(ctx, cgpath)
	}

	err := conn.DeleteDirectory(ctx, filesystemName, pathDir, false)
	if err != nil {
		if !(strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist")) { // directory is already deleted
			return false, status.Error(codes.Internal, fmt.Sprintf("unable to Delete Dir using FS [%v] at path [%v]. Error [%v]", filesystemName, pathDir, err))
		}
	}

	// Now check if consistency group snapshot metadata directory can be deleted
	pathDir = fmt.Sprintf("%s/%s", consistencyGroup, cgSnapName)
	statInfo, err := conn.StatDirectory(ctx, filesystemName, pathDir)
	if err != nil {
		if !(strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist")) { // directory is already deleted
			return false, status.Error(codes.Internal, fmt.Sprintf("unable to stat directory using FS [%v] at path [%v]. Error [%v]", filesystemName, pathDir, err))
		}
		return true, nil
	}

	nlink, err := parseStatDirInfo(statInfo)
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("invalid number of links [%d] returned in stat output for FS [%v] at path [%v]. Error [%v]", nlink, filesystemName, pathDir, err))
	}

	klog.Infof("[%s] DelSnapMetadataDir - number of links for directory in FS [%v] at path [%v] is [%v]", loggerId, filesystemName, pathDir, nlink)

	if nlink == 2 {
		// directory can be deleted
		err := conn.DeleteDirectory(ctx, filesystemName, pathDir, true)
		if err != nil {
			if !(strings.Contains(err.Error(), "EFSSG0264C") ||
				strings.Contains(err.Error(), "does not exist")) {
				return false, status.Error(codes.Internal, fmt.Sprintf("unable to delete directory for FS [%s] at path [%s]. Error: [%v]", filesystemName, pathDir, err))
			}
		}
		return true, nil
	}

	return false, nil
}

func parseStatDirInfo(statInfo string) (int, error) {
	statSplit := strings.Split(statInfo, "\n")
	thirdLineSplit := strings.Split(statSplit[2], " ")
	lenSplit := len(thirdLineSplit)
	linkStr := strings.TrimRight(thirdLineSplit[lenSplit-1], "\n")
	nlink, err := strconv.Atoi(linkStr)
	return nlink, err
}

// DeleteSnapshot - Delete snapshot
func (cs *ScaleControllerServer) DeleteSnapshot(newctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	loggerId := utils.GetLoggerId(newctx)
	ctx := utils.SetModuleName(newctx, deleteSnapshot)
	klog.Infof("[%s] DeleteSnapshot - delete snapshot req: %v", loggerId, req)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		klog.Errorf("[%s] DeleteSnapshot - invalid delete snapshot req %v: %v", loggerId, req, err)
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
		klog.Errorf("[%s] Invalid snapshot ID %s [%v]", loggerId, snapID, err)
		return nil, err
	}

	conn, err := cs.getConnFromClusterID(ctx, snapIdMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	filesystemName, err := conn.GetFilesystemName(ctx, snapIdMembers.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get filesystem Name for Filesystem UID [%v] and clusterId [%v]. Error [%v]", snapIdMembers.FsUUID, snapIdMembers.ClusterId, err))
	}

	filesetExist := false
	if snapIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
		filesetExist, err = conn.CheckIfFilesetExist(ctx, filesystemName, snapIdMembers.ConsistencyGroup)
	} else {
		filesetExist, err = conn.CheckIfFilesetExist(ctx, filesystemName, snapIdMembers.FsetName)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the fileset %s details details. Error [%v]", snapIdMembers.FsetName, err))
	}

	var shallowCopyRefPath string
	//skip delete if snapshot not exist, return success
	if filesetExist {
		snapExist := false
		if snapIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
			klog.V(4).Infof("[%s] DeleteSnapshot - for advanced storageClass check if snapshot [%s] exist in fileset [%s] under filesystem [%s]", loggerId, snapIdMembers.SnapName, snapIdMembers.ConsistencyGroup, filesystemName)
			chkSnapExist, err := conn.CheckIfSnapshotExist(ctx, filesystemName, snapIdMembers.ConsistencyGroup, snapIdMembers.SnapName)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the snapshot details. Error [%v]", err))
			}
			snapExist = chkSnapExist
		} else {
			klog.V(4).Infof("[%s] DeleteSnapshot - for classic storageClass check if snapshot [%s] exist in fileset [%s] under filesystem [%s]", loggerId, snapIdMembers.SnapName, snapIdMembers.FsetName, filesystemName)
			chkSnapExist, err := conn.CheckIfSnapshotExist(ctx, filesystemName, snapIdMembers.FsetName, snapIdMembers.SnapName)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to get the snapshot details. Error [%v]", err))
			}
			snapExist = chkSnapExist
		}

		// skip delete snapshot if not exist, return success
		if snapExist {
			if snapIdMembers.StorageClassType == STORAGECLASS_CLASSIC {
				shallowCopyRefPath = fmt.Sprintf("%s/%s", snapIdMembers.FsetName, snapIdMembers.SnapName)
			}

			deleteSnapshot := true
			filesetName := snapIdMembers.FsetName
			if snapIdMembers.StorageClassType == STORAGECLASS_ADVANCED {
				delSnap, snaperr := cs.DelSnapMetadataDir(ctx, conn, filesystemName, snapIdMembers.ConsistencyGroup, snapIdMembers.FsetName, snapIdMembers.SnapName, snapIdMembers.MetaSnapName)
				if snaperr != nil {
					klog.Errorf("[%s] DeleteSnapshot - error while deleting snapshot %s: Error: %v", loggerId, snapIdMembers.SnapName, snaperr)
					return nil, snaperr
				}
				if delSnap {
					filesetName = snapIdMembers.ConsistencyGroup
					klog.V(4).Infof("[%s] DeleteSnapshot - for advanced storageClass we can delete snapshot [%s] from fileset [%s] under filesystem [%s]", loggerId, snapIdMembers.SnapName, filesetName, filesystemName)
				} else {
					deleteSnapshot = false
				}
			} else {
				dirExists, err := conn.CheckIfFileDirPresent(ctx, filesystemName, shallowCopyRefPath)
				if err != nil {
					if !(strings.Contains(err.Error(), "EFSSG0264C") ||
						strings.Contains(err.Error(), "does not exist")) {
						klog.Errorf("[%s] unable to check if directory path [%s] exists in filesystem [%s]. Error : %v", loggerId, shallowCopyRefPath, filesystemName, err)
						deleteSnapshot = false
					}
				}

				if dirExists {
					statInfo, err := conn.StatDirectory(ctx, filesystemName, shallowCopyRefPath)
					if err != nil {
						klog.Errorf("[%s] unable to stat directory using FS [%s] at path [%s]. Error [%v]", loggerId, filesystemName, shallowCopyRefPath, err)
						deleteSnapshot = false
					} else {
						nlink, err := parseStatDirInfo(statInfo)
						if err != nil {
							klog.Errorf("[%s] invalid number of links [%d] returned in stat output for FS [%s] at path [%s]", loggerId, nlink, filesystemName, shallowCopyRefPath)
							deleteSnapshot = false
						}
						if nlink > 2 {
							deleteSnapshot = false
							return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteSnapshot - unable to delete snapshot [%s] as there is a reference for shallowcopy volume", snapIdMembers.SnapName))
						}
					}
				}
			}

			if deleteSnapshot {
				klog.Infof("[%s] DeleteSnapshot - deleting snapshot [%s] from fileset [%s] under filesystem [%s]", loggerId, snapIdMembers.SnapName, filesetName, filesystemName)
				snaperr := conn.DeleteSnapshot(ctx, filesystemName, filesetName, snapIdMembers.SnapName)
				if snaperr != nil {
					klog.Errorf("[%s] DeleteSnapshot - error deleting snapshot %s: %v", loggerId, snapIdMembers.SnapName, snaperr)
					return nil, snaperr
				}
				klog.Infof("[%s] DeleteSnapshot - successfully deleted snapshot [%s] from fileset [%s] under filesystem [%s]", loggerId, snapIdMembers.SnapName, filesetName, filesystemName)
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

func (cs *ScaleControllerServer) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ScaleControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *ScaleControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] ControllerExpandVolume - Volume expand req: %v", loggerId, req)

	if err := cs.Driver.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_EXPAND_VOLUME); err != nil {
		klog.Errorf("[%s] ControllerExpandVolume - invalid expand volume req: %v", loggerId, req)
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
	capacity := uint64(capRange.GetRequiredBytes()) // #nosec G115 -- false positive

	volumeIDMembers, err := getVolIDMembers(volID)

	if err != nil {
		klog.Errorf("[%s] ControllerExpandVolume - Error in source Volume ID %v: %v", loggerId, volID, err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("ControllerExpandVolume - Error in source Volume ID %v: %v", volID, err))
	}

	if volumeIDMembers.VolType == FILE_SHALLOWCOPY_VOLUME {
		klog.Errorf("[%s] ControllerExpandVolume - volume expansion is not supported for shallow copy volume", loggerId)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerExpandVolume - volume expansion is not supported for shallow copy volume %s", volID))
	}

	// For lightweight return volume expanded as no action is required
	if !volumeIDMembers.IsFilesetBased {
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(capacity),
			NodeExpansionRequired: false,
		}, nil
	}

	conn, err := cs.getConnFromClusterID(ctx, volumeIDMembers.ClusterId)
	if err != nil {
		return nil, err
	}

	filesystemName, err := conn.GetFilesystemName(ctx, volumeIDMembers.FsUUID)
	if err != nil {
		klog.Errorf("[%s] ControllerExpandVolume - unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", loggerId, volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerExpandVolume - unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", volumeIDMembers.FsUUID, volumeIDMembers.ClusterId, err))
	}

	filesetName := volumeIDMembers.FsetName

	fsetExist, err := conn.CheckIfFilesetExist(ctx, filesystemName, filesetName)
	if err != nil {
		klog.Errorf("[%s] unable to check fileset [%v] existance in filesystem [%v]. Error [%v]", loggerId, filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to check fileset [%v] existance in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	if !fsetExist {
		klog.Errorf("[%s] Fileset [%v] does not exist in filesystem [%v]. Error [%v]", loggerId, filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("fileset [%v] does not exist in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	quota, err := conn.ListFilesetQuota(ctx, filesystemName, filesetName)
	if err != nil {
		klog.Errorf("[%s] unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", loggerId, filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list quota for fileset [%v] in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	filesetQuotaBytes, err := ConvertToBytes(quota)
	if err != nil {
		klog.Errorf("[%s] unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", loggerId, filesetName, filesystemName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to convert quota for fileset [%v] in filesystem [%v]. Error [%v]", filesetName, filesystemName, err))
	}

	if filesetQuotaBytes < capacity {
		var hardLimit, softLimit string
		hardLimit = strconv.FormatUint(capacity, 10)
		if volumeIDMembers.StorageClassType == STORAGECLASS_CACHE {
			softLimit = strconv.FormatUint(uint64(math.Round(float64(capacity)*float64(softQuotaPercent)/float64(100))), 10)
		} else {
			softLimit = hardLimit
		}
		err = conn.SetFilesetQuota(ctx, filesystemName, filesetName, hardLimit, softLimit)
		if err != nil {
			klog.Errorf("[%s] unable to update the quota. Error [%v]", loggerId, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to expand the volume. Error [%v]", err))
		}
	}

	fsetDetails, err := conn.ListFileset(ctx, filesystemName, filesetName)
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
				fseterr := conn.UpdateFileset(ctx, filesystemName, filesetName, opt)
				if fseterr != nil {
					klog.Errorf("[%s] Volume:[%v] - unable to update fileset [%v] in filesystem [%v]. Error: %v", loggerId, filesetName, filesetName, filesystemName, fseterr)
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
func (cs *ScaleControllerServer) getRemoteClusterID(ctx context.Context, clusterName string) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] Fetching cluster details from cache map for cluster %s", loggerId, clusterName)
	clusterDetails, found := cs.Driver.clusterMap.Load(ClusterName{clusterName})
	if found {
		klog.V(4).Infof("[%s] Checking if cluster details found from cache map for cluster %s has expired.", loggerId, clusterName)
		if expired := checkExpiry(clusterDetails); !expired { // cluster details are not expired.
			klog.V(4).Infof("[%s] Cluster details found from cache map for cluster %s are valid.", loggerId, clusterName)
			return clusterDetails.(ClusterDetails).id, nil
		} else { // cluster details are expired
			klog.V(4).Infof("[%s] cluster details found from cache map for cluster %s are expired.", loggerId, clusterName)
			cID := clusterDetails.(ClusterDetails).id
			conn, err := cs.getConnFromClusterID(ctx, cID)
			if err != nil {
				return "", err
			}
			clusterSummary, err := conn.GetClusterSummary(ctx)
			if err != nil {
				return "", err
			}
			cName := clusterSummary.ClusterName
			if cName == clusterName {
				klog.V(4).Infof("[%s] updating cluster details in cache map for cluster %s.", loggerId, clusterName)
				cs.Driver.clusterMap.Store(ClusterName{cName}, ClusterDetails{cID, cName, time.Now(), 24})
				cs.Driver.clusterMap.Store(ClusterID{cID}, ClusterDetails{cID, cName, time.Now(), 24})
				klog.V(4).Infof("[%s] ClusterMap updated, [%s : %s]", loggerId, cID, cName)
				return cID, nil
			} else {
				found = false
			}
		}
	}
	var err error
	cName := ""
	updated := false
	if !found {
		klog.V(4).Infof("[%s] Cluster details are either expired or not found in cache map for cluster %s. Updating the cache map.", loggerId, clusterName)
		scaleconfig := settings.LoadScaleConfigSettings(ctx)

		for i := range scaleconfig.Clusters {

			cID := scaleconfig.Clusters[i].ID
			klog.V(4).Infof("[%s] Fetching cluster details from cache map for cluster %s", loggerId, scaleconfig.Clusters[i].ID)
			clusterDetails, found := cs.Driver.clusterMap.Load(ClusterID{cID})
			if found {
				klog.V(4).Infof("[%s] Checking if cluster details found from cache map for cluster %s has expired.", loggerId, scaleconfig.Clusters[i].ID)
				if expired := checkExpiry(clusterDetails); !expired {
					klog.V(4).Infof("[%s] Cluster details found from cache map for cluster %s are valid.", loggerId, scaleconfig.Clusters[i].ID)
					cName := clusterDetails.(ClusterDetails).name
					if cName == clusterName {
						return cID, nil
					}
				} else {
					klog.V(4).Infof("[%s] Cluster details found from cache map for cluster %s are expired.", loggerId, scaleconfig.Clusters[i].ID)
					klog.V(4).Infof("[%s] Updating cluster details in cache map for cluster %s.", loggerId, scaleconfig.Clusters[i].ID)
					cName, updated, err = cs.updateClusterMap(ctx, cID)
					if !updated {
						continue
					}
					if cName == clusterName {
						return cID, nil
					}
				}
			} else { // if !found
				klog.V(4).Infof("[%s] Cluster details not found in cache map for cluster %s.", loggerId, scaleconfig.Clusters[i].ID)
				klog.V(4).Infof("[%s] adding cluster details in cache map for cluster %s.", loggerId, scaleconfig.Clusters[i].ID)
				cName, updated, err = cs.updateClusterMap(ctx, cID)
				if !updated {
					continue
				}
				if cName == clusterName {
					return cID, nil
				}
			}
		}
	}

	return "", status.Error(codes.Internal, fmt.Sprintf("unable to get cluster ID for cluster %s. Error %v", clusterName, err))
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
func (cs *ScaleControllerServer) updateClusterMap(ctx context.Context, cID string) (string, bool, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] Creating new connector for the cluster %s", loggerId, cID)
	clusterConnector, err := cs.getConnFromClusterID(ctx, cID)
	// clusterConnector, err := connectors.NewSpectrumRestV2(cluster)
	if err != nil {
		klog.V(4).Infof("[%s] unable to create new connector for the cluster %s", loggerId, cID)
		return "", false, err
	}

	clusterSummary, err := clusterConnector.GetClusterSummary(ctx)
	if err != nil {
		klog.V(4).Infof("[%s] unable to get cluster summary for cluster %s", loggerId, cID)
		return "", false, err
	}

	cName := clusterSummary.ClusterName
	// cID = fmt.Sprint(clusterSummary.ClusterID)
	cs.Driver.clusterMap.Store(ClusterName{cName}, ClusterDetails{cID, cName, time.Now(), 24})
	cs.Driver.clusterMap.Store(ClusterID{cID}, ClusterDetails{cID, cName, time.Now(), 24})
	klog.V(4).Infof("[%s] ClusterMap updated: [%s : %s]", loggerId, cID, cName)
	return cName, true, nil
}

func lockBucket(loggerId string, volName string, bucket string) bool {
	bucketMutex.Lock()
	defer bucketMutex.Unlock()

	if _, exists := bucketLock[bucket]; exists {
		return false
	}
	bucketLock[bucket] = true
	klog.V(4).Infof("[%s] The bucket [%s] is locked for creation of a volume: [%s]", loggerId, bucket, volName)
	return true
}

func unlockBucket(loggerId string, volName string, bucket string) {
	bucketMutex.Lock()
	defer bucketMutex.Unlock()
	delete(bucketLock, bucket)
	klog.V(4).Infof("[%s] The bucket [%s] is unlocked for creation of a volume: [%s]", loggerId, bucket, volName)
}
