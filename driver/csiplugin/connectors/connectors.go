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
	//Filesystem operations
	GetFilesystemMountDetails(filesystemName string) (MountInfo, error)
	IsFilesystemMounted(filesystemName string) (bool, error)
	ListFilesystems() ([]string, error)
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
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(filesystemName string, filesetName string) (string, error)
	SetFilesetQuota(filesystemName string, filesetName string, quota string) error
	CheckIfFSQuotaEnabled(filesystem string) error
	//Directory operations
	MakeDirectory(filesystemName string, relativePath string, uid string, gid string) error
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
)

func GetSpectrumScaleConnector(config settings.Clusters) (SpectrumScaleConnector, error) {
	glog.V(4).Infof("connector GetSpectrumScaleConnector")
	return NewSpectrumRestV2(config)
}
