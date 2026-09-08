/**
 * Copyright 2019, 2024 IBM Corp.
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
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"k8s.io/klog/v2"

	"context"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"k8s.io/mount-utils"
)

type ScaleNodeServer struct {
	Driver *ScaleDriver
	// TODO: Only lock mutually exclusive calls and make locking more fine grained
	//mux sync.Mutex
	csi.UnimplementedNodeServer
}

const hostDir = "/host"
const mountPath = "/mnt"
const errStaleNFSFileHandle = "stale NFS file handle"

const nodePublishMethod = "NODEPUBLISH_METHOD"
const nodePublishMethodSymlink = "SYMLINK"

const mountPathLength = 6

const ENVClusterCNSAPresenceCheck = "CNSADeployment"
const ENVClusterConfigurationType = "ClusterConfigurationType"
const ENVClusterTypeOpenshift = "OpenShiftPlatform"
const gpfsmagicNumber uint64 = 0x47504653

// A map for locking/unlocking a target path for NodePublish/NodeUnpublish
// calls. The key is target path and value is a boolean true in case there
// is any NodePublishVolume or NodeUnpublishVolume request in progress for
// the target path.
var nodePublishUnpublishLock map[string]bool

// a mutex var used to make sure certain code blocks are executed by
// only one goroutine at a time.
var mutex sync.Mutex

func lock(targetPath string, ctx context.Context) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if len(nodePublishUnpublishLock) == 0 {
		nodePublishUnpublishLock = make(map[string]bool)
	}

	if _, exists := nodePublishUnpublishLock[targetPath]; exists {
		return false
	}
	nodePublishUnpublishLock[targetPath] = true
	klog.V(4).Infof("[%s] The target path is locked for NodePublish/NodeUnpublish: [%s]", utils.GetLoggerId(ctx), targetPath)
	return true
}

func unlock(targetPath string, ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(nodePublishUnpublishLock, targetPath)
	klog.V(4).Infof("[%s] The target path is unlocked for NodePublish/NodeUnpublish: [%s]", utils.GetLoggerId(ctx), targetPath)
}

// checkGpfsType checks if a given path is of type gpfs and
// returns nil if it is a gpfs type, otherwise returns
// corresponding error.
//
// Path namespace note:
//   - `path` is always passed as /host/<kernel-path> (container view).
//   - filepath.EvalSymlinks may or may not resolve through the /host bind-mount
//     depending on the kernel/container runtime; it may return either
//     /host/<kernel-path> or <kernel-path> directly.
//   - getGpfsPaths returns kernel-visible paths with no /host prefix.
//
// To compare both sides in the same namespace, we explicitly strip the /host
// prefix from resolvedPath before the mount containment check.
func checkGpfsType(ctx context.Context, path string) error {
	loggerId := utils.GetLoggerId(ctx)

	// Resolve the path to its canonical form (resolves symlinks and normalizes).
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If symlink resolution fails, fall back to a clean path.
		resolvedPath = filepath.Clean(path)
		klog.V(4).Infof("[%s] checkGpfsType: could not resolve symlinks for [%s], using normalized path [%s]: %v",
			loggerId, path, resolvedPath, err)
	}

	// Verify the resolved path is absolute.
	if !filepath.IsAbs(resolvedPath) {
		return fmt.Errorf("checkGpfsType: resolved path [%s] is not absolute", resolvedPath)
	}

	// Normalise to kernel-visible namespace: strip /host prefix regardless of
	// whether EvalSymlinks resolved through the bind-mount or not.
	kernelPath := strings.TrimPrefix(resolvedPath, hostDir)

	gpfsPaths := getGpfsPaths(ctx)
	if len(gpfsPaths) == 0 {
		return fmt.Errorf("checkGpfsType: no GPFS mount points found on the system")
	}
	klog.V(4).Infof("[%s] checkGpfsType: resolvedPath: %s kernelPath: %s", loggerId, resolvedPath, kernelPath)

	// gpfsPaths entries are kernel-visible mount roots (e.g. /var/mnt/remote-sample).
	// kernelPath is always a deep volume path under one of those roots.
	for _, gpfsPath := range gpfsPaths {
		klog.V(4).Infof("[%s] checkGpfsType: gpfsPath %s", loggerId, gpfsPath)
		if kernelPath != gpfsPath && !strings.HasPrefix(kernelPath, gpfsPath+"/") {
			klog.V(4).Infof("[%s] checkGpfsType: kernelPath [%s] is not under mount [%s], skipping", loggerId, kernelPath, gpfsPath)
			continue
		}
		ok, err := isGPFS(ctx, resolvedPath)
		if err != nil {
			return fmt.Errorf("checkGpfsType: failed to validate volume mount path [%s]: %w", resolvedPath, err)
		}
		if ok {
			return nil
		}
	}

	klog.Errorf("[%s] checkGpfsType: path [%s] resolved to [%s] is not within any valid GPFS mount point. Available GPFS mounts: %v",
		loggerId, strings.TrimPrefix(path, hostDir), kernelPath, gpfsPaths)
	return fmt.Errorf("checkGpfsType: the path [%s] (resolves to [%s]) is not within any valid GPFS mount point",
		strings.TrimPrefix(path, hostDir), kernelPath)
}

// isGPFS reports whether the filesystem containing path is GPFS.
func isGPFS(ctx context.Context, path string) (bool, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] isGPFS: path %s", loggerId, path)

	st, err := utils.StatfsWithTimeout(ctx, path)
	if err != nil {
		klog.Errorf("[%s] isGPFS: statfs %q failed: %v", loggerId, path, err)
		return false, err
	}
	klog.V(4).Infof("[%s] isGPFS: fsType 0x%x for path %s", loggerId, uint64(st.Type), path)
	// GPFS magic number: 0x47504653 ("GPFS")
	return uint64(st.Type) == gpfsmagicNumber, nil
}

// getGpfsPaths returns GPFS mount points as kernel-visible paths (no /host prefix).
//
// /proc/mounts inside the node-plugin container shows host paths prefixed with
// /host because the host rootfs is bind-mounted there. EvalSymlinks resolves
// through that bind mount and returns the real kernel path (e.g. /mnt/... or
// /var/mnt/... on CNSA/OpenShift). To make the comparison in checkGpfsType
// work across all environments, we store the path exactly as the kernel sees
// it: strip the leading /host prefix and keep the rest unchanged.
func getGpfsPaths(ctx context.Context) []string {
	loggerId := utils.GetLoggerId(ctx)
	var gpfsPaths []string
	gpfsPathCmd := `cat /proc/mounts | grep "gpfs"`
	cmd := exec.Command("bash", "-c", gpfsPathCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("[%s] Error in executing command: [%s]", utils.GetLoggerId(ctx), err)
		return gpfsPaths
	}

	strOutput := strings.Split(string(output), "\n")
	for _, out := range strOutput {
		fields := strings.Split(out, " ")
		if len(fields) != mountPathLength {
			continue
		}
		mountPoint := fields[1]
		fsType := fields[2]
		if mountPoint == "" || fsType != "gpfs" {
			continue
		}
		// Strip the container-side /host prefix to get the kernel-visible path.
		// This is the path EvalSymlinks will resolve to, regardless of whether
		// the cluster is CNSA/OpenShift or plain Kubernetes.
		kernelPath := strings.TrimPrefix(mountPoint, hostDir)
		gpfsPaths = append(gpfsPaths, kernelPath)
		klog.V(4).Infof("[%s] getGpfsPaths: added kernel path [%s] (from mount entry [%s])", loggerId, kernelPath, mountPoint)
	}
	klog.V(4).Infof("[%s] getGpfsPaths: %v", loggerId, gpfsPaths)
	return gpfsPaths
}

func (ns *ScaleNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	reqToLog := proto.Clone(req).(*csi.NodePublishVolumeRequest)
	reqToLog.Secrets = nil
	klog.Infof("[%s] NodePublishVolume - request: %#v", loggerId, reqToLog)

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

	lockSuccess := lock(targetPath, ctx)
	if !lockSuccess {
		message := fmt.Sprintf("NodePublishVolume - another NodePublish/NodeUnpublish is in progress for the targetPath: [%s]", targetPath)
		klog.Errorf("[%s] "+message, loggerId)
		return nil, status.Error(codes.Internal, message)
	} else {
		defer unlock(targetPath, ctx)
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
		// Validate the symlink target to prevent path traversal attacks.
		// A malicious or misconfigured symlink pointing to e.g. "../../etc"
		// would otherwise escape the GPFS mount and expose arbitrary host paths.
		if err := validatePath(symlinkTarget); err != nil {
			klog.Errorf("[%s] NodePublishVolume - symlink target [%s] failed path validation: [%v]", loggerId, symlinkTarget, err)
			return nil, status.Errorf(codes.InvalidArgument, "NodePublishVolume - symlink target [%s] is not a valid path: %v", symlinkTarget, err)
		}
		volScalePathInContainer = hostDir + symlinkTarget
		volScalePath = symlinkTarget
		klog.V(4).Infof("[%s] NodePublishVolume - symlink targetPath is [%s]", loggerId, volScalePathInContainer)
	}

	err = checkGpfsType(ctx, volScalePathInContainer)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(os.Getenv(nodePublishMethod))
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
		err = checkGpfsType(ctx, volScalePathInContainer)
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
				} else {
					klog.Infof("[%s] NodePublishVolume - the target directory [%s] is created successfully", loggerId, targetPath)
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
		err = checkGpfsType(ctx, volScalePathInContainer)
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
	klog.V(4).Infof("[%s] unmount and delete targetPath:[%s], forceful:[%v]", loggerId, targetPath, forceful)
	isMP := false
	var err error
	mounter := &mount.Mounter{}
	if !forceful {
		isMP, err = mounter.IsMountPoint(targetPath)
		if err != nil {
			klog.Errorf("[%s] checking the targetPath: [%s] for mountPoint,failed with error [%v]", loggerId, targetPath, err)
			if os.IsNotExist(err) {
				klog.V(4).Infof("[%s] NodeUnpublishVolume - targetPath [%s] is not found when !forceful, returning success ", loggerId, targetPath)
				return true, &csi.NodeUnpublishVolumeResponse{}, nil
			}
			klog.Errorf("[%s] mount point check on targetPath:[%s] failed with error [%v]", loggerId, targetPath, err)
			return true, nil, status.Error(codes.Internal, fmt.Sprintf("NodeUnpublishVolume - mount point check on targetPath:[%s] failed with error [%v]", targetPath, err))
		}
		klog.V(4).Infof("[%s] isMP value for the target path [%s] is [%t]", loggerId, targetPath, isMP)
	}
	if forceful || isMP {
		// Unmount the targetPath
		err = mounter.Unmount(targetPath)
		if err != nil {
			klog.Errorf("[%s] unmount [%s] failed with error [%v]", loggerId, targetPath, err)
			return true, nil, status.Error(codes.Internal, fmt.Sprintf("NodeUnpublishVolume - unmount [%s] failed with error [%v]", targetPath, err))
		}
		klog.V(4).Infof("[%s] %v is unmounted successfully", loggerId, targetPath)
	}
	// Delete the mount point
	klog.V(4).Infof("[%s] remove mount point ", loggerId)
	if err = os.Remove(targetPath); err != nil {
		klog.Errorf("[%s] remove targetPath:[%s] failed with error [%v]", loggerId, targetPath, err)
		if os.IsNotExist(err) {
			klog.V(4).Infof("[%s] targetPath [%s] is not present", loggerId, targetPath)
			return false, nil, nil
		}
		klog.V(4).Infof("[%s] mount point:[%s] removal failed with error [%v]", loggerId, targetPath, err)
		return true, nil, status.Error(codes.Internal, fmt.Sprintf("NodeUnpublishVolume - mount point [%s] removal failed with error [%v]", targetPath, err))
	}
	klog.V(4).Infof("[%s] Path [%s] is deleted", loggerId, targetPath)
	return false, nil, nil
}

func (ns *ScaleNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] NodeUnpublishVolume - request: %#v", loggerId, req)
	defer klog.Infof("[%s] NodeUnpublishVolume has completed", loggerId)
	// Validate Arguments
	targetPath := req.GetTargetPath()
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "targetPath must be provided")
	}

	lockSuccess := lock(targetPath, ctx)
	if !lockSuccess {
		message := fmt.Sprintf("NodeUnpublishVolume - another NodePublish/NodeUnpublish is in progress for the targetPath: [%s]", targetPath)
		klog.Errorf("[%s] "+message, loggerId)
		return nil, status.Error(codes.Internal, message)
	} else {
		defer unlock(targetPath, ctx)
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
			klog.Warningf("[%s] NodeUnpublishVolume - unmount [%s] failed with error [%v]. trying forceful unmount", loggerId, targetPath, err)
			needReturn, response, error := unmountAndDelete(ctx, targetPath, true)
			if needReturn {
				return response, error
			}
			klog.Infof("[%s] NodeUnpublishVolume - forced unmount [%s] is successful", loggerId, targetPath)
			return &csi.NodeUnpublishVolumeResponse{}, nil
		} else {
			klog.Errorf("[%s] NodeUnpublishVolume - lstat [%s] failed with error [%v]", loggerId, targetPath, err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("NodeUnpublishVolume - lstat [%s] failed with error [%v]", targetPath, err))
		}
	}
	if f.Mode()&os.ModeSymlink != 0 {
		klog.V(6).Infof("[%s] %v is a symlink", loggerId, targetPath)
		if err := os.Remove(targetPath); err != nil {
			if os.IsNotExist(err) {
				klog.Infof("[%s] NodeUnpublishVolume - symlink [%s] is not present, returning success ", loggerId, targetPath)
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

	volumePath := req.GetVolumePath()

	fileInfo, err := os.Lstat(volumePath)
	if err != nil {
		klog.Errorf("[%s] NodeGetVolumeStats - Lstat failed with error [%v]", loggerId, err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("FsStatInfo [%s] failed with error [%v]", volumePath, err))
	}
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		// This is a symlink
		dst, err := os.Readlink(volumePath)
		if err != nil {
			klog.Errorf("[%s] NodeGetVolumeStats - destination [%s] failed with error [%v]", loggerId, volumePath, err)
		} else if len(dst) > 0 {
			// Validate the symlink target to prevent path traversal.
			// A symlink pointing to e.g. "../../etc" would otherwise let
			// FsStatInfo read arbitrary host filesystem paths.
			if err := validatePath(dst); err != nil {
				klog.Errorf("[%s] NodeGetVolumeStats - symlink target [%s] failed path validation: [%v]", loggerId, dst, err)
				return nil, status.Errorf(codes.InvalidArgument, "NodeGetVolumeStats - symlink target [%s] is not a valid path: %v", dst, err)
			}
			volumePath = hostDir + dst
			klog.V(4).Infof("[%s] %s links to (%s) is a SYMLINK", loggerId, req.GetVolumePath(), volumePath)
		}
	}

	available, capacity, used, inodes, inodesFree, inodesUsed, err := utils.FsStatInfo(volumePath)
	if err != nil {
		klog.Errorf("[%s] NodeGetVolumeStats - FsStatInfo [%s] failed with error [%v]", loggerId, volumePath, err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("FsStatInfo [%s] failed with error [%v]", volumePath, err))
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
