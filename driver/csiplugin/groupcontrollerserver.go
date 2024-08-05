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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/klog/v2"
)

const ()

type ScaleGroupControllerServer struct {
	Driver *ScaleDriver
}

// GroupControllerGetCapabilities implements the default GRPC callout.
func (gs *ScaleGroupControllerServer) GroupControllerGetCapabilities(ctx context.Context, req *csi.GroupControllerGetCapabilitiesRequest) (*csi.GroupControllerGetCapabilitiesResponse, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] GroupControllerGetCapabilities called with req: %#v", loggerId, req)
	return &csi.GroupControllerGetCapabilitiesResponse{
		Capabilities: gs.Driver.gcscap,
	}, nil
}

// CreateVolumeGroupSnapshot Create VolumeGroup Snapshot
func (gs *ScaleGroupControllerServer) CreateVolumeGroupSnapshot(ctx context.Context, req *csi.CreateVolumeGroupSnapshotRequest) (*csi.CreateVolumeGroupSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] CreateVolumeGroupSnapshot - create CreateVolumeGroupSnapshot req: %v", loggerId, req)

	// req.SourceVolumeIds: [1;1;16603246530329299476;F0070B0A:6683CB02;0f7c070e-6183-462f-9573-38e7ae124e2a-ibm-spectrum-scale-csi-driver;pvc-1cae06c9-a419-43c4-a9d0-32673e50eeb3;/ibm/fs1/0f7c070e-6183-462f-9573-38e7ae124e2a-ibm-spectrum-scale-csi-driver/pvc-1cae06c9-a419-43c4-a9d0-32673e50eeb3]
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "CreateVolumeGroupSnapshot - Request cannot be empty")
	}

	volIDs := req.GetSourceVolumeIds()
	if len(volIDs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CreateVolumeGroupSnapshot - Source Volume IDs is a required field")
	}
	var volIDMemberStr []string
	for _, volID := range volIDs {

		volumeIDMember, err := volIDGroupParse(volID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateVolumeGroupSnapshot - Error in parsing source Volume ID %v: %v", volID, err))
		}
		volIDMemberStr = append(volIDMemberStr, volumeIDMember)
	}
	klog.Infof("[%s] CreateVolumeGroupSnapshot - SourceVolumeParsed: %v", loggerId, volIDMemberStr)
	if !volGroupMemberValidation(volIDMemberStr) {
		return nil, status.Error(codes.InvalidArgument, "CreateVolumeGroupSnapshot - Source Volume IDs must belong to same consistency group")
	}
	var Snapshots []*csi.Snapshot
	//for _, volID := range volIDs {
	volID := volIDs[0]
	scaleVolId, err := getVolIDMembers(volID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateVolumeGroupSnapshot - Error in source Volume ID %v: %v", volID, err))
	}
	klog.Infof("[%s] CreateVolumeGroupSnapshot - volIDs: %v", loggerId, volIDs)
	klog.Infof("[%s] CreateVolumeGroupSnapshot - scaleVolId: %v", loggerId, scaleVolId)
	snapshot, err := gs.commonSnapshotFunction(ctx, scaleVolId, volID, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateVolumeGroupSnapshot - Error in snapshot create %v: %v", volID, err))
	}
	klog.Infof("[%s] CreateVolumeGroupSnapshot - snapshot response : %v", loggerId, snapshot)
	//Snapshots = append(Snapshots, snapshot)
	for _, sourceVolID := range volIDs {
		snapshot.SourceVolumeId = sourceVolID
		Snapshots = append(Snapshots, snapshot)
	}
	//}
	return &csi.CreateVolumeGroupSnapshotResponse{
		GroupSnapshot: &csi.VolumeGroupSnapshot{
			GroupSnapshotId: req.GetName(),
			Snapshots:       Snapshots,
			ReadyToUse:      true,
			CreationTime:    timestamppb.Now(),
		},
	}, nil
}

// GetVolumeGroupSnapshot Get VolumeGroup Snapshot
func (gs *ScaleGroupControllerServer) GetVolumeGroupSnapshot(ctx context.Context, req *csi.GetVolumeGroupSnapshotRequest) (*csi.GetVolumeGroupSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] GetVolumeGroupSnapshot -  GetVolumeGroupSnapshot req: %v", loggerId, req)

	return &csi.GetVolumeGroupSnapshotResponse{
		GroupSnapshot: &csi.VolumeGroupSnapshot{},
	}, nil
}

// DeleteVolumeGroupSnapshot Delete VolumeGroup Snapshot
func (gs *ScaleGroupControllerServer) DeleteVolumeGroupSnapshot(ctx context.Context, req *csi.DeleteVolumeGroupSnapshotRequest) (*csi.DeleteVolumeGroupSnapshotResponse, error) { //nolint:gocyclo,funlen
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] DeleteVolumeGroupSnapshot -  DeleteVolumeGroupSnapshot req: %v", loggerId, req)

	return &csi.DeleteVolumeGroupSnapshotResponse{}, nil
}

func (gs *ScaleGroupControllerServer) commonSnapshotFunction(ctx context.Context, scaleVolId scaleVolId, volID string, req *csi.CreateVolumeGroupSnapshotRequest) (*csi.Snapshot, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] CreateVolumeGroupSnapshot - commonSnapshotFunction scaleVolId: %v  volID: %v", loggerId, scaleVolId, volID)
	if scaleVolId.StorageClassType != STORAGECLASS_ADVANCED {

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateVolumeGroupSnapshot - volume [%s] - Volume snapshot can only be created when source volume is version 2 fileset", volID))

	}
	conn, err := gs.getConnFromClusterID(ctx, scaleVolId.ClusterId)
	if err != nil {
		return nil, err
	}
	assembledScaleversion, err := gs.assembledScaleVersion(ctx, conn)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("the  IBM Storage Scale version check for permissions failed with error %s", err))
	}
	/* Check if IBM Storage Scale supports Snapshot */
	chkSnapshotErr := checkSnapshotSupport(assembledScaleversion)
	if chkSnapshotErr != nil {
		return nil, chkSnapshotErr
	}

	primaryConn, isprimaryConnPresent := gs.Driver.connmap["primary"]
	if !isprimaryConnPresent {
		klog.Errorf("[%s] CreateSnapshot - unable to get connector for primary cluster", loggerId)
		return nil, status.Error(codes.Internal, "CreateSnapshot - unable to find primary cluster details in custom resource")
	}

	filesystemName, err := primaryConn.GetFilesystemName(ctx, scaleVolId.FsUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get filesystem Name for Filesystem Uid [%v] and clusterId [%v]. Error [%v]", scaleVolId.FsUUID, scaleVolId.ClusterId, err))
	}

	mountInfo, err := primaryConn.GetFilesystemMountDetails(ctx, filesystemName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - unable to get mount info for FS [%v] in primary cluster", filesystemName))
	}

	filesetResp := connectors.Fileset_v2{}
	filesystemName = getRemoteFsName(mountInfo.RemoteDeviceName)
	if scaleVolId.FsetName != "" {
		filesetResp, err = conn.GetFileSetResponseFromName(ctx, filesystemName, scaleVolId.FsetName)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get Fileset response for Fileset [%v] FS [%v] ClusterId [%v]", scaleVolId.FsetName, filesystemName, scaleVolId.ClusterId))
		}
	} else {
		filesetResp, err = conn.GetFileSetResponseFromId(ctx, filesystemName, scaleVolId.FsetId)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - Unable to get Fileset response for Fileset Id [%v] FS [%v] ClusterId [%v]", scaleVolId.FsetId, filesystemName, scaleVolId.ClusterId))
		}
	}

	filesetName := filesetResp.FilesetName
	relPath := ""
	if scaleVolId.StorageClassType == STORAGECLASS_ADVANCED {
		klog.V(4).Infof("[%s] CreateSnapshot - creating snapshot for advanced storageClass", loggerId)
		relPath = strings.Replace(scaleVolId.Path, mountInfo.MountPoint, "", 1)
	}
	relPath = strings.Trim(relPath, "!/")

	/* Confirm it is same fileset which was created for this PV */
	pvName := filepath.Base(relPath)
	if pvName != filesetName {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot - PV name from path [%v] does not match with filesetName [%v].", pvName, filesetName))
	}

	filesetName = scaleVolId.ConsistencyGroup

	snapName := req.GetName()
	snapWindowInt := 0

	snapParams := req.GetParameters()
	snapWindow, snapWindowSpecified := snapParams[connectors.UserSpecifiedSnapWindow]
	if !snapWindowSpecified {
		// use default snapshot window for consistency group
		snapWindow = defaultSnapWindow
		klog.Infof("[%s] SnapWindow not specified. Using default snapWindow: [%s] for for fileset[%s:%s]", loggerId, snapWindow, filesetResp.FilesetName, filesystemName)
	}
	snapWindowInt, err = strconv.Atoi(snapWindow)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateSnapshot [%s] - invalid snapWindow value: [%v]", snapName, snapWindow))
	}

	// Additional check for RDR fileset in secondary mode
	AFMMode, err := gs.GetAFMMode(ctx, filesystemName, filesetName, conn)
	if err != nil {
		return nil, err
	}
	if AFMMode == connectors.AFMModeSecondary {
		klog.Errorf("[%s] snapshot is not supported for AFM Secondary mode of ConsistencyGroup fileset [%v]", loggerId, filesetName)
		return nil, status.Error(codes.Internal, fmt.Sprintf("snapshot is not supported for AFM Secondary mode of ConsistencyGroup fileset [%v]", filesetName))
	}

	snapExist, err := conn.CheckIfSnapshotExist(ctx, filesystemName, filesetName, snapName)
	if err != nil {
		klog.Errorf("[%s] CreateSnapshot [%s] - Unable to get the snapshot details. Error [%v]", loggerId, snapName, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to get the snapshot details for [%s]. Error [%v]", snapName, err))
	}

	if !snapExist {
		/* For new storageClass check last snapshot creation time, if time passed is less than
		 * snapWindow then return existing snapshot */
		createNewSnap := true

		cgSnapName, err := gs.CheckNewSnapRequired(ctx, conn, filesystemName, filesetName, snapWindowInt)
		if err != nil {
			klog.Errorf("[%s] CreateSnapshot [%s] - unable to check if snapshot is required for new storageClass for fileset [%s:%s]. Error: [%v]", loggerId, snapName, filesystemName, filesetName, err)
			return nil, err
		}
		if cgSnapName != "" {
			usable, err := gs.isExistingSnapUseableForVol(ctx, conn, filesystemName, filesetName, filesetResp.FilesetName, cgSnapName)
			if !usable {
				return nil, err
			}
			createNewSnap = false
			snapName = cgSnapName
		} else {
			klog.Infof("[%s] CreateSnapshot - creating new snapshot for consistency group for fileset: [%s:%s]", loggerId, filesystemName, filesetName)
		}

		if createNewSnap {
			snapshotList, err := conn.ListFilesetSnapshots(ctx, filesystemName, filesetName)
			if err != nil {
				klog.Errorf("[%s] CreateSnapshot [%s] - unable to list snapshots for fileset [%s:%s]. Error: [%v]", loggerId, snapName, filesystemName, filesetName, err)
				return nil, status.Error(codes.Internal, fmt.Sprintf("unable to list snapshots for fileset [%s:%s]. Error: [%v]", filesystemName, filesetName, err))
			}

			if len(snapshotList) >= 256 {
				klog.Errorf("[%s] CreateSnapshot [%s] - max limit of snapshots reached for fileset [%s:%s]. No more snapshots can be created for this fileset.", loggerId, snapName, filesystemName, filesetName)
				return nil, status.Error(codes.OutOfRange, fmt.Sprintf("max limit of snapshots reached for fileset [%s:%s]. No more snapshots can be created for this fileset.", filesystemName, filesetName))
			}
			klog.Infof("[%s] commonSnapshotFunction - creating new snapshot CreateSnapshot filesystemName: %s, filesetName:%s ,snapName: %s", loggerId, filesystemName, filesetName, snapName)
			snaperr := conn.CreateSnapshot(ctx, filesystemName, filesetName, snapName)
			if snaperr != nil {
				klog.Errorf("[%s] Snapshot [%s] - Unable to create snapshot. Error [%v]", loggerId, snapName, snaperr)
				return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create snapshot [%s]. Error [%v]", snapName, snaperr))
			}
		}
	}

	snapID := ""
	// storageclass_type;volumeType;clusterId;FSUUID;consistency_group;filesetName;snapshotName;metaSnapshotName
	snapID = fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s;%s", scaleVolId.StorageClassType, scaleVolId.VolType, scaleVolId.ClusterId, scaleVolId.FsUUID, filesetName, filesetResp.FilesetName, snapName, req.GetName())

	timestamp, err := gs.getSnapshotCreateTimestamp(ctx, conn, filesystemName, filesetName, snapName)
	if err != nil {
		klog.Errorf("[%s] Error getting create timestamp for snapshot %s:%s:%s", loggerId, filesystemName, filesetName, snapName)
		return nil, err
	}

	restoreSize, err := gs.getSnapRestoreSize(ctx, conn, filesystemName, filesetResp.FilesetName)
	if err != nil {
		klog.Errorf("[%s] Error getting the snapshot restore size for snapshot %s:%s:%s", loggerId, filesystemName, filesetResp.FilesetName, snapName)
		return nil, err
	}

	err = gs.MakeSnapMetadataDir(ctx, conn, filesystemName, filesetResp.FilesetName, filesetName, snapName, req.GetName())
	if err != nil {
		klog.Errorf("[%s] Error in creating directory for storing metadata information for advanced storageClass. Error: [%v]", loggerId, err)
		return nil, err
	}

	return &csi.Snapshot{
		SnapshotId:     snapID,
		SourceVolumeId: volID,
		ReadyToUse:     true,
		CreationTime:   &timestamp,
		SizeBytes:      restoreSize,
	}, nil
}

func volIDGroupParse(vID string) (string, error) {
	splitVid := strings.Split(vID, ";")
	//var vIdMem scaleVolId
	toValidateSameCGVolMember := ""

	if len(splitVid) == 7 {
		/* Volume ID created from CSI 2.5.0 onwards  */
		/* VolID: <storageclass_type>;<type_of_volume>;<cluster_id>;<filesystem_uuid>;<consistency_group>;<fileset_name>;<path> */

		toValidateSameCGVolMember = splitVid[0] + splitVid[1] + splitVid[2] + splitVid[3] + splitVid[4]
		return toValidateSameCGVolMember, nil

	}

	return toValidateSameCGVolMember, status.Error(codes.Internal, fmt.Sprintf("Invalid Volume Id : [%v]", vID))
}

func volGroupMemberValidation(volIDMemberStr []string) bool {

	for _, v := range volIDMemberStr {
		if v != volIDMemberStr[0] {
			return false
		}
	}
	return true

}

func (gs *ScaleGroupControllerServer) getConnFromClusterID(ctx context.Context, cid string) (connectors.SpectrumScaleConnector, error) {
	loggerId := utils.GetLoggerId(ctx)
	connector, isConnPresent := gs.Driver.connmap[cid]
	if isConnPresent {
		return connector, nil
	}
	klog.Errorf("[%s] unable to get connector for cluster ID %v", loggerId, cid)
	return nil, status.Error(codes.Internal, fmt.Sprintf("unable to find cluster [%v] details in custom resource", cid))
}

func (gs *ScaleGroupControllerServer) assembledScaleVersion(ctx context.Context, conn connectors.SpectrumScaleConnector) (string, error) {
	assembledScaleVer := ""
	scaleVersion, err := conn.GetScaleVersion(ctx)
	if err != nil {
		return assembledScaleVer, err
	}
	/* Assuming IBM Storage Scale version is in a format like 5.0.0-0_170818.165000 */
	// "serverVersion" : "5.1.1.1-developer build",
	splitScaleVer := strings.Split(scaleVersion, ".")
	if len(splitScaleVer) < 3 {
		return assembledScaleVer, status.Error(codes.Internal, fmt.Sprintf("invalid IBM Storage Scale version - %s", scaleVersion))
	}
	var splitMinorVer []string
	if len(splitScaleVer) == 4 {
		//dev build e.g. "5.1.5.0-developer build"
		splitMinorVer = strings.Split(splitScaleVer[3], "-")
		assembledScaleVer = splitScaleVer[0] + splitScaleVer[1] + splitScaleVer[2] + splitMinorVer[0]
	} else {
		//GA build e.g. "5.1.5-0"
		splitMinorVer = strings.Split(splitScaleVer[2], "-")
		assembledScaleVer = splitScaleVer[0] + splitScaleVer[1] + splitMinorVer[0] + splitMinorVer[1][0:1]
	}
	return assembledScaleVer, nil
}

func checkSnapshotSupport(assembledScaleversion string) error {
	/* Verify IBM Storage Scale Version is not below 5.1.1-0 */
	versionCheck := checkMinScaleVersionValid(assembledScaleversion, "5110")
	if !versionCheck {
		return status.Error(codes.FailedPrecondition, "the minimum required IBM Storage Scale version for snapshot support with CSI is 5.1.1-0")
	}
	return nil
}

func (gs *ScaleGroupControllerServer) getPrimaryFSMountPoint(ctx context.Context) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	klog.Infof("[%s] getPrimaryFSMountPoint", loggerId)

	primaryConn := gs.Driver.connmap["primary"]
	primaryFS := gs.Driver.primary.GetPrimaryFs()
	fsMountInfo, err := primaryConn.GetFilesystemMountDetails(ctx, primaryFS)
	if err != nil {
		klog.Errorf("[%s] Failed to get details of primary filesystem %s:Error: %v", loggerId, primaryFS, err)
		return "", status.Error(codes.NotFound, fmt.Sprintf("Failed to get details of primary filesystem %s. Error: %v", primaryFS, err))

	}
	return fsMountInfo.MountPoint, nil
}

// GetAFMMode returns the AFM mode of the fileset and also the error
// if there is any (including the fileset not found error) while getting
// the fileset info
func (gs *ScaleGroupControllerServer) GetAFMMode(ctx context.Context, filesystemName string, filesetName string, conn connectors.SpectrumScaleConnector) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	filesetDetails, err := conn.ListFileset(ctx, filesystemName, filesetName)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("failed to get fileset info, filesystem: [%v], fileset: [%v], error: [%v]", filesystemName, filesetName, err))
	}

	klog.V(4).Infof("[%s] AFM mode of the fileset [%v] is [%v]", loggerId, filesetName, filesetDetails.AFM.AFMMode)
	return filesetDetails.AFM.AFMMode, nil
}

func (gs *ScaleGroupControllerServer) CheckNewSnapRequired(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string, snapWindow int) (string, error) {
	loggerId := utils.GetLoggerId(ctx)
	latestSnapList, err := conn.GetLatestFilesetSnapshots(ctx, filesystemName, filesetName)
	if err != nil {
		klog.Errorf("[%s] CheckNewSnapRequired - getting latest snapshot list failed for fileset: [%s:%s]. Error: [%v]", loggerId, filesystemName, filesetName, err)
		return "", err
	}

	if len(latestSnapList) == 0 {
		// No snapshot exists, so create new one
		return "", nil
	}

	timestamp, err := gs.getSnapshotCreateTimestamp(ctx, conn, filesystemName, filesetName, latestSnapList[0].SnapshotName)
	if err != nil {
		klog.Errorf("[%s] Error getting create timestamp for snapshot %s:%s:%s", loggerId, filesystemName, filesetName, latestSnapList[0].SnapshotName)
		return "", err
	}

	var timestampSecs int64 = timestamp.GetSeconds()
	lastSnapTime := time.Unix(timestampSecs, 0)
	passedTime := time.Since(lastSnapTime).Seconds()
	klog.Infof("[%s] Fileset [%s:%s], last snapshot time: [%v], current time: [%v], passed time: %v seconds, snapWindow: %v minutes", loggerId, filesystemName, filesetName, lastSnapTime, time.Now(), int64(passedTime), snapWindow)

	snapWindowSeconds := snapWindow * 60

	if passedTime < float64(snapWindowSeconds) {
		// we don't need to take new snapshot
		klog.Infof("[%s] CheckNewSnapRequired - for fileset [%s:%s], using existing snapshot [%s]", loggerId, filesystemName, filesetName, latestSnapList[0].SnapshotName)
		return latestSnapList[0].SnapshotName, nil
	}

	klog.Infof("[%s] CheckNewSnapRequired - for fileset [%s:%s] we need to create new snapshot", loggerId, filesystemName, filesetName)
	return "", nil
}

func (gs *ScaleGroupControllerServer) MakeSnapMetadataDir(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string, indepFileset string, cgSnapName string, metaSnapName string) error {
	loggerId := utils.GetLoggerId(ctx)
	path := fmt.Sprintf("%s/%s/%s", indepFileset, cgSnapName, metaSnapName)
	klog.Infof("[%s] MakeSnapMetadataDir - creating directory [%s] for fileset: [%s:%s]", loggerId, path, filesystemName, filesetName)
	err := conn.MakeDirectory(ctx, filesystemName, path, "0", "0")
	if err != nil {
		// Directory creation failed
		klog.Errorf("[%s] Volume:[%v] - unable to create directory [%v] in filesystem [%v]. Error : %v", loggerId, filesetName, path, filesystemName, err)
		return fmt.Errorf("unable to create directory [%v] in filesystem [%v]. Error : %v", path, filesystemName, err)
	}
	return nil
}

func (gs *ScaleGroupControllerServer) isExistingSnapUseableForVol(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, consistencyGroup string, filesetName string, cgSnapName string) (bool, error) {
	pathDir := fmt.Sprintf("%s/.snapshots/%s/%s", consistencyGroup, cgSnapName, filesetName)
	_, err := conn.StatDirectory(ctx, filesystemName, pathDir)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0264C") ||
			strings.Contains(err.Error(), "does not exist") { // directory does not exist
			return false, status.Error(codes.Internal, fmt.Sprintf("snapshot for volume [%v] in filesystem [%v] is not taken. Wait till current snapWindow expires.", filesetName, filesystemName))
		} else {
			return false, err
		}
	}
	return true, nil
}

func (gs *ScaleGroupControllerServer) getSnapshotCreateTimestamp(ctx context.Context, conn connectors.SpectrumScaleConnector, fs string, fset string, snap string) (timestamppb.Timestamp, error) {
	var timestamp timestamppb.Timestamp

	createTS, err := conn.GetSnapshotCreateTimestamp(ctx, fs, fset, snap)
	if err != nil {
		klog.Errorf("[%s]snapshot [%s] - Unable to get snapshot create timestamp", utils.GetLoggerId(ctx), snap)
		return timestamp, err
	}

	timezoneOffset, err := conn.GetTimeZoneOffset(ctx)
	if err != nil {
		klog.Errorf("[%s] snapshot [%s] - Unable to get cluster timezone", utils.GetLoggerId(ctx), snap)
		return timestamp, err
	}

	// for GMT, REST API returns Z instead of 00:00
	if timezoneOffset == "Z" {
		timezoneOffset = "+00:00"
	}

	// Rest API returns create timestamp in the format 2006-01-02 15:04:05,000
	// irrespective of the cluster timezone. We replace the last part of this date
	// with the timezone offset returned by cluster config REST API and then parse
	// the timestamp with correct zone info
	const longForm = "2006-01-02 15:04:05-07:00"
	//nolint::staticcheck

	createTSTZ := strings.Replace(createTS, ",000", timezoneOffset, 1)
	t, err := time.Parse(longForm, createTSTZ)
	if err != nil {
		klog.Errorf("[%s] snapshot - for fileset [%s:%s] error in parsing timestamp: [%v]. Error: [%v]", utils.GetLoggerId(ctx), fs, fset, createTS, err)
		return timestamp, err
	}
	timestamp.Seconds = t.Unix()
	timestamp.Nanos = 0

	klog.Infof("[%s] getSnapshotCreateTimestamp: for fileset [%s:%s] snapshot creation timestamp: [%v]", utils.GetLoggerId(ctx), fs, fset, createTSTZ)
	return timestamp, nil
}

func (gs *ScaleGroupControllerServer) getSnapRestoreSize(ctx context.Context, conn connectors.SpectrumScaleConnector, filesystemName string, filesetName string) (int64, error) {
	quotaResp, err := conn.GetFilesetQuotaDetails(ctx, filesystemName, filesetName)

	if err != nil {
		return 0, err
	}

	if quotaResp.BlockLimit < 0 {
		klog.Errorf("[%s] getSnapRestoreSize: Invalid block limit [%v] for fileset [%s:%s] found", utils.GetLoggerId(ctx), quotaResp.BlockLimit, filesystemName, filesetName)
		return 0, status.Error(codes.Internal, fmt.Sprintf("invalid block limit [%v] for fileset [%s:%s] found", quotaResp.BlockLimit, filesystemName, filesetName))
	}

	// REST API returns block limit in kb, convert it to bytes and return
	return int64(quotaResp.BlockLimit * 1024), nil
}
