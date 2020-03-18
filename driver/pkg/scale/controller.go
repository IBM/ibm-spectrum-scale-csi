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

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	no       = "no"
	yes      = "yes"
	notFound = "NOT_FOUND"
)

type ControllerService struct {
	fab     connectors.ConnectorFactory
	cscap   []*csi.ControllerServiceCapability
	primary *settings.Primary

	reqmap map[string]int64
}

func newControllerService(
	config *settings.ConfigMap,
	fab connectors.ConnectorFactory,
) ControllerService {
	return ControllerService{
		fab:     fab,
		primary: config.Primary,
		reqmap:  make(map[string]int64),
	}
}

func (cs *ControllerService) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	glog.V(3).Infof("gpfs ValidateControllerServiceRequest")
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}
	for _, cap := range cs.cscap {
		if c == cap.GetRpc().Type {
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, "Invalid controller service request")
}

func (cs *ControllerService) IfSameVolReqInProcess(scVol *scaleVolume) (bool, error) {
	cap, volpresent := cs.reqmap[scVol.VolName]
	glog.Infof("reqmap: %v", cs.reqmap)
	if volpresent {
		if cap == int64(scVol.VolSize) {
			return true, nil
		} else {
			return false, status.Error(codes.Internal, fmt.Sprintf("Volume %v present in map but requested size %v does not match with size %v in map", scVol.VolName, scVol.VolSize, cap))
		}
	}
	return false, nil
}

func (cs *ControllerService) IfFileSetBasedVolExist(scVol *scaleVolume) (bool, error) {
	remote := cs.fab.NewConnector(scVol)
	primary := cs.fab.NewConnector(cs.primary)

	/* Check if fileset is there. Check if quota matches and see if symlink exists*/
	_, err := remote.ListFileset(scVol.VolBackendFs, scVol.VolName)
	if err != nil {
		return false, nil
	}

	if scVol.VolSize != 0 {
		quota, err := remote.ListFilesetQuota(scVol.VolBackendFs, scVol.VolName)
		if err != nil {
			return false, status.Error(codes.Internal, fmt.Sprintf("Unable to list quota for Fset [%v] in FS [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err))
		}

		filesetQuotaBytes, err := ConvertToBytes(quota)
		if err != nil {
			return false, err
		}
		if filesetQuotaBytes != scVol.VolSize {
			return false, status.Error(codes.AlreadyExists, fmt.Sprintf("Fileset %v present but quota %v does not match with requested size %v", scVol.VolName, filesetQuotaBytes, scVol.VolSize))
		}
	}

	/* Check if Symlink Present */
	volSlnkPath := fmt.Sprintf("%s/%s", scVol.PrimarySLnkRelPath, scVol.VolName)
	symLinkExists, err := primary.CheckIfFileDirPresent(scVol.PrimaryFS, volSlnkPath)
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("Unable to check if symlink path [%v] exists in FS [%v]. Error [%v]", volSlnkPath, scVol.PrimaryFS, err))
	}

	if symLinkExists {
		return true, nil
	}

	return false, nil
}

func (cs *ControllerService) IfLwVolExist(scVol *scaleVolume) (bool, error) {
	/* Check if Dir present and see if symlink exists*/
	volPath := fmt.Sprintf("%s/%s", scVol.VolDirBasePath, scVol.VolName)
	primary := cs.fab.NewConnector(cs.primary)
	dirPresent, err := primary.CheckIfFileDirPresent(scVol.VolBackendFs, volPath)
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("Unable to check if path [%v] exists in FS [%v]. Error [%v]", volPath, scVol.VolBackendFs, err))
	}

	if dirPresent {
		/* Check if Symlink Present */

		volSlnkPath := fmt.Sprintf("%s/%s", scVol.PrimarySLnkRelPath, scVol.VolName)
		glog.Infof("Symlink fs [%v] slinkpath [%v]", scVol.PrimaryFS, volSlnkPath)
		symLinkExists, err := primary.CheckIfFileDirPresent(scVol.PrimaryFS, volSlnkPath)

		if err != nil {
			return false, status.Error(codes.Internal, fmt.Sprintf("Unable to check if symlink [%v] exists in FS [%v]. Error [%v]", volSlnkPath, scVol.PrimaryFS, err))
		}

		if symLinkExists {
			return true, nil
		}
	}
	glog.Infof("returning false for isPresent")
	return false, nil
}

func (cs *ControllerService) CreateLWVol(scVol *scaleVolume) error {
	primary := cs.fab.NewConnector(cs.primary)
	baseDirExists, err := primary.CheckIfFileDirPresent(scVol.VolBackendFs, scVol.VolDirBasePath)

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Unable to check if DirBasePath %v is present in FS %v", scVol.VolDirBasePath, scVol.VolBackendFs))
	}

	if !baseDirExists {
		return status.Error(codes.Internal, fmt.Sprintf("Directory base path %v not present in FS %v", scVol.VolDirBasePath, scVol.VolBackendFs))
	}

	dirPath := fmt.Sprintf("%s/%s", scVol.VolDirBasePath, scVol.VolName)
	/* FS from sc */
	err = primary.MakeDirectory(scVol.VolBackendFs, dirPath, scVol.VolUid, scVol.VolGid)

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Unable to create dir [%v] in FS [%v] with uid:gid [%v:%v]. Error [%v]", dirPath, scVol.VolBackendFs, scVol.VolUid, scVol.VolGid, err))
	}
	return nil
}

func (cs *ControllerService) GenerateVolId(scVol *scaleVolume) (string, error) {
	var volId string
	primary := cs.fab.NewConnector(cs.primary)

	/* We need to put FSUUID for localFS in volID */
	uid, err := primary.GetFsUid(scVol.LocalFS)
	glog.Infof("GetFsUID error [%v] uid [%v]", err, uid)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to get FS UUID for FS [%v]. Error [%v]", scVol.VolBackendFs, err))
	}

	if scVol.IsFilesetBased {
		remote := cs.fab.NewConnector(scVol)
		fSetuid, err := remote.GetFileSetUid(scVol.VolBackendFs, scVol.VolName)

		if err != nil {
			return "", status.Error(codes.Internal, fmt.Sprintf("Unable to get Fset UID for [%v] in FS [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err))
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

func (cs *ControllerService) GetFsMntPt(scVol *scaleVolume) (string, error) {
	remote := cs.fab.NewConnector(scVol)
	fsMount, err := remote.GetFilesystemMountDetails(scVol.VolBackendFs)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to fetch mount details for FS %v", scVol.VolBackendFs))
	}

	if fsMount.NodesMounted == nil || len(fsMount.NodesMounted) == 0 {
		return "", status.Error(codes.Internal, fmt.Sprintf("filesystem %v not mounted on any node", scVol.VolBackendFs))
	}
	fsMountPt := fsMount.MountPoint
	return fsMountPt, err
}

func (cs *ControllerService) GetFsetLnkPath(scVol *scaleVolume) (string, error) {
	remote := cs.fab.NewConnector(scVol)
	fsetResponse, err := remote.ListFileset(scVol.VolBackendFs, scVol.VolName)
	if err != nil {
		_ = cs.Cleanup(scVol)
		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to list Fset [%v] in FS [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, err))
	}

	linkpath := fsetResponse.Config.Path
	return linkpath, err
}

func (cs *ControllerService) GetTargetPathforFset(scVol *scaleVolume) (string, error) {
	linkpath, err := cs.GetFsetLnkPath(scVol)
	if err != nil {
		return "", err
	}
	fsMountPt, err := cs.GetFsMntPt(scVol)
	if err != nil {
		return "", err
	}
	targetPath := strings.Replace(linkpath, fsMountPt, "", 1)
	targetPath = strings.Trim(targetPath, "!/")
	targetPath = fmt.Sprintf("%s/%s-data", targetPath, scVol.VolName)
	return targetPath, nil
}

func (cs *ControllerService) CreateFilesetBasedVol(scVol *scaleVolume) (string, error) { //nolint:gocyclo,funlen
	opt := make(map[string]interface{})

	remote := cs.fab.NewConnector(scVol)
	isFsMounted, err := remote.IsFilesystemMounted(scVol.VolBackendFs)

	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to check if FS [%v] is mounted. Error [%v]", scVol.VolBackendFs, err))
	}

	if !isFsMounted {
		return "", status.Error(codes.Internal, fmt.Sprintf("Filesystem %v in cluster %v is not mounted", scVol.VolBackendFs, scVol.ClusterId))
	}

	if scVol.VolSize != 0 {
		err = remote.CheckIfFSQuotaEnabled(scVol.VolBackendFs)
		if err != nil {
			return "", status.Error(codes.Internal, fmt.Sprintf("Quota not enabled for Filesystem %v inside cluster %v", scVol.VolBackendFs, scVol.ClusterId))
		}
	}

	if scVol.VolUid != "" {
		opt[settings.Uid] = scVol.VolUid
	}
	if scVol.VolGid != "" {
		opt[settings.Gid] = scVol.VolGid
	}
	if scVol.FilesetType != "" {
		opt[settings.FilesetType] = scVol.FilesetType
	}
	if scVol.InodeLimit != "" {
		opt[settings.InodeLimit] = scVol.InodeLimit
	}
	if scVol.ParentFileset != "" {
		opt[settings.ParentFset] = scVol.ParentFileset
	}

	fseterr := remote.CreateFileset(scVol.VolBackendFs, scVol.VolName, opt)

	if fseterr != nil {
		/* Fileset creation failed, but in some cases GUI returns failure when fileset was created but not linked. So delete a incomplete created fileset, so that in next iteration we can create fresh one. */

		_, err := remote.ListFileset(scVol.VolBackendFs, scVol.VolName)

		if err == nil {
			_ = cs.Cleanup(scVol)
		}

		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to create fileset [%v] in FS [%v]. Error [%v]", scVol.VolName, scVol.VolBackendFs, fseterr))
	}

	isFilesetLinked, err := remote.IsFilesetLinked(scVol.VolBackendFs, scVol.VolName)

	if err != nil {
		_ = cs.Cleanup(scVol)
		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to check if Fset [%v] in FS [%v] is linked. Error [%v]", scVol.VolName, scVol.VolBackendFs, err))
	}

	if !isFilesetLinked {
		_ = cs.Cleanup(scVol)
		return "", status.Error(codes.Internal, fmt.Sprintf("Fileset [%v] was created in FS [%v] but was not linked", scVol.VolName, scVol.VolBackendFs))
	}

	if scVol.VolSize != 0 {
		volsiz := strconv.FormatUint(scVol.VolSize, 10)

		err = remote.SetFilesetQuota(scVol.VolBackendFs, scVol.VolName, volsiz)

		if err != nil {
			_ = cs.Cleanup(scVol)
			return "", status.Error(codes.Internal, fmt.Sprintf("Fileset [%v] was created in FS [%v] but not able to set quota [%v]", scVol.VolName, scVol.VolBackendFs, scVol.VolSize))
		}
	}

	/* Now we need to create a dir inside a fileset */
	targetBasePath, err := cs.GetTargetPathforFset(scVol)

	if err != nil {
		glog.Infof("Unable to get target Path for [%v]\n", scVol)
		_ = cs.Cleanup(scVol)
		return "", err
	}

	err = remote.MakeDirectory(scVol.VolBackendFs, targetBasePath, scVol.VolUid, scVol.VolGid)

	if err != nil {
		_ = cs.Cleanup(scVol)
		return "", status.Error(codes.Internal, fmt.Sprintf("Unable to create dir [%v] in FS [%v]", targetBasePath, scVol.VolBackendFs))
	}

	return targetBasePath, err
}

func (cs *ControllerService) GetVolumeSizeInBytes(req *csi.CreateVolumeRequest) (int64, error) {
	cap := req.GetCapacityRange()
	return cap.GetRequiredBytes(), nil
}

func (cs *ControllerService) Cleanup(scVol *scaleVolume) error {
	var err error
	if scVol.IsFilesetBased {
		remote := cs.fab.NewConnector(scVol)
		err = remote.DeleteFileset(scVol.VolBackendFs, scVol.VolName)
	} else {
		primary := cs.fab.NewConnector(cs.primary)
		dirPath := fmt.Sprintf("%s/%s", scVol.VolDirBasePath, scVol.VolName)
		glog.Infof("Directory path to be deleted [%v]", dirPath)
		err = primary.DeleteDirectory(scVol.VolBackendFs, dirPath)
	}
	return err
}

func (cs *ControllerService) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) { //nolint:gocyclo,funlen
	glog.V(3).Infof("create volume req: %v", req)

	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
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
	volSize, err := cs.GetVolumeSizeInBytes(req)

	if err != nil {
		return nil, err
	}

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

	scaleVol.PrimarySLnkRelPath = cs.primary.SymlinkRelativePath
	scaleVol.PrimaryFS = cs.primary.GetPrimaryFs()
	scaleVol.PrimaryFSMount = cs.primary.PrimaryFSMount
	scaleVol.PrimarySLnkPath = cs.primary.SymlinkAbsolutePath

	/* scaleVol.VolBackendFs will always be local cluster FS. So we need to find a
	   remote cluster FS in case local cluster FS is remotely mounted. We will find    local FS RemoteDeviceName on local cluster, will use that as VolBackendFs and   create fileset on that FS. */

	if scaleVol.IsFilesetBased {
		remote := cs.fab.NewConnector(scaleVol)
		mountInfo, err := remote.GetFilesystemMountDetails(scaleVol.VolBackendFs)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get Mount Details for FS [%v] in Primary cluster", scaleVol.VolBackendFs))
		}

		remoteDeviceName := mountInfo.RemoteDeviceName
		splitDevName := strings.Split(remoteDeviceName, ":")
		remDevFs := splitDevName[len(splitDevName)-1]

		scaleVol.LocalFS = scaleVol.VolBackendFs
		scaleVol.VolBackendFs = remDevFs
	} else {
		scaleVol.ClusterId = cs.primary.PrimaryCid
		scaleVol.LocalFS = scaleVol.VolBackendFs
	}

	glog.Infof("Scale vol create params : %v\n", scaleVol)

	volReqInProcess, err := cs.IfSameVolReqInProcess(scaleVol)
	if err != nil {
		return nil, err
	}

	if volReqInProcess {
		return nil, status.Error(codes.Aborted, fmt.Sprintf("Volume creation already in process : %v", scaleVol.VolName))
	}

	/* Update driver map with new volume. Make sure to defer delete */

	cs.reqmap[scaleVol.VolName] = volSize
	defer delete(cs.reqmap, scaleVol.VolName)

	glog.Infof("reqmap After: %v", cs.reqmap)

	/* Check if Volume already present */
	var isPresent bool
	if scaleVol.IsFilesetBased {
		isPresent, err = cs.IfFileSetBasedVolExist(scaleVol)
		if err != nil {
			return nil, err
		}
	} else {
		isPresent, err = cs.IfLwVolExist(scaleVol)
		if err != nil {
			return nil, err
		}
	}

	if isPresent {
		volId, err := cs.GenerateVolId(scaleVol)
		if err != nil {
			return nil, err
		}

		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				VolumeId:      volId,
				CapacityBytes: int64(scaleVol.VolSize),
				VolumeContext: req.GetParameters(),
			},
		}, nil
	}
	/* If we reach here we need to create a volume */
	var targetPath string

	if scaleVol.IsFilesetBased {
		targetPath, err = cs.CreateFilesetBasedVol(scaleVol)
	} else {
		err = cs.CreateLWVol(scaleVol)
	}

	if err != nil {
		return nil, err
	}

	if !scaleVol.IsFilesetBased {
		targetPath = fmt.Sprintf("%s/%s", scaleVol.VolDirBasePath, scaleVol.VolName)
	}

	/* Create a Symlink */

	lnkPath := fmt.Sprintf("%s/%s", scaleVol.PrimarySLnkRelPath, scaleVol.VolName)

	glog.Infof("Symlink info FS [%v] TargetFS [%v]  target Path [%v] lnkPath [%v]", scaleVol.PrimaryFS, scaleVol.LocalFS, targetPath, lnkPath)

	primary := cs.fab.NewConnector(cs.primary)
	err = primary.CreateSymLink(scaleVol.PrimaryFS, scaleVol.LocalFS, targetPath, lnkPath)

	if err != nil {
		_ = cs.Cleanup(scaleVol)
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to create symlink [%v] in FS [%v], for target [%v] in FS [%v]. Error [%v]", lnkPath, scaleVol.PrimaryFS, targetPath, scaleVol.LocalFS, err))
	}

	volId, err := cs.GenerateVolId(scaleVol)
	if err != nil {
		_ = cs.Cleanup(scaleVol)
		return nil, err
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volId,
			CapacityBytes: int64(scaleVol.VolSize),
			VolumeContext: req.GetParameters(),
		},
	}, nil
}

func (cs *ControllerService) GetVolIdMembers(vId string) (scaleVolId, error) {
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

func (cs *ControllerService) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Warningf("invalid delete volume req: %v", req)
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("invalid delete volume req (%v): %v", req, err))
	}
	// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeID := req.GetVolumeId()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume Id is a required field")
	}

	volumeIdMembers, err := cs.GetVolIdMembers(volumeID)

	if err != nil {
		return &csi.DeleteVolumeResponse{}, nil
	}

	glog.Infof("Volume Id Members [%v]", volumeIdMembers)

	remote := cs.fab.NewConnector(volumeIdMembers)

	if err != nil {
		return nil, err
	}

	primary := cs.fab.NewConnector(cs.primary)

	/* FsUUID in volumeIdMembers will be of Primary cluster. So lets get Name of it
	   from Primary cluster */

	FilesystemName, err := primary.GetFilesystemName(volumeIdMembers.FsUUID)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get filesystem Name for Id [%v] and clusterId [%v]. Error [%v]", volumeIdMembers.FsUUID, volumeIdMembers.ClusterId, err))
	}

	mountInfo, err := primary.GetFilesystemMountDetails(FilesystemName)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get mount info for FS [%v] in primary cluster", FilesystemName))
	}

	remoteDeviceName := mountInfo.RemoteDeviceName
	splitDevName := strings.Split(remoteDeviceName, ":")
	remDevFs := splitDevName[len(splitDevName)-1]

	FilesystemName = remDevFs

	sLinkRelPath := strings.Replace(volumeIdMembers.SymLnkPath, cs.primary.PrimaryFSMount, "", 1)
	sLinkRelPath = strings.Trim(sLinkRelPath, "!/")

	if volumeIdMembers.IsFilesetBased {
		FilesetName, err := remote.GetFileSetNameFromId(FilesystemName, volumeIdMembers.FsetId)

		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get Fileset Name for Id [%v] FS [%v] ClusterId [%v]", volumeIdMembers.FsetId, FilesystemName, volumeIdMembers.ClusterId))
		}

		if FilesetName != "" {
			/* Confirm it is same fileset which was created for this PV */
			pvName := filepath.Base(sLinkRelPath)
			if pvName == FilesetName {
				err = remote.DeleteFileset(FilesystemName, FilesetName)

				if err != nil {
					return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to Delete Fileset [%v] for FS [%v] and clusterId [%v]: %v", FilesetName, FilesystemName, volumeIdMembers.ClusterId, err))
				}
			} else {
				glog.Infof("PV name from path [%v] does not match with filesetName [%v]. Skipping delete of fileset", pvName, FilesetName)
			}
		}
	} else {
		/* Delete Dir for Lw volume */
		err = primary.DeleteDirectory(cs.primary.GetPrimaryFs(), sLinkRelPath)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to Delete Dir using FS [%v] Relative SymLink [%v]: %v", cs.primary.GetPrimaryFs(), sLinkRelPath, err))
		}
	}

	err = primary.DeleteSymLnk(cs.primary.GetPrimaryFs(), sLinkRelPath)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to delete symlnk [%v:%v] Error [%v]", cs.primary.GetPrimaryFs(), sLinkRelPath, err))
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerGetCapabilities implements the default GRPC callout.
func (cs *ControllerService) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	glog.V(4).Infof("ControllerGetCapabilities called with req: %#v", req)
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.cscap,
	}, nil
}

func (cs *ControllerService) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
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

func (cs *ControllerService) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	glog.V(3).Infof("controllerserver ControllerUnpublishVolume")
	glog.V(4).Infof("ControllerUnpublishVolume : req %#v", req)

	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); err != nil {
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

func (cs *ControllerService) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) { //nolint:gocyclo,funlen
	glog.V(3).Infof("controllerserver ControllerPublishVolume")
	glog.V(4).Infof("ControllerPublishVolume : req %#v", req)

	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); err != nil {
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

	primary := cs.fab.NewConnector(cs.primary)

	//Get filesystem name from UUID
	fsName, err := primary.GetFilesystemName(filesystemID)
	if err != nil {
		glog.Errorf("ControllerPublishVolume : Error in getting filesystem Name for filesystem ID of %s.", filesystemID)
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in getting filesystem Name for filesystem ID of %s. Error [%v]", filesystemID, err))
	}

	//Check if primary filesystem is mounted.
	primaryfsName := cs.primary.GetPrimaryFs()
	pfsMount, err := primary.GetFilesystemMountDetails(primaryfsName)
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
		fsMount, err := primary.GetFilesystemMountDetails(fsName)
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
		err = primary.MountFilesystem(primaryfsName, scalenodeID)
		if err != nil {
			glog.Errorf("ControllerPublishVolume : Error in mounting filesystem %s on node %s", primaryfsName, scalenodeID)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume :  Error in mounting filesystem %s on node %s. Error [%v]", primaryfsName, scalenodeID, err))
		}
	}

	//mount the volume filesystem if mounted
	if !(isFsMounted) && skipMountUnmount == no && primaryfsName != fsName {
		glog.V(4).Infof("ControllerPublishVolume : mounting %s on %s", fsName, scalenodeID)
		err = primary.MountFilesystem(fsName, scalenodeID)
		if err != nil {
			glog.Errorf("ControllerPublishVolume : Error in mounting filesystem %s on node %s", fsName, scalenodeID)
			return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerPublishVolume : Error in mounting filesystem %s on node %s. Error [%v]", fsName, scalenodeID, err))
		}
	}
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (cs *ControllerService) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerService) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerService) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerService) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *ControllerService) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *ControllerService) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
