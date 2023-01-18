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
	"strings"
	"sync"
	"time"

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
	glog.V(3).Infof("nodeserver NodePublishVolume")
	glog.V(4).Infof("NodePublishVolume called with req: %#v", req)
	start := time.Now()
	defer glog.V(4).Infof("NodePublishVolume : req %#v time spent : %v", req, time.Since(start))

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

	volScalePathInContainer := hostDir + volScalePath
	f, err := os.Lstat(volScalePathInContainer)
	if err != nil {
		glog.V(4).Infof("NodePublishVolume - failed to get lstat of [%s]. Error [%v]", volScalePathInContainer, err)
		return nil, fmt.Errorf("NodePublishVolume - failed to get lstat of [%s]. Error [%v]", volScalePathInContainer, err)
	}
	if f.Mode()&os.ModeSymlink != 0 {
		symlinkTarget, readlinkErr := os.Readlink(volScalePathInContainer)
		if readlinkErr != nil {
			glog.V(4).Infof("NodePublishVolume - failed to get symlink target for [%s]. Error [%v]", volScalePathInContainer, readlinkErr)
			return nil, fmt.Errorf("NodePublishVolume - failed to get symlink target for [%s]. Error [%v]", volScalePathInContainer, readlinkErr)
		}
		volScalePathInContainer = hostDir + symlinkTarget
		volScalePath = symlinkTarget
		glog.V(4).Infof("NodePublishVolume - symlink target path is [%s]\n", volScalePathInContainer)
	}

	err = checkGpfsType(volScalePathInContainer)
	if err != nil {
		glog.V(4).Infof("NodePublishVolume - the path [%v] is not a valid gpfs path", volScalePathInContainer)
		return nil, err
	}
	//Use symlink by default for 2.7.0 fixpack2

	//There can be 2 symlinks here:
	//1. symlink1 (volScalePath): User provides a symlink as path for volume
	//and this symlink must point to a GPFS path. To mount volumes, instead
	//of symlink we are using target of the symlink already. volScalePath may
	//or may not be a symlink.
	//2. symlink2 (targetPath): this is the one we create for version 1 volumes.
	//This symlink will always be there for 2.7.0 fixpack2.
	const useSymlink = true
	if useSymlink {
		//Check if mount dir/slink exists, if yes delete it
		_, err := os.Lstat(targetPath)
		if err != nil {
			//It is ok if the target path does not exist, it will be created as part
			//of NodePublishVolume.
			if !os.IsNotExist(err) {
				glog.V(4).Infof("NodePublishVolume - failed to get lstat of targetPath [%s]. Error [%v]", targetPath, err)
			}
		} else {
			glog.V(4).Infof("NodePublishVolume - deleting the targetPath - [%v]", targetPath)
			err := os.Remove(targetPath)
			if err != nil && !os.IsNotExist(err) {
				glog.V(4).Infof("NodePublishVolume - failed to delete the target path - [%s]. Error [%v]", targetPath, err)
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete the target path - [%s]. Error [%v]", targetPath, err))
			}
		}

		//Ceate a new symlink (symlink2) pointing to volScalePath
		glog.V(4).Infof("NodePublishVolume - creating symlink [%v] -> [%v]", targetPath, volScalePath)
		symlinkerr := os.Symlink(volScalePath, targetPath)
		if symlinkerr != nil {
			glog.V(4).Infof("NodePublishVolume - failed to create symlink [%s] -> [%s]. Error [%v]", targetPath, volScalePath, symlinkerr)
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create symlink [%s] -> [%s]. Error [%v]", targetPath, volScalePath, symlinkerr))
		}

		//check for the gpfs type again, if not gpfs type, delete the symlink and return error
		err = checkGpfsType(volScalePathInContainer)
		if err != nil {
			rerr := os.Remove(targetPath)
			if rerr != nil && !os.IsNotExist(rerr) {
				glog.V(4).Infof("NodePublishVolume - failed to delete the targetPath - [%s]. Error [%v]", targetPath, rerr)
				return nil, status.Error(codes.Internal, fmt.Sprintf("NodePublishVolume - failed to delete the targetPath - [%s]. Error [%v]", targetPath, rerr))
			}

			//gpfs type check has failed, return error
			glog.V(4).Infof("NodePublishVolume - the path [%v] is not a valid gpfs path", volScalePathInContainer)
			return nil, err
		}
	} else {
		notMP, err := mount.IsNotMountPoint(mount.New(""), targetPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err = os.Mkdir(targetPath, 0750); err != nil {
					glog.V(4).Infof("NodePublishVolume - failed to create target path [%s]. Error [%v]", targetPath, err)
					return nil, fmt.Errorf("NodePublishVolume -failed to create target path [%s]. Error [%v]", targetPath, err)
				}
			} else {
				glog.V(4).Infof("NodePublishVolume - failed to check target path [%s]. Error [%v]", targetPath, err)
				return nil, fmt.Errorf("NodePublishVolume -failed to check target path [%s]. Error [%v]", targetPath, err)
			}
		}
		if !notMP {
			glog.V(4).Infof("NodePublishVolume - returning success as the path [%s] is already a mount point", targetPath)
			return &csi.NodePublishVolumeResponse{}, nil
		}

		// create bind mount
		options := []string{"bind"}
		mounter := mount.New("")
		glog.V(4).Infof("NodePublishVolume - creating bind mount [%v] -> [%v]", targetPath, volScalePath)
		if err := mounter.Mount(volScalePath, targetPath, "", options); err != nil {
			glog.V(4).Infof("NodePublishVolume - failed to mount: [%s] at [%s]. Error [%v]", volScalePath, targetPath, err)
			return nil, fmt.Errorf("NodePublishVolume -failed to mount: [%s] at [%s]. Error [%v]", volScalePath, targetPath, err)
		}

		//check for the gpfs type again, if not gpfs type, unmount and return error.
		err = checkGpfsType(volScalePathInContainer)
		if err != nil {
			uerr := mount.New("").Unmount(targetPath)
			if uerr != nil {
				glog.V(4).Infof("NodePublishVolume - failed to unmount the path [%s]. Error %v", targetPath, uerr)
				return nil, fmt.Errorf("NodePublishVolume - failed to unmount the path [%s]. Error %v", targetPath, uerr)
			}
			return nil, err
		}
	}
	glog.V(4).Infof("NodePublishVolume - successfully mounted %s", targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// unmountAndDelete unmounts and deletes a targetPath (forcefully if
// foreceful=true is passed) and returns a bool which tells if a
// calling function should return, along with the response and error
// to be returned if there are any.
func unmountAndDelete(targetPath string, forceful bool) (bool, *csi.NodeUnpublishVolumeResponse, error) {
	glog.V(3).Infof("nodeserver unmountAndDelete")
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
		glog.V(4).Infof("%v is unmounted successfully", targetPath)
	}
	// Delete the mount point
	if err = os.Remove(targetPathInContainer); err != nil {
		if os.IsNotExist(err) {
			glog.V(4).Infof("target path %v is already deleted", targetPath)
			return false, nil, nil
		}
		return true, nil, fmt.Errorf("failed to remove the mount point [%s]. Error %v", targetPathInContainer, err)
	}
	return false, nil, nil
}

func (ns *ScaleNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	glog.V(3).Infof("nodeserver NodeUnpublishVolume")
	glog.V(4).Infof("NodeUnpublishVolume called with args: %v", req)
	start := time.Now()
	defer glog.V(4).Infof("NodeUnpublishVolume : req %#v time spent : %v", req, time.Since(start))
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
		//Handling for target path (softlink or bindmount) is already deleted/not present
		if os.IsNotExist(err) {
			glog.V(4).Infof("NodeUnpublishVolume - returning success as targetpath %v is not found ", targetPath)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		//Handling for bindmount is gpfs is unmounted/unlinked
		if strings.Contains(err.Error(), errStaleNFSFileHandle) {
			glog.V(4).Infof("NodeUnpublishVolume - error [%v] is observed, trying forceful unmount of [%s]", err, targetPath)
			needReturn, response, error := unmountAndDelete(targetPath, true)
			if needReturn {
				glog.V(4).Infof("NodeUnpublishVolume - returning response and error from unmountAndDelete. reponse [%v], error [%v]", response, error)
				return response, error
			}
			glog.V(4).Infof("NodeUnpublishVolume - Forceful unmount is successful")
			return &csi.NodeUnpublishVolumeResponse{}, nil
		} else {
			glog.V(4).Infof("NodeUnpublishVolume - failed to get lstat of target path [%s]. Error %v", targetPath, err)
			return nil, fmt.Errorf("NodeUnpublishVolume - failed to get lstat of target path [%s]. Error %v", targetPath, err)
		}
	}
	if f.Mode()&os.ModeSymlink != 0 {
		glog.V(4).Infof("%v is a symlink", targetPath)
		if err := os.Remove(targetPath); err != nil {
			if os.IsNotExist(err) {
				glog.V(4).Infof("symlink %v is already deleted", targetPath)
				return &csi.NodeUnpublishVolumeResponse{}, nil
			}
			glog.V(4).Infof("NodeUnpublishVolume - failed to remove symlink targetPath [%v]. Error [%v]", targetPath, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("NodeUnpublishVolume - failed to remove symlink targetPath [%v]. Error [%v]", targetPath, err))
		}
	} else {
		glog.V(4).Infof("%v is a bind mount", targetPath)
		needReturn, response, error := unmountAndDelete(targetPath, false)
		if needReturn {
			glog.V(4).Infof("NodeUnpublishVolume - returning response and error from unmountAndDelete. reponse [%v], error [%v]", response, error)
			return response, error
		}
	}
	glog.V(4).Infof("NodeUnpublishVolume - successfully unpublished %s", targetPath)
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
