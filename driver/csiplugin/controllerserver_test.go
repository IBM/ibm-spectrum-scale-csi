/**
 * Copyright 2024 IBM Corp.
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

//go:build linux

package scale

import (
	"testing"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Test scaffold
// ---------------------------------------------------------------------------

// newTestControllerServer builds a minimal ScaleControllerServer whose Driver
// is pre-populated with the supplied connector map.  It mirrors the
// newTestNodeServer helper that lives in nodeserver_test.go.
func newTestControllerServer(connMap map[string]connectors.SpectrumScaleConnector) *ScaleControllerServer {
	d := &ScaleDriver{
		nodeID:  "test-node",
		cscap:   []*csi.ControllerServiceCapability{},
		nscap:   []*csi.NodeServiceCapability{},
		reqmap:  make(map[string]int64),
		connmap: connMap,
	}
	return &ScaleControllerServer{Driver: d}
}

// withCap returns a controller server that advertises the given capability
// types, so that ValidateControllerServiceRequest succeeds for those types.
func withCap(caps ...csi.ControllerServiceCapability_RPC_Type) *ScaleControllerServer {
	cs := newTestControllerServer(nil)
	for _, c := range caps {
		cs.Driver.cscap = append(cs.Driver.cscap, NewControllerServiceCapability(c))
	}
	return cs
}

// ---------------------------------------------------------------------------
// IfSameVolReqInProcess
// ---------------------------------------------------------------------------

// TestIfSameVolReqInProcess_NotInMap verifies that a volume not yet in the
// reqmap returns (false, nil).
func TestIfSameVolReqInProcess_NotInMap(t *testing.T) {
	cs := newTestControllerServer(nil)
	scVol := &scaleVolume{VolName: "vol-1", VolSize: 1024}
	inProc, err := cs.IfSameVolReqInProcess(scVol)
	require.NoError(t, err)
	assert.False(t, inProc, "volume not in map should return false")
}

// TestIfSameVolReqInProcess_SameCapacity verifies that a matching in-flight
// request (same volume name AND same size) returns (true, nil).
func TestIfSameVolReqInProcess_SameCapacity(t *testing.T) {
	cs := newTestControllerServer(nil)
	cs.Driver.reqmap["vol-2"] = 2048
	scVol := &scaleVolume{VolName: "vol-2", VolSize: 2048}
	inProc, err := cs.IfSameVolReqInProcess(scVol)
	require.NoError(t, err)
	assert.True(t, inProc, "same volume with matching capacity must be in-process")
}

// TestIfSameVolReqInProcess_DifferentCapacity verifies that a conflicting
// in-flight request (same name, different size) returns (false, Internal error).
func TestIfSameVolReqInProcess_DifferentCapacity(t *testing.T) {
	cs := newTestControllerServer(nil)
	cs.Driver.reqmap["vol-3"] = 4096
	scVol := &scaleVolume{VolName: "vol-3", VolSize: 8192}
	inProc, err := cs.IfSameVolReqInProcess(scVol)
	require.Error(t, err)
	assert.False(t, inProc)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ---------------------------------------------------------------------------
// getTargetPath
// ---------------------------------------------------------------------------

func TestGetTargetPath_MissingBothParams(t *testing.T) {
	// Both fsetLinkPath and fsMountPoint are empty → error.
	cs := newTestControllerServer(nil)
	_, err := cs.getTargetPath(freshCtx(), "", "", "vol", false, false, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing details")
}

func TestGetTargetPath_MissingFsetLinkPath(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.getTargetPath(freshCtx(), "", "/ibm/fs1", "vol", false, false, false)
	require.Error(t, err)
}

func TestGetTargetPath_MissingFsMountPoint(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.getTargetPath(freshCtx(), "/ibm/fs1/fset", "", "vol", false, false, false)
	require.Error(t, err)
}

// TestGetTargetPath_NormalPath verifies that when both parameters are present
// the function strips the mount point prefix and returns the relative path.
func TestGetTargetPath_NormalPath(t *testing.T) {
	cs := newTestControllerServer(nil)
	path, err := cs.getTargetPath(freshCtx(),
		"/ibm/fs1/fset", "/ibm/fs1", "vol", false, false, false)
	require.NoError(t, err)
	// Trailing slashes/bangs are trimmed; mount point is stripped.
	assert.Equal(t, "fset", path)
}

// TestGetTargetPath_CreateDataDir verifies that the -data suffix is appended
// when createDataDir=true and the volume is neither CG nor cache.
func TestGetTargetPath_CreateDataDir(t *testing.T) {
	cs := newTestControllerServer(nil)
	path, err := cs.getTargetPath(freshCtx(),
		"/ibm/fs1/fset", "/ibm/fs1", "myvol", true, false, false)
	require.NoError(t, err)
	assert.Equal(t, "fset/myvol-data", path)
}

// TestGetTargetPath_CreateDataDir_CG verifies that -data is NOT appended
// for CG volumes even when createDataDir=true.
func TestGetTargetPath_CreateDataDir_CG(t *testing.T) {
	cs := newTestControllerServer(nil)
	path, err := cs.getTargetPath(freshCtx(),
		"/ibm/fs1/fset", "/ibm/fs1", "myvol", true, true, false)
	require.NoError(t, err)
	assert.Equal(t, "fset", path)
}

// TestGetTargetPath_CreateDataDir_Cache verifies that -data is NOT appended
// for cache volumes.
func TestGetTargetPath_CreateDataDir_Cache(t *testing.T) {
	cs := newTestControllerServer(nil)
	path, err := cs.getTargetPath(freshCtx(),
		"/ibm/fs1/fset", "/ibm/fs1", "myvol", true, false, true)
	require.NoError(t, err)
	assert.Equal(t, "fset", path)
}

// ---------------------------------------------------------------------------
// getVolumeSizeInBytes
// ---------------------------------------------------------------------------

// TestGetVolumeSizeInBytes_NoCapacityRange verifies that a nil CapacityRange
// returns 0 (not smallestVolSize — that enforcement is the caller's job).
func TestGetVolumeSizeInBytes_NoCapacityRange(t *testing.T) {
	cs := newTestControllerServer(nil)
	req := &csi.CreateVolumeRequest{Name: "vol"}
	size := cs.getVolumeSizeInBytes(req)
	assert.Equal(t, int64(0), size, "no CapacityRange → 0 bytes returned by helper")
}

// TestGetVolumeSizeInBytes_WithRequiredBytes verifies that the exact required
// bytes value is returned unchanged.
func TestGetVolumeSizeInBytes_WithRequiredBytes(t *testing.T) {
	cs := newTestControllerServer(nil)
	wantBytes := int64(5 * 1024 * 1024 * 1024) // 5 GiB
	req := &csi.CreateVolumeRequest{
		Name:          "vol",
		CapacityRange: &csi.CapacityRange{RequiredBytes: wantBytes},
	}
	size := cs.getVolumeSizeInBytes(req)
	assert.Equal(t, wantBytes, size)
}

// ---------------------------------------------------------------------------
// checkMinScaleVersionValid
// ---------------------------------------------------------------------------

func TestCheckMinScaleVersionValid(t *testing.T) {
	tests := []struct {
		name      string
		assembled string
		minimum   string
		want      bool
	}{
		// Version equal to minimum → valid.
		{"equal", "5110", "5110", true},
		// Version above minimum → valid.
		{"above", "5120", "5110", true},
		// Version below minimum → invalid.
		{"below", "5100", "5110", false},
		// Much higher version → valid.
		{"higher", "6010", "5110", true},
		// Much lower version → invalid.
		{"lower", "4999", "5110", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkMinScaleVersionValid(tc.assembled, tc.minimum)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// checkMinFsVersion
// ---------------------------------------------------------------------------

func TestCheckMinFsVersion(t *testing.T) {
	cs := newTestControllerServer(nil)
	tests := []struct {
		name      string
		fsVersion string // e.g. "27.00"
		minimum   string // e.g. "2700"
		want      bool
	}{
		// Exactly at boundary.
		{"at boundary", "27.00", "2700", true},
		// Above boundary.
		{"above", "28.00", "2700", true},
		// Below boundary.
		{"below", "26.99", "2700", false},
		// Well above.
		{"well above", "30.00", "2700", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cs.checkMinFsVersion(tc.fsVersion, tc.minimum)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// check*Support helpers — each requires min version or returns FailedPrecondition
// ---------------------------------------------------------------------------

// TestCheckSnapshotSupport verifies version boundary for snapshot support (5.1.1-0).
func TestCheckSnapshotSupport(t *testing.T) {
	cs := newTestControllerServer(nil)
	// Exactly at minimum → nil.
	assert.NoError(t, cs.checkSnapshotSupport("5110"))
	// Above → nil.
	assert.NoError(t, cs.checkSnapshotSupport("5120"))
	// Below → FailedPrecondition.
	err := cs.checkSnapshotSupport("5100")
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// TestCheckVolCloneSupport verifies version boundary for volume clone (5.1.2-1).
func TestCheckVolCloneSupport(t *testing.T) {
	cs := newTestControllerServer(nil)
	assert.NoError(t, cs.checkVolCloneSupport("5121"))
	assert.NoError(t, cs.checkVolCloneSupport("6000"))
	err := cs.checkVolCloneSupport("5120")
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// TestCheckCGSupport verifies version boundary for consistency group (5.1.3-0).
func TestCheckCGSupport(t *testing.T) {
	cs := newTestControllerServer(nil)
	assert.NoError(t, cs.checkCGSupport("5130"))
	assert.NoError(t, cs.checkCGSupport("5140"))
	err := cs.checkCGSupport("5129")
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// TestCheckCacheVolumeSupport verifies version boundary for cache volume (5.2.3-0).
func TestCheckCacheVolumeSupport(t *testing.T) {
	cs := newTestControllerServer(nil)
	assert.NoError(t, cs.checkCacheVolumeSupport("5230"))
	assert.NoError(t, cs.checkCacheVolumeSupport("6000"))
	err := cs.checkCacheVolumeSupport("5229")
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// TestCheckVMDiskCloningSupport verifies version boundary for VM disk cloning (6.0.1-0).
func TestCheckVMDiskCloningSupport(t *testing.T) {
	cs := newTestControllerServer(nil)
	assert.NoError(t, cs.checkVMDiskCloningSupport("6010"))
	assert.NoError(t, cs.checkVMDiskCloningSupport("6020"))
	err := cs.checkVMDiskCloningSupport("6009")
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// TestCheckVolTierSupport verifies version boundary for tiering (fs version 27.00).
func TestCheckVolTierSupport(t *testing.T) {
	cs := newTestControllerServer(nil)
	assert.NoError(t, cs.checkVolTierSupport("27.00"))
	assert.NoError(t, cs.checkVolTierSupport("28.00"))
	err := cs.checkVolTierSupport("26.99")
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// ---------------------------------------------------------------------------
// parseStatDirInfo
// ---------------------------------------------------------------------------

// TestParseStatDirInfo_ValidInput verifies that a three-line stat output
// with the link count as the last word on line 3 is parsed correctly.
func TestParseStatDirInfo_ValidInput(t *testing.T) {
	statInfo := "File: /ibm/fs1/dir\nSize: 4096\nLinks: 5"
	nlink, err := parseStatDirInfo(statInfo)
	require.NoError(t, err)
	assert.Equal(t, 5, nlink)
}

// TestParseStatDirInfo_MultiWordLastLine verifies parsing when the third line
// has multiple words and the link count is the last token.
func TestParseStatDirInfo_MultiWordLastLine(t *testing.T) {
	statInfo := "File: /ibm/fs1/dir\nSize: 4096\nInode: 12345 Links: 3"
	nlink, err := parseStatDirInfo(statInfo)
	require.NoError(t, err)
	assert.Equal(t, 3, nlink)
}

// TestParseStatDirInfo_NonNumericLastWord verifies that a non-numeric last
// word on line 3 returns a strconv error.
func TestParseStatDirInfo_NonNumericLastWord(t *testing.T) {
	statInfo := "File: /ibm/fs1/dir\nSize: 4096\nLinks: abc"
	_, err := parseStatDirInfo(statInfo)
	require.Error(t, err, "non-numeric link count must return error")
}

// ---------------------------------------------------------------------------
// checkExpiry
// ---------------------------------------------------------------------------

// TestCheckExpiry_Expired verifies that a ClusterDetails whose lastupdated
// timestamp is further in the past than expiryDuration returns true.
func TestCheckExpiry_Expired(t *testing.T) {
	// 48 hours ago, expiry window 24 hours → expired.
	cd := ClusterDetails{
		id:             "c1",
		name:           "cluster1",
		lastupdated:    time.Now().Add(-48 * time.Hour),
		expiryDuration: 24,
	}
	assert.True(t, checkExpiry(cd), "entry older than expiry window must be expired")
}

// TestCheckExpiry_Fresh verifies that a recently-updated ClusterDetails
// returns false (not expired).
func TestCheckExpiry_Fresh(t *testing.T) {
	// Updated 1 hour ago, expiry window 24 hours → still fresh.
	cd := ClusterDetails{
		id:             "c2",
		name:           "cluster2",
		lastupdated:    time.Now().Add(-1 * time.Hour),
		expiryDuration: 24,
	}
	assert.False(t, checkExpiry(cd), "recently-updated entry must not be expired")
}

// ---------------------------------------------------------------------------
// GetSnapIdMembers
// ---------------------------------------------------------------------------

func TestGetSnapIdMembers(t *testing.T) {
	cs := newTestControllerServer(nil)

	tests := []struct {
		name      string
		snapID    string
		wantErr   bool
		wantCode  codes.Code
		checkFunc func(t *testing.T, s scaleSnapId)
	}{
		{
			// Classic 4-part snap ID: clusterId;FSUUID;filesetName;snapshotName
			name:   "classic 4-part",
			snapID: "clus1;fsuuid1;fset1;snap1",
			checkFunc: func(t *testing.T, s scaleSnapId) {
				assert.Equal(t, "clus1", s.ClusterId)
				assert.Equal(t, "fsuuid1", s.FsUUID)
				assert.Equal(t, "fset1", s.FsetName)
				assert.Equal(t, "snap1", s.SnapName)
				assert.Equal(t, "/", s.Path, "missing path segment defaults to /")
				assert.Equal(t, STORAGECLASS_CLASSIC, s.StorageClassType)
			},
		},
		{
			// Classic 5-part snap ID with explicit path.
			name:   "classic 5-part with path",
			snapID: "clus1;fsuuid1;fset1;snap1;/my/custom/path",
			checkFunc: func(t *testing.T, s scaleSnapId) {
				assert.Equal(t, "/my/custom/path", s.Path)
				assert.Equal(t, STORAGECLASS_CLASSIC, s.StorageClassType)
			},
		},
		{
			// Advanced 8-part snap ID.
			// storageclass_type;volumeType;clusterId;FSUUID;cg;filesetName;snapshotName;metaSnapName
			name:   "advanced 8-part",
			snapID: "1;2;clus2;fsuuid2;cgname;fset2;snap2;metasnap",
			checkFunc: func(t *testing.T, s scaleSnapId) {
				assert.Equal(t, "1", s.StorageClassType)
				assert.Equal(t, "clus2", s.ClusterId)
				assert.Equal(t, "fsuuid2", s.FsUUID)
				assert.Equal(t, "cgname", s.ConsistencyGroup)
				assert.Equal(t, "fset2", s.FsetName)
				assert.Equal(t, "snap2", s.SnapName)
				assert.Equal(t, "metasnap", s.MetaSnapName)
				assert.Equal(t, "/", s.Path)
			},
		},
		{
			// Advanced 9-part snap ID with explicit path.
			name:   "advanced 9-part with path",
			snapID: "1;2;clus2;fsuuid2;cgname;fset2;snap2;metasnap;/custom/path",
			checkFunc: func(t *testing.T, s scaleSnapId) {
				assert.Equal(t, "/custom/path", s.Path)
			},
		},
		{
			// Too short → Internal error.
			name:     "too short",
			snapID:   "clus1;fsuuid1;fset1",
			wantErr:  true,
			wantCode: codes.Internal,
		},
		{
			// Empty string → Internal error.
			name:     "empty",
			snapID:   "",
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := cs.GetSnapIdMembers(tc.snapID)
			if tc.wantErr {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, st.Code())
				return
			}
			require.NoError(t, err)
			if tc.checkFunc != nil {
				tc.checkFunc(t, s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ControllerGetCapabilities
// ---------------------------------------------------------------------------

// TestControllerGetCapabilities_Empty verifies the response when no capabilities
// have been added to the driver.
func TestControllerGetCapabilities_Empty(t *testing.T) {
	cs := newTestControllerServer(nil)
	resp, err := cs.ControllerGetCapabilities(freshCtx(), &csi.ControllerGetCapabilitiesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Capabilities)
}

// TestControllerGetCapabilities_WithCaps verifies that the capabilities added
// to the driver are reflected in the response.
func TestControllerGetCapabilities_WithCaps(t *testing.T) {
	cs := withCap(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	)
	resp, err := cs.ControllerGetCapabilities(freshCtx(), &csi.ControllerGetCapabilitiesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Capabilities, 2)
	types := []csi.ControllerServiceCapability_RPC_Type{
		resp.Capabilities[0].GetRpc().Type,
		resp.Capabilities[1].GetRpc().Type,
	}
	assert.Contains(t, types, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	assert.Contains(t, types, csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT)
}

// ---------------------------------------------------------------------------
// ValidateVolumeCapabilities
// ---------------------------------------------------------------------------

// TestValidateVolumeCapabilities_MissingVolumeID verifies that an empty
// volumeID returns InvalidArgument.
func TestValidateVolumeCapabilities_MissingVolumeID(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.ValidateVolumeCapabilities(freshCtx(), &csi.ValidateVolumeCapabilitiesRequest{})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// TestValidateVolumeCapabilities_EmptyCapabilities verifies that a missing
// capabilities list returns InvalidArgument.
func TestValidateVolumeCapabilities_EmptyCapabilities(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.ValidateVolumeCapabilities(freshCtx(), &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId: "cluster1;fsuuid1;fset1;path=/ibm/fs1/vol",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// TestValidateVolumeCapabilities_MultiNodeMultiWriter verifies that
// MULTI_NODE_MULTI_WRITER capabilities are confirmed.
func TestValidateVolumeCapabilities_MultiNodeMultiWriter(t *testing.T) {
	cs := newTestControllerServer(nil)
	caps := []*csi.VolumeCapability{stdVolCap()}
	resp, err := cs.ValidateVolumeCapabilities(freshCtx(), &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           "cluster1;fsuuid1;fset1;path=/ibm/fs1/vol",
		VolumeCapabilities: caps,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	// MULTI_NODE_MULTI_WRITER must be confirmed.
	require.NotNil(t, resp.Confirmed, "MULTI_NODE_MULTI_WRITER must be in confirmed capabilities")
}

// TestValidateVolumeCapabilities_SingleNodeWriter verifies that a
// SINGLE_NODE_WRITER access mode returns an empty (unconfirmed) response.
func TestValidateVolumeCapabilities_SingleNodeWriter(t *testing.T) {
	cs := newTestControllerServer(nil)
	singleCap := &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
	}
	resp, err := cs.ValidateVolumeCapabilities(freshCtx(), &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           "cluster1;fsuuid1;fset1;path=/ibm/fs1/vol",
		VolumeCapabilities: []*csi.VolumeCapability{singleCap},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Nil(t, resp.Confirmed, "non MULTI_NODE_MULTI_WRITER mode must return unconfirmed response")
}

// ---------------------------------------------------------------------------
// ControllerUnpublishVolume
// ---------------------------------------------------------------------------

// TestControllerUnpublishVolume_CapabilityNotAdvertised verifies that calling
// ControllerUnpublishVolume without the PUBLISH_UNPUBLISH_VOLUME capability
// returns an Internal error.
func TestControllerUnpublishVolume_CapabilityNotAdvertised(t *testing.T) {
	// Server with no capabilities.
	cs := newTestControllerServer(nil)
	_, err := cs.ControllerUnpublishVolume(freshCtx(), &csi.ControllerUnpublishVolumeRequest{
		VolumeId: "cluster1;fsuuid1;fset1;path=/ibm/fs1/vol",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// TestControllerUnpublishVolume_BadVolumeID verifies that a volume ID that
// does not parse returns InvalidArgument.
func TestControllerUnpublishVolume_BadVolumeID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	_, err := cs.ControllerUnpublishVolume(freshCtx(), &csi.ControllerUnpublishVolumeRequest{
		VolumeId: "not-valid",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// TestControllerUnpublishVolume_Success verifies that a well-formed volume ID
// returns success when the capability is advertised.
func TestControllerUnpublishVolume_Success(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	// Use a valid 3-part LW volume ID (cluster;fsuuid;path=<abs>).
	resp, err := cs.ControllerUnpublishVolume(freshCtx(), &csi.ControllerUnpublishVolumeRequest{
		VolumeId: "cluster1;fsuuid1;path=/ibm/fs1/vol",
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// ---------------------------------------------------------------------------
// ControllerPublishVolume — input-validation paths
// ---------------------------------------------------------------------------

// TestControllerPublishVolume_CapabilityNotAdvertised verifies that calling
// without the PUBLISH_UNPUBLISH_VOLUME capability returns Internal.
func TestControllerPublishVolume_CapabilityNotAdvertised(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.ControllerPublishVolume(freshCtx(), &csi.ControllerPublishVolumeRequest{
		VolumeId:         "cluster1;fsuuid1;path=/ibm/fs1/vol",
		NodeId:           "node1",
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// TestControllerPublishVolume_MissingNodeID verifies that an empty nodeID
// returns InvalidArgument.
func TestControllerPublishVolume_MissingNodeID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	_, err := cs.ControllerPublishVolume(freshCtx(), &csi.ControllerPublishVolumeRequest{
		VolumeId:         "cluster1;fsuuid1;path=/ibm/fs1/vol",
		VolumeCapability: stdVolCap(),
		// NodeId intentionally omitted.
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "NodeID")
}

// TestControllerPublishVolume_MissingVolumeID verifies that an empty volumeID
// returns InvalidArgument.
func TestControllerPublishVolume_MissingVolumeID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	_, err := cs.ControllerPublishVolume(freshCtx(), &csi.ControllerPublishVolumeRequest{
		NodeId:           "node1",
		VolumeCapability: stdVolCap(),
		// VolumeId intentionally omitted.
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// TestControllerPublishVolume_MissingCapability verifies that a nil
// VolumeCapability returns InvalidArgument.
func TestControllerPublishVolume_MissingCapability(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	_, err := cs.ControllerPublishVolume(freshCtx(), &csi.ControllerPublishVolumeRequest{
		VolumeId: "cluster1;fsuuid1;path=/ibm/fs1/vol",
		NodeId:   "node1",
		// VolumeCapability intentionally nil.
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// TestControllerPublishVolume_BadVolumeID verifies that a malformed volumeID
// (passes the nil check but fails parsing) returns InvalidArgument.
func TestControllerPublishVolume_BadVolumeID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	_, err := cs.ControllerPublishVolume(freshCtx(), &csi.ControllerPublishVolumeRequest{
		VolumeId:         "bad-volume-id",
		NodeId:           "node1",
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ---------------------------------------------------------------------------
// DeleteSnapshot — input-validation paths
// ---------------------------------------------------------------------------

// TestDeleteSnapshot_CapabilityNotAdvertised verifies that calling without the
// CREATE_DELETE_SNAPSHOT capability returns Internal.
func TestDeleteSnapshot_CapabilityNotAdvertised(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.DeleteSnapshot(freshCtx(), &csi.DeleteSnapshotRequest{
		SnapshotId: "clus1;fsuuid1;fset1;snap1",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// TestDeleteSnapshot_MissingSnapshotID verifies that an empty snapshotId
// returns InvalidArgument.
func TestDeleteSnapshot_MissingSnapshotID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT)
	_, err := cs.DeleteSnapshot(freshCtx(), &csi.DeleteSnapshotRequest{
		SnapshotId: "",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "snapshot Id")
}

// TestDeleteSnapshot_InvalidSnapID verifies that a snap ID that is too short
// (< 4 parts) returns an Internal error from GetSnapIdMembers.
func TestDeleteSnapshot_InvalidSnapID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT)
	_, err := cs.DeleteSnapshot(freshCtx(), &csi.DeleteSnapshotRequest{
		SnapshotId: "clus1;fsuuid1;fset1", // only 3 parts → invalid
	})
	require.Error(t, err)
	// The error propagates from GetSnapIdMembers with codes.Internal.
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ---------------------------------------------------------------------------
// ControllerModifyVolume — input-validation paths
// ---------------------------------------------------------------------------

// TestControllerModifyVolume_CapabilityNotAdvertised verifies that calling
// without MODIFY_VOLUME capability returns Internal.
func TestControllerModifyVolume_CapabilityNotAdvertised(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.ControllerModifyVolume(freshCtx(), &csi.ControllerModifyVolumeRequest{
		VolumeId:          "cluster1;fsuuid1;path=/ibm/fs1/vol",
		MutableParameters: map[string]string{"key": "val"},
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// TestControllerModifyVolume_MissingVolumeID verifies that a missing volumeId
// returns InvalidArgument.
func TestControllerModifyVolume_MissingVolumeID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_MODIFY_VOLUME)
	_, err := cs.ControllerModifyVolume(freshCtx(), &csi.ControllerModifyVolumeRequest{
		MutableParameters: map[string]string{"key": "val"},
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// TestControllerModifyVolume_EmptyParams verifies that empty mutable
// parameters return InvalidArgument.
func TestControllerModifyVolume_EmptyParams(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_MODIFY_VOLUME)
	_, err := cs.ControllerModifyVolume(freshCtx(), &csi.ControllerModifyVolumeRequest{
		VolumeId: "cluster1;fsuuid1;path=/ibm/fs1/vol",
		// MutableParameters intentionally empty.
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ---------------------------------------------------------------------------
// DeleteVolume — input-validation paths
// ---------------------------------------------------------------------------

// TestDeleteVolume_CapabilityNotAdvertised verifies that calling without the
// CREATE_DELETE_VOLUME capability returns an error.
func TestDeleteVolume_CapabilityNotAdvertised(t *testing.T) {
	cs := newTestControllerServer(nil)
	_, err := cs.DeleteVolume(freshCtx(), &csi.DeleteVolumeRequest{
		VolumeId: "cluster1;fsuuid1;path=/ibm/fs1/vol",
	})
	require.Error(t, err)
}

// TestDeleteVolume_MissingVolumeID verifies that an empty volumeId
// returns InvalidArgument.
func TestDeleteVolume_MissingVolumeID(t *testing.T) {
	cs := withCap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	_, err := cs.DeleteVolume(freshCtx(), &csi.DeleteVolumeRequest{
		VolumeId: "",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}
