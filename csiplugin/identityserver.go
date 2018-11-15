package scale

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
        "github.com/golang/glog"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
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
	glog.V(4).Infof("Probe called with args: %#v", req)
	return &csi.ProbeResponse{}, nil
}

// GetPluginInfo(context.Context, *GetPluginInfoRequest) (*GetPluginInfoResponse, error)
func (is *ScaleIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	glog.V(5).Infof("Using default GetPluginInfo")

	if is.Driver.name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	return &csi.GetPluginInfoResponse{
		Name:          is.Driver.name,
		VendorVersion: is.Driver.vendorVersion,
	}, nil
}
