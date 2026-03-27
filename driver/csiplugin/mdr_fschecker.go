/**
 * Copyright 2024 IBM Corp.
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
	"fmt"
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	// CNSA Filesystem CR constants
	filesystemCRGroup     = "scale.spectrum.ibm.com"
	filesystemCRVersion   = "v1beta1"
	filesystemCRResource  = "filesystems"
	filesystemCRNamespace = "ibm-spectrum-scale"
	externalReplication   = "external"
)

// FilesystemSpec represents the spec section of Filesystem CR
type FilesystemSpec struct {
	Local *LocalSpec `json:"local,omitempty"`
}

// LocalSpec represents the local section in Filesystem spec
type LocalSpec struct {
	Replication string `json:"replication,omitempty"`
}

// Filesystem represents the CNSA Filesystem Custom Resource
type Filesystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FilesystemSpec `json:"spec,omitempty"`
}

// isMDREnabledOnFS checks if MetroDR replication is enabled for the given filesystem. This is determined
// by querying the Filesystem CR in ibm-spectrum-scale namespace and checking
// if fs.spec.local.replication == "external"
func (cs *ScaleControllerServer) isMDREnabledOnFS(ctx context.Context, filesystemName string) bool {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] isMDREnabledOnFS for filesystemName [%s], check if MetroDR replication is enabled", loggerId, filesystemName)
	// check the Filesystem CR for the specified filesystem

	// Define the GVR for Filesystem CR
	filesystemGVR := schema.GroupVersionResource{
		Group:    filesystemCRGroup,
		Version:  filesystemCRVersion,
		Resource: filesystemCRResource,
	}

	// Try to get the Filesystem CR
	filesystemCR, err := cs.Driver.dynamicClient.Resource(filesystemGVR).Namespace(filesystemCRNamespace).Get(ctx, filesystemName, metav1.GetOptions{})
	if err != nil {
		klog.V(4).Infof("[%s] Filesystem CR '%s' not found in namespace '%s', isMDREnabledOnFS is false: %v",
			loggerId, filesystemName, filesystemCRNamespace, err)
		return false
	}

	// Extract replication value from spec.local.replication
	replication, found, err := unstructured.NestedString(filesystemCR.Object, "spec", "local", "replication")
	if err != nil {
		klog.V(4).Infof("[%s] Error reading spec.local.replication from Filesystem CR '%s', isMDREnabledOnFS is false: %v",
			loggerId, filesystemName, err)
		return false
	}

	if !found {
		klog.V(4).Infof("[%s] spec.local.replication not found in Filesystem CR '%s', isMDREnabledOnFS is false",
			loggerId, filesystemName)
		return false
	}

	// Check if replication is "external"
	isExternal := strings.ToLower(replication) == externalReplication
	if isExternal {
		klog.Infof("[%s] Filesystem CR '%s' has spec.local.replication='%s', isMDREnabledOnFS is true",
			loggerId, filesystemName, replication)
	} else {
		klog.V(4).Infof("[%s] Filesystem CR '%s' has spec.local.replication='%s' (not 'external'), isMDREnabledOnFS is false",
			loggerId, filesystemName, replication)
	}

	return isExternal
}

func (cs *ScaleControllerServer) validateCG(ctx context.Context, connector connectors.SpectrumScaleConnector, volBackendFs string, consistencyGroup string) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.V(4).Infof("[%s] Validate CG for volBackendFs [%v]", loggerId, volBackendFs)

	fsetlist, err := connector.ListCSIIndependentFilesets(ctx, volBackendFs)
	if err != nil {
		return "", err
	}

	klog.V(4).Infof("[%s] Validate CG response fsetlist [%v]", loggerId, fsetlist)
	var flist []string
	pvcns := consistencyGroup[cgPrefixLen:]

	for _, fset := range fsetlist {
		if len(fset.FilesetName) > cgPrefixLen {
			if fset.FilesetName[cgPrefixLen:] == pvcns {
				flist = append(flist, fset.FilesetName)
			}
		}
	}

	klog.Infof("[%s] Filesets with namespace [%s] as suffix: [%v]", loggerId, pvcns, flist)

	// no fileset with this namespace found
	if len(flist) == 0 {
		return consistencyGroup, nil
	}

	// multiple filesets with this namespace found
	if len(flist) > 1 {
		return "", status.Error(codes.Internal, fmt.Sprintf("conflicting filesets found %+v", flist))
	}

	// this is either local CG or Remote CG
	return flist[0], nil
}
