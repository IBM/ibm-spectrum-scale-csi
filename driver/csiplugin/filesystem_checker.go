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
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// ShouldDiscoverCGFileset checks if DISCOVER_CG_FILESET should be enabled
// by querying the Filesystem CR in ibm-spectrum-scale namespace and checking
// if fs.spec.local.replication == "external"
func ShouldDiscoverCGFileset(ctx context.Context, clientset *kubernetes.Clientset, filesystemName string) bool {
	loggerId := utils.GetLoggerId(ctx)

	// Check if DISCOVER_CG_FILESET is explicitly disabled via environment variable
	discoverCGEnv := strings.ToUpper(utils.GetEnv(settings.DiscoverCGFileset, ""))
	if discoverCGEnv == "DISABLED" {
		klog.V(4).Infof("[%s] DISCOVER_CG_FILESET is explicitly disabled via environment variable", loggerId)
		return false
	}

	// If explicitly enabled via environment variable, return true
	if discoverCGEnv == "ENABLED" || discoverCGEnv == "TRUE" {
		klog.V(4).Infof("[%s] DISCOVER_CG_FILESET is explicitly enabled via environment variable", loggerId)
		return true
	}

	// Otherwise, check the Filesystem CR
	// Get the REST config from the clientset
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.V(4).Infof("[%s] Unable to get in-cluster config for Filesystem CR check, defaulting DISCOVER_CG_FILESET to false: %v", loggerId, err)
		return false
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.V(4).Infof("[%s] Unable to create dynamic client for Filesystem CR check, defaulting DISCOVER_CG_FILESET to false: %v", loggerId, err)
		return false
	}

	// Define the GVR for Filesystem CR
	filesystemGVR := schema.GroupVersionResource{
		Group:    filesystemCRGroup,
		Version:  filesystemCRVersion,
		Resource: filesystemCRResource,
	}

	// Try to get the Filesystem CR
	filesystemCR, err := dynamicClient.Resource(filesystemGVR).Namespace(filesystemCRNamespace).Get(ctx, filesystemName, metav1.GetOptions{})
	if err != nil {
		klog.V(4).Infof("[%s] Filesystem CR '%s' not found in namespace '%s', defaulting DISCOVER_CG_FILESET to false: %v",
			loggerId, filesystemName, filesystemCRNamespace, err)
		return false
	}

	// Extract replication value from spec.local.replication
	replication, found, err := unstructured.NestedString(filesystemCR.Object, "spec", "local", "replication")
	if err != nil {
		klog.V(4).Infof("[%s] Error reading spec.local.replication from Filesystem CR '%s', defaulting DISCOVER_CG_FILESET to false: %v",
			loggerId, filesystemName, err)
		return false
	}

	if !found {
		klog.V(4).Infof("[%s] spec.local.replication not found in Filesystem CR '%s', defaulting DISCOVER_CG_FILESET to false",
			loggerId, filesystemName)
		return false
	}

	// Check if replication is "external"
	isExternal := strings.ToLower(replication) == externalReplication
	if isExternal {
		klog.Infof("[%s] Filesystem CR '%s' has spec.local.replication='%s', enabling DISCOVER_CG_FILESET",
			loggerId, filesystemName, replication)
	} else {
		klog.V(4).Infof("[%s] Filesystem CR '%s' has spec.local.replication='%s' (not 'external'), DISCOVER_CG_FILESET remains disabled",
			loggerId, filesystemName, replication)
	}

	return isExternal
}
