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
	"os"
	"sync"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/golang/glog"
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

func (ns *ScaleNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.V(3).Infof("nodeserver NodePublishVolume")

	glog.V(4).Infof("NodePublishVolume called with req: %#v", req)

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

	glog.V(4).Infof("Target SpectrumScale Path : %v\n", volScalePath)

	// Check if /host directory exists, if exists use bind mount,
	// otherwise use symlink
	hostDirMounted := false
	if _, err = os.Stat(hostDir); err == nil {
		hostDirMounted = true
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to get stat of [%s]. Error [%v]", hostDir, err)
	}

	if !hostDirMounted {
		//Use symlink
		//Check if mount dir/slink exists, if yes delete it
		if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
			glog.V(4).Infof("NodePublishVolume - deleting the targetPath - [%v]", targetPath)
			err := os.Remove(targetPath)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete the target path - [%s]. Error [%v]", targetPath, err.Error()))
			}
		}

		// create symlink
		glog.V(4).Infof("NodePublishVolume - creating symlink [%v] -> [%v]", targetPath, volScalePath)
		symlinkerr := os.Symlink(volScalePath, targetPath)
		if symlinkerr != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create symlink [%s] -> [%s]. Error [%v]", targetPath, volScalePath, symlinkerr.Error()))
		}
	} else {
		//Use bind mount
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
		glog.V(4).Infof("NodePublishVolume - creating bind mount [%v] -> [%v]", targetPath, volScalePath)
		if err := mounter.Mount(volScalePath, targetPath, "", options); err != nil {
			return nil, fmt.Errorf("failed to mount: [%s] at [%s]. Error [%v]", volScalePath, targetPath, err)
		}
	}

	glog.V(4).Infof("successfully mounted %s", targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *ScaleNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	glog.V(3).Infof("nodeserver NodeUnpublishVolume")
	glog.V(4).Infof("NodeUnpublishVolume called with args: %v", req)
	// Validate Arguments
	targetPath := req.GetTargetPath()
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "target path must be provided")
	}

	glog.V(4).Infof("NodeUnpublishVolume - deleting the targetPath - [%v]", targetPath)

	//Check if target is a symlink or bind mount and cleanup accordingly
	f, err := os.Lstat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get lstat of target path [%s]. Error %v", targetPath, err)
	}
	if f.Mode()&os.ModeSymlink != 0 {
		glog.V(4).Infof("%v is a symlink", targetPath)
		if err := os.Remove(targetPath); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove symlink targetPath [%v]. Error [%v]", targetPath, err.Error()))
		}
	} else {
		glog.V(4).Infof("%v is a bind mount", targetPath)
		targetPathInContainer := hostDir + targetPath
		notMP, err := mount.IsNotMountPoint(mount.New(""), targetPathInContainer)
		if err != nil {
			if os.IsNotExist(err) {
				glog.V(4).Infof("target path %v is already deleted", targetPathInContainer)
				return &csi.NodeUnpublishVolumeResponse{}, nil
			}
			return nil, fmt.Errorf("failed to check if target path [%s] is mount point. Error %v", targetPathInContainer, err)
		}
		if !notMP {
			// Unmount the targetPath
			err = mount.New("").Unmount(targetPath)
			if err != nil {
				return nil, fmt.Errorf("failed to unmount the mount point [%s]. Error %v", targetPath, err)
			}
			glog.V(4).Infof("%v is a unmounted successfully", targetPath)
		}
		// Delete the mount point
		if err = os.Remove(targetPathInContainer); err != nil {
			return nil, fmt.Errorf("failed to remove the mount point [%s]. Error %v", targetPathInContainer, err)
		}
	}
	glog.V(4).Infof("successfully unpublished %s", targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *ScaleNodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	glog.V(3).Infof("nodeserver NodeStageVolume")
	ns.mux.Lock()
	defer ns.mux.Unlock()
	glog.V(4).Infof("NodeStageVolume called with req: %#v", req)

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
	glog.V(3).Infof("nodeserver NodeUnstageVolume")
	ns.mux.Lock()
	defer ns.mux.Unlock()
	glog.V(4).Infof("NodeUnstageVolume called with req: %#v", req)

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
	glog.V(4).Infof("NodeGetCapabilities called with req: %#v", req)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *ScaleNodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	glog.V(4).Infof("NodeGetInfo called with req: %#v", req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *ScaleNodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *ScaleNodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	glog.V(4).Infof("NodeGetVolumeStats called with req: %#v", req)

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
		glog.V(4).Infof("incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			volumeIDMembers.FsetName, available, capacity)

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			volumeIDMembers.FsetName, available, capacity))
	}

	glog.V(4).Infof("stat for volume:%v, Total:%v, Used:%v Available:%v, Total Inodes:%v, Used Inodes:%v, Available Inodes:%v,",
		volumeIDMembers.FsetName, capacity, used, available, inodes, inodesUsed, inodesFree)

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
