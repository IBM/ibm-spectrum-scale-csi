/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gpfs

import (
	"fmt"
	"sync"
	//"os"
	//"strings"

	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"k8s.io/kubernetes/pkg/util/mount"
)

type GPFSNodeServer struct {
	Driver          *GPFSDriver
	// TODO: Only lock mutually exclusive calls and make locking more fine grained
	mux sync.Mutex
}

func (ns *GPFSNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	glog.V(4).Infof("NodePublishVolume called with req: %#v", req)

	// Validate Arguments
	targetPath := req.GetTargetPath()
	stagingTargetPath := req.GetStagingTargetPath()
	readOnly := req.GetReadonly()
	volumeID := req.GetVolumeId()
	volumeCapability := req.GetVolumeCapability()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume ID must be provided")
	}
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Staging Target Path must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Target Path must be provided")
	}
	if volumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume Capability must be provided")
	}

	options := []string{}
	if readOnly {
		options = append(options, "ro")
	}
	options = append(options, "bind")

	fsType := ""
	glog.V(4).Infof("Bind mount %s at %s, fsType %s, options %v ...", stagingTargetPath, targetPath, fsType, options)
	mounter := mount.New("")
	if err := mounter.Mount(stagingTargetPath, targetPath, fsType, options); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.V(4).Infof("Mount bind %s at %s succeed", stagingTargetPath, targetPath)

	/*fsType := "gpfs"
	diskMounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}
	if err := diskMounter.FormatAndMount(stagingTargetPath, targetPath, fsType, options); err != nil {
		return nil, err
	}*/

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *GPFSNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	glog.V(4).Infof("NodeUnpublishVolume called with args: %v", req)
	// Validate Arguments
	targetPath := req.GetTargetPath()
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Volume ID must be provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Target Path must be provided")
	}

	// TODO: Check volume still exists

	mounter := mount.New("")
	err :=  mounter.Unmount(targetPath) //ns.Mounter.Interface.Unmount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unmount failed: %v\nUnmounting arguments: %s\n", err, targetPath))
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}
func (ns *GPFSNodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
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

	err := ops.MountFS(volumeID, stagingTargetPath, ns.Driver.nodeID)
	if err != nil {
		return nil, status.Error(codes.Internal,
			fmt.Sprintf("Failed to mount device")) // from (%q) to (%q) with fstype (%q) and options (%q): %v",
				//devicePath, stagingTargetPath, fstype, options, err))
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *GPFSNodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {
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

	err := ops.UnmountFS(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("NodeUnstageVolume failed to unmount at path %s: %v", stagingTargetPath, err))
	}
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *GPFSNodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	glog.V(4).Infof("NodeGetCapabilities called with req: %#v", req)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *GPFSNodeServer) NodeGetId(ctx context.Context, req *csi.NodeGetIdRequest) (*csi.NodeGetIdResponse, error) {
	glog.V(4).Infof("NodeGetId called with req: %#v", req)
	return &csi.NodeGetIdResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *GPFSNodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	glog.V(4).Infof("NodeGetInfo called with req: %#v", req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}
