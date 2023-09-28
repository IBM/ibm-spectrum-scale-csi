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
	"context"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type ScaleIdentityServer struct {
	Driver *ScaleDriver
}

func (is *ScaleIdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (is *ScaleIdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] Probe called with args: %#v", loggerId, req)

	// Node mapping check
	scalenodeID := getNodeMapping(is.Driver.nodeID)
	klog.V(6).Infof("[%s] Probe: scalenodeID:%s --known as-- k8snodeName: %s", loggerId, scalenodeID, is.Driver.nodeID)
	// IsNodeComponentHealthy accepts nodeName as admin node name, daemon node name, etc.
	ghealthy, err := is.Driver.connmap["primary"].IsNodeComponentHealthy(ctx, scalenodeID, "GPFS")
	if !ghealthy {
		// Even gpfs health is unhealthy, success is return because restarting csi driver is not going help fix the issue
		klog.Errorf("[%s] Probe: IBM Storage Scale on node %v is unhealthy. Error: %v", loggerId, scalenodeID, err)
		return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
	}

	klog.V(4).Infof("[%s] Probe: IBM Storage Scale on node %v is healthy", loggerId, scalenodeID)
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
}

func (is *ScaleIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] Using default GetPluginInfo", loggerId)

	if is.Driver.name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	return &csi.GetPluginInfoResponse{
		Name:          is.Driver.name,
		VendorVersion: is.Driver.vendorVersion,
	}, nil
}
