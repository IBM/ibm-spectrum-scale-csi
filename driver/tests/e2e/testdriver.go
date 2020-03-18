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

package e2e

import (
	"fmt"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
)

type scaleDriver struct {
	driverInfo testsuites.DriverInfo
}

//Confirm we match required interfaces
var _ testsuites.TestDriver = &scaleDriver{}

//var _ testsuites.PreprovisionedVolumeTestDriver = &scaleDriver{}
//var _ testsuites.PreprovisionedPVTestDriver = &scaleDriver{}
var _ testsuites.DynamicPVTestDriver = &scaleDriver{}

type filesetVolume struct {
}

func NewTestDriver() testsuites.TestDriver {
	return &scaleDriver{
		driverInfo: testsuites.DriverInfo{
			Name:        "csi-scale-driver", //must match what we deploy as
			MaxFileSize: testpatterns.FileSizeLarge,
			SupportedFsType: sets.NewString(
				"", //Default fsType
			),
			Capabilities: map[testsuites.Capability]bool{
				testsuites.CapPersistence:  true, //does data persiste across pod restarts
				testsuites.CapExec:         true, //can file be executed within volume
				testsuites.CapRWX:          true, //can volume be ReadWriteMany
				testsuites.CapMultiPODs:    true, //can volume publish to multiple pods
				testsuites.CapVolumeLimits: true, //*slow test*
				testsuites.CapFsGroup:      true, //can volume ownership be set
				//testsuites.CapSingleNodeVolume: true, //can volume be ReadWriteSingle
			},
		},
	}
}

func (d *scaleDriver) GetDriverInfo() *testsuites.DriverInfo {
	return &d.driverInfo
}

func (d *scaleDriver) SkipUnsupportedTest(testpatterns.TestPattern) {
	//nothing for now?
}

func (d *scaleDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	config := &testsuites.PerTestConfig{
		Driver:    d,
		Prefix:    "gpfs",
		Framework: f,
	}

	return config, func() {}
}

/*
// CreateVolume creates a pre-provisioned volume of the desired volume type.
func (d *scaleDriver) CreateVolume(config *testsuites.PerTestConfig, volumeType testpatterns.TestVolType) testsuites.TestVolume {
	return nil
}
*/
/* needed for pre-provisioned
func (d *filesetVolume) DeleteVolume() {

}*/

// GetPersistentVolumeSource returns a PersistentVolumeSource with volume node affinity for pre-provisioned Persistent Volume.
// It will set readOnly and fsType to the PersistentVolumeSource, if TestDriver supports both of them.
// It will return nil, if the TestDriver doesn't support either of the parameters.
/*func (d *scaleDriver) GetPersistentVolumeSource(readOnly bool, fsType string, testVolume testsuites.TestVolume) (*v1.PersistentVolumeSource, *v1.VolumeNodeAffinity) {
	vol, _ := testVolume.(*filesetVolume)
	return &v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver:       d.driverInfo.Name,
			VolumeHandle: "nfs-vol",
			VolumeAttributes: map[string]string{
				"server":   fmt.Sprintf("%v", vol), //TODO
				"share":    "/",
				"readOnly": "true",
			},
		},
	}, nil
}*/

/* GetDynamicProvisionStorageClass returns a StorageClass dynamic provision Persistent Volume.
 * It will set fsType to the StorageClass, if TestDriver supports it.
 * It will return nil, if the TestDriver doesn't support it.
 */
func (d *scaleDriver) GetDynamicProvisionStorageClass(config *testsuites.PerTestConfig, fsType string) *storagev1.StorageClass {

	provisioner := config.GetUniqueDriverName()
	parameters := map[string]string{
		settings.FilesetType:  "independent",
		settings.ClusterId:    "16482346744146153652",
		settings.VolBackendFs: "fs1",
		settings.ParentFset:   ".csi",
		settings.VolDirPath:   "volumes/",
		settings.InodeLimit:   "1024",
	}
	bindingMode := storagev1.VolumeBindingImmediate
	ns := config.Framework.Namespace.Name
	suffix := fmt.Sprintf("%s-sc", provisioner)

	return testsuites.GetStorageClass(
		"csi-spectrum-scale", //TODO: hardcode deploy'd for now
		parameters,
		&bindingMode,
		ns,
		suffix,
	)
}

/* GetClaimSize returns the size of the volume that is to be provisioned ("5Gi", "1Mi").
 * The size must be chosen so that the resulting volume is large enough for all
 * enabled tests and within the range supported by the underlying storage.
 */
func (d *scaleDriver) GetClaimSize() string {
	return "5Gi"
}
