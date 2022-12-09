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
	"context"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
)

var logger *utils.CsiLogger

//go:generate counterfeiter -o ../../../fakes/fake_spectrum.go . SpectrumScaleConnector
type SpectrumScaleConnector interface {
	//Cluster operations
	GetClusterId() (string, error)
	GetClusterSummary() (ClusterSummary, error)
	GetTimeZoneOffset() (string, error)
	GetScaleVersion() (string, error)
	//Filesystem operations
	GetFilesystemMountDetails(filesystemName string) (MountInfo, error)
	IsFilesystemMountedOnGUINode(filesystemName string) (bool, error)
	ListFilesystems() ([]string, error)
	GetFilesystemDetails(ctx context.Context, filesystemName string) (FileSystem_v2, error)
	GetFilesystemMountpoint(filesystemName string) (string, error)
	//Fileset operations
	CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error
	UpdateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error
	DeleteFileset(filesystemName string, filesetName string) error
	//LinkFileset(filesystemName string, filesetName string) error
	LinkFileset(filesystemName string, filesetName string, linkpath string) error
	UnlinkFileset(filesystemName string, filesetName string) error
	//ListFilesets(filesystemName string) ([]resources.Volume, error)
	ListFileset(filesystemName string, filesetName string) (Fileset_v2, error)
	GetFilesetsInodeSpace(filesystemName string, inodeSpace int) ([]Fileset_v2, error)
	IsFilesetLinked(ctx context.Context, filesystemName string, filesetName string) (bool, error)
	FilesetRefreshTask() error
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(ctx context.Context, filesystemName string, filesetName string) (string, error)
	GetFilesetQuotaDetails(filesystemName string, filesetName string) (Quota_v2, error)
	SetFilesetQuota(ctx context.Context, filesystemName string, filesetName string, quota string) error
	CheckIfFSQuotaEnabled(filesystem string) error
	CheckIfFilesetExist(ctx context.Context, filesystemName string, filesetName string) (bool, error)
	//Directory operations
	MakeDirectory(filesystemName string, relativePath string, uid string, gid string) error
	MakeDirectoryV2(filesystemName string, relativePath string, uid string, gid string, permissions string) error
	MountFilesystem(filesystemName string, nodeName string) error
	UnmountFilesystem(filesystemName string, nodeName string) error
	GetFilesystemName(ctx context.Context, filesystemUUID string) (string, error)
	CheckIfFileDirPresent(filesystemName string, relPath string) (bool, error)
	CreateSymLink(SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error
	GetFsUid(filesystemName string) (string, error)
	DeleteDirectory(ctx context.Context, filesystemName string, dirName string, safe bool) error
	StatDirectory(filesystemName string, dirName string) (string, error)
	GetFileSetUid(filesystemName string, filesetName string) (string, error)
	GetFileSetNameFromId(ctx context.Context, filesystemName string, Id string) (string, error)
	DeleteSymLnk(ctx context.Context, filesystemName string, LnkName string) error
	GetFileSetResponseFromId(ctx context.Context, filesystemName string, Id string) (Fileset_v2, error)
	GetFileSetResponseFromName(filesystemName string, filesetName string) (Fileset_v2, error)
	SetFilesystemPolicy(ctx context.Context, policy *Policy, filesystemName string) error
	DoesTierExist(ctx context.Context, tierName string, filesystemName string) error
	GetTierInfoFromName(tierName string, filesystemName string) (*StorageTier, error)
	GetFirstDataTier(ctx context.Context, filesystemName string) (string, error)
	IsValidNodeclass(nodeclass string) (bool, error)
	IsSnapshotSupported() (bool, error)
	CheckIfDefaultPolicyPartitionExists(ctx context.Context, partitionName string, filesystemName string) bool

	//Snapshot operations
	WaitForJobCompletion(statusCode int, jobID uint64) error
	WaitForJobCompletionWithResp(statusCode int, jobID uint64) (GenericResponse, error)
	CreateSnapshot(ctx context.Context, filesystemName string, filesetName string, snapshotName string) error
	DeleteSnapshot(ctx context.Context, filesystemName string, filesetName string, snapshotName string) error
	GetLatestFilesetSnapshots(filesystemName string, filesetName string) ([]Snapshot_v2, error)
	GetSnapshotUid(filesystemName string, filesetName string, snapName string) (string, error)
	GetSnapshotCreateTimestamp(filesystemName string, filesetName string, snapName string) (string, error)
	CheckIfSnapshotExist(ctx context.Context, filesystemName string, filesetName string, snapshotName string) (bool, error)
	ListFilesetSnapshots(ctx context.Context, filesystemName string, filesetName string) ([]Snapshot_v2, error)
	CopyFsetSnapshotPath(filesystemName string, filesetName string, snapshotName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
	CopyFilesetPath(filesystemName string, filesetName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
	CopyDirectoryPath(filesystemName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
	IsNodeComponentHealthy(ctx context.Context, nodeName string, component string) (bool, error)
}

const (
	UserSpecifiedFilesetType      string = "filesetType"
	UserSpecifiedFilesetTypeDep   string = "fileset-type"
	UserSpecifiedInodeLimit       string = "inodeLimit"
	UserSpecifiedInodeLimitDep    string = "inode-limit"
	UserSpecifiedUid              string = "uid"
	UserSpecifiedGid              string = "gid"
	UserSpecifiedClusterId        string = "clusterId"
	UserSpecifiedParentFset       string = "parentFileset"
	UserSpecifiedVolBackendFs     string = "volBackendFs"
	UserSpecifiedVolDirPath       string = "volDirBasePath"
	UserSpecifiedNodeClass        string = "nodeClass"
	UserSpecifiedPermissions      string = "permissions"
	UserSpecifiedStorageClassType string = "version"
	UserSpecifiedCompression      string = "compression"
	UserSpecifiedTier             string = "tier"
	UserSpecifiedSnapWindow       string = "snapWindow"
	UserSpecifiedConsistencyGroup string = "consistencyGroup"
	UserSpecifiedShared           string = "shared"

	FilesetComment string = "Fileset created by IBM Container Storage Interface driver"
)

func GetSpectrumScaleConnector(config settings.Clusters) (SpectrumScaleConnector, error) {
	logger.Infof("connector GetSpectrumScaleConnector")
	return NewSpectrumRestV2(config)
}
