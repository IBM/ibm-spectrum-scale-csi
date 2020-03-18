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
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

/*ParseEndpoint raw string to scheme and address, to be consumed by net.Listen
Note: Attempts to remove existing unix domain sockets, if they exist.
*/
func ParseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("could not parse endpoint: %v", err)
	}

	var scheme = strings.ToLower(u.Scheme)
	var addr string

	switch scheme {
	case "tcp":
		addr = u.Host
	case "unix":
		addr = u.Path
		if err := os.Remove(addr); err != nil && os.IsExist(err) {
			return "", "", fmt.Errorf("could not remove unix domain socket %q: %v", addr, err)
		}
	default:
		return "", "", fmt.Errorf("unsupported protocol: %s", scheme)
	}

	return scheme, addr, nil
}

/*LogGRPC to klog
 */
func LogGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	klog.V(3).Infof("gRPC call: %s", info.FullMethod)
	klog.V(5).Infof("gRPC request: %+v", req)
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("gRPC error: %v", err)
	} else {
		klog.V(5).Infof("gRPC response: %+v", resp)
	}
	return resp, err
}
