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

package integration

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	"github.com/kubernetes-csi/csi-test/v3/pkg/sanity"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

var (
	address = os.TempDir() + "csi.sock"
	driver  *scale.Driver
)

var (
	cid         = "14100825633803609761"
	fsName      = "fs1"
	primaryFset = ".csi"
	nodeName    = "edunn-master.fyre.ibm.com"
	basePath    = "volumes"
	parentFset  = ".csi"
)

const (
	SCALE_GUI      = "SCALE_GUI"
	SCALE_USER     = "SCALE_USER"
	SCALE_PASSWORD = "SCALE_PASSWORD"
	SCALE_CONFIG   = "SCALE_CONFIG"
	JUNIT_FILE     = "JUNIT_FILE"
)

var _ = SynchronizedBeforeSuite(func() []byte {

	os.Setenv("SCALE_HOSTPATH", "/ibm/"+fsName)

	var configMap *settings.ConfigMap
	if configMapFile := os.Getenv(SCALE_CONFIG); configMapFile != "" {
		var err error
		configMap, err = settings.LoadScaleConfig()
		if err != nil {
			Fail(fmt.Sprintf("could not load testsuite configMap: %v", err))
		}
	} else {

		//defaults
		scaleGuiHost := "localhost"
		scaleGuiPort := 443
		//attempt to parse "SCALE_GUI" env
		if scaleGui := os.Getenv(SCALE_GUI); scaleGui != "" {
			url, err := url.Parse(scaleGui)
			if err != nil {
				Fail(fmt.Sprintf("could not parse URL from \"SCALE_GUI\": %v", err))
			}
			scaleGuiHost = url.Hostname()
			scaleGuiPort, err = strconv.Atoi(url.Port())
			if err != nil {
				scaleGuiPort = 443
			}
		}

		configMap = &settings.ConfigMap{
			InsecureSkipTLSVerify: true,
			Clusters: []settings.Cluster{
				{
					ID: cid,
					Primary: settings.Primary{
						PrimaryCid:  cid,
						PrimaryFs:   fsName,
						PrimaryFset: primaryFset,
					},
					RestAPI: []settings.RestAPI{
						{
							GuiHost: scaleGuiHost,
							GuiPort: scaleGuiPort,
						},
					},
					SecureSslMode: false,
					Secrets:       "k8s-fake-secret-name",
					MgmtUsername:  os.Getenv(SCALE_USER),
					MgmtPassword:  os.Getenv(SCALE_PASSWORD),
				},
			},
		}
	}

	//test our Validation
	err := configMap.Validate()
	if err != nil {
		Fail(fmt.Sprintf("could not validate testsuite configMap: %v", err))
	}

	driver = scale.NewDriver(
		"csi-sanity",
		"0.0.0",
		nodeName,
		configMap,
	)

	//test our PluginInitialize
	err = driver.PluginInitialize(configMap)
	if err != nil {
		Fail(fmt.Sprintf("could not initialize plugin: %v", err))
	}

	go driver.Run("unix://" + address)

	return []byte{}
}, func(_ []byte) {})

var (
	dirCtx *sanity.TestContext
	depCtx *sanity.TestContext
	indCtx *sanity.TestContext
)

var _ = Describe("CSI Integration Sanity", func() {

	Context("Directory-based StorageClass", func() {

		config := sanity.NewTestConfig()
		config.Address = address
		config.TestVolumeParameters = map[string]string{
			settings.ClusterId:    cid,
			settings.VolBackendFs: fsName,
			settings.VolDirPath:   basePath,
		}

		dirCtx = sanity.GinkgoTest(&config)
	})

	Context("Dependent Fileset-based StorageClass", func() {

		config := sanity.NewTestConfig()
		config.Address = address
		config.TestVolumeParameters = map[string]string{
			settings.ClusterId:    cid,
			settings.VolBackendFs: fsName,
			settings.ParentFset:   parentFset,
			settings.FilesetType:  "dependent",
		}

		sanity.GinkgoTest(&config)

		depCtx = sanity.GinkgoTest(&config)
	})

	Context("Independent Fileset-based StorageClass", func() {

		config := sanity.NewTestConfig()
		config.Address = address
		config.TestVolumeParameters = map[string]string{
			settings.ClusterId:    cid,
			settings.VolBackendFs: fsName,
			settings.FilesetType:  "independent",
			settings.InodeLimit:   "1024",
		}

		indCtx = sanity.GinkgoTest(&config)
	})
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	//close test gRPC client
	dirCtx.Finalize()
	depCtx.Finalize()
	indCtx.Finalize()

	//close under-test gRPC server (our CSI driver)
	driver.Stop()
})

func TestDriver(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	if junitFile := os.Getenv(JUNIT_FILE); junitFile != "" {
		junitReporter := reporters.NewJUnitReporter(junitFile)
		specReporters = append(specReporters, junitReporter)
	}
	RunSpecsWithDefaultAndCustomReporters(t, "CSI Driver Test Suite", specReporters)
}
