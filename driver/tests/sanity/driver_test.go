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

package sanity

import (
	"os"
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
	cid         = "123"
	fsName      = "fs_test"
	primaryFset = ".csi"
	nodeName    = "edunn-master.fyre.ibm.com"
	basePath    = "volumes"
	parentFset  = ".csi"
)

const (
	JUNIT_FILE = "JUNIT_FILE"
)

var _ = BeforeSuite(func() {
	nodeID := "unittest"

	//not the best way to mock, but w/e
	os.Setenv("SCALE_HOSTPATH", "/ibm/"+fsName)

	configMap := &settings.ConfigMap{
		Clusters: []settings.Cluster{
			{
				ID: cid,
				Primary: settings.Primary{
					PrimaryCid:  cid,
					PrimaryFs:   fsName,
					PrimaryFset: "fset_test",
				},
				RestAPI: []settings.RestAPI{
					{
						GuiHost: "not.a.real.endpoint.ibm.com",
						GuiPort: 443,
					},
				},
				SecureSslMode: false,
				Secrets:       "k8s-fake-secret-name",
				MgmtUsername:  "fake",
				MgmtPassword:  "fake",
			},
		},
	}
	//test our Validation
	err := configMap.Validate()
	Expect(err).NotTo(HaveOccurred())

	fab := newFakeConnectorFactory()
	primary := fab.NewConnector(configMap.Primary)
	primary.MountFilesystem(configMap.Primary.PrimaryFs, nodeID)

	//make basePath for directory volume
	primary.MakeDirectory(fsName, basePath, "0", "0")

	driver = scale.NewFakeDriver(
		"csi-sanity",
		"0.0.0",
		nodeID,
		configMap,
		fab,
	)

	//test our PluginInitialize
	err = driver.PluginInitialize(configMap)
	Expect(err).NotTo(HaveOccurred())

	go driver.Run("unix://" + address)
})

var (
	dirCtx *sanity.TestContext
	depCtx *sanity.TestContext
	indCtx *sanity.TestContext
)

var _ = Describe("CSI Mock Sanity", func() {

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

var _ = AfterSuite(func() {
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
