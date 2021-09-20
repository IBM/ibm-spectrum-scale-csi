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

package connectors

import (
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/golang/glog"
)

//go:generate counterfeiter -o ../../../fakes/fake_spectrum.go . SpectrumScaleConnector
type SpectrumScaleConnector interface {
	//Cluster operations
	GetClusterId() (string, error)
	GetTimeZoneOffset() (string, error)
	GetScaleVersion() (string, error)
	//Filesystem operations
	GetFilesystemMountDetails(filesystemName string) (MountInfo, error)
	IsFilesystemMountedOnGUINode(filesystemName string) (bool, error)
	ListFilesystems() ([]string, error)
	GetFilesystemDetails(filesystemName string) (FileSystem_v2, error)
	GetFilesystemMountpoint(filesystemName string) (string, error)
	//Fileset operations
	CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error
	DeleteFileset(filesystemName string, filesetName string) error
	//LinkFileset(filesystemName string, filesetName string) error
	LinkFileset(filesystemName string, filesetName string, linkpath string) error
	UnlinkFileset(filesystemName string, filesetName string) error
	//ListFilesets(filesystemName string) ([]resources.Volume, error)
	ListFileset(filesystemName string, filesetName string) (Fileset_v2, error)
	IsFilesetLinked(filesystemName string, filesetName string) (bool, error)
	FilesetRefreshTask() error
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(filesystemName string, filesetName string) (string, error)
	GetFilesetQuotaDetails(filesystemName string, filesetName string) (Quota_v2, error)
	SetFilesetQuota(filesystemName string, filesetName string, quota string) error
	CheckIfFSQuotaEnabled(filesystem string) error
	CheckIfFilesetExist(filesystemName string, filesetName string) (bool, error)
	//Directory operations
	MakeDirectory(filesystemName string, relativePath string, uid string, gid string) error
	MakeDirectoryV2(filesystemName string, relativePath string, uid string, gid string, permissions string) error
	MountFilesystem(filesystemName string, nodeName string) error
	UnmountFilesystem(filesystemName string, nodeName string) error
	GetFilesystemName(filesystemUUID string) (string, error)
	CheckIfFileDirPresent(filesystemName string, relPath string) (bool, error)
	CreateSymLink(SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error
	GetFsUid(filesystemName string) (string, error)
	DeleteDirectory(filesystemName string, dirName string) error
	GetFileSetUid(filesystemName string, filesetName string) (string, error)
	GetFileSetNameFromId(filesystemName string, Id string) (string, error)
	DeleteSymLnk(filesystemName string, LnkName string) error
	GetFileSetResponseFromId(filesystemName string, Id string) (Fileset_v2, error)
	GetFileSetResponseFromName(filesystemName string, filesetName string) (Fileset_v2, error)

	IsValidNodeclass(nodeclass string) (bool, error)
	IsSnapshotSupported() (bool, error)

	//Snapshot operations
	WaitForSnapshotCopy(statusCode int, jobID uint64) error
	CreateSnapshot(filesystemName string, filesetName string, snapshotName string) error
	DeleteSnapshot(filesystemName string, filesetName string, snapshotName string) error
	GetSnapshotUid(filesystemName string, filesetName string, snapName string) (string, error)
	GetSnapshotCreateTimestamp(filesystemName string, filesetName string, snapName string) (string, error)
	CheckIfSnapshotExist(filesystemName string, filesetName string, snapshotName string) (bool, error)
	ListFilesetSnapshots(filesystemName string, filesetName string) ([]Snapshot_v2, error)
	CopyFsetSnapshotPath(filesystemName string, filesetName string, snapshotName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
	CopyFilesetPath(filesystemName string, filesetName string, srcPath string, targetPath string, nodeclass string) error
	IsNodeComponentHealthy(nodeName string, component string) (bool, error)
}

const (
	UserSpecifiedFilesetType    string = "filesetType"
	UserSpecifiedFilesetTypeDep string = "fileset-type"
	UserSpecifiedInodeLimit     string = "inodeLimit"
	UserSpecifiedInodeLimitDep  string = "inode-limit"
	UserSpecifiedUid            string = "uid"
	UserSpecifiedGid            string = "gid"
	UserSpecifiedClusterId      string = "clusterId"
	UserSpecifiedParentFset     string = "parentFileset"
	UserSpecifiedVolBackendFs   string = "volBackendFs"
	UserSpecifiedVolDirPath     string = "volDirBasePath"
	UserSpecifiedNodeClass      string = "nodeClass"
	UserSpecifiedPermissions    string = "permissions"

	FilesetComment string = "Fileset created by IBM Container Storage Interface driver"
)

func GetSpectrumScaleConnector(config settings.Clusters) (SpectrumScaleConnector, error) {
	glog.V(4).Infof("connector GetSpectrumScaleConnector")
	return NewSpectrumRestV2(config)
}
