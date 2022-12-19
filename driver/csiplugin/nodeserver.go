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
	"github.com/golang/glog"
	"os"
	"strings"
	"sync"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"golang.org/x/net/context"
	"k8s.io/mount-utils"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScaleNodeServer struct {
	Driver *ScaleDriver
	// TODO: Only lock mutually exclusive calls and make locking more fine grained
	mux sync.Mutex
}

const hostDir = "/host"
const errStaleNFSFileHandle = "stale NFS file handle"

// checkGpfsType checks if a given path is of type gpfs and
// returns nil if it is a gpfs type, otherwise returns
// corresponding error.
func checkGpfsType(path string) (bool error) {
	args := []string{"-f", "-c", "%T", path}
	out, err := executeCmd("stat", args)
	if err != nil {
		return fmt.Errorf("checkGpfsType: failed to get type of file with stat of [%s]. Error [%v]", path, err)
	}
	outString := fmt.Sprintf("%s", out)
	outString = strings.TrimRight(outString, "\n")
	if outString != "gpfs" {
		return fmt.Errorf("checkGpfsType: the path [%s] is not a valid gpfs path, the path is of type [%s]", strings.TrimPrefix(path, hostDir), outString)
	}
	return nil
}

func (ns *ScaleNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.Infof("[%s] nodeserver NodePublishVolume", loggerId)

	glog.V(4).Infof("[%s] NodePublishVolume called with req: %#v", loggerId, req)

	// Validate Arguments
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()
	volumeCapability := req.GetVolumeCapability()

	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "target path must be provided")
	}
	if volumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "volume capability must be provided")
	}

	volumeIDMembers, err := getVolIDMembers(volumeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume : VolumeID is not in proper format")
	}
	volScalePath := volumeIDMembers.Path

	glog.V(4).Infof("[%s] Target SpectrumScale Path : %v\n", loggerId, volScalePath)

	volScalePathInContainer := hostDir + volScalePath
	f, err := os.Lstat(volScalePathInContainer)
	if err != nil {
		return nil, fmt.Errorf("NodePublishVolume: failed to get lstat of [%s]. Error [%v]", volScalePathInContainer, err)
	}
	if f.Mode()&os.ModeSymlink != 0 {
		symlinkTarget, readlinkErr := os.Readlink(volScalePathInContainer)
		if readlinkErr != nil {
			return nil, fmt.Errorf("NodePublishVolume: failed to get symlink target for [%s]. Error [%v]", volScalePathInContainer, readlinkErr)
		}
		volScalePathInContainer = hostDir + symlinkTarget
		volScalePath = symlinkTarget
		glog.Infof("[%s] NodePublishVolume: symlink tarrget path is [%s]\n", loggerId, volScalePathInContainer)
	}

	err = checkGpfsType(volScalePathInContainer)
	if err != nil {
		return nil, err
	}
	notMP, err := mount.IsNotMountPoint(mount.New(""), targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(targetPath, 0750); err != nil {
				return nil, fmt.Errorf("failed to create target path [%s]. Error [%v]", targetPath, err)
			}
		} else {
			return nil, fmt.Errorf("failed to check target path [%s]. Error [%v]", targetPath, err)
		}
	}
	if !notMP {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// create bind mount
	options := []string{"bind"}
	mounter := mount.New("")
	glog.Infof("[%s] NodePublishVolume - creating bind mount [%v] -> [%v]", loggerId, targetPath, volScalePath)
	if err := mounter.Mount(volScalePath, targetPath, "", options); err != nil {
		return nil, fmt.Errorf("failed to mount: [%s] at [%s]. Error [%v]", volScalePath, targetPath, err)
	}

	//check for the gpfs type again, if not gpfs type, unmount and return error.
	err = checkGpfsType(volScalePathInContainer)
	if err != nil {
		uerr := mount.New("").Unmount(targetPath)
		if uerr != nil {
			return nil, fmt.Errorf("NodePublishVolume - failed to unmount the path [%s]. Error %v", targetPath, uerr)
		}
		return nil, err
	}
	glog.Infof("[%s] successfully mounted %s", loggerId, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// unmountAndDelete unmounts and deletes a targetPath (forcefully if
// foreceful=true is passed) and returns a bool which tells if a
// calling function should return, along with the response and error
// to be returned if there are any.
func unmountAndDelete(targetPath string, forceful bool) (bool, *csi.NodeUnpublishVolumeResponse, error) {
	glog.Infof("nodeserver unmountAndDelete")
	targetPathInContainer := hostDir + targetPath
	isMP := false
	var err error
	if !forceful {
		isMP, err = mount.New("").IsMountPoint(targetPathInContainer)
		if err != nil {
			if os.IsNotExist(err) {
				glog.V(4).Infof("target path %v is already deleted", targetPathInContainer)
				return true, &csi.NodeUnpublishVolumeResponse{}, nil
			}
			return true, nil, fmt.Errorf("failed to check if target path [%s] is a mount point. Error %v", targetPathInContainer, err)
		}
	}
	if forceful || isMP {
		// Unmount the targetPath
		err = mount.New("").Unmount(targetPath)
		if err != nil {
			return true, nil, fmt.Errorf("failed to unmount the mount point [%s]. Error %v", targetPath, err)
		}
		glog.Infof("%v is unmounted successfully", targetPath)
	}
	// Delete the mount point
	if err = os.Remove(targetPathInContainer); err != nil {
		return true, nil, fmt.Errorf("failed to remove the mount point [%s]. Error %v", targetPathInContainer, err)
	}
	return false, nil, nil
}

func (ns *ScaleNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.Infof("[%s] nodeserver NodeUnpublishVolume", loggerId)
	glog.V(4).Infof("[%s] NodeUnpublishVolume called with args: %v", loggerId, req)
	// Validate Arguments
	targetPath := req.GetTargetPath()
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "target path must be provided")
	}

	glog.Infof("[%s] NodeUnpublishVolume - deleting the targetPath - [%v]", loggerId, targetPath)

	//Check if target is a symlink or bind mount and cleanup accordingly
	f, err := os.Lstat(targetPath)
	if err != nil {
		if strings.Contains(err.Error(), errStaleNFSFileHandle) {
			glog.Errorf("[%s] Error [%v] is observed, trying forceful unmount of [%s]", loggerId, err, targetPath)
			needReturn, response, error := unmountAndDelete(targetPath, true)
			if needReturn {
				return response, error
			}
			return &csi.NodeUnpublishVolumeResponse{}, nil
		} else {
			return nil, fmt.Errorf("failed to get lstat of target path [%s]. Error %v", targetPath, err)
		}
	}
	if f.Mode()&os.ModeSymlink != 0 {
		glog.Infof("[%s] %v is a symlink", loggerId, targetPath)
		if err := os.Remove(targetPath); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove symlink targetPath [%v]. Error [%v]", targetPath, err.Error()))
		}
	} else {
		glog.Infof("[%s] %v is a bind mount", loggerId, targetPath)
		needReturn, response, error := unmountAndDelete(targetPath, false)
		if needReturn {
			return response, error
		}
	}
	glog.Infof("[%s] successfully unpublished %s", loggerId, targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *ScaleNodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.Infof("[%s] nodeserver NodeStageVolume", loggerId)
	ns.mux.Lock()
	defer ns.mux.Unlock()
	glog.V(4).Infof("[%s] NodeStageVolume called with req: %#v", loggerId, req)

	// Validate Arguments
	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	volumeCapability := req.GetVolumeCapability()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume ID must be provided")
	}
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Staging Target Path must be provided")
	}
	if volumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume Capability must be provided")
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *ScaleNodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.Infof("[%s] nodeserver NodeUnstageVolume", loggerId)
	ns.mux.Lock()
	defer ns.mux.Unlock()
	glog.V(4).Infof("[%s] NodeUnstageVolume called with req: %#v", loggerId, req)

	// Validate arguments
	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume Volume ID must be provided")
	}
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume Staging Target Path must be provided")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *ScaleNodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.V(4).Infof("[%s] NodeGetCapabilities called with req: %#v", loggerId, req)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *ScaleNodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.V(4).Infof("[%s] NodeGetInfo called with req: %#v", loggerId, req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *ScaleNodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *ScaleNodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	glog.V(4).Infof("[%s] NodeGetVolumeStats called with req: %#v", loggerId, req)

	if len(req.VolumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume ID must be provided")
	}
	if len(req.VolumePath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Target Path must be provided")
	}

	if _, err := os.Lstat(req.VolumePath); err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "path %s does not exist", req.VolumePath)
		}
		return nil, status.Errorf(codes.Internal, "failed to stat path %s: %v", req.VolumePath, err)
	}

	volumeIDMembers, err := getVolIDMembers(req.GetVolumeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats : VolumeID is not in proper format")
	}

	if !volumeIDMembers.IsFilesetBased {
		return nil, status.Error(codes.InvalidArgument, "volume stats are not supported for lightweight volumes")
	}

	available, capacity, used, inodes, inodesFree, inodesUsed, err := utils.FsStatInfo(req.GetVolumePath())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume stat failed with error %v", err))
	}

	if available > capacity || used > capacity {
		glog.Infof("[%s] Incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			loggerId, volumeIDMembers.FsetName, available, capacity)

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			volumeIDMembers.FsetName, available, capacity))
	}

	glog.V(4).Infof("[%s] Stat for volume:%v, Total:%v, Used:%v Available:%v, Total Inodes:%v, Used Inodes:%v, Available Inodes:%v,",
		loggerId, volumeIDMembers.FsetName, capacity, used, available, inodes, inodesUsed, inodesFree)

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: available,
				Total:     capacity,
				Used:      used,
				Unit:      csi.VolumeUsage_BYTES,
			}, {
				Available: inodesFree,
				Used:      inodesUsed,
				Total:     inodes,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil

}
