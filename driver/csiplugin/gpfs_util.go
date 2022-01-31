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
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	dependentFileset     = "dependent"
	independentFileset   = "independent"
	storageClassAdvanced = "advanced"
)

type scaleVolume struct {
	VolName            string                            `json:"volName"`
	VolSize            uint64                            `json:"volSize"`
	VolBackendFs       string                            `json:"volBackendFs"`
	IsFilesetBased     bool                              `json:"isFilesetBased"`
	VolDirBasePath     string                            `json:"volDirBasePath"`
	VolUid             string                            `json:"volUid"`
	VolGid             string                            `json:"volGid"`
	VolPermissions     string                            `json:"volPermissions"`
	ClusterId          string                            `json:"clusterId"`
	FilesetType        string                            `json:"filesetType"`
	InodeLimit         string                            `json:"inodeLimit"`
	Connector          connectors.SpectrumScaleConnector `json:"connector"`
	PrimaryConnector   connectors.SpectrumScaleConnector `json:"primaryConnector"`
	PrimarySLnkRelPath string                            `json:"primarySLnkRelPath"`
	PrimarySLnkPath    string                            `json:"primarySLnkPath"`
	PrimaryFS          string                            `json:"primaryFS"`
	PrimaryFSMount     string                            `json:"primaryFSMount"`
	ParentFileset      string                            `json:"parentFileset"`
	LocalFS            string                            `json:"localFS"`
	TargetPath         string                            `json:"targetPath"`
	FsetLinkPath       string                            `json:"fsetLinkPath"`
	FsMountPoint       string                            `json:"fsMountPoint"`
	NodeClass          string                            `json:"nodeClass"`
	StorageClassType   string                            `json:"storageClassType"`
	ConsistencyGroup   string                            `json:"consistencyGroup"`
	Compression        string                            `json:"compression"`
	Tier               string                            `json:"tier"`
}

type scaleVolId struct {
	ClusterId        string
	FsUUID           string
	FsName           string
	FsetId           string
	FsetName         string
	DirPath          string
	Path		 string
	IsFilesetBased   bool
	StorageClassType string
	ConsistencyGroup string
	VolType          string
}

type scaleSnapId struct {
	ClusterId string
	FsUUID    string
	FsetName  string
	SnapName  string
	Path      string
	FsName    string
}

//nolint
type scaleVolSnapshot struct {
	SnapName   string `json:"snapName"`
	SourceVol  string `json:"sourceVol"`
	Filesystem string `json:"filesystem"`
	Fileset    string `json:"fileset"`
	ClusterId  string `json:"clusterId"`
	SnapSize   uint64 `json:"snapSize"`
} //nolint

//nolint
type scaleVolSnapId struct {
	ClusterId string
	FsUUID    string
	FsetId    string
	SnapId    string
} //nolint

func IsValidCompressionAlgorithm(input string) bool {
	switch strings.ToLower(input) {
	case
		"z",
		"lz4",
		"zfast",
		"alphae",
		"alphah":
		return true
	}
	return false
}

func getRemoteFsName(remoteDeviceName string) string {
	splitDevName := strings.Split(remoteDeviceName, ":")
	remDevFs := splitDevName[len(splitDevName)-1]
	return remDevFs
}

func getScaleVolumeOptions(volOptions map[string]string) (*scaleVolume, error) { //nolint:gocyclo,funlen
	//var err error
	scaleVol := &scaleVolume{}

	volBckFs, fsSpecified := volOptions[connectors.UserSpecifiedVolBackendFs]
	volDirPath, volDirPathSpecified := volOptions[connectors.UserSpecifiedVolDirPath]
	clusterID, clusterIDSpecified := volOptions[connectors.UserSpecifiedClusterId]
	uid, uidSpecified := volOptions[connectors.UserSpecifiedUid]
	gid, gidSpecified := volOptions[connectors.UserSpecifiedGid]
	fsType, fsTypeSpecified := volOptions[connectors.UserSpecifiedFilesetType]
	inodeLim, inodeLimSpecified := volOptions[connectors.UserSpecifiedInodeLimit]
	parentFileset, isparentFilesetSpecified := volOptions[connectors.UserSpecifiedParentFset]
	nodeClass, isNodeClassSpecified := volOptions[connectors.UserSpecifiedNodeClass]
	permissions, isPermissionsSpecified := volOptions[connectors.UserSpecifiedPermissions]
	storageClassType, isSCTypeSpecified := volOptions[connectors.UserSpecifiedStorageClassType]
	compression, isCompressionSpecified := volOptions[connectors.UserSpecifiedCompression]
	tier, isTierSpecified := volOptions[connectors.UserSpecifiedTier]

	// Handling empty values
	scaleVol.VolDirBasePath = ""
	scaleVol.InodeLimit = ""
	scaleVol.FilesetType = ""
	scaleVol.ClusterId = ""
	scaleVol.NodeClass = ""
	scaleVol.ConsistencyGroup = ""
	scaleVol.StorageClassType = ""
	scaleVol.Compression = ""
	scaleVol.Tier = ""

	if isSCTypeSpecified && storageClassType == "" {
		isSCTypeSpecified = false
	}
	if isSCTypeSpecified {
		//This is a new type of StorageClass
		if storageClassType != storageClassAdvanced {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "storageClassType must be \""+storageClassAdvanced+"\" if specified.")
		}
		scaleVol.StorageClassType = storageClassType
	}

	if fsSpecified && volBckFs == "" {
		fsSpecified = false
	}

	if fsSpecified {
		scaleVol.VolBackendFs = volBckFs
	} else {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "volBackendFs must be specified in storageClass")
	}

	if fsTypeSpecified && fsType == "" {
		fsTypeSpecified = false
	}

	if volDirPathSpecified && volDirPath == "" {
		volDirPathSpecified = false
	}

	if !fsTypeSpecified && !volDirPathSpecified && !isSCTypeSpecified {
		fsTypeSpecified = true
		fsType = independentFileset
	}

	if uidSpecified && uid == "" {
		uidSpecified = false
	}

	if gidSpecified && gid == "" {
		gidSpecified = false
	}

	if gidSpecified && !uidSpecified {
		uidSpecified = true
		uid = "0"
	}

	if inodeLimSpecified && inodeLim == "" {
		inodeLimSpecified = false
	}

	if isparentFilesetSpecified && parentFileset == "" {
		isparentFilesetSpecified = false
	}
	if clusterIDSpecified && clusterID != "" {
		scaleVol.ClusterId = clusterID
	}

	if isPermissionsSpecified && permissions == "" {
		isPermissionsSpecified = false
	}

	if volDirPathSpecified {
		if fsTypeSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "filesetType and volDirBasePath must not be specified together in storageClass")
		}
		if isparentFilesetSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "parentFileset and volDirBasePath must not be specified together in storageClass")
		}
		if inodeLimSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "inodeLimit and volDirBasePath must not be specified together in storageClass")
		}
	}

	if fsTypeSpecified {
		if fsType == dependentFileset {
			if inodeLimSpecified {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "inodeLimit and filesetType=dependent must not be specified together in storageClass")
			}
		} else if fsType == independentFileset {
			if isparentFilesetSpecified {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "parentFileset and filesetType=independent(Default) must not be specified together in storageClass")
			}
		} else {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "Invalid value specified for filesetType in storageClass")
		}
	}

	if fsTypeSpecified && inodeLimSpecified {
		inodelimit, err := strconv.Atoi(inodeLim)
		if err != nil {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "Invalid value specified for inodeLimit in storageClass")
		}
		if inodelimit < 1024 {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "inodeLimit specified in storageClass must be equal to or greater than 1024")
		}
	}

	/* Check if either fileset based or LW volume. */

	if volDirPathSpecified {
		scaleVol.VolDirBasePath = volDirPath
		scaleVol.IsFilesetBased = false
	}

	if fsTypeSpecified && isSCTypeSpecified {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"type\" and \"filesetType\" are mutually exclusive")
	}
	if fsTypeSpecified && inodeLimSpecified {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"type\" and \"inodeLimit\" are mutually exclusive")
	}
	if fsTypeSpecified && volDirPathSpecified {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"type\" and \"volDirBasePath\" are mutually exclusive")
	}
 
	if fsTypeSpecified || isSCTypeSpecified {
		scaleVol.IsFilesetBased = true
	}

	if isCompressionSpecified && compression == "" {
		isCompressionSpecified = false
	}
	if isTierSpecified && tier == "" {
		isTierSpecified = false
	}
	if scaleVol.IsFilesetBased {
		if isCompressionSpecified {
			scaleVol.Compression = compression
		}
		if isTierSpecified {
			scaleVol.Tier = tier
		}
	} else {
		if isCompressionSpecified || isTierSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"compression\" and \"tier\" are not supported in storageClass for lightweight volumes")
		}
	}

	/* Get UID/GID */
	if uidSpecified {
		scaleVol.VolUid = uid
	}

	if gidSpecified {
		scaleVol.VolGid = gid
	}

	if isPermissionsSpecified {
		_, err := strconv.Atoi(permissions)
		if err != nil || len(permissions) != 3 {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "invalid value specified for permissions")
		}

		for _, n := range permissions {
			if n < 48 || n > 55 {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "invalid value specified for permissions")
			}
		}

		scaleVol.VolPermissions = permissions
	}

	if scaleVol.IsFilesetBased {
		if fsTypeSpecified {
			scaleVol.FilesetType = fsType
		}
		if isparentFilesetSpecified {
			scaleVol.ParentFileset = parentFileset
		}
		if inodeLimSpecified {
			scaleVol.InodeLimit = inodeLim
		}
	}

	if isNodeClassSpecified {
		scaleVol.NodeClass = nodeClass
	}

	scaleVol.ConsistencyGroup = volOptions["csi.storage.k8s.io/pvc/namespace"]

	if isCompressionSpecified {
		// Default compression will be Z if set but not specified
		if strings.ToLower(compression) == "true" {
			glog.V(5).Infof("gpfs_util compression was set to true. Defaulting to Z")
			compression = "z"
		}

		if !IsValidCompressionAlgorithm(compression) {
			glog.V(5).Infof("gpfs_util invalid compression algorithm specified: %s",
				compression)
			return &scaleVolume{}, status.Errorf(codes.InvalidArgument,
				"invalid compression algorithm specified: %s", compression)
		}
		scaleVol.Compression = compression
		glog.V(5).Infof("gpfs_util compression was set to %s", compression)
	}

	if isTierSpecified && tier != "" {
		scaleVol.Tier = tier
		glog.V(5).Infof("gpfs_util tier was set: %s", tier)
	}

	return scaleVol, nil
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

const (
	SCALE_NODE_MAPPING_PREFIX = "SCALE_NODE_MAPPING_PREFIX"
	DefaultScaleNodeMapPrefix = "K8sNodePrefix_"
)

// getNodeMapping returns the configured mapping to GPFS Admin Node Name given Kubernetes Node ID.
func getNodeMapping(kubernetesNodeID string) (gpfsAdminName string) {
	gpfsAdminName = utils.GetEnv(kubernetesNodeID, notFound)
	// Additional node mapping check in case of k8s node id start with number.
	if gpfsAdminName == notFound {
		prefix := utils.GetEnv(SCALE_NODE_MAPPING_PREFIX, DefaultScaleNodeMapPrefix)
		gpfsAdminName = utils.GetEnv(prefix+kubernetesNodeID, notFound)
		if gpfsAdminName == notFound {
			glog.V(4).Infof("getNodeMapping: scale node mapping not found for %s using %s", prefix+kubernetesNodeID, kubernetesNodeID)
			gpfsAdminName = kubernetesNodeID
		}
	}
	return gpfsAdminName
}

const (
	SHORTNAME_NODE_MAPPING = "SHORTNAME_NODE_MAPPING"
	SKIP_MOUNT_UNMOUNT     = "SKIP_MOUNT_UNMOUNT"
)

func shortnameInSlice(shortname string, nodeNames []string) bool {
	glog.V(6).Infof("gpfs_util shortnameInSlice. string: %s, slice: %v", shortname, nodeNames)
	for _, name := range nodeNames {
		short := strings.SplitN(name, ".", 2)[0]
		if strings.EqualFold(short, shortname) {
			return true
		}
	}
	return false
}

func numberInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getVolIDMembers(vID string) (scaleVolId, error) {
	splitVid := strings.Split(vID, ";")
	var vIdMem scaleVolId

	if len(splitVid) == 3 {
		/* This is LW volume */
		/* <cluster_id>;<filesystem_uuid>;path=<symlink_path> */
		vIdMem.ClusterId = splitVid[0]
		vIdMem.FsUUID = splitVid[1]
		SlnkPart := splitVid[2]
		slnkSplit := strings.Split(SlnkPart, "=")
		if len(slnkSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vID))
		}
		vIdMem.Path = slnkSplit[1]
		vIdMem.IsFilesetBased = false
		return vIdMem, nil
	}

	if len(splitVid) == 4 {
		/* This is fileset Based volume */
		/* <cluster_id>;<filesystem_uuid>;fileset=<fileset_id>;path=<symlink_path> */
		vIdMem.ClusterId = splitVid[0]
		vIdMem.FsUUID = splitVid[1]
		fileSetPart := splitVid[2]
		fileSetSplit := strings.Split(fileSetPart, "=")
		if len(fileSetSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vID))
		}

		if fileSetSplit[0] == "filesetName" {
			vIdMem.FsetName = fileSetSplit[1]
		} else {
			vIdMem.FsetId = fileSetSplit[1]
		}

		SlnkPart := splitVid[3]
		slnkSplit := strings.Split(SlnkPart, "=")
		if len(slnkSplit) < 2 {
			return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vID))
		}
		vIdMem.Path = slnkSplit[1]
		vIdMem.IsFilesetBased = true
		return vIdMem, nil
	}

	if len(splitVid) == 7 {
		/* Volume ID created from 2.5.0 onwards  */
		/* VolID: <storageclass_type>;<type_of_volume>;<cluster_id>;<filesystem_uuid>;<consistency_group>;<fileset_name>;<path> */
		vIdMem.StorageClassType = splitVid[0]
		vIdMem.VolType = splitVid[1]
		vIdMem.ClusterId = splitVid[2]
		vIdMem.FsUUID = splitVid[3]
		vIdMem.ConsistencyGroup = splitVid[4]
		vIdMem.FsetName = splitVid[5]
		if vIdMem.StorageClassType == STORAGECLASS_CLASSIC {
			if vIdMem.VolType == FILE_DIRECTORYBASED_VOLUME {
				vIdMem.IsFilesetBased = false
			} else {
				vIdMem.IsFilesetBased = true
			}
		} else {
			vIdMem.IsFilesetBased = true
		}
		vIdMem.Path = splitVid[6]
		return vIdMem, nil

	}

	return scaleVolId{}, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vID))
}
