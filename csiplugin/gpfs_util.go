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
	"strconv"
	//"os/exec"
	//"strings"
	//"errors"

	"github.com/golang/glog"
)

type gpfsVolume struct {
	VolName       string `json:"volName"`
	VolID         string `json:"volID"`
	VolSize       int64  `json:"volSize"`
	VolIscsi      bool   `json:"volIscsi"`
	VolIscsiVid   string `json:"volIscsiVid"`
}


// CreateImage creates a new volume with provision and volume options.
func createGpfsImage(pOpts *gpfsVolume, volSzMb int) error {
	var err error

	volName := pOpts.VolID //Name
	var device string = ""

	/* 
	* Check if device for NSD must be backed by iSCSI
	* and if so, create/attach an iSCSI volume.
	*/
	if (pOpts.VolIscsi) {
		// Create iSCSI volume.
		glog.V(4).Infof("create iSCSI volume (%s)", volName)
		pbVol, err := iscsiOps.CreateVolume(volName, volSzMb)
		if err != nil {
			glog.Errorf("Failed to create volume (%s): %v", volName, err)
			return err
		}
		pOpts.VolIscsiVid = pbVol.Vid.Val

		// Attach volume to all nodes in gpfs cluster.
		nodes := []string{"fin27p","fin31p","fin57p"} // TODO
		for _, node := range nodes {
			err = ops.AttachIscsiDevicesToNode(pbVol.Iphost, pbVol.Iqn, node)
			if err != nil {
				glog.Errorf("Failed to attach devices to node (%s): %v", node, err)
				return err
			}
		}
		device = iscsiOps.GetDevicePath(pbVol.Iphost, pbVol.Iqn, strconv.Itoa(int(pbVol.Lun)))
		// Pick node for NSD
		device = "fin31p:"+device
		
	} else {
	        // Pick raw device based on some policy.
        	device, err = ops.Sched.PickNextRawDevice()
	        if err != nil {
			glog.Errorf("Failed to pick backing device: %v", err)
                	return err
	        }
	}

	glog.V(4).Infof("create NSD (%s) backed by %s", volName, device)
	err = ops.CreateNSD(volName, device)
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
	var err error
	gpfsVol := &gpfsVolume{}
	gpfsVol.VolIscsi, err = strconv.ParseBool(volOptions["volIscsi"])
	if err != nil {
		return nil, fmt.Errorf("Missing required parameter volIscsi, %v", err)
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
	//var output []byte
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

	//glog.Errorf("failed to delete gpfs image: %v, command output: %s", err, string(output))
	err = iscsiOps.DeleteVolume(pOpts.VolIscsiVid)
	return err
}
