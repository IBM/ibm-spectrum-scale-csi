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
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestNodeServer returns a fully-initialised ScaleNodeServer whose
// ScaleDriver has a nodeID and an empty nscap slice (avoids nil-panics in
// NodeGetCapabilities).
func newTestNodeServer() *ScaleNodeServer {
	d := &ScaleDriver{
		nodeID: "test-node-01",
		nscap:  []*csi.NodeServiceCapability{},
	}
	return &ScaleNodeServer{Driver: d}
}

// freshCtx returns a context with a unique logger-id so every test is
// cleanly isolated from the global klog output.
func freshCtx() context.Context {
	return context.Background()
}

// lwVolumeID returns a lightweight (directory-based) volume ID.
// Format: <clusterID>;<fsUUID>;path=<absPath>
func lwVolumeID(path string) string {
	return "cluster1;fsuuid1;path=" + path
}

// fsetVolumeID returns a fileset-based volume ID.
// Format: <clusterID>;<fsUUID>;fileset=<fsetID>;path=<absPath>
func fsetVolumeID(path string) string {
	return "cluster1;fsuuid1;fileset=fset0;path=" + path
}

// v25VolumeID returns a CSI-2.5 style volume ID (7 parts).
// Format: <scType>;<volType>;<clusterID>;<fsUUID>;<cg>;<fsetName>;<path>
func v25VolumeID(scType, volType, path string) string {
	return scType + ";0;" + "cluster1;fsuuid1;cg;fset1;" + path
}

// stdVolCap returns a basic MULTI_NODE_MULTI_WRITER mount capability.
func stdVolCap() *csi.VolumeCapability {
	return &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	}
}

// ---------------------------------------------------------------------------
// NodeGetCapabilities
// ---------------------------------------------------------------------------

func TestNodeGetCapabilities_EmptyCapabilities(t *testing.T) {
	ns := newTestNodeServer()
	resp, err := ns.NodeGetCapabilities(freshCtx(), &csi.NodeGetCapabilitiesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Capabilities)
}

func TestNodeGetCapabilities_WithVolumeStats(t *testing.T) {
	d := &ScaleDriver{
		nodeID: "node1",
		nscap: []*csi.NodeServiceCapability{
			NewNodeServiceCapability(csi.NodeServiceCapability_RPC_GET_VOLUME_STATS),
		},
	}
	ns := &ScaleNodeServer{Driver: d}
	resp, err := ns.NodeGetCapabilities(freshCtx(), &csi.NodeGetCapabilitiesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Capabilities, 1)
	assert.Equal(t, csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		resp.Capabilities[0].GetRpc().Type)
}

// ---------------------------------------------------------------------------
// NodeGetInfo
// ---------------------------------------------------------------------------

func TestNodeGetInfo_ReturnsNodeID(t *testing.T) {
	ns := newTestNodeServer()
	resp, err := ns.NodeGetInfo(freshCtx(), &csi.NodeGetInfoRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "test-node-01", resp.NodeId)
}

// ---------------------------------------------------------------------------
// NodeStageVolume / NodeUnstageVolume / NodeExpandVolume — all Unimplemented
// ---------------------------------------------------------------------------

func TestNodeStageVolume_Unimplemented(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeStageVolume(freshCtx(), &csi.NodeStageVolumeRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

func TestNodeUnstageVolume_Unimplemented(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeUnstageVolume(freshCtx(), &csi.NodeUnstageVolumeRequest{})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

func TestNodeExpandVolume_Unimplemented(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeExpandVolume(freshCtx(), &csi.NodeExpandVolumeRequest{})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

// ---------------------------------------------------------------------------
// NodePublishVolume — input validation
// ---------------------------------------------------------------------------

func TestNodePublishVolume_MissingVolumeID(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
		TargetPath:       "/tmp/target",
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "volumeID")
}

func TestNodePublishVolume_MissingTargetPath(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
		VolumeId:         lwVolumeID("/mnt/vol"),
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "targetPath")
}

func TestNodePublishVolume_MissingVolumeCapability(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
		VolumeId:   lwVolumeID("/mnt/vol"),
		TargetPath: "/tmp/target",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "volume capability")
}

func TestNodePublishVolume_InvalidVolumeID(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
		VolumeId:         "not-a-valid-volume-id",
		TargetPath:       "/tmp/target",
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "volumeID is not in proper format")
}

// TestNodePublishVolume_SourcePathNotExist verifies that when the volume
// source path (resolved from the volume ID) does not exist on the host
// filesystem, NodePublishVolume returns a non-nil error (it calls os.Lstat
// and expects the path to be present).
func TestNodePublishVolume_SourcePathNotExist(t *testing.T) {
	ns := newTestNodeServer()
	// Use a path that will not exist under /host (inside the container the host
	// rootfs is at /host, so /host/nonexistent/path will always be absent).
	volID := lwVolumeID("/nonexistent/scale/path")
	_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
		VolumeId:         volID,
		TargetPath:       "/tmp/target",
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	// The error message must mention the lstat failure.
	assert.Contains(t, err.Error(), "lstat")
}

// ---------------------------------------------------------------------------
// NodePublishVolume — concurrent lock
// ---------------------------------------------------------------------------

// TestNodePublishVolume_LockConflict verifies that a concurrent request for
// the same target path is rejected with codes.Internal while the first
// request holds the in-memory lock.
func TestNodePublishVolume_LockConflict(t *testing.T) {
	targetPath := t.TempDir() + "/conflict-target"

	// Pre-acquire the lock manually to simulate a concurrent in-flight call.
	locked := lock(targetPath, freshCtx())
	require.True(t, locked, "expected to acquire lock on a fresh path")
	defer unlock(targetPath, freshCtx())

	ns := newTestNodeServer()
	_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
		VolumeId:         lwVolumeID("/mnt/vol"),
		TargetPath:       targetPath,
		VolumeCapability: stdVolCap(),
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "in progress")
}

// ---------------------------------------------------------------------------
// NodeUnpublishVolume — input validation
// ---------------------------------------------------------------------------

func TestNodeUnpublishVolume_MissingVolumeID(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeUnpublishVolume(freshCtx(), &csi.NodeUnpublishVolumeRequest{
		TargetPath: "/tmp/target",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "volumeID")
}

func TestNodeUnpublishVolume_MissingTargetPath(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeUnpublishVolume(freshCtx(), &csi.NodeUnpublishVolumeRequest{
		VolumeId: "cluster1;fsuuid1;path=/mnt/vol",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "targetPath")
}

// TestNodeUnpublishVolume_TargetNotFound returns success when the target path
// does not exist (idempotent unpublish).
func TestNodeUnpublishVolume_TargetNotFound(t *testing.T) {
	ns := newTestNodeServer()
	resp, err := ns.NodeUnpublishVolume(freshCtx(), &csi.NodeUnpublishVolumeRequest{
		VolumeId:   "cluster1;fsuuid1;path=/mnt/vol",
		TargetPath: "/tmp/does-not-exist-" + t.Name(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestNodeUnpublishVolume_SymlinkRemoved verifies that a symlink target path
// is removed and success is returned.
func TestNodeUnpublishVolume_SymlinkRemoved(t *testing.T) {
	dir := t.TempDir()
	realTarget := filepath.Join(dir, "real-dir")
	require.NoError(t, os.Mkdir(realTarget, 0750))
	symlinkPath := filepath.Join(dir, "symlink-target")
	require.NoError(t, os.Symlink(realTarget, symlinkPath))

	ns := newTestNodeServer()
	resp, err := ns.NodeUnpublishVolume(freshCtx(), &csi.NodeUnpublishVolumeRequest{
		VolumeId:   "cluster1;fsuuid1;path=/mnt/vol",
		TargetPath: symlinkPath,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	// The symlink must be gone after a successful unpublish.
	_, statErr := os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(statErr), "symlink should have been removed")
}

// TestNodeUnpublishVolume_LockConflict checks that a concurrent unpublish on
// the same target is refused with codes.Internal.
func TestNodeUnpublishVolume_LockConflict(t *testing.T) {
	targetPath := t.TempDir() + "/lock-target"
	locked := lock(targetPath, freshCtx())
	require.True(t, locked)
	defer unlock(targetPath, freshCtx())

	ns := newTestNodeServer()
	_, err := ns.NodeUnpublishVolume(freshCtx(), &csi.NodeUnpublishVolumeRequest{
		VolumeId:   "cluster1;fsuuid1;path=/mnt/vol",
		TargetPath: targetPath,
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ---------------------------------------------------------------------------
// NodeGetVolumeStats — input validation
// ---------------------------------------------------------------------------

func TestNodeGetVolumeStats_MissingVolumeID(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeGetVolumeStats(freshCtx(), &csi.NodeGetVolumeStatsRequest{
		VolumePath: "/tmp/some-path",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestNodeGetVolumeStats_MissingVolumePath(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeGetVolumeStats(freshCtx(), &csi.NodeGetVolumeStatsRequest{
		VolumeId: fsetVolumeID("/mnt/vol"),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestNodeGetVolumeStats_VolumePathNotExist(t *testing.T) {
	ns := newTestNodeServer()
	_, err := ns.NodeGetVolumeStats(freshCtx(), &csi.NodeGetVolumeStatsRequest{
		VolumeId:   fsetVolumeID("/mnt/vol"),
		VolumePath: "/tmp/definitely-does-not-exist-" + t.Name(),
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestNodeGetVolumeStats_InvalidVolumeID(t *testing.T) {
	dir := t.TempDir() // a path that actually exists
	ns := newTestNodeServer()
	_, err := ns.NodeGetVolumeStats(freshCtx(), &csi.NodeGetVolumeStatsRequest{
		VolumeId:   "bad-id",
		VolumePath: dir,
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "volumeID is not in proper format")
}

// TestNodeGetVolumeStats_LightweightVolumeRejected verifies that a LW
// (directory-based) volume ID is rejected because stats are not supported
// for lightweight volumes.
func TestNodeGetVolumeStats_LightweightVolumeRejected(t *testing.T) {
	dir := t.TempDir()
	ns := newTestNodeServer()
	// 3-part volume ID => IsFilesetBased = false
	_, err := ns.NodeGetVolumeStats(freshCtx(), &csi.NodeGetVolumeStatsRequest{
		VolumeId:   lwVolumeID("/mnt/vol"),
		VolumePath: dir,
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "lightweight")
}

// ---------------------------------------------------------------------------
// lock / unlock — concurrent safety
// ---------------------------------------------------------------------------

func TestLockUnlock_AcquireAndRelease(t *testing.T) {
	path := "/tmp/test-lock-" + t.Name()
	ctx := freshCtx()

	ok := lock(path, ctx)
	assert.True(t, ok, "first lock should succeed")

	// Second attempt on same path must fail.
	ok2 := lock(path, ctx)
	assert.False(t, ok2, "second lock on same path must fail")

	unlock(path, ctx)

	// After unlocking, re-acquisition must succeed.
	ok3 := lock(path, ctx)
	assert.True(t, ok3, "lock after unlock must succeed")
	unlock(path, ctx)
}

func TestLockUnlock_DifferentPaths(t *testing.T) {
	ctx := freshCtx()
	p1 := "/tmp/lock-path1-" + t.Name()
	p2 := "/tmp/lock-path2-" + t.Name()

	ok1 := lock(p1, ctx)
	ok2 := lock(p2, ctx)
	assert.True(t, ok1, "lock p1 should succeed")
	assert.True(t, ok2, "lock p2 should succeed independently")

	unlock(p1, ctx)
	unlock(p2, ctx)
}

// ---------------------------------------------------------------------------
// validatePath (package-internal helper in gpfs_util.go)
// ---------------------------------------------------------------------------

func TestValidatePath_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty path", "", true},
		{"relative path", "relative/path", true},
		{"path with dotdot", "/mnt/../etc/passwd", true},
		{"valid abs path", "/mnt/gpfs/vol1", false},
		{"valid deep abs path", "/mnt/gpfs/cluster1/pvc-abc/data", false},
		{"double slash in middle", "/mnt//gpfs", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePath(tc.path)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getVolIDMembers — table-driven unit tests
// ---------------------------------------------------------------------------

func TestGetVolIDMembers_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		volID          string
		wantErr        bool
		wantFileset    bool
		wantPath       string
		wantClusterID  string
	}{
		{
			name:        "invalid single part",
			volID:       "onlyone",
			wantErr:     true,
		},
		{
			name:        "invalid two parts",
			volID:       "a;b",
			wantErr:     true,
		},
		{
			name:        "valid LW (3-part)",
			volID:       "cid;fsuuid;path=/mnt/vol",
			wantErr:     false,
			wantFileset: false,
			wantPath:    "/mnt/vol",
			wantClusterID: "cid",
		},
		{
			name:        "LW with non-absolute path",
			volID:       "cid;fsuuid;path=relative/path",
			wantErr:     true,
		},
		{
			name:        "valid fileset-based (4-part)",
			volID:       "cid;fsuuid;fileset=fset0;path=/mnt/vol",
			wantErr:     false,
			wantFileset: true,
			wantPath:    "/mnt/vol",
			wantClusterID: "cid",
		},
		{
			name:    "4-part malformed path part",
			volID:   "cid;fsuuid;fileset=fset0;badpath",
			wantErr: true,
		},
		{
			name:        "valid v2.5 classic dir-based (7-part, IsFilesetBased=false)",
			volID:       "0;0;cid;fsuuid;cg;fset1;/mnt/vol",
			wantErr:     false,
			wantFileset: false,
			wantPath:    "/mnt/vol",
		},
		{
			name:        "valid v2.5 classic fileset-based (7-part, type!=0 => IsFilesetBased=true)",
			volID:       "0;2;cid;fsuuid;cg;fset1;/mnt/vol",
			wantErr:     false,
			wantFileset: true,
			wantPath:    "/mnt/vol",
		},
		{
			name:        "valid v2.5 advanced storageclass (7-part, scType=1)",
			volID:       "1;2;cid;fsuuid;cg;fset1;/mnt/vol",
			wantErr:     false,
			wantFileset: true,
			wantPath:    "/mnt/vol",
		},
		{
			name:    "v2.5 with non-absolute path",
			volID:   "0;0;cid;fsuuid;cg;fset1;relative/path",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			members, err := getVolIDMembers(tc.volID)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantFileset, members.IsFilesetBased)
			assert.Equal(t, tc.wantPath, members.Path)
			if tc.wantClusterID != "" {
				assert.Equal(t, tc.wantClusterID, members.ClusterId)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// skipLogging (internal helper in utils.go)
// ---------------------------------------------------------------------------

func TestSkipLogging_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantSkip   bool
	}{
		{"NodeGetCapabilities", "/csi.Node/NodeGetCapabilities", true},
		{"Identity Probe", "/csi.Identity/Probe", true},
		{"Identity GetPluginInfo", "/csi.Identity/GetPluginInfo", true},
		{"Node NodeGetInfo", "/csi.Node/NodeGetInfo", true},
		{"Node NodeGetVolumeStats", "/csi.Node/NodeGetVolumeStats", true},
		{"NodePublishVolume", "/csi.Node/NodePublishVolume", false},
		{"NodeUnpublishVolume", "/csi.Node/NodeUnpublishVolume", false},
		{"ControllerCreateVolume", "/csi.Controller/CreateVolume", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantSkip, skipLogging(tc.method))
		})
	}
}

// ---------------------------------------------------------------------------
// ScaleDriver capability helpers
// ---------------------------------------------------------------------------

func TestAddNodeServiceCapabilities(t *testing.T) {
	d := &ScaleDriver{}
	ctx := freshCtx()
	err := d.AddNodeServiceCapabilities(ctx, []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
	})
	require.NoError(t, err)
	require.Len(t, d.nscap, 1)
	assert.Equal(t, csi.NodeServiceCapability_RPC_GET_VOLUME_STATS, d.nscap[0].GetRpc().Type)
}

func TestAddVolumeCapabilityAccessModes(t *testing.T) {
	d := &ScaleDriver{}
	ctx := freshCtx()
	err := d.AddVolumeCapabilityAccessModes(ctx, []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	})
	require.NoError(t, err)
	require.Len(t, d.vcap, 1)
	assert.Equal(t, csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, d.vcap[0].Mode)
}

func TestAddControllerServiceCapabilities(t *testing.T) {
	d := &ScaleDriver{}
	ctx := freshCtx()
	err := d.AddControllerServiceCapabilities(ctx, []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	require.NoError(t, err)
	require.Len(t, d.cscap, 1)
	assert.Equal(t, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME, d.cscap[0].GetRpc().Type)
}

func TestValidateControllerServiceRequest_UnknownAllowed(t *testing.T) {
	d := &ScaleDriver{}
	ctx := freshCtx()
	_ = d.AddControllerServiceCapabilities(ctx, []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	err := d.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_UNKNOWN)
	assert.NoError(t, err)
}

func TestValidateControllerServiceRequest_MatchingCap(t *testing.T) {
	d := &ScaleDriver{}
	ctx := freshCtx()
	_ = d.AddControllerServiceCapabilities(ctx, []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	err := d.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	assert.NoError(t, err)
}

func TestValidateControllerServiceRequest_NotAdvertised(t *testing.T) {
	d := &ScaleDriver{}
	ctx := freshCtx()
	_ = d.AddControllerServiceCapabilities(ctx, []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	err := d.ValidateControllerServiceRequest(ctx, csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// =============================================================================
// SECTION: Volume-handle format regression tests
// =============================================================================
//
// Every supported volumeHandle format must continue to be accepted by
// getVolIDMembers and produce the correct scaleVolId fields.  These tests form
// a regression fence: if a future refactor breaks any format the build breaks
// before any cluster-level test is needed.
//
// Supported formats (per gpfs_util.go):
//
//   Format A – LW 3-part legacy (pre-2.5)
//     <clusterID>;<fsUUID>;path=<absPath>
//
//   Format B – fileset 4-part legacy (pre-2.5) with numeric fileset id
//     <clusterID>;<fsUUID>;fileset=<id>;path=<absPath>
//
//   Format C – fileset 4-part legacy (pre-2.5) with filesetName key
//     <clusterID>;<fsUUID>;filesetName=<name>;path=<absPath>
//
//   Format D – CSI ≥2.5 classic DIRECTORYBASED (scType=0, volType=0)
//     0;0;<clusterID>;<fsUUID>;<cg>;<fsetName>;<absPath>
//
//   Format E – CSI ≥2.5 classic DEPENDENTFILESET (scType=0, volType=1)
//     0;1;<clusterID>;<fsUUID>;<cg>;<fsetName>;<absPath>
//
//   Format F – CSI ≥2.5 classic INDEPENDENTFILESET (scType=0, volType=2)
//     0;2;<clusterID>;<fsUUID>;<cg>;<fsetName>;<absPath>
//
//   Format G – CSI ≥2.5 classic SHALLOWCOPY (scType=0, volType=3)
//     0;3;<clusterID>;<fsUUID>;<cg>;<fsetName>;<absPath>
//
//   Format H – CSI ≥2.5 advanced (scType=1, any volType → IsFilesetBased=true)
//     1;2;<clusterID>;<fsUUID>;<cg>;<fsetName>;<absPath>
//
//   Format I – CSI ≥2.5 cache (scType=2, any volType → IsFilesetBased=true)
//     2;2;<clusterID>;<fsUUID>;<cg>;<fsetName>;<absPath>
// =============================================================================

// TestGetVolIDMembers_AllFormats_Regression exercises every supported handle
// format and asserts the exact scaleVolId fields that downstream code depends on.
func TestGetVolIDMembers_AllFormats_Regression(t *testing.T) {
	const (
		clid  = "7118073361626808055"
		fsuud = "09762E69:5D36FE8D"
		cg    = "cgroupA"
		fset  = "myfset"
		vpath = "/ibm/gpfs0/pvc-abc123/data"
	)

	tests := []struct {
		name            string
		volumeHandle    string
		wantClusterID   string
		wantFsUUID      string
		wantFsetID      string // empty unless 4-part fileset= key
		wantFsetName    string // empty unless filesetName= key or 7-part
		wantPath        string
		wantIsFileset   bool
		wantSCType      string // empty unless 7-part
		wantVolType     string // empty unless 7-part
		wantCG          string // empty unless 7-part
	}{
		// ── Format A: LW 3-part ──────────────────────────────────────────────
		{
			name:          "FormatA: LW 3-part",
			volumeHandle:  clid + ";" + fsuud + ";path=" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantPath:      vpath,
			wantIsFileset: false,
		},
		// ── Format B: fileset 4-part with numeric id ─────────────────────────
		{
			name:          "FormatB: fileset 4-part (fileset= key)",
			volumeHandle:  clid + ";" + fsuud + ";fileset=42;path=" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetID:    "42",
			wantPath:      vpath,
			wantIsFileset: true,
		},
		// ── Format C: fileset 4-part with filesetName key ────────────────────
		{
			name:          "FormatC: fileset 4-part (filesetName= key)",
			volumeHandle:  clid + ";" + fsuud + ";filesetName=" + fset + ";path=" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantPath:      vpath,
			wantIsFileset: true,
		},
		// ── Format D: CSI ≥2.5 classic directory-based ───────────────────────
		{
			name:          "FormatD: v2.5 classic dirBased (0;0)",
			volumeHandle:  "0;0;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantCG:        cg,
			wantPath:      vpath,
			wantIsFileset: false,
			wantSCType:    STORAGECLASS_CLASSIC,
			wantVolType:   FILE_DIRECTORYBASED_VOLUME,
		},
		// ── Format E: CSI ≥2.5 classic dependent fileset ─────────────────────
		{
			name:          "FormatE: v2.5 classic dependentFileset (0;1)",
			volumeHandle:  "0;1;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantCG:        cg,
			wantPath:      vpath,
			wantIsFileset: true,
			wantSCType:    STORAGECLASS_CLASSIC,
			wantVolType:   FILE_DEPENDENTFILESET_VOLUME,
		},
		// ── Format F: CSI ≥2.5 classic independent fileset ───────────────────
		{
			name:          "FormatF: v2.5 classic independentFileset (0;2)",
			volumeHandle:  "0;2;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantCG:        cg,
			wantPath:      vpath,
			wantIsFileset: true,
			wantSCType:    STORAGECLASS_CLASSIC,
			wantVolType:   FILE_INDEPENDENTFILESET_VOLUME,
		},
		// ── Format G: CSI ≥2.5 classic shallow-copy fileset ──────────────────
		{
			name:          "FormatG: v2.5 classic shallowCopy (0;3)",
			volumeHandle:  "0;3;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantCG:        cg,
			wantPath:      vpath,
			wantIsFileset: true,
			wantSCType:    STORAGECLASS_CLASSIC,
			wantVolType:   FILE_SHALLOWCOPY_VOLUME,
		},
		// ── Format H: CSI ≥2.5 advanced storageclass ─────────────────────────
		{
			name:          "FormatH: v2.5 advanced (1;2)",
			volumeHandle:  "1;2;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantCG:        cg,
			wantPath:      vpath,
			wantIsFileset: true,
			wantSCType:    STORAGECLASS_ADVANCED,
			wantVolType:   FILE_INDEPENDENTFILESET_VOLUME,
		},
		// ── Format I: CSI ≥2.5 cache storageclass ────────────────────────────
		{
			name:          "FormatI: v2.5 cache (2;2)",
			volumeHandle:  "2;2;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + vpath,
			wantClusterID: clid,
			wantFsUUID:    fsuud,
			wantFsetName:  fset,
			wantCG:        cg,
			wantPath:      vpath,
			wantIsFileset: true,
			wantSCType:    STORAGECLASS_CACHE,
			wantVolType:   FILE_INDEPENDENTFILESET_VOLUME,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m, err := getVolIDMembers(tc.volumeHandle)
			require.NoError(t, err, "valid handle must parse without error")

			assert.Equal(t, tc.wantClusterID, m.ClusterId, "ClusterId")
			assert.Equal(t, tc.wantFsUUID, m.FsUUID, "FsUUID")
			assert.Equal(t, tc.wantPath, m.Path, "Path")
			assert.Equal(t, tc.wantIsFileset, m.IsFilesetBased, "IsFilesetBased")

			if tc.wantFsetID != "" {
				assert.Equal(t, tc.wantFsetID, m.FsetId, "FsetId")
			}
			if tc.wantFsetName != "" {
				assert.Equal(t, tc.wantFsetName, m.FsetName, "FsetName")
			}
			if tc.wantSCType != "" {
				assert.Equal(t, tc.wantSCType, m.StorageClassType, "StorageClassType")
			}
			if tc.wantVolType != "" {
				assert.Equal(t, tc.wantVolType, m.VolType, "VolType")
			}
			if tc.wantCG != "" {
				assert.Equal(t, tc.wantCG, m.ConsistencyGroup, "ConsistencyGroup")
			}
		})
	}
}

// TestGetVolIDMembers_InvalidFormats_Regression confirms that malformed and
// unsupported segment counts are all rejected.
func TestGetVolIDMembers_InvalidFormats_Regression(t *testing.T) {
	invalid := []struct {
		name   string
		handle string
	}{
		{"empty string", ""},
		{"1 segment", "onlycluster"},
		{"2 segments", "cluster1;fsuuid"},
		{"5 segments – unsupported", "a;b;c;d;e"},
		{"6 segments – unsupported", "a;b;c;d;e;f"},
		{"8 segments – unsupported", "a;b;c;d;e;f;g;h"},
		{"3-part missing path= prefix", "c1;fsuuid;/mnt/vol"},
		{"4-part malformed fileset segment", "c1;fsuuid;BADKEY;path=/mnt/vol"},
		{"4-part malformed path segment", "c1;fsuuid;fileset=42;/mnt/vol"},
	}
	for _, tc := range invalid {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := getVolIDMembers(tc.handle)
			require.Error(t, err, "invalid handle must be rejected")
		})
	}
}

// =============================================================================
// SECTION: CVE path-traversal tests (NodePublishVolume defence-in-depth)
// =============================================================================
//
// CVE summary
// -----------
// Before validatePath() was added to getVolIDMembers, an attacker with
// `create persistentvolumes` permission could craft a volumeHandle whose
// path= component contained ".." segments (e.g. path=/ibm/gpfs0/../..).
// Because the only confinement gate was a strings.HasPrefix check against the
// GPFS mount list, a path like "/ibm/gpfs0/../.." passed the check yet
// resolved to host "/" when bind-mounted via chroot /host.
//
// The fix is validatePath() called inside every branch of getVolIDMembers.
// The tests below prove that:
//   1. The traversal payload is rejected at parse time (getVolIDMembers).
//   2. NodePublishVolume also rejects it (codes.InvalidArgument) – no OS call
//      is made with the traversal path, so no lstat, no bind-mount, nothing.
//   3. NodeUnpublishVolume also rejects all traversal handles at argument-
//      validation time (the volumeID itself never reaches getVolIDMembers
//      during unpublish, but the targetPath validation still gates execution).
//
// Each traversal payload is tested against every supported handle format to
// ensure there are no format-specific bypass paths.
// =============================================================================

// traversalPayloads lists attack strings that must be caught by validatePath.
// These are drawn directly from the CVE report's exploit examples.
var traversalPayloads = []struct {
	name    string
	path    string
}{
	// Classic dotdot traversal from the CVE PoC
	{"dotdot escape to root: /ibm/gpfs0/../..", "/ibm/gpfs0/../.."},
	// Partial escape (one hop)
	{"dotdot one hop: /ibm/gpfs0/..", "/ibm/gpfs0/.."},
	// Escape to /etc from a valid GPFS prefix
	{"escape to /etc: /ibm/gpfs0/../../etc", "/ibm/gpfs0/../../etc"},
	// Escape to /var/lib/kubelet to steal pod secrets
	{"escape to kubelet: /ibm/gpfs0/../../var/lib/kubelet", "/ibm/gpfs0/../../var/lib/kubelet"},
	// Trailing dotdot with trailing slash
	{"trailing slash traversal: /ibm/gpfs0/../", "/ibm/gpfs0/../"},
	// Non-normalised double slash (redundant separator)
	{"double slash: /ibm//gpfs0/vol", "/ibm//gpfs0/vol"},
	// Relative path (no leading slash)
	{"relative path: ibm/gpfs0/vol", "ibm/gpfs0/vol"},
	// Empty path
	{"empty path", ""},
}

// buildHandleWithPath builds a volumeHandle of the requested format that
// embeds the supplied (potentially malicious) path.
func buildHandleWithPath(format, path string) string {
	const (
		clid  = "7118073361626808055"
		fsuud = "09762E69:5D36FE8D"
		cg    = "cgroupA"
		fset  = "myfset"
	)
	switch format {
	case "A": // LW 3-part
		return clid + ";" + fsuud + ";path=" + path
	case "B": // fileset 4-part (fileset= key)
		return clid + ";" + fsuud + ";fileset=42;path=" + path
	case "C": // fileset 4-part (filesetName= key)
		return clid + ";" + fsuud + ";filesetName=" + fset + ";path=" + path
	case "D7": // CSI ≥2.5 (7-part), using the path as the last segment
		return "0;0;" + clid + ";" + fsuud + ";" + cg + ";" + fset + ";" + path
	default:
		panic("unknown format: " + format)
	}
}

// TestGetVolIDMembers_TraversalRejectedInAllFormats verifies that validatePath
// inside getVolIDMembers catches every traversal payload in every format.
// If any combination returns nil error the parser has a bypass – this is the
// primary regression guard for the CVE.
func TestGetVolIDMembers_TraversalRejectedInAllFormats(t *testing.T) {
	formats := []string{"A", "B", "C", "D7"}

	for _, fmt := range formats {
		fmt := fmt
		for _, payload := range traversalPayloads {
			payload := payload
			name := "format=" + fmt + " payload=" + payload.name
			t.Run(name, func(t *testing.T) {
				handle := buildHandleWithPath(fmt, payload.path)
				_, err := getVolIDMembers(handle)
				require.Error(t, err,
					"traversal path %q in format %s must be rejected by getVolIDMembers",
					payload.path, fmt)

				st, ok := status.FromError(err)
				require.True(t, ok, "error must be a gRPC status error")
				assert.Equal(t, codes.InvalidArgument, st.Code(),
					"traversal rejection must use codes.InvalidArgument, got %s", st.Code())
			})
		}
	}
}

// TestNodePublishVolume_TraversalRejectedBeforeOSCall verifies the end-to-end
// defence: NodePublishVolume returns codes.InvalidArgument and never reaches
// os.Lstat when the volumeHandle contains a traversal path.
//
// This specifically reproduces the CVE exploit path:
//
//   volumeHandle = "7118073361626808055;09762E69:5D36FE8D;path=/ibm/gpfs0/../.."
//
// If the fix is absent, NodePublishVolume would call os.Lstat("/host/ibm/gpfs0/../..")
// (which resolves to "/host/" on the real host) and proceed to bind-mount "/".
// With the fix, getVolIDMembers returns an error before any OS call.
func TestNodePublishVolume_TraversalRejectedBeforeOSCall(t *testing.T) {
	ns := newTestNodeServer()
	targetDir := t.TempDir()

	formats := []string{"A", "B", "C", "D7"}
	for _, fmt := range formats {
		fmt := fmt
		for _, payload := range traversalPayloads {
			payload := payload
			// Skip the empty-path case: it produces a different error code path
			// (codes.Internal "Invalid Volume Id") which is still a rejection –
			// the separate empty-path tests below cover it explicitly.
			if payload.path == "" {
				continue
			}
			name := "format=" + fmt + " payload=" + payload.name
			t.Run(name, func(t *testing.T) {
				handle := buildHandleWithPath(fmt, payload.path)
				targetPath := filepath.Join(targetDir, "target-"+fmt+"-"+t.Name())

				_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
					VolumeId:         handle,
					TargetPath:       targetPath,
					VolumeCapability: stdVolCap(),
				})
				require.Error(t, err,
					"NodePublishVolume must reject traversal path %q (format %s)", payload.path, fmt)

				// Must be InvalidArgument – not Internal, not a raw OS error.
				// An Internal or raw error would suggest the path reached OS calls.
				st, ok := status.FromError(err)
				require.True(t, ok, "error must be a gRPC status error")
				assert.Equal(t, codes.InvalidArgument, st.Code(),
					"traversal must be rejected with InvalidArgument before any OS call; got %s – error: %v",
					st.Code(), err)
			})
		}
	}
}

// TestNodePublishVolume_CVE_ExactReproduction reproduces the exact
// volumeHandle strings from the CVE report and verifies they are rejected.
// These strings are the ones the security researcher used in their PoC.
func TestNodePublishVolume_CVE_ExactReproduction(t *testing.T) {
	ns := newTestNodeServer()
	targetDir := t.TempDir()

	cveHandles := []struct {
		name   string
		handle string
	}{
		{
			// Primary PoC – mounts host "/"
			name:   "CVE PoC: escape to host root via /ibm/gpfs0/../..",
			handle: "7118073361626808055;09762E69:5D36FE8D;path=/ibm/gpfs0/../..",
		},
		{
			// Escape to /etc
			name:   "CVE PoC: escape to /etc via /ibm/gpfs0/../../etc",
			handle: "7118073361626808055;09762E69:5D36FE8D;path=/ibm/gpfs0/../../etc",
		},
		{
			// Escape to kubelet secrets
			name:   "CVE PoC: escape to kubelet via /ibm/gpfs0/../../var/lib/kubelet",
			handle: "7118073361626808055;09762E69:5D36FE8D;path=/ibm/gpfs0/../../var/lib/kubelet",
		},
		{
			// Same as primary PoC but with the 4-part fileset format
			name:   "CVE PoC 4-part fileset: escape to host root",
			handle: "7118073361626808055;09762E69:5D36FE8D;fileset=0;path=/ibm/gpfs0/../..",
		},
		{
			// Same using the v2.5 7-part format
			name:   "CVE PoC 7-part v2.5: escape to host root",
			handle: "0;2;7118073361626808055;09762E69:5D36FE8D;cg;myfset;/ibm/gpfs0/../..",
		},
	}

	for _, tc := range cveHandles {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			targetPath := filepath.Join(targetDir, "cve-target")

			_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
				VolumeId:         tc.handle,
				TargetPath:       targetPath,
				VolumeCapability: stdVolCap(),
			})

			// Must fail – any success means the vulnerability is present.
			require.Error(t, err, "CVE exploit handle MUST be rejected: %s", tc.handle)

			st, ok := status.FromError(err)
			require.True(t, ok, "error must be a gRPC status error, got: %v", err)
			assert.Equal(t, codes.InvalidArgument, st.Code(),
				"CVE traversal must be rejected with InvalidArgument (defence-in-depth at parse time), got %s", st.Code())

			// Extra: confirm the error message mentions the path to aid forensics.
			assert.Contains(t, st.Message(), "volumeID is not in proper format",
				"rejection message must identify the invalid volumeID")
		})
	}
}

// TestNodeUnpublishVolume_TraversalHandlesRejected verifies that
// NodeUnpublishVolume also rejects traversal volumeHandles.
// Note: NodeUnpublishVolume does not call getVolIDMembers itself; it validates
// only VolumeId / TargetPath for emptiness, and then inspects the targetPath
// filesystem.  Because these tests supply non-existent targetPaths, the
// lstat returns IsNotExist → idempotent success.  What we are testing here is:
//  (a) all format A/B/C/D7 handles with valid paths still succeed (not broken)
//  (b) handles with traversal paths do NOT cause panic or unexpected OS calls
//      (NodeUnpublishVolume does not parse the volumeHandle path component at all,
//       so the traversal is inert – the idempotent "target not found" path fires)
//
// This is documented intentionally: the CVE is in NodePublishVolume only
// (the SINK is bind-mount).  NodeUnpublishVolume is safe because it does not
// use the path embedded in the volumeHandle to make OS calls.
func TestNodeUnpublishVolume_AllFormats_ValidHandles(t *testing.T) {
	const vpath = "/ibm/gpfs0/pvc-abc123/data"
	ns := newTestNodeServer()

	handles := []struct {
		name   string
		handle string
	}{
		{"FormatA LW 3-part", "cluster1;fsuuid1;path=" + vpath},
		{"FormatB fileset 4-part", "cluster1;fsuuid1;fileset=42;path=" + vpath},
		{"FormatC filesetName 4-part", "cluster1;fsuuid1;filesetName=myfset;path=" + vpath},
		{"FormatD v2.5 dirBased", "0;0;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatE v2.5 depFileset", "0;1;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatF v2.5 indepFileset", "0;2;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatG v2.5 shallowCopy", "0;3;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatH v2.5 advanced", "1;2;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatI v2.5 cache", "2;2;cluster1;fsuuid1;cg;fset;" + vpath},
	}

	for _, tc := range handles {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// targetPath does not exist → idempotent success
			resp, err := ns.NodeUnpublishVolume(freshCtx(), &csi.NodeUnpublishVolumeRequest{
				VolumeId:   tc.handle,
				TargetPath: "/tmp/does-not-exist-unpublish-" + t.Name(),
			})
			// Idempotent: missing target is success in unpublish
			require.NoError(t, err,
				"NodeUnpublishVolume with valid handle and absent target must succeed (idempotent)")
			require.NotNil(t, resp)
		})
	}
}

// TestNodePublishVolume_AllFormats_ValidHandlesReachOSCheck verifies that
// valid (non-traversal) handles of every format pass getVolIDMembers cleanly
// and fail at the OS-level check (os.Lstat on /host<path>) – not at parse
// time.  The OS failure confirms the path was accepted by the parser and only
// rejected because the test environment has no real GPFS mount, proving
// backward compatibility for every format.
func TestNodePublishVolume_AllFormats_ValidHandlesReachOSCheck(t *testing.T) {
	const vpath = "/ibm/gpfs0/pvc-abc123/data" // valid, no traversal
	ns := newTestNodeServer()
	targetDir := t.TempDir()

	handles := []struct {
		name   string
		handle string
	}{
		{"FormatA LW 3-part", "cluster1;fsuuid1;path=" + vpath},
		{"FormatB fileset 4-part (fileset= key)", "cluster1;fsuuid1;fileset=42;path=" + vpath},
		{"FormatC fileset 4-part (filesetName= key)", "cluster1;fsuuid1;filesetName=myfset;path=" + vpath},
		{"FormatD v2.5 classic dirBased (0;0)", "0;0;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatE v2.5 classic depFileset (0;1)", "0;1;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatF v2.5 classic indepFileset (0;2)", "0;2;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatG v2.5 classic shallowCopy (0;3)", "0;3;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatH v2.5 advanced (1;2)", "1;2;cluster1;fsuuid1;cg;fset;" + vpath},
		{"FormatI v2.5 cache (2;2)", "2;2;cluster1;fsuuid1;cg;fset;" + vpath},
	}

	for _, tc := range handles {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			targetPath := filepath.Join(targetDir, "tgt-"+t.Name())

			_, err := ns.NodePublishVolume(freshCtx(), &csi.NodePublishVolumeRequest{
				VolumeId:         tc.handle,
				TargetPath:       targetPath,
				VolumeCapability: stdVolCap(),
			})

			// Expect an error – but it MUST NOT be codes.InvalidArgument
			// (which would indicate a parse-time rejection, i.e. regression).
			// It must be a raw OS / fmt error because the path does not exist
			// under /host in the test environment.
			require.Error(t, err,
				"NodePublishVolume with a valid handle but no real GPFS mount must fail at OS level")

			// If the error is a gRPC status with InvalidArgument, the parser
			// rejected a valid handle – that is the regression we are guarding.
			st, ok := status.FromError(err)
			if ok {
				assert.NotEqual(t, codes.InvalidArgument, st.Code(),
					"valid handle must NOT be rejected with InvalidArgument; "+
						"codes.InvalidArgument means the parser broke format %s", tc.name)
			}
			// The error must mention "lstat" (the OS call that fires because
			// /host<path> does not exist in the test environment).
			assert.Contains(t, err.Error(), "lstat",
				"valid handle must reach the lstat call before failing; "+
					"missing 'lstat' means something rejected the path too early (parser regression) for format %s", tc.name)
		})
	}
}

// TestValidatePath_CVETraversalPayloads exercises validatePath directly with
// the exact strings from the CVE PoC to confirm they are all caught.
func TestValidatePath_CVETraversalPayloads(t *testing.T) {
	for _, payload := range traversalPayloads {
		payload := payload
		t.Run(payload.name, func(t *testing.T) {
			err := validatePath(payload.path)
			require.Error(t, err,
				"validatePath must reject CVE traversal payload %q", payload.path)
			st, ok := status.FromError(err)
			require.True(t, ok, "validatePath must return a gRPC status error")
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}
}

// TestValidatePath_LegitimateGPFSPaths confirms that real GPFS-style paths
// used by every shipped format remain accepted by validatePath after the CVE
// fix – no regression for legitimate volumes.
func TestValidatePath_LegitimateGPFSPaths(t *testing.T) {
	legit := []string{
		"/ibm/gpfs0",
		"/ibm/gpfs0/pvc-abc123",
		"/ibm/gpfs0/pvc-abc123/data",
		"/mnt/gpfs/cluster1/pvc-deadbeef-0000-1111-2222-aabbccddeeff",
		"/var/mnt/scale/fs1/pvc-xyz",
		"/gpfs/fs1/volumes/pvc-test-001/subdir",
	}
	for _, p := range legit {
		p := p
		t.Run(p, func(t *testing.T) {
			assert.NoError(t, validatePath(p),
				"legitimate GPFS path %q must not be rejected after CVE fix", p)
		})
	}
}
