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

package sanity

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors/api/v2"
)

type fakeConnectorFactory struct {
	clusters map[string]*fakeConnector
}

type fakeConnector struct {
	hasClusterID  connectors.HasClusterID
	filesystems   map[string]*api.FileSystem_v2
	filesets      map[fsFsetKey]*api.Fileset_v2
	filesetIDs    map[int]fsFsetKey
	filesetQuotas map[fsFsetKey]*api.Quota_v2
	directories   map[dir]api.OwnerInfo
	symLinks      map[dir]dir
}

type fsFsetKey struct {
	fs   string
	fset string
}

type dir struct {
	fs      string
	relPath string
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func newFakeConnectorFactory() fakeConnectorFactory {
	return fakeConnectorFactory{
		clusters: make(map[string]*fakeConnector),
	}
}

func (fake fakeConnectorFactory) NewConnector(in connectors.HasClusterID) connectors.Connector {
	cid := in.ClusterID()
	if conn, ok := fake.clusters[cid]; ok {
		return conn
	}
	conn := &fakeConnector{
		hasClusterID:  in,
		filesystems:   make(map[string]*api.FileSystem_v2),
		filesets:      make(map[fsFsetKey]*api.Fileset_v2),
		filesetIDs:    make(map[int]fsFsetKey),
		filesetQuotas: make(map[fsFsetKey]*api.Quota_v2),
		directories:   make(map[dir]api.OwnerInfo),
		symLinks:      make(map[dir]dir),
	}
	fake.clusters[cid] = conn
	return conn
}

func (fake *fakeConnector) ClusterID() string {
	return fake.hasClusterID.ClusterID()
}
func (fake *fakeConnector) GetClusterId() (string, error) {
	return fake.hasClusterID.ClusterID(), nil
}

//unused by implementation
func (fake *fakeConnector) ListFilesystems() ([]string, error) {
	keys := make([]string, 0, len(fake.filesystems))
	for key := range fake.filesystems {
		keys = append(keys, key)
	}
	return keys, nil
}
func (fake *fakeConnector) GetFilesystemName(fsuuid string) (string, error) {
	//fsuuid == filesystemName in our fake env
	_, ok := fake.filesystems[fsuuid]
	if !ok {
		return "", fmt.Errorf("Unable to fetch filesystem name details for %s", fsuuid)
	}

	return fsuuid, nil
}
func (fake *fakeConnector) GetFsUid(filesystemName string) (string, error) {
	//fsuuid == filesystemName in our fake env
	_, ok := fake.filesystems[filesystemName]
	if !ok {
		return "", fmt.Errorf("Unable to fetch filesystem name details for %s", filesystemName)
	}

	return filesystemName, nil
}

func (fake *fakeConnector) GetFilesystemMountpoint(filesystemName string) (string, error) {
	if fs, ok := fake.filesystems[filesystemName]; ok {
		return fs.Mount.MountPoint, nil
	}
	return "", fmt.Errorf("filesystem not found")
}
func (fake *fakeConnector) GetFilesystemMountDetails(filesystemName string) (api.MountInfo, error) {
	if fs, ok := fake.filesystems[filesystemName]; ok {
		return fs.Mount, nil
	}
	return api.MountInfo{}, fmt.Errorf("filesystem not found")
}
func (fake *fakeConnector) IsFilesystemMounted(filesystemName string) (bool, error) {
	if fs, ok := fake.filesystems[filesystemName]; ok {
		//we assume single node possible...
		return len(fs.Mount.NodesMounted) > 0, nil
	}
	return false, fmt.Errorf("filesystem not found")
}

func (fake *fakeConnector) MountFilesystem(filesystemName string, nodeName string) error {
	if fs, ok := fake.filesystems[filesystemName]; ok {
		//we assume single node possible...
		fs.Mount.NodesMounted = []string{nodeName}
	}
	fake.filesystems[filesystemName] = &api.FileSystem_v2{
		Mount: api.MountInfo{
			MountPoint:   "default", //assumed default path
			NodesMounted: []string{nodeName},
		},
	}
	return nil
}
func (fake *fakeConnector) UnmountFilesystem(filesystemName string, nodeName string) error {
	if fs, ok := fake.filesystems[filesystemName]; ok {
		//we assume single node possible...
		fs.Mount.NodesMounted = []string{}
	}
	return fmt.Errorf("filesystem not found")
}

func (fake *fakeConnector) CreateFileset(filesystemName string, fsetName string, opts map[string]interface{}) error {
	fsetNameHash := int(hash(fsetName))
	fsFsetKey := fsFsetKey{filesystemName, fsetName}

	fake.filesetIDs[fsetNameHash] = fsFsetKey
	fake.filesets[fsFsetKey] = &api.Fileset_v2{
		FilesetName: fsetName,
		Config: api.FilesetConfig_v2{
			Id: fsetNameHash,
		},
	}
	return nil
}
func (fake *fakeConnector) DeleteFileset(filesystemName string, filesetName string) error {
	if _, ok := fake.filesets[fsFsetKey{filesystemName, filesetName}]; ok {
		delete(fake.filesets, fsFsetKey{filesystemName, filesetName})
		return nil
	}
	return fmt.Errorf("fileset not found")
}

func (fake *fakeConnector) LinkFileset(filesystemName string, filesetName string, linkpath string) error {
	if fset, ok := fake.filesets[fsFsetKey{filesystemName, filesetName}]; ok {
		fset.Config.Path = linkpath
		return nil
	}
	return fmt.Errorf("fileset not found")
}
func (fake *fakeConnector) UnlinkFileset(filesystemName string, filesetName string) error {
	if fset, ok := fake.filesets[fsFsetKey{filesystemName, filesetName}]; ok {
		fset.Config.Path = ""
		return nil
	}
	return fmt.Errorf("fileset not found")
}

func (fake *fakeConnector) ListFileset(filesystemName string, filesetName string) (api.Fileset_v2, error) {
	if fset, ok := fake.filesets[fsFsetKey{filesystemName, filesetName}]; ok {
		return *fset, nil
	}
	return api.Fileset_v2{}, fmt.Errorf("fileset not found")
}
func (fake *fakeConnector) GetFileSetUid(filesystemName string, filesetName string) (string, error) {
	if fset, ok := fake.filesets[fsFsetKey{filesystemName, filesetName}]; ok {
		return strconv.Itoa(fset.Config.Id), nil
	}
	return "", fmt.Errorf("fileset not found")
}
func (fake *fakeConnector) GetFileSetNameFromId(filesystemName string, Id string) (string, error) {
	id, err := strconv.Atoi(Id)
	if err != nil {
		return "", err
	}
	if fsetKey, ok := fake.filesetIDs[id]; ok {
		return fsetKey.fset, nil
	}
	return "", fmt.Errorf("fileset by id not found")
}
func (fake *fakeConnector) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	if fset, ok := fake.filesets[fsFsetKey{filesystemName, filesetName}]; ok {
		return fset.Config.Path == "", nil
	}
	return false, fmt.Errorf("fileset not found")
}

func (fake *fakeConnector) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	quota, ok := fake.filesetQuotas[fsFsetKey{filesystemName, filesetName}]
	if !ok {
		return "", fmt.Errorf("fileset not found")
	}

	return strconv.Itoa(quota.BlockLimit), nil
}
func (fake *fakeConnector) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	quotaBytes, err := strconv.Atoi(quota)
	if err != nil {
		return fmt.Errorf("quota bytes error: %w", err)
	}
	fake.filesetQuotas[fsFsetKey{filesystemName, filesetName}] = &api.Quota_v2{
		BlockLimit: quotaBytes,
	}
	return nil
}

//assume enabled
func (fake *fakeConnector) CheckIfFSQuotaEnabled(filesystemName string) error {
	return nil
}

func (fake *fakeConnector) MakeDirectory(filesystemName string, relPath string, user string, group string) error {
	dirreq := api.OwnerInfo{}

	if user != "" {
		uid, err := strconv.Atoi(user)
		if err != nil {
			dirreq.User = user
		} else {
			dirreq.UID = uid
		}
	} else {
		dirreq.UID = 0
	}

	if group != "" {
		gid, err := strconv.Atoi(group)
		if err != nil {
			dirreq.Group = group
		} else {
			dirreq.GID = gid
		}
	} else {
		dirreq.GID = 0
	}

	fake.directories[dir{filesystemName, relPath}] = dirreq
	fmt.Printf("mkdir: %v", dir{filesystemName, relPath})
	return nil
}
func (fake *fakeConnector) DeleteDirectory(filesystemName string, relPath string) error {
	if _, ok := fake.directories[dir{filesystemName, relPath}]; ok {
		fmt.Printf("rmdir: %v", dir{filesystemName, relPath})
		delete(fake.directories, dir{filesystemName, relPath})
		return nil
	}
	return fmt.Errorf("dir not found")
}
func (fake *fakeConnector) CheckIfFileDirPresent(filesystemName string, relPath string) (bool, error) {
	_, ok := fake.directories[dir{filesystemName, relPath}]
	return ok, nil
}
func (fake *fakeConnector) CreateSymLink(SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error {
	if _, ok := fake.symLinks[dir{SlnkfilesystemName, LnkPath}]; !ok {
		fake.symLinks[dir{SlnkfilesystemName, LnkPath}] = dir{TargetFs, relativePath}
		return nil
	}
	return nil //our connector returns nil for: fmt.Errorf("symLink already exists")
}
func (fake *fakeConnector) DeleteSymLnk(filesystemName string, LnkName string) error {
	if _, ok := fake.symLinks[dir{filesystemName, LnkName}]; ok {
		delete(fake.symLinks, dir{filesystemName, LnkName})
		return nil
	}
	return nil //our connector returns nil for: fmt.Errorf("symLink doesn't exists")
}
