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

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	logger.Debugf("Probe called with args: %#v", req)

	// Determine plugin health
	// If unhealthy return gRPC error code
	// more on error codes https://github.com/container-storage-interface/spec/blob/master/spec.md#probe-errors

	// Node mapping check
	scalenodeID := getNodeMapping(is.Driver.nodeID)
	logger.Debugf("Probe: scalenodeID:%s --known as-- k8snodeName: %s", scalenodeID, is.Driver.nodeID)
	// IsNodeComponentHealthy accepts nodeName as admin node name, daemon node name, etc.
	ghealthy, err := is.Driver.connmap["primary"].IsNodeComponentHealthy(scalenodeID, "GPFS")
	if ghealthy == false {
		logger.Errorf("Probe: GPFS component on node %v is not healthy. Error: %v", scalenodeID, err)
		return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
	}

	// nhealthy, err := is.Driver.connmap["primary"].IsNodeComponentHealthy(scalenodeID, "NODE")
	// if nhealthy == false {
	// 	logger.Errorf("Probe: NODE component on node %v is not healthy. Error: %v", scalenodeID, err)
	// 	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, err
	// }

	logger.Infof("Probe: GPFS on node %v is healthy", scalenodeID)

	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
}

func (is *ScaleIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	logger.Infof("Using default GetPluginInfo")

	if is.Driver.name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	return &csi.GetPluginInfoResponse{
		Name:          is.Driver.name,
		VendorVersion: is.Driver.vendorVersion,
	}, nil
}
