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
        "encoding/json"
	"fmt"
	"strings"
        "os"
        "path"
	"bytes"
	"strconv"
	"os/exec"
	"github.com/golang/glog"
       "google.golang.org/grpc/status"
       "google.golang.org/grpc/codes"
       "github.com/IBM/ibm-spectrum-scale-csi-driver/csiplugin/connectors"
)

type scaleVolume struct {
        VolName        string `json:"volName"`
        VolSize        uint64  `json:"volSize"`
        VolBackendFs   string `json:"volBackendFs"`
        IsFilesetBased bool   `json:"isFilesetBased"`
        VolDirBasePath string `json:"volDirBasePath"`
        VolUid         string `json:"volUid"`
        VolGid         string `json:"volGid"`
        ClusterId      string `json:"clusterId"`
        FilesetType    string `json:"filesetType"`
        InodeLimit     string `json:"inodeLimit"`
        Connector      connectors.SpectrumScaleConnector `json:"connector"`
        PrimaryConnector      connectors.SpectrumScaleConnector `json:"primaryConnector"`
        PrimarySLnkRelPath string `json:"primarySLnkRelPath"`
        PrimarySLnkPath string `json:"primarySLnkPath"`
        PrimaryFS      string `json:"primaryFS"`
        PrimaryFSMount string `json:"primaryFSMount"`
        ParentFileset string `json:"parentFileset"`
        LocalFS          string `json:"localFS"`
}

type scaleVolId struct {
        ClusterId string
        FsUUID    string
        FsetId    string
        DirPath   string
        SymLnkPath string
        IsFilesetBased bool
}

func getScaleVolumeOptions(volOptions map[string]string) (*scaleVolume, error) {
        //var err error
        scaleVol := &scaleVolume{}

        volBckFs, fsSpecified := volOptions[connectors.UserSpecifiedVolBackendFs]
        if fsSpecified {
                scaleVol.VolBackendFs = volBckFs
        } else {
                return &scaleVolume{}, status.Error(codes.InvalidArgument, "Volume Backend Filesystem not specified in request parameters")
        }

/* Check if either fileset based or LW volume. */
        volDirPath, volDirPathSpecified := volOptions[connectors.UserSpecifiedVolDirPath]
        if volDirPathSpecified {
                scaleVol.VolDirBasePath = volDirPath
                scaleVol.IsFilesetBased = false
        } else {
                scaleVol.VolDirBasePath = ""
                scaleVol.IsFilesetBased = true
        }

/* cluster Id not mandatory for LW volumes */

        if scaleVol.IsFilesetBased {
                clusterId, clusterIdSpecified := volOptions[connectors.UserSpecifiedClusterId]
                if clusterIdSpecified {
                        scaleVol.ClusterId = clusterId
                } else {
                        return &scaleVolume{}, status.Error(codes.InvalidArgument, "clusterId not specified in request parameters")
                }
        }

/* Get UID/GID */
        uid, uidSpecified := volOptions[connectors.UserSpecifiedUid]
        if uidSpecified {
                scaleVol.VolUid = uid
        } else {
                scaleVol.VolUid = ""
        }

        gid, gidSpecified := volOptions[connectors.UserSpecifiedGid]
        if gidSpecified {
                scaleVol.VolGid = gid
        } else {
                scaleVol.VolGid = ""
        }

        if scaleVol.IsFilesetBased {

                fsType, fsTypeSpecified := volOptions[connectors.UserSpecifiedFilesetType]
                if fsTypeSpecified {
                        scaleVol.FilesetType = fsType
                } else {
                        scaleVol.FilesetType = ""
                }
                inodeLim, inodeLimSpecified := volOptions[connectors.UserSpecifiedInodeLimit]
                if  inodeLimSpecified {
                        scaleVol.InodeLimit = inodeLim
                } else {
                        scaleVol.InodeLimit = ""
                }

                parentFileset, isparentFilesetSpecified := volOptions[connectors.UserSpecifiedParentFset]
                if isparentFilesetSpecified {
                        scaleVol.ParentFileset = parentFileset
                } else {
                        scaleVol.ParentFileset = ""
                }
        }
        return scaleVol, nil
}

func getScaleVolumeByName(volName string) (*scaleVolume, error) {
        glog.V(4).Infof("gpfs_util getScaleVolumeByName")

	for _, scaleVol := range scaleVolumes {
		if scaleVol.VolName == volName {
			return scaleVol, nil
		}
	}
	return nil, fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}

func persistVolInfo(image string, persistentStoragePath string, volInfo *scaleVolume) error {
        glog.V(4).Infof("gpfs_util persistVolInfo")

	file := path.Join(persistentStoragePath, image+".json")
	fp, err := os.Create(file)
	if err != nil {
		glog.Errorf("scale: failed to create persistent storage file %s with error: %v\n", file, err)
		return fmt.Errorf("scale: create err %s/%s", file, err)
	}
	defer fp.Close()
	encoder := json.NewEncoder(fp)
	if err = encoder.Encode(volInfo); err != nil {
		glog.Errorf("scale: failed to encode volInfo: %+v for file: %s with error: %v\n", volInfo, file, err)
		return fmt.Errorf("scale: encode err: %v", err)
	}
	glog.Infof("scale: successfully saved volInfo: %+v into file: %s\n", volInfo, file)
	return nil
}

func loadVolInfo(image string, persistentStoragePath string, volInfo *scaleVolume) error {
        glog.V(4).Infof("gpfs_util loadVolInfo")

	file := path.Join(persistentStoragePath, image+".json")
	fp, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("scale: open err %s/%s", file, err)
	}
	defer fp.Close()

	decoder := json.NewDecoder(fp)
	if err = decoder.Decode(volInfo); err != nil {
		return fmt.Errorf("scale: decode err: %v.", err)
	}

	return nil
}

func deleteVolInfo(image string, persistentStoragePath string) error {
        glog.V(4).Infof("gpfs_util deleteVolInfo")

	file := path.Join(persistentStoragePath, image+".json")
	glog.Infof("scale: Deleting file for Volume: %s at: %s resulting path: %+v\n", image, persistentStoragePath, file)
	err := os.Remove(file)
	if err != nil {
		if err != os.ErrNotExist {
			return fmt.Errorf("scale: error removing file: %s/%s", file, err)
		}
	}
	return nil
}

func executeCmd(command string, args []string) ([]byte, error) {
        glog.V(5).Infof("gpfs_util executeCmd")

        cmd := exec.Command(command, args...)
        var stdout bytes.Buffer
        var stderr bytes.Buffer

        cmd.Stdout = &stdout
        cmd.Stderr = &stderr
        err := cmd.Run()
        stdOut := stdout.Bytes()
        return stdOut, err
}

func ConvertToBytes(inputStr string) (uint64, error) {
        var Iter int
        var byteSlice []byte
        var retValue uint64
        var uintMax64 uint64

        byteSlice = []byte(inputStr)
        uintMax64 = (1 << 64) - 1

        for Iter = 0; Iter < len(byteSlice); Iter++ {
                if ('0' <= byteSlice[Iter]) &&
                        (byteSlice[Iter] <= '9') {
                        continue
                } else {
                        break
                }
        }
       if Iter == 0 {
                return 0, fmt.Errorf("Invalid number specified %v", inputStr)
        }

        retValue, err := strconv.ParseUint(inputStr[:Iter], 10, 64)

        if err != nil {
                return 0, fmt.Errorf("ParseUint Failed for %v", inputStr[:Iter])
        }

        if Iter == len(inputStr) {
               return retValue, nil
        }

        unit := strings.TrimSpace(string(byteSlice[Iter:]))
        unit = strings.ToLower(unit)

        switch unit {
        case "b", "bytes":
                /* Nothing to do here */
        case "k", "kb", "kilobytes", "kilobyte":
                retValue *= 1024
        case "m", "mb", "megabytes", "megabyte":
                retValue *= (1024 * 1024)
        case "g", "gb", "gigabytes", "gigabyte":
                retValue *= (1024 * 1024 * 1024)
        case "t", "tb", "terabytes", "terabyte":
                retValue *= (1024 * 1024 * 1024 * 1024)
        default:
                return 0, fmt.Errorf("Invalid Unit %v supplied with %v", unit, inputStr)
        }

        if retValue > uintMax64 {
                return 0, fmt.Errorf("Overflow detected %v", inputStr)
        }

        return retValue, nil
}
