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

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const loggerId = "logger_id"

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
	newCtx := setLoggerId(ctx)
	glog.V(3).Infof("[%s] GRPC call: %s", GetLoggerId(newCtx), info.FullMethod)
	glog.V(5).Infof("[%s] GRPC request: %+v", GetLoggerId(newCtx), req)
	resp, err := handler(newCtx, req)
	if err != nil {
		glog.Errorf("[%s] GRPC error: %v", GetLoggerId(newCtx), err)
	} else {
		glog.V(5).Infof("[%s] GRPC response: %+v", GetLoggerId(newCtx), resp)
	}
	return resp, err
}

func setLoggerId(ctx context.Context) context.Context {
	id := uuid.New().String()
	glog.V(3).Infof("uuid: %s", id.String())
	return context.WithValue(ctx, loggerId, id)
}

func GetLoggerId(ctx context.Context) (string, bool) {
	logger, ok := ctx.Value(loggerId).(string)
	return logger, ok
}
