/**
 * Copyright 2020 IBM Corp.
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
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"

	"k8s.io/klog"
)

/*NewDriver creates new CSI plugin driver from ConfigMap
 */
func NewFakeDriver(
	driverName string,
	vendorVersion string,
	nodeID string,
	config *settings.ConfigMap,
	connFactory connectors.ConnectorFactory,
) *Driver {
	klog.Infof(`Driver: %v Version: %v`, driverName, vendorVersion)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(LogGRPC),
	}
	d := &Driver{
		IdentityService: newIdentityService(
			driverName,
			vendorVersion,
		),
		NodeService: newNodeService(
			nodeID,
			connFactory,
		),
		ControllerService: newControllerService(
			config,
			connFactory,
		),
		gRPC: grpc.NewServer(opts...),
	}

	d.AddVolumeCapabilityAccessModes(
		[]csi.VolumeCapability_AccessMode_Mode{
			csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	)

	d.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		},
	)

	d.AddNodeServiceCapabilities(
		[]csi.NodeServiceCapability_RPC_Type{},
	)

	csi.RegisterIdentityServer(d.gRPC, d)
	csi.RegisterControllerServer(d.gRPC, d)
	csi.RegisterNodeServer(d.gRPC, d)

	return d
}
