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

package config

import (
	"fmt"
	"io/ioutil"
	"os"

	v1 "github.com/IBM/ibm-spectrum-scale-csi/api/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

var csiLog = log.Log.WithName("settings")

const (
	EnvNameCrYaml string = "CR_YAML"
	NodeAgentTag         = "1.0.0"

	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	NodeAgentPort = "10086"

	IBMRegistryUsername    = "ibmcom"
	K8SRegistryUsername    = "k8s.gcr.io/sig-storage"
	QuayRegistryUsername   = "quay.io/k8scsi"
	RedHatRegistryUsername = "registry.redhat.io/openshift4"
)

var DefaultCr v1.CSIScaleOperator

// var DefaultSidecarsByName map[string]v1.CSISidecar

var OfficialRegistriesUsernames = sets.NewString(IBMRegistryUsername, K8SRegistryUsername,
	QuayRegistryUsername, RedHatRegistryUsername)

func LoadDefaultsOfCSIScaleOperator() error {
	crYamlPath := os.Getenv(EnvNameCrYaml)

	if crYamlPath == "" {
		return fmt.Errorf("environment variable %q was not set", EnvNameCrYaml)
	}

	yamlFile, err := ioutil.ReadFile(crYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %v", yamlFile, err)
	}

	err = yaml.Unmarshal(yamlFile, &DefaultCr)
	if err != nil {
		return fmt.Errorf("error unmarshaling yaml: %v", err)
	}

	logger := csiLog.WithName("LoadDefaultsOfCSIScaleOperator")
	logger.Info("getting DefaultSidecarImageByName")

	// DefaultSidecarsByName = make(map[string]v1.CSISidecar)
	// for _, sidecar := range DefaultCr.Spec.Sidecars {
	// 	logger.Info("printing sidecar names", "name", sidecar)
	// 	DefaultSidecarsByName[sidecar.Name] = sidecar
	// }

	return nil
}
