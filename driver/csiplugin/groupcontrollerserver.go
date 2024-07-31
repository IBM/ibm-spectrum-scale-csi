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
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
)

const ()

type ScaleGroupControllerServer struct {
	Driver *ScaleDriver
}

// GroupControllerGetCapabilities implements the default GRPC callout.
func (gcs *ScaleGroupControllerServer) GroupControllerGetCapabilities(ctx context.Context, req *csi.GroupControllerGetCapabilitiesRequest) (*csi.GroupControllerGetCapabilitiesResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] GroupControllerGetCapabilities called with req: %#v", loggerId, req)
	return &csi.GroupControllerGetCapabilitiesResponse{
		Capabilities: gcs.Driver.gcscap,
	}, nil
}

// CreateVolumeGroupSnapshot Create VolumeGroup Snapshot
func (gcs *ScaleGroupControllerServer) CreateVolumeGroupSnapshot(ctx context.Context, req *csi.CreateVolumeGroupSnapshotRequest) (*csi.CreateVolumeGroupSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] CreateVolumeGroupSnapshot - create CreateVolumeGroupSnapshot req: %v", loggerId, req)

	return &csi.CreateVolumeGroupSnapshotResponse{
		GroupSnapshot: &csi.VolumeGroupSnapshot{},
	}, nil
}

// GetVolumeGroupSnapshot Get VolumeGroup Snapshot
func (gcs *ScaleGroupControllerServer) GetVolumeGroupSnapshot(ctx context.Context, req *csi.GetVolumeGroupSnapshotRequest) (*csi.GetVolumeGroupSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] GetVolumeGroupSnapshot -  GetVolumeGroupSnapshot req: %v", loggerId, req)

	return &csi.GetVolumeGroupSnapshotResponse{
		GroupSnapshot: &csi.VolumeGroupSnapshot{},
	}, nil
}

// DeleteVolumeGroupSnapshot Delete VolumeGroup Snapshot
func (gcs *ScaleGroupControllerServer) DeleteVolumeGroupSnapshot(ctx context.Context, req *csi.DeleteVolumeGroupSnapshotRequest) (*csi.DeleteVolumeGroupSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] DeleteVolumeGroupSnapshot -  DeleteVolumeGroupSnapshot req: %v", loggerId, req)

	return &csi.DeleteVolumeGroupSnapshotResponse{}, nil
}
