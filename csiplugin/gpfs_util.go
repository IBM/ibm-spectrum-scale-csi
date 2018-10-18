/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gpfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	//"os/exec"
	//"strings"
	//"errors"

	"github.com/golang/glog"
)

type gpfsVolume struct {
	VolName       string `json:"volName"`
	VolID         string `json:"volID"`
	VolSize       int64  `json:"volSize"`
	VolFormat     string `json:"volFormat"`
}


// CreateImage creates a new volume with provision and volume options.
func createGpfsImage(pOpts *gpfsVolume, volSz int) error {
	var err error

	volName := pOpts.VolID //Name

	glog.V(4).Infof("create NSD (%s)", volName)
	err = ops.CreateNSD(volName)
	if err != nil {
		glog.Errorf("Failed to create nsd (%s): %v", volName, err)
		return err
	}
	
	glog.V(4).Infof("create FS (%s)", volName)
	err = ops.CreateFS(volName, volName)
	if err != nil {
		glog.Errorf("Failed to create fs (%s): %v", volName, err)
		return err
	}
	return nil
}

func getGpfsVolumeOptions(volOptions map[string]string) (*gpfsVolume, error) {
	var ok bool
	gpfsVol := &gpfsVolume{}
	gpfsVol.VolFormat, ok = volOptions["volFormat"]
	if !ok {
		return nil, fmt.Errorf("Missing required parameter volFormat")
	}
	return gpfsVol, nil
}

func getGpfsVolumeByName(volName string) (*gpfsVolume, error) {
	for _, gpfsVol := range gpfsVolumes {
		if gpfsVol.VolName == volName {
			return gpfsVol, nil
		}
	}
	return nil, fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}

func persistVolInfo(image string, persistentStoragePath string, volInfo *gpfsVolume) error {
	file := path.Join(persistentStoragePath, image+".json")
	fp, err := os.Create(file)
	if err != nil {
		glog.Errorf("gpfs: failed to create persistent storage file %s with error: %v\n", file, err)
		return fmt.Errorf("gpfs: create err %s/%s", file, err)
	}
	defer fp.Close()
	encoder := json.NewEncoder(fp)
	if err = encoder.Encode(volInfo); err != nil {
		glog.Errorf("gpfs: failed to encode volInfo: %+v for file: %s with error: %v\n", volInfo, file, err)
		return fmt.Errorf("gpfs: encode err: %v", err)
	}
	glog.Infof("gpfs: successfully saved volInfo: %+v into file: %s\n", volInfo, file)
	return nil
}
func loadVolInfo(image string, persistentStoragePath string, volInfo *gpfsVolume) error {
	file := path.Join(persistentStoragePath, image+".json")
	fp, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("gpfs: open err %s/%s", file, err)
	}
	defer fp.Close()

	decoder := json.NewDecoder(fp)
	if err = decoder.Decode(volInfo); err != nil {
		return fmt.Errorf("gpfs: decode err: %v.", err)
	}

	return nil
}

func deleteVolInfo(image string, persistentStoragePath string) error {
	file := path.Join(persistentStoragePath, image+".json")
	glog.Infof("gpfs: Deleting file for Volume: %s at: %s resulting path: %+v\n", image, persistentStoragePath, file)
	err := os.Remove(file)
	if err != nil {
		if err != os.ErrNotExist {
			return fmt.Errorf("gpfs: error removing file: %s/%s", file, err)
		}
	}
	return nil
}

// DeleteImage deletes a volume with provision and volume options.
func deleteGpfsImage(pOpts *gpfsVolume) error {
	var output []byte
	var err error
	image := pOpts.VolID //Name
	glog.V(4).Infof("gpfs: rm %s", image)

        // Delete FS
        glog.V(4).Infof("Delete filesystem (%s)", image)
        err = ops.DeleteFS(image)
        if err != nil {
                return fmt.Errorf("failed to delete gpfs: %v", err)
        }

        // Delete NSD
        glog.V(4).Infof("Delete NSD (%s)", image)
        err = ops.DeleteNSD(image)
        if err != nil {
                return fmt.Errorf("failed to delete gpfs nsd: %v", err)
        }

	glog.Errorf("failed to delete gpfs image: %v, command output: %s", err, string(output))
	return err
}
