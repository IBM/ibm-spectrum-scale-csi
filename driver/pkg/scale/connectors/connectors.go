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

import "github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors/api/v2"

//ConnectorFactory create new connectors to specific clusters
type ConnectorFactory interface {
	NewConnector(in HasClusterID) Connector
}

//Connector all operations
type Connector interface {
	HasClusterID
	ClusterConnector
	FilesystemConnector
	FilesetConnector
	DirectoryConnector
}

//HasClusterID operation
type HasClusterID interface {
	ClusterID() string //uint64
}

//ClusterConnector operations
type ClusterConnector interface {
	GetClusterId() (string, error)
}

//FilesystemConnector operations
type FilesystemConnector interface {
	ListFilesystems() ([]string, error)
	GetFilesystemName(filesystemUUID string) (string, error)
	GetFsUid(filesystemName string) (string, error)

	GetFilesystemMountpoint(filesystemName string) (string, error)
	GetFilesystemMountDetails(filesystemName string) (api.MountInfo, error)
	IsFilesystemMounted(filesystemName string) (bool, error)

	MountFilesystem(filesystemName string, nodeName string) error
	UnmountFilesystem(filesystemName string, nodeName string) error
}

//FilesetConnector operations
type FilesetConnector interface {
	CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error
	DeleteFileset(filesystemName string, filesetName string) error

	LinkFileset(filesystemName string, filesetName string, linkpath string) error
	UnlinkFileset(filesystemName string, filesetName string) error

	ListFileset(filesystemName string, filesetName string) (api.Fileset_v2, error)
	GetFileSetUid(filesystemName string, filesetName string) (string, error)
	GetFileSetNameFromId(filesystemName string, Id string) (string, error)
	IsFilesetLinked(filesystemName string, filesetName string) (bool, error)

	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(filesystemName string, filesetName string) (string, error)
	SetFilesetQuota(filesystemName string, filesetName string, quota string) error
	CheckIfFSQuotaEnabled(filesystem string) error
}

//DirectoryConnector operations
type DirectoryConnector interface {
	MakeDirectory(filesystemName string, relativePath string, uid string, gid string) error
	DeleteDirectory(filesystemName string, dirName string) error

	CheckIfFileDirPresent(filesystemName string, relPath string) (bool, error)

	CreateSymLink(SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error
	DeleteSymLnk(filesystemName string, LnkName string) error
}
