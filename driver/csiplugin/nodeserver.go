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
		return fmt.Errorf("checkGpfsType - stat [%s] failed with error [%v]", path, err)
	}
	outString := string(out[:])
	outString = strings.TrimRight(outString, "\n")
	if outString != "gpfs" {
		return fmt.Errorf("checkGpfsType - [%s] is not a valid gpfs path. reported type is [%s]", strings.TrimPrefix(path, hostDir), outString)
	}
	return nil
}

func (ns *ScaleNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] NodePublishVolume - request: %#v", loggerId, req)

	// Validate Arguments
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()
	volumeCapability := req.GetVolumeCapability()

	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "targetPath must be provided")
	}
	if volumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "volume capability must be provided")
	}

	volumeIDMembers, err := getVolIDMembers(volumeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume : volumeID is not in proper format")
	}
	volScalePath := volumeIDMembers.Path

	volScalePathInContainer := hostDir + volScalePath
	f, err := os.Lstat(volScalePathInContainer)
	if err != nil {
		klog.Errorf("[%s] NodePublishVolume - lstat [%s] failed with error [%v]", loggerId, volScalePathInContainer, err)
		return nil, fmt.Errorf("NodePublishVolume - lstat [%s] failed with error [%v]", volScalePathInContainer, err)
	}
	if f.Mode()&os.ModeSymlink != 0 {
		symlinkTarget, readlinkErr := os.Readlink(volScalePathInContainer)
		if readlinkErr != nil {
			klog.Errorf("[%s] NodePublishVolume - readlink [%s] failed with error [%v]", loggerId, volScalePathInContainer, readlinkErr)
			return nil, fmt.Errorf("NodePublishVolume - readlink [%s] failed with error [%v]", volScalePathInContainer, readlinkErr)
		}
		volScalePathInContainer = hostDir + symlinkTarget
		volScalePath = symlinkTarget
		klog.V(4).Infof("[%s] NodePublishVolume - symlink targetPath is [%s]", loggerId, volScalePathInContainer)
	}

	err = checkGpfsType(volScalePathInContainer)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(os.Getenv(nodePublishMethod))
	if !(method == nodePublishMethodSymlink) {
		method = nodePublishMethodBindMount
	}
	klog.V(4).Infof("[%s] NodePublishVolume - NodePublishVolume method used: %s", loggerId, method)

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
				klog.V(4).Infof("[%s] NodePublishVolume - lstat [%s] failed with error [%v]", loggerId, targetPath, err)
			}
		} else {
			klog.V(4).Infof("[%s] NodePublishVolume - deleting the targetPath [%v]", loggerId, targetPath)
			err := os.Remove(targetPath)
			if err != nil && !os.IsNotExist(err) {
				klog.Errorf("[%s] NodePublishVolume - delete [%s] failed with error [%v]", loggerId, targetPath, err)
				return nil, status.Error(codes.Internal, fmt.Sprintf("delete [%s] failed with error [%v]", targetPath, err))
			}
		}

		//Create a new symlink (symlink2) pointing to volScalePath
		klog.V(4).Infof("[%s] NodePublishVolume - creating symlink [%v] -> [%v]", loggerId, targetPath, volScalePath)
		symlinkerr := os.Symlink(volScalePath, targetPath)
		if symlinkerr != nil {
			klog.Errorf("[%s] NodePublishVolume - symlink [%s] -> [%s] creation failed with error [%v]", loggerId, targetPath, volScalePath, symlinkerr)
			return nil, status.Error(codes.Internal, fmt.Sprintf("symlink [%s] -> [%s] creation failed with error [%v]", targetPath, volScalePath, symlinkerr))
		}

		//check for the gpfs type again, if not gpfs type, delete the symlink and return error
		err = checkGpfsType(volScalePathInContainer)
		if err != nil {
			rerr := os.Remove(targetPath)
			if rerr != nil && !os.IsNotExist(rerr) {
				klog.Errorf("[%s] NodePublishVolume - targetPath [%s] deletion failed with error [%v]", loggerId, targetPath, rerr)
				return nil, status.Error(codes.Internal, fmt.Sprintf("NodePublishVolume - targetPath [%s] deletion failed with error [%v]", targetPath, rerr))
			}

			//gpfs type check has failed, return error
			return nil, err
		}
	} else {
		mounter := &mount.Mounter{}
		mntPoint, err := mounter.IsMountPoint(targetPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err = os.Mkdir(targetPath, 0750); err != nil {
					klog.Errorf("[%s] NodePublishVolume - targetPath [%s] creation failed with error [%v]", loggerId, targetPath, err)
					return nil, fmt.Errorf("NodePublishVolume - targetPath [%s] creation failed with error [%v]", targetPath, err)
				}
			} else {
				klog.Errorf("[%s] NodePublishVolume - targetPath [%s] check failed with error [%v]", loggerId, targetPath, err)
				return nil, fmt.Errorf("NodePublishVolume - targetPath [%s] check failed with error [%v]", targetPath, err)
			}
		}
		if mntPoint {
			klog.V(4).Infof("[%s] NodePublishVolume - [%s] is already a mount point", loggerId, targetPath)
			return &csi.NodePublishVolumeResponse{}, nil
		}

		// create bind mount
		options := []string{"bind"}
		klog.V(4).Infof("[%s] NodePublishVolume - creating bind mount [%v] -> [%v]", loggerId, targetPath, volScalePath)
		if err := mounter.Mount(volScalePath, targetPath, "", options); err != nil {
			klog.Errorf("[%s] NodePublishVolume - mounting [%s] at [%s] failed with error [%v]", loggerId, volScalePath, targetPath, err)
			return nil, fmt.Errorf("NodePublishVolume - mounting [%s] at [%s] failed with error [%v]", volScalePath, targetPath, err)
		}

		//check for the gpfs type again, if not gpfs type, unmount and return error.
		err = checkGpfsType(volScalePathInContainer)
		if err != nil {
			uerr := mounter.Unmount(targetPath)
			if uerr != nil {
				klog.Errorf("[%s] NodePublishVolume - unmount [%s] failed with error [%v]", loggerId, targetPath, uerr)
				return nil, fmt.Errorf("NodePublishVolume - unmount [%s] failed with error [%v]", targetPath, uerr)
			}
			return nil, err
		}
	}
	klog.Infof("[%s] NodePublishVolume - successfully mounted [%s] using %s", loggerId, targetPath, method)
	return &csi.NodePublishVolumeResponse{}, nil
}

// unmountAndDelete unmounts and deletes a targetPath (forcefully if
// foreceful=true is passed) and returns a bool which tells if a
// calling function should return, along with the response and error
// to be returned if there are any.
func unmountAndDelete(ctx context.Context, targetPath string, forceful bool) (bool, *csi.NodeUnpublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] unmount and delete %s", loggerId, targetPath)
	targetPathInContainer := hostDir + targetPath
	isMP := false
	var err error
	mounter := &mount.Mounter{}
	if !forceful {
		isMP, err = mounter.IsMountPoint(targetPathInContainer)
		if err != nil {
			if os.IsNotExist(err) {
				klog.V(4).Infof("[%s] targetPath %v is not present", loggerId, targetPathInContainer)
				return true, &csi.NodeUnpublishVolumeResponse{}, nil
			}
			klog.Errorf("[%s] mount point check on [%s] failed with error [%v]", loggerId, targetPathInContainer, err)
			return true, nil, fmt.Errorf("mount point check on [%s] failed with error [%v]", targetPathInContainer, err)
		}
	}
	if forceful || isMP {
		// Unmount the targetPath
		err = mounter.Unmount(targetPath)
		if err != nil {
			klog.Errorf("[%s] unmount [%s] failed with error [%v]", loggerId, targetPath, err)
			return true, nil, fmt.Errorf("unmount [%s] failed with error [%v]", targetPath, err)
		}
		klog.V(4).Infof("[%s] %v is unmounted successfully", loggerId, targetPath)
	}
	// Delete the mount point
	if err = os.Remove(targetPathInContainer); err != nil {
		if os.IsNotExist(err) {
			klog.V(4).Infof("[%s] targetPath [%s] is not present", loggerId, targetPath)
			return false, nil, nil
		}
		klog.V(4).Infof("[%s] mount point [%s] removal failed with error [%v]", loggerId, targetPathInContainer, err)
		return true, nil, fmt.Errorf("mount point [%s] removal failed with error [%v]", targetPathInContainer, err)
	}
	klog.V(4).Infof("[%s] Path [%s] is deleted", loggerId, targetPath)
	return false, nil, nil
}

func (ns *ScaleNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] NodeUnpublishVolume - request: %#v", loggerId, req)
	// Validate Arguments
	targetPath := req.GetTargetPath()
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "targetPath must be provided")
	}

	//Check if target is a symlink or bind mount and cleanup accordingly
	f, err := os.Lstat(targetPath)
	if err != nil {
		//Handling for target path is already deleted/not present
		if os.IsNotExist(err) {
			klog.Infof("[%s] NodeUnpublishVolume - targetPath [%s] is not found, returning success ", loggerId, targetPath)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		//Handling for bindmount if filesystem is unmounted or fileset is unlinked
		if strings.Contains(err.Error(), errStaleNFSFileHandle) {
			klog.Warning("[%s] NodeUnpublishVolume - unmount [%s] failed with error [%v]. trying forceful unmount", loggerId, targetPath, err)
			needReturn, response, error := unmountAndDelete(ctx, targetPath, true)
			if needReturn {
				return response, error
			}
			klog.Infof("[%s] NodeUnpublishVolume - forced unmount [%s] is successful", loggerId, targetPath)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		} else {
			klog.Errorf("[%s] NodeUnpublishVolume - lstat [%s] failed with error [%v]", loggerId, targetPath, err)
			return nil, fmt.Errorf("NodeUnpublishVolume - lstat [%s] failed with error [%v]", targetPath, err)
		}
	}
	if f.Mode()&os.ModeSymlink != 0 {
		klog.V(6).Infof("[%s] %v is a symlink", loggerId, targetPath)
		if err := os.Remove(targetPath); err != nil {
			if os.IsNotExist(err) {
				klog.Infof("[%s] NodeUnpublishVolume - symlink [%s] is not present]", loggerId, targetPath)
				return &csi.NodeUnpublishVolumeResponse{}, nil
			}
			return nil, status.Error(codes.Internal, fmt.Sprintf("removal of symlink [%s] failed with error [%v]", targetPath, err.Error()))
		}
	} else {
		klog.V(6).Infof("[%s] %v is a bind mount", loggerId, targetPath)
		needReturn, response, error := unmountAndDelete(ctx, targetPath, false)
		if needReturn {
			return response, error
		}
	}
	klog.Infof("[%s] NodeUnpublishVolume - successfully unpublished [%s]", loggerId, targetPath)
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
	klog.V(4).Infof("[%s] NodeGetCapabilities - request: %#v", loggerId, req)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *ScaleNodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] NodeGetInfo - request: %#v", loggerId, req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *ScaleNodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *ScaleNodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] NodeGetVolumeStats - request: %#v", loggerId, req)

	if len(req.VolumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats - volumeID must be provided")
	}
	if len(req.VolumePath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats - targetPath must be provided")
	}

	if _, err := os.Lstat(req.VolumePath); err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "path %s does not exist", req.VolumePath)
		}
		return nil, status.Errorf(codes.Internal, "stat [%s] failed with error [%v]", req.VolumePath, err)
	}

	volumeIDMembers, err := getVolIDMembers(req.GetVolumeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats - volumeID is not in proper format")
	}

	if !volumeIDMembers.IsFilesetBased {
		return nil, status.Error(codes.InvalidArgument, "volume stats are not supported for lightweight volumes")
	}

	available, capacity, used, inodes, inodesFree, inodesUsed, err := utils.FsStatInfo(req.GetVolumePath())
	if err != nil {
		klog.Errorf("[%s] NodeGetVolumeStats - FsStatInfo [%s] failed with error [%v]", loggerId, req.GetVolumePath(), err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("FsStatInfo [%s] failed with error [%v]", req.GetVolumePath(), err))
	}

	if available > capacity || used > capacity {
		klog.V(4).Infof("[%s] Incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			loggerId, volumeIDMembers.FsetName, available, capacity)

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incorrect values reported for volume (%v) against Available(%v) or Capacity(%v)",
			volumeIDMembers.FsetName, available, capacity))
	}

	klog.Infof("[%s] NodeGetVolumeStats - stat for volume:%v, Total:%v, Used:%v Available:%v, Total Inodes:%v, Used Inodes:%v, Available Inodes:%v,",
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
