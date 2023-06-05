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
	"k8s.io/klog/v2"
)

//go:generate counterfeiter -o ../../../fakes/fake_spectrum.go . SpectrumScaleConnector
type SpectrumScaleConnector interface {
	//Cluster operations
	GetClusterId(ctx context.Context) (string, error)
	GetClusterSummary(ctx context.Context) (ClusterSummary, error)
	GetTimeZoneOffset(ctx context.Context) (string, error)
	GetScaleVersion(ctx context.Context) (string, error)
	//Filesystem operations
	GetFilesystemMountDetails(ctx context.Context, filesystemName string) (MountInfo, error)
	IsFilesystemMountedOnGUINode(ctx context.Context, filesystemName string) (bool, error)
	ListFilesystems(ctx context.Context) ([]string, error)
	GetFilesystemDetails(ctx context.Context, filesystemName string) (FileSystem_v2, error)
	GetFilesystemMountpoint(ctx context.Context, filesystemName string) (string, error)
	//Fileset operations
	CreateFileset(ctx context.Context, filesystemName string, filesetName string, opts map[string]interface{}) error
	UpdateFileset(ctx context.Context, filesystemName string, filesetName string, opts map[string]interface{}) error
	DeleteFileset(ctx context.Context, filesystemName string, filesetName string) error
	//LinkFileset(filesystemName string, filesetName string) error
	LinkFileset(ctx context.Context, filesystemName string, filesetName string, linkpath string) error
	UnlinkFileset(ctx context.Context, filesystemName string, filesetName string) error
	//ListFilesets(filesystemName string) ([]resources.Volume, error)
	ListFileset(ctx context.Context, filesystemName string, filesetName string) (Fileset_v2, error)
	GetFilesetsInodeSpace(ctx context.Context, filesystemName string, inodeSpace int) ([]Fileset_v2, error)
	IsFilesetLinked(ctx context.Context, filesystemName string, filesetName string) (bool, error)
	FilesetRefreshTask(ctx context.Context) error
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(ctx context.Context, filesystemName string, filesetName string) (string, error)
	GetFilesetQuotaDetails(ctx context.Context, filesystemName string, filesetName string) (Quota_v2, error)
	SetFilesetQuota(ctx context.Context, filesystemName string, filesetName string, quota string) error
	CheckIfFSQuotaEnabled(ctx context.Context, filesystem string) error
	CheckIfFilesetExist(ctx context.Context, filesystemName string, filesetName string) (bool, error)
	//Directory operations
	MakeDirectory(ctx context.Context, filesystemName string, relativePath string, uid string, gid string) error
	MakeDirectoryV2(ctx context.Context, filesystemName string, relativePath string, uid string, gid string, permissions string) error
	MountFilesystem(ctx context.Context, filesystemName string, nodeName string) error
	UnmountFilesystem(ctx context.Context, filesystemName string, nodeName string) error
	GetFilesystemName(ctx context.Context, filesystemUUID string) (string, error)
	CheckIfFileDirPresent(ctx context.Context, filesystemName string, relPath string) (bool, error)
	CreateSymLink(ctx context.Context, SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error
	GetFsUid(ctx context.Context, filesystemName string) (string, error)
	DeleteDirectory(ctx context.Context, filesystemName string, dirName string, safe bool) error
	StatDirectory(ctx context.Context, filesystemName string, dirName string) (string, error)
	GetFileSetUid(ctx context.Context, filesystemName string, filesetName string) (string, error)
	GetFileSetNameFromId(ctx context.Context, filesystemName string, Id string) (string, error)
	DeleteSymLnk(ctx context.Context, filesystemName string, LnkName string) error
	GetFileSetResponseFromId(ctx context.Context, filesystemName string, Id string) (Fileset_v2, error)
	GetFileSetResponseFromName(ctx context.Context, filesystemName string, filesetName string) (Fileset_v2, error)
	SetFilesystemPolicy(ctx context.Context, policy *Policy, filesystemName string) error
	DoesTierExist(ctx context.Context, tierName string, filesystemName string) error
	GetTierInfoFromName(ctx context.Context, tierName string, filesystemName string) (*StorageTier, error)
	GetFirstDataTier(ctx context.Context, filesystemName string) (string, error)
	IsValidNodeclass(ctx context.Context, nodeclass string) (bool, error)
	IsSnapshotSupported(ctx context.Context) (bool, error)
	CheckIfDefaultPolicyPartitionExists(ctx context.Context, partitionName string, filesystemName string) bool

	//Snapshot operations
	WaitForJobCompletion(ctx context.Context, statusCode int, jobID uint64) error
	WaitForJobCompletionWithResp(ctx context.Context, statusCode int, jobID uint64) (GenericResponse, error)
	CreateSnapshot(ctx context.Context, filesystemName string, filesetName string, snapshotName string) error
	DeleteSnapshot(ctx context.Context, filesystemName string, filesetName string, snapshotName string) error
	GetLatestFilesetSnapshots(ctx context.Context, filesystemName string, filesetName string) ([]Snapshot_v2, error)
	GetSnapshotUid(ctx context.Context, filesystemName string, filesetName string, snapName string) (string, error)
	GetSnapshotCreateTimestamp(ctx context.Context, filesystemName string, filesetName string, snapName string) (string, error)
	CheckIfSnapshotExist(ctx context.Context, filesystemName string, filesetName string, snapshotName string) (bool, error)
	ListFilesetSnapshots(ctx context.Context, filesystemName string, filesetName string) ([]Snapshot_v2, error)
	CopyFsetSnapshotPath(ctx context.Context, filesystemName string, filesetName string, snapshotName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
	CopyFilesetPath(ctx context.Context, filesystemName string, filesetName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
	CopyDirectoryPath(ctx context.Context, filesystemName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error)
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

func GetSpectrumScaleConnector(ctx context.Context, config settings.Clusters) (SpectrumScaleConnector, error) {
	klog.V(4).Infof("[%s] connector GetSpectrumScaleConnector", utils.GetLoggerId(ctx))
	return NewSpectrumRestV2(ctx, config)
}
