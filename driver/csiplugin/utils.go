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
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

func NewVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	newCtx := utils.SetLoggerId(ctx)
	loggerId := utils.GetLoggerId(newCtx)

	skipLog := skipLogging(info.FullMethod)
	if skipLog {
		klog.V(4).Infof("[%s] GRPC call: %s", loggerId, info.FullMethod)
	} else {
		klog.Infof("[%s] GRPC call: %s", loggerId, info.FullMethod)
	}

	klog.V(4).Infof("[%s] GRPC request: %+v", loggerId, req)

	startTime := utils.GetExecutionTime()
	resp, err := handler(newCtx, req)
	if err != nil {
		klog.Errorf("[%s] GRPC error: %v", loggerId, err)
	} else {
		klog.V(4).Infof("[%s] GRPC response: %+v", loggerId, resp)
	}
	endTime := utils.GetExecutionTime()
	diffTime := endTime - startTime

	if skipLog {
		klog.V(4).Infof("[%s] Time taken to execute %s request(in milliseconds): %d", loggerId, info.FullMethod, diffTime)
	} else {
		klog.Infof("[%s] Time taken to execute %s request(in milliseconds): %d", loggerId, info.FullMethod, diffTime)
	}
	return resp, err
}

func skipLogging(methodName string) bool {
	method := [...]string{"NodeGetCapabilities", "Identity/Probe", "Identity/GetPluginInfo", "Node/NodeGetInfo", "Node/NodeGetVolumeStats"}
	for _, m := range method {
		if strings.Contains(methodName, m) {
			return true
		}
	}
	return false
}
