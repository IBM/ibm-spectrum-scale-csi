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

	"k8s.io/klog/v2"

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
	//mux sync.Mutex
}

const hostDir = "/host"
const errStaleNFSFileHandle = "stale NFS file handle"

const nodePublishMethod = "NODEPUBLISH_METHOD"
const nodePublishMethodSymlink = "SYMLINK"
const nodePublishMethodBindMount = "BINDMOUNT"

// checkGpfsType checks if a given path is of type gpfs and
// returns nil if it is a gpfs type, otherwise returns
// corresponding error.
func checkGpfsType(path string) (bool error) {
	args := []string{"-f", "-c", "%T", path}
	out, err := executeCmd("stat", args)
	if err != nil {
		return fmt.Errorf("checkGpfsType: failed to get type of file with stat of [%s]. Error [%v]", path, err)
	}
	outString := string(out[:])
	outString = strings.TrimRight(outString, "\n")
	if outString != "gpfs" {
		return fmt.Errorf("checkGpfsType: the path [%s] is not a valid gpfs path, the path is of type [%s]", strings.TrimPrefix(path, hostDir), outString)
	}
	return nil
}

func (ns *ScaleNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] nodeserver NodePublishVolume", loggerId)

	klog.V(4).Infof("[%s] NodePublishVolume called with req: %#v", loggerId, req)

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

	klog.V(4).Infof("[%s] Target IBM Storage Scale Path : %v\n", loggerId, volScalePath)

	volScalePathInContainer := hostDir + volScalePath
	f, err := os.Lstat(volScalePathInContainer)
	if err != nil {
		klog.Errorf("[%s] NodePublishVolume - failed to get lstat of [%s]. Error [%v]", loggerId, volScalePathInContainer, err)
		return nil, fmt.Errorf("NodePublishVolume - failed to get lstat of [%s]. Error [%v]", volScalePathInContainer, err)
	}
	if f.Mode()&os.ModeSymlink != 0 {
		symlinkTarget, readlinkErr := os.Readlink(volScalePathInContainer)
		if readlinkErr != nil {
			klog.Errorf("[%s] NodePublishVolume - failed to get symlink target for [%s]. Error [%v]", loggerId, volScalePathInContainer, readlinkErr)
			return nil, fmt.Errorf("NodePublishVolume - failed to get symlink target for [%s]. Error [%v]", volScalePathInContainer, readlinkErr)
		}
		volScalePathInContainer = hostDir + symlinkTarget
		volScalePath = symlinkTarget
		klog.Infof("[%s] NodePublishVolume - symlink tarrget path is [%s]\n", loggerId, volScalePathInContainer)
	}

	err = checkGpfsType(volScalePathInContainer)
	if err != nil {
		klog.Errorf("[%s] NodePublishVolume - the path [%v] is not a valid gpfs path", loggerId, volScalePathInContainer)
		return nil, err
	}

	method := strings.ToUpper(os.Getenv(nodePublishMethod))
	klog.Infof("[%s] NodePublishVolume - NodePublishVolume method used: %s", loggerId, method)

	if method == nodePublishMethodSymlink {
		//There can be 2 symlinks here:
		//1. symlink1 (volScalePath): User provides a symlink as path for volume
		//and this symlink must point to a GPFS path. To mount volumes, instead
		//of symlink we are using target of the symlink already. volScalePath may
		//or may not be a symlink.
		//2. symlink2 (targetPath): this is the one we create for version 1 volumes.
		//This symlink will always be there when nodePublishMethod is SYMLINK otherwise
		//bind mount will be used.

		//Check if targetPath exists, if yes delete it
		_, err := os.Lstat(targetPath)
		if err != nil {
			//It is ok if the target path does not exist, it will be created as part
			//of NodePublishVolume
			if !os.IsNotExist(err) {
				klog.Errorf("[%s] NodePublishVolume - failed to get lstat of targetPath [%s]. Error [%v]", loggerId, targetPath, err)
			}
		} else {
			klog.Infof("[%s] NodePublishVolume - deleting the targetPath - [%v]", loggerId, targetPath)
			err := os.Remove(targetPath)
			if err != nil && !os.IsNotExist(err) {
				klog.Errorf("[%s] NodePublishVolume - failed to delete the target path - [%s]. Error [%v]", loggerId, targetPath, err)
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete the target path - [%s]. Error [%v]", targetPath, err))
			}
		}

		//Create a new symlink (symlink2) pointing to volScalePath
		klog.Infof("[%s] NodePublishVolume - creating symlink [%v] -> [%v]", loggerId, targetPath, volScalePath)
		symlinkerr := os.Symlink(volScalePath, targetPath)
		if symlinkerr != nil {
			klog.Errorf("[%s] NodePublishVolume - failed to create symlink [%s] -> [%s]. Error [%v]", loggerId, targetPath, volScalePath, symlinkerr)
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create symlink [%s] -> [%s]. Error [%v]", targetPath, volScalePath, symlinkerr))
		}

		//check for the gpfs type again, if not gpfs type, delete the symlink and return error
		err = checkGpfsType(volScalePathInContainer)
		if err != nil {
			rerr := os.Remove(targetPath)
			if rerr != nil && !os.IsNotExist(rerr) {
				klog.Errorf("[%s] NodePublishVolume - failed to delete the targetPath - [%s]. Error [%v]", loggerId, targetPath, rerr)
				return nil, status.Error(codes.Internal, fmt.Sprintf("NodePublishVolume - failed to delete the targetPath - [%s]. Error [%v]", targetPath, rerr))
			}

			//gpfs type check has failed, return error
			klog.Errorf("[%s] NodePublishVolume - the path [%v] is not a valid gpfs path", loggerId, volScalePathInContainer)
			return nil, err
		}
	} else {
		mounter := &mount.Mounter{}
		mntPoint, err := mounter.IsMountPoint(targetPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err = os.Mkdir(targetPath, 0750); err != nil {
					klog.Errorf("[%s] NodePublishVolume - failed to create target path [%s]. Error [%v]", loggerId, targetPath, err)
					return nil, fmt.Errorf("NodePublishVolume - failed to create target path [%s]. Error [%v]", targetPath, err)
				}
			} else {
				klog.Errorf("[%s] NodePublishVolume - failed to check target path [%s]. Error [%v]", loggerId, targetPath, err)
				return nil, fmt.Errorf("NodePublishVolume - failed to check target path [%s]. Error [%v]", targetPath, err)
			}
		}
		if mntPoint {
			klog.V(4).Infof("[%s] NodePublishVolume - returning success as the path [%s] is already a mount point", loggerId, targetPath)
			return &csi.NodePublishVolumeResponse{}, nil
		}

		// create bind mount
		options := []string{"bind"}
		klog.Infof("[%s] NodePublishVolume - creating bind mount [%v] -> [%v]", loggerId, targetPath, volScalePath)
		if err := mounter.Mount(volScalePath, targetPath, "", options); err != nil {
			klog.Errorf("[%s] NodePublishVolume - failed to mount: [%s] at [%s]. Error [%v]", loggerId, volScalePath, targetPath, err)
			return nil, fmt.Errorf("NodePublishVolume - failed to mount: [%s] at [%s]. Error [%v]", volScalePath, targetPath, err)
		}

		//check for the gpfs type again, if not gpfs type, unmount and return error.
		err = checkGpfsType(volScalePathInContainer)
		if err != nil {
			uerr := mounter.Unmount(targetPath)
			if uerr != nil {
				klog.Errorf("[%s] NodePublishVolume - failed to unmount the path [%s]. Error %v", loggerId, targetPath, uerr)
				return nil, fmt.Errorf("NodePublishVolume - failed to unmount the path [%s]. Error %v", targetPath, uerr)
			}
			return nil, err
		}
	}
	klog.Infof("[%s] NodePublishVolume - successfully mounted %s", loggerId, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// unmountAndDelete unmounts and deletes a targetPath (forcefully if
// foreceful=true is passed) and returns a bool which tells if a
// calling function should return, along with the response and error
// to be returned if there are any.
func unmountAndDelete(ctx context.Context, targetPath string, forceful bool) (bool, *csi.NodeUnpublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] nodeserver unmountAndDelete", loggerId)
	targetPathInContainer := hostDir + targetPath
	isMP := false
	var err error
	mounter := &mount.Mounter{}
	if !forceful {
		isMP, err = mounter.IsMountPoint(targetPathInContainer)
		if err != nil {
			if os.IsNotExist(err) {
				klog.V(4).Infof("[%s] target path %v is already deleted", loggerId, targetPathInContainer)
				return true, &csi.NodeUnpublishVolumeResponse{}, nil
			}
			klog.Errorf("[%s] failed to check if target path [%s] is a mount point. Error %v", loggerId, targetPathInContainer, err)
			return true, nil, fmt.Errorf("failed to check if target path [%s] is a mount point. Error %v", targetPathInContainer, err)
		}
	}
	if forceful || isMP {
		// Unmount the targetPath
		err = mounter.Unmount(targetPath)
		if err != nil {
			klog.Errorf("[%s] failed to unmount the mount point [%s]. Error %v", loggerId, targetPath, err)
			return true, nil, fmt.Errorf("failed to unmount the mount point [%s]. Error %v", targetPath, err)
		}
		klog.Infof("[%s] %v is unmounted successfully", loggerId, targetPath)
	}
	// Delete the mount point
	if err = os.Remove(targetPathInContainer); err != nil {
		if os.IsNotExist(err) {
			klog.V(4).Infof("[%s] target path %v is already deleted", loggerId, targetPath)
			return false, nil, nil
		}
		klog.V(4).Infof("[%s] failed to remove the mount point [%s]. Error %v", loggerId, targetPathInContainer, err)
		return true, nil, fmt.Errorf("failed to remove the mount point [%s]. Error %v", targetPathInContainer, err)
	}
	klog.Infof("[%s] %v is deleted successfully", loggerId, targetPath)
	return false, nil, nil
}

func (ns *ScaleNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] nodeserver NodeUnpublishVolume", loggerId)
	klog.V(4).Infof("[%s] NodeUnpublishVolume called with args: %v", loggerId, req)
	// Validate Arguments
	targetPath := req.GetTargetPath()
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "target path must be provided")
	}

	klog.Infof("[%s] NodeUnpublishVolume - deleting the targetPath - [%v]", loggerId, targetPath)

	//Check if target is a symlink or bind mount and cleanup accordingly
	f, err := os.Lstat(targetPath)
	if err != nil {
		//Handling for target path is already deleted/not present
		if os.IsNotExist(err) {
			klog.V(4).Infof("[%s] NodeUnpublishVolume - returning success as targetpath %v is not found ", loggerId, targetPath)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		//Handling for bindmount if filesystem is unmounted or fileset is unlinked
		if strings.Contains(err.Error(), errStaleNFSFileHandle) {
			klog.Errorf("[%s] NodeUnpublishVolume - Error [%v] is observed, trying forceful unmount of [%s]", loggerId, err, targetPath)
			needReturn, response, error := unmountAndDelete(ctx, targetPath, true)
			if needReturn {
				klog.Infof("[%s] NodeUnpublishVolume - returning response and error from unmountAndDelete. reponse [%v], error [%v]", loggerId, response, error)
				return response, error
			}
			klog.Infof("[%s] NodeUnpublishVolume - Forceful unmount is successful", loggerId)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		} else {
			klog.Errorf("[%s] NodeUnpublishVolume - failed to get lstat of target path [%s]. Error %v", loggerId, targetPath, err)
			return nil, fmt.Errorf("NodeUnpublishVolume - failed to get lstat of target path [%s]. Error %v", targetPath, err)
		}
	}
	if f.Mode()&os.ModeSymlink != 0 {
		klog.Infof("[%s] %v is a symlink", loggerId, targetPath)
		if err := os.Remove(targetPath); err != nil {
			if os.IsNotExist(err) {
				klog.V(4).Infof("[%s] NodeUnpublishVolume - symlink %v is already deleted", loggerId, targetPath)
				return &csi.NodeUnpublishVolumeResponse{}, nil
			}
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove symlink targetPath [%v]. Error [%v]", targetPath, err.Error()))
		}
	} else {
		klog.Infof("[%s] %v is a bind mount", loggerId, targetPath)
		needReturn, response, error := unmountAndDelete(ctx, targetPath, false)
		if needReturn {
			klog.Infof("[%s] NodeUnpublishVolume - returning response and error from unmountAndDelete. reponse [%v], error [%v]", loggerId, response, error)
			return response, error
		}
	}
	klog.Infof("[%s] NodeUnpublishVolume - successfully unpublished %s", loggerId, targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *ScaleNodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *ScaleNodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *ScaleNodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] NodeGetCapabilities called with req: %#v", loggerId, req)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *ScaleNodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] NodeGetInfo called with req: %#v", loggerId, req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *ScaleNodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *ScaleNodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] NodeGetVolumeStats called with req: %#v", loggerId, req)

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
		klog.Infof("[%s] Incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			loggerId, volumeIDMembers.FsetName, available, capacity)

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			volumeIDMembers.FsetName, available, capacity))
	}

	klog.V(4).Infof("[%s] Stat for volume:%v, Total:%v, Used:%v Available:%v, Total Inodes:%v, Used Inodes:%v, Available Inodes:%v,",
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
