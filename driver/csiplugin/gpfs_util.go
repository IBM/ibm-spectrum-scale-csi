/**
 * Copyright 2019, 2024 IBM Corp.
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
	"os"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	dependentFileset         = "dependent"
	independentFileset       = "independent"
	scversion1               = "1"
	scversion2               = "2"
	sharedPermissions        = "777"
	defaultVolNamePrefix     = "pvc"
	VolNamePrefixEnvKey      = "VOLUME_NAME_PREFIX"
	existingVolumeAllowedVal = "yes"
	AFMCacheSharedPermission = "0777"
)

// AFM caching constants
const (
	cacheVolume = "cache"

	// AFM cache modes
	afmModeRO = "ro" // Read-Only
	afmModeIW = "iw" // Independent-Writer
	afmModeSW = "sw" // Single-Writer
	afmModeLU = "lu" // Local-Update

	// User input cache modes
	inputModeRO = "readonly"
	inputModeIW = "parallel"
	inputModeSW = "exclusive"
	inputModeLU = "detached"
)

// A map for mapping the user input mode to actual AFM mode
var inputToAFMMode = map[string]string{
	inputModeRO: afmModeRO,
	inputModeIW: afmModeIW,
	inputModeSW: afmModeSW,
	inputModeLU: afmModeLU,
}

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
	Shared             bool                              `json:"shared"`
	VolumeType         string                            `json:"volumeType"`
	CacheMode          string                            `json:"cacheMode"`
	VolNamePrefix      string                            `json:"volNamePrefix"`
	ExistingVolume     string                            `json:"existingVolume"`
	IsStaticPVBased    bool                              `json:"isStaticPV"`
	PVCName            string                            `json:"pvcName"`
	Namespace          string                            `json:"namespace"`
}

type scaleVolId struct {
	ClusterId        string
	FsUUID           string
	FsName           string
	FsetId           string
	FsetName         string
	DirPath          string
	Path             string
	IsFilesetBased   bool
	StorageClassType string
	ConsistencyGroup string
	VolType          string
	IsStaticPVBased  bool
}

type scaleSnapId struct {
	ClusterId        string
	FsUUID           string
	FsetName         string
	SnapName         string
	MetaSnapName     string
	Path             string
	FsName           string
	StorageClassType string
	ConsistencyGroup string
	VolType          string
	IsStaticPVBased  bool
}

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

func getScaleVolumeOptions(ctx context.Context, volOptions map[string]string) (*scaleVolume, error) { //nolint:gocyclo,funlen
	//var err error
	scaleVol := &scaleVolume{}
	loggerId := utils.GetLoggerId(ctx)

	volBckFs, fsSpecified := volOptions[connectors.UserSpecifiedVolBackendFs]
	volDirPath, volDirPathSpecified := volOptions[connectors.UserSpecifiedVolDirPath]
	clusterID, clusterIDSpecified := volOptions[connectors.UserSpecifiedClusterId]
	uid, uidSpecified := volOptions[connectors.UserSpecifiedUid]
	gid, gidSpecified := volOptions[connectors.UserSpecifiedGid]
	fsetType, fsetTypeSpecified := volOptions[connectors.UserSpecifiedFilesetType]
	inodeLim, inodeLimSpecified := volOptions[connectors.UserSpecifiedInodeLimit]
	parentFileset, isparentFilesetSpecified := volOptions[connectors.UserSpecifiedParentFset]
	nodeClass, isNodeClassSpecified := volOptions[connectors.UserSpecifiedNodeClass]
	permissions, isPermissionsSpecified := volOptions[connectors.UserSpecifiedPermissions]
	storageClassType, isSCTypeSpecified := volOptions[connectors.UserSpecifiedStorageClassType]
	compression, isCompressionSpecified := volOptions[connectors.UserSpecifiedCompression]
	tier, isTierSpecified := volOptions[connectors.UserSpecifiedTier]
	cg, isCGSpecified := volOptions[connectors.UserSpecifiedConsistencyGroup]
	shared, isSharedSpecified := volOptions[connectors.UserSpecifiedShared]
	volNamePrefix, isVolNamePrefixSpecified := volOptions[connectors.UserSpecifiedVolNamePrefix]

	volumeType, volumeTypeSpecified := volOptions[connectors.UserSpecifiedVolumeType]
	cacheMode, cacheModeSpecified := volOptions[connectors.UserSpecifiedCacheMode]

	// for static pv
	scaleVol.IsStaticPVBased = false
	existingVolume, existingVolumeSpecified := volOptions[connectors.UserSpecifiedExistingVolume]
	if existingVolumeSpecified && existingVolume == existingVolumeAllowedVal {
		scaleVol.IsStaticPVBased = true
	}

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
	scaleVol.PVCName = ""

	if isSCTypeSpecified && storageClassType == "" {
		isSCTypeSpecified = false
	}
	if volOptions["csi.storage.k8s.io/pvc/name"] != "" {
		scaleVol.PVCName = volOptions["csi.storage.k8s.io/pvc/name"]
	}
	if volOptions["csi.storage.k8s.io/pvc/namespace"] != "" {
		scaleVol.Namespace = volOptions["csi.storage.k8s.io/pvc/namespace"]
	}

	isSCAdvanced := false
	if isSCTypeSpecified {
		if storageClassType != scversion1 && storageClassType != scversion2 {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameter \"version\" can have values only "+
				"\""+scversion1+"\" or \""+scversion2+"\"")
		}
		if storageClassType == scversion2 && scaleVol.IsStaticPVBased {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameter \"existingVolume\" is not allowed for version "+""+scversion2+"\"")
		} else if storageClassType == scversion2 {
			isSCAdvanced = true
			scaleVol.StorageClassType = STORAGECLASS_ADVANCED
		}

		if storageClassType == scversion1 {
			scaleVol.StorageClassType = STORAGECLASS_CLASSIC
		}
	} else {
		scaleVol.StorageClassType = STORAGECLASS_CLASSIC
	}

	if fsSpecified && volBckFs == "" {
		fsSpecified = false
	}

	if fsSpecified {
		scaleVol.VolBackendFs = volBckFs
	} else {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "volBackendFs must be specified in storageClass")
	}

	if fsetTypeSpecified && fsetType == "" {
		fsetTypeSpecified = false
	}

	if isCGSpecified && cg == "" {
		isCGSpecified = false
	}

	if volDirPathSpecified && volDirPath == "" {
		volDirPathSpecified = false
	}
	// VolNamePrefix
	if isVolNamePrefixSpecified {
		scaleVol.VolNamePrefix = volNamePrefix
	} else if volNamePrefixEnvVal, valFound := os.LookupEnv(VolNamePrefixEnvKey); valFound {
		scaleVol.VolNamePrefix = volNamePrefixEnvVal
	} else {
		scaleVol.VolNamePrefix = defaultVolNamePrefix
	}
	klog.Infof("[%s] getScaleVolumeOptions:  VolNamePrefix assigned: %s", loggerId, scaleVol.VolNamePrefix)

	isUserInputFsetType := fsetTypeSpecified
	if !fsetTypeSpecified && !volDirPathSpecified && !isSCAdvanced {
		fsetTypeSpecified = true
		fsetType = independentFileset
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

	/* Check if either fileset based or LW volume. */
	if volDirPathSpecified {
		if (fsetTypeSpecified && (fsetType == dependentFileset || fsetType == independentFileset)) || isSCAdvanced {
			scaleVol.IsFilesetBased = true
		} else {
			if inodeLimSpecified {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "inodeLimit and volDirBasePath must not be specified together in storageClass")
			}
			if isparentFilesetSpecified {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "parentFileset and volDirBasePath must not be specified together in storageClass")
			}
			scaleVol.IsFilesetBased = false
		}
		scaleVol.VolDirBasePath = volDirPath
	}

	if fsetTypeSpecified {
		if fsetType == dependentFileset {
			if inodeLimSpecified {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "inodeLimit and filesetType=dependent must not be specified together in storageClass")
			}
		} else if fsetType == independentFileset {
			if isparentFilesetSpecified {
				return &scaleVolume{}, status.Error(codes.InvalidArgument, "parentFileset and filesetType=independent(Default) must not be specified together in storageClass")
			}
		} else {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "Invalid value specified for filesetType in storageClass")
		}
	}

	if fsetTypeSpecified && inodeLimSpecified {
		inodelimit, err := strconv.Atoi(inodeLim)
		if err != nil {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "Invalid value specified for inodeLimit in storageClass")
		}
		if inodelimit < 1024 {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "inodeLimit specified in storageClass must be equal to or greater than 1024")
		}
	}

	if isSCAdvanced && fsetTypeSpecified {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "filesetType and version="+scversion2+" must not be specified together in storageClass")
	}
	if isSCAdvanced && isparentFilesetSpecified {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "parentFileset and version="+scversion2+" must not be specified together in storageClass")
	}
	if isSCAdvanced && volDirPathSpecified {
		//return &scaleVolume{}, status.Error(codes.InvalidArgument, "volDirBasePath and version="+scversion2+" must not be specified together in storageClass")
		scaleVol.VolDirBasePath = volDirPath
	}

	if fsetTypeSpecified || isSCAdvanced {
		scaleVol.IsFilesetBased = true
	}

	if isCompressionSpecified && (compression == "" || compression == "false") {
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

	if scaleVol.IsStaticPVBased {
		if uidSpecified || gidSpecified || isSharedSpecified || inodeLimSpecified || isPermissionsSpecified || isNodeClassSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"uid\" , \"gid\" , \"inodeLimit\" , \"shared\" , \"nodeClass\" and \"permissions\" are not allowed in storageClass for static volumes i.e. with \"existingVolume\"")
		}
		if volDirPathSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "volDirBasePath is not allowed in storageClass for static volumes i.e. with \"existingVolume\"")
		}
		if isCompressionSpecified || isTierSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"compression\" and \"tier\" are not supported in storageClass for static volumes i.e. with \"existingVolume\"")
		}
	}

	/* Get UID/GID */
	if uidSpecified {
		scaleVol.VolUid = uid
	}

	if gidSpecified {
		scaleVol.VolGid = gid
	}

	if isSharedSpecified {
		//ignore case of passed "shared" parameter
		icShared := strings.ToLower(shared)
		if !(icShared == "true" || icShared == "false") {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "invalid value specified for parameter shared")
		}
		if icShared == "false" {
			isSharedSpecified = false
			scaleVol.Shared = false
		} else {
			scaleVol.Shared = true
		}
	}

	if isSharedSpecified && isPermissionsSpecified {
		return &scaleVolume{}, status.Error(codes.InvalidArgument, "shared=true and permissions must not be specified together in storageClass")
	}
	if isSharedSpecified {
		scaleVol.VolPermissions = sharedPermissions
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
		if fsetTypeSpecified {
			scaleVol.FilesetType = fsetType
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

	if isCGSpecified {
		scaleVol.ConsistencyGroup = cg
	} else {
		cgPrefix := utils.GetEnv("CSI_CG_PREFIX", notFound)
		if cgPrefix == notFound {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "Failed to extract the consistencyGroup prefix")
		}
		scaleVol.ConsistencyGroup = fmt.Sprintf("%s-%s", cgPrefix, volOptions["csi.storage.k8s.io/pvc/namespace"])
	}

	if isCompressionSpecified {
		// Default compression will be Z if set but not specified
		if strings.ToLower(compression) == "true" {
			klog.V(6).Infof("[%s] gpfs_util compression was set to true. Defaulting to Z", loggerId)
			compression = "z"
		}

		if !IsValidCompressionAlgorithm(compression) {
			klog.V(4).Infof("[%s] gpfs_util invalid compression algorithm specified: %s",
				loggerId, compression)
			return &scaleVolume{}, status.Errorf(codes.InvalidArgument,
				"invalid compression algorithm specified: %s", compression)
		}
		scaleVol.Compression = compression
		klog.V(4).Infof("[%s] gpfs_util compression was set to %s", loggerId, compression)
	}

	if isTierSpecified && tier != "" {
		scaleVol.Tier = tier
		klog.V(6).Infof("[%s] gpfs_util tier was set: %s", loggerId, tier)
	}

	if volumeTypeSpecified {
		if isSCTypeSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"version\" and \"volumeType\" in storage class are mutually exclusive")
		}

		if isUserInputFsetType {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"filesetType\" and \"volumeType\" in storage class are mutually exclusive")
		}

		if isparentFilesetSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"parentFileset\" and \"volumeType\" in storage class are mutually exclusive")
		}

		/*if volDirPathSpecified {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, "The parameters \"volDirBasePath\" and \"volumeType\" in storage class are mutually exclusive")
		}*/

		volumeType = strings.ToLower(volumeType)
		if volumeType == cacheVolume {
			scaleVol.StorageClassType = STORAGECLASS_CACHE
			scaleVol.VolumeType = cacheVolume
			if volDirPathSpecified {
				scaleVol.VolDirBasePath = volDirPath
				scaleVol.IsFilesetBased = true
			}
		} else {
			return &scaleVolume{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid volumeType is specified: %s, only allowed value is: %s", volumeType, cacheVolume))
		}
	}

	if cacheModeSpecified && scaleVol.VolumeType != cacheVolume {
		return &scaleVolume{}, status.Errorf(codes.InvalidArgument,
			"The storage class parameter cacheMode can only be specified with volumeType=\"cache\"")
	}

	if cacheModeSpecified {
		cacheMode = strings.ToLower(cacheMode)
		switch cacheMode {
		case inputModeIW, inputModeRO, inputModeSW, inputModeLU:
			scaleVol.CacheMode = inputToAFMMode[cacheMode]
		default:
			allowedCacheModes := inputModeRO + ", " + inputModeIW + ", " + inputModeSW + " or " + inputModeLU
			return &scaleVolume{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid cache mode is specified: %s, allowed cache modes are: %s", cacheMode, allowedCacheModes))
		}
	}

	return scaleVol, nil
}

/*func executeCmd(command string, args []string) ([]byte, error) {
	klog.V(6).Infof("gpfs_util executeCmd")

	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stdOut := stdout.Bytes()
	return stdOut, err
}*/

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
		return 0, fmt.Errorf("invalid number specified %v", inputStr)
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
		return 0, fmt.Errorf("invalid Unit %v supplied with %v", unit, inputStr)
	}

	if retValue > uintMax64 {
		return 0, fmt.Errorf("overflow detected %v", inputStr)
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
			klog.V(4).Infof("getNodeMapping: scale node mapping not found for %s using %s", prefix+kubernetesNodeID, kubernetesNodeID)
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
	klog.V(6).Infof("gpfs_util shortnameInSlice. string: %s, slice: %v", shortname, nodeNames)
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
		/* Volume ID created from CSI 2.5.0 onwards  */
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

func isSubset(subset []string, superset []string) bool {
	checkset := make(map[string]bool)
	for _, element := range subset {
		checkset[element] = true
	}
	for _, value := range superset {
		if checkset[value] {
			delete(checkset, value)
		}
	}
	return len(checkset) == 0 //this implies that set is subset of superset
}
