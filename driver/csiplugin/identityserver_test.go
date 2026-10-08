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

// ---------------------------------------------------------------------------
// identityserver_test.go
//
// Unit tests for ScaleIdentityServer:
//   - GetPluginCapabilities
//   - Probe  (3 paths: no-primary-conn, healthy, unhealthy)
//   - GetPluginInfo
//
// Strategy
// --------
//  • ScaleIdentityServer is constructed directly — no real gRPC server needed.
//  • IsNodeComponentHealthy is the only connector method exercised by Probe.
//    We satisfy the full SpectrumScaleConnector interface with
//    fakeConnectorForIdentity, which panics on any method that should never
//    be called in these tests (documents the expected call boundary clearly).
//  • getNodeMapping() reads env-vars; we set them with t.Setenv so cleanup
//    is automatic.
// ---------------------------------------------------------------------------

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Minimal fake connector — only IsNodeComponentHealthy is wired up.
// Every other method panics so accidental calls are caught immediately.
// ---------------------------------------------------------------------------

type fakeConnectorForIdentity struct {
	healthy bool
	err     error
}

func (f *fakeConnectorForIdentity) IsNodeComponentHealthy(_ context.Context, _ string, _ string) (bool, error) {
	return f.healthy, f.err
}

// All remaining SpectrumScaleConnector methods — panic (not under test).
func (f *fakeConnectorForIdentity) GetClusterId(_ context.Context) (string, error) {
	panic("unexpected call: GetClusterId")
}
func (f *fakeConnectorForIdentity) GetClusterSummary(_ context.Context) (connectors.ClusterSummary, error) {
	panic("unexpected call: GetClusterSummary")
}
func (f *fakeConnectorForIdentity) GetTimeZoneOffset(_ context.Context) (string, error) {
	panic("unexpected call: GetTimeZoneOffset")
}
func (f *fakeConnectorForIdentity) GetScaleVersion(_ context.Context) (string, error) {
	panic("unexpected call: GetScaleVersion")
}
func (f *fakeConnectorForIdentity) GetFilesystemMountDetails(_ context.Context, _ string) (connectors.MountInfo, error) {
	panic("unexpected call: GetFilesystemMountDetails")
}
func (f *fakeConnectorForIdentity) IsFilesystemMountedOnGUINode(_ context.Context, _ string) (bool, error) {
	panic("unexpected call: IsFilesystemMountedOnGUINode")
}
func (f *fakeConnectorForIdentity) ListFilesystems(_ context.Context) (map[string]string, error) {
	panic("unexpected call: ListFilesystems")
}
func (f *fakeConnectorForIdentity) GetFilesystemDetails(_ context.Context, _ string) (connectors.FileSystem_v2, error) {
	panic("unexpected call: GetFilesystemDetails")
}
func (f *fakeConnectorForIdentity) GetFilesystemMountpoint(_ context.Context, _ string) (string, error) {
	panic("unexpected call: GetFilesystemMountpoint")
}
func (f *fakeConnectorForIdentity) GetGatewayNode(_ context.Context) (string, error) {
	panic("unexpected call: GetGatewayNode")
}
func (f *fakeConnectorForIdentity) ListGatewayNodes(_ context.Context) ([]string, error) {
	panic("unexpected call: ListGatewayNodes")
}
func (f *fakeConnectorForIdentity) CreateFileset(_ context.Context, _, _, _ string, _ map[string]interface{}, _, _ string, _ map[string]string) error {
	panic("unexpected call: CreateFileset")
}
func (f *fakeConnectorForIdentity) CheckFilesetWithAFMTarget(_ context.Context, _, _ string) (string, error) {
	panic("unexpected call: CheckFilesetWithAFMTarget")
}
func (f *fakeConnectorForIdentity) SetBucketKeys(_ context.Context, _ map[string]string, _ string) error {
	panic("unexpected call: SetBucketKeys")
}
func (f *fakeConnectorForIdentity) DeleteBucketKeys(_ context.Context, _ string) error {
	panic("unexpected call: DeleteBucketKeys")
}
func (f *fakeConnectorForIdentity) DeleteCacheVolumeNodeMapping(_ context.Context, _ string) error {
	panic("unexpected call: DeleteCacheVolumeNodeMapping")
}
func (f *fakeConnectorForIdentity) CreateS3CacheFileset(_ context.Context, _, _, _ string, _ map[string]interface{}, _ map[string]string, _ string, _ *url.URL) error {
	panic("unexpected call: CreateS3CacheFileset")
}
func (f *fakeConnectorForIdentity) CreateCacheVolumeNodeMapping(_ context.Context, _ string, _ string, _, _ map[string]string, _ bool) error {
	panic("unexpected call: CreateCacheVolumeNodeMapping")
}
func (f *fakeConnectorForIdentity) UpdateFileset(_ context.Context, _, _, _ string, _ map[string]interface{}, _ string) error {
	panic("unexpected call: UpdateFileset")
}
func (f *fakeConnectorForIdentity) DeleteFileset(_ context.Context, _, _ string) error {
	panic("unexpected call: DeleteFileset")
}
func (f *fakeConnectorForIdentity) LinkFileset(_ context.Context, _, _, _ string) error {
	panic("unexpected call: LinkFileset")
}
func (f *fakeConnectorForIdentity) UnlinkFileset(_ context.Context, _, _ string, _ bool) error {
	panic("unexpected call: UnlinkFileset")
}
func (f *fakeConnectorForIdentity) ListFileset(_ context.Context, _, _ string) (connectors.Fileset_v2, error) {
	panic("unexpected call: ListFileset")
}
func (f *fakeConnectorForIdentity) ListCSIIndependentFilesets(_ context.Context, _ string) ([]connectors.Fileset_v2, error) {
	panic("unexpected call: ListCSIIndependentFilesets")
}
func (f *fakeConnectorForIdentity) GetFilesetsInodeSpace(_ context.Context, _ string, _ int) ([]connectors.Fileset_v2, error) {
	panic("unexpected call: GetFilesetsInodeSpace")
}
func (f *fakeConnectorForIdentity) IsFilesetLinked(_ context.Context, _, _ string) (bool, error) {
	panic("unexpected call: IsFilesetLinked")
}
func (f *fakeConnectorForIdentity) FilesetRefreshTask(_ context.Context) error {
	panic("unexpected call: FilesetRefreshTask")
}
func (f *fakeConnectorForIdentity) ListFilesetQuota(_ context.Context, _, _ string) (string, error) {
	panic("unexpected call: ListFilesetQuota")
}
func (f *fakeConnectorForIdentity) GetFilesetQuotaDetails(_ context.Context, _, _ string) (connectors.Quota_v2, error) {
	panic("unexpected call: GetFilesetQuotaDetails")
}
func (f *fakeConnectorForIdentity) SetFilesetQuota(_ context.Context, _, _, _, _ string) error {
	panic("unexpected call: SetFilesetQuota")
}
func (f *fakeConnectorForIdentity) CheckIfFSQuotaEnabled(_ context.Context, _ string) error {
	panic("unexpected call: CheckIfFSQuotaEnabled")
}
func (f *fakeConnectorForIdentity) CheckIfFilesetExist(_ context.Context, _, _ string) (bool, error) {
	panic("unexpected call: CheckIfFilesetExist")
}
func (f *fakeConnectorForIdentity) MakeDirectory(_ context.Context, _, _, _, _ string) error {
	panic("unexpected call: MakeDirectory")
}
func (f *fakeConnectorForIdentity) MakeDirectoryV2(_ context.Context, _, _, _, _, _ string) error {
	panic("unexpected call: MakeDirectoryV2")
}
func (f *fakeConnectorForIdentity) MountFilesystem(_ context.Context, _ string, _ []string) error {
	panic("unexpected call: MountFilesystem")
}
func (f *fakeConnectorForIdentity) UnmountFilesystem(_ context.Context, _, _ string) error {
	panic("unexpected call: UnmountFilesystem")
}
func (f *fakeConnectorForIdentity) GetFilesystemName(_ context.Context, _ string) (string, error) {
	panic("unexpected call: GetFilesystemName")
}
func (f *fakeConnectorForIdentity) CheckIfFileDirPresent(_ context.Context, _, _ string) (bool, error) {
	panic("unexpected call: CheckIfFileDirPresent")
}
func (f *fakeConnectorForIdentity) CreateSymLink(_ context.Context, _, _, _, _ string) error {
	panic("unexpected call: CreateSymLink")
}
func (f *fakeConnectorForIdentity) GetFsUid(_ context.Context, _ string) (string, error) {
	panic("unexpected call: GetFsUid")
}
func (f *fakeConnectorForIdentity) DeleteDirectory(_ context.Context, _, _ string, _ bool) error {
	panic("unexpected call: DeleteDirectory")
}
func (f *fakeConnectorForIdentity) StatDirectory(_ context.Context, _, _ string) (string, error) {
	panic("unexpected call: StatDirectory")
}
func (f *fakeConnectorForIdentity) GetFileSetUid(_ context.Context, _, _ string) (string, error) {
	panic("unexpected call: GetFileSetUid")
}
func (f *fakeConnectorForIdentity) GetFileSetNameFromId(_ context.Context, _, _ string) (string, error) {
	panic("unexpected call: GetFileSetNameFromId")
}
func (f *fakeConnectorForIdentity) DeleteSymLnk(_ context.Context, _, _ string) error {
	panic("unexpected call: DeleteSymLnk")
}
func (f *fakeConnectorForIdentity) GetFileSetResponseFromId(_ context.Context, _, _ string) (connectors.Fileset_v2, error) {
	panic("unexpected call: GetFileSetResponseFromId")
}
func (f *fakeConnectorForIdentity) GetFileSetResponseFromName(_ context.Context, _, _ string) (connectors.Fileset_v2, error) {
	panic("unexpected call: GetFileSetResponseFromName")
}
func (f *fakeConnectorForIdentity) SetFilesystemPolicy(_ context.Context, _ *connectors.Policy, _ string) error {
	panic("unexpected call: SetFilesystemPolicy")
}
func (f *fakeConnectorForIdentity) DoesTierExist(_ context.Context, _, _ string) error {
	panic("unexpected call: DoesTierExist")
}
func (f *fakeConnectorForIdentity) GetTierInfoFromName(_ context.Context, _, _ string) (*connectors.StorageTier, error) {
	panic("unexpected call: GetTierInfoFromName")
}
func (f *fakeConnectorForIdentity) GetFirstDataTier(_ context.Context, _ string) (string, error) {
	panic("unexpected call: GetFirstDataTier")
}
func (f *fakeConnectorForIdentity) IsValidNodeclass(_ context.Context, _ string) (bool, error) {
	panic("unexpected call: IsValidNodeclass")
}
func (f *fakeConnectorForIdentity) IsSnapshotSupported(_ context.Context) (bool, error) {
	panic("unexpected call: IsSnapshotSupported")
}
func (f *fakeConnectorForIdentity) CheckIfDefaultPolicyPartitionExists(_ context.Context, _, _ string) bool {
	panic("unexpected call: CheckIfDefaultPolicyPartitionExists")
}
func (f *fakeConnectorForIdentity) WaitForJobCompletion(_ context.Context, _ int, _ uint64) error {
	panic("unexpected call: WaitForJobCompletion")
}
func (f *fakeConnectorForIdentity) WaitForJobCompletionWithResp(_ context.Context, _ int, _ uint64) (connectors.GenericResponse, error) {
	panic("unexpected call: WaitForJobCompletionWithResp")
}
func (f *fakeConnectorForIdentity) CreateSnapshot(_ context.Context, _, _, _ string) error {
	panic("unexpected call: CreateSnapshot")
}
func (f *fakeConnectorForIdentity) DeleteSnapshot(_ context.Context, _, _, _ string) error {
	panic("unexpected call: DeleteSnapshot")
}
func (f *fakeConnectorForIdentity) CreateSnapshotCloneCopy(_ context.Context, _, _, _, _, _, _, _ string) error {
	panic("unexpected call: CreateSnapshotCloneCopy")
}
func (f *fakeConnectorForIdentity) CreateSnapshotCloneSplit(_ context.Context, _, _ string) error {
	panic("unexpected call: CreateSnapshotCloneSplit")
}
func (f *fakeConnectorForIdentity) GetSnapshotCloneChild(_ context.Context, _, _, _, _ string) (string, error) {
	panic("unexpected call: GetSnapshotCloneChild")
}
func (f *fakeConnectorForIdentity) GetLatestFilesetSnapshots(_ context.Context, _, _ string) ([]connectors.Snapshot_v2, error) {
	panic("unexpected call: GetLatestFilesetSnapshots")
}
func (f *fakeConnectorForIdentity) GetSnapshotUid(_ context.Context, _, _, _ string) (string, error) {
	panic("unexpected call: GetSnapshotUid")
}
func (f *fakeConnectorForIdentity) GetSnapshotCreateTimestamp(_ context.Context, _, _, _ string) (string, error) {
	panic("unexpected call: GetSnapshotCreateTimestamp")
}
func (f *fakeConnectorForIdentity) CheckIfSnapshotExist(_ context.Context, _, _, _ string) (bool, error) {
	panic("unexpected call: CheckIfSnapshotExist")
}
func (f *fakeConnectorForIdentity) ListFilesetSnapshots(_ context.Context, _, _ string) ([]connectors.Snapshot_v2, error) {
	panic("unexpected call: ListFilesetSnapshots")
}
func (f *fakeConnectorForIdentity) CopyFsetSnapshotPath(_ context.Context, _, _, _, _, _ string, _ string) (int, uint64, error) {
	panic("unexpected call: CopyFsetSnapshotPath")
}
func (f *fakeConnectorForIdentity) CopyFilesetPath(_ context.Context, _, _, _, _ string, _ string) (int, uint64, error) {
	panic("unexpected call: CopyFilesetPath")
}
func (f *fakeConnectorForIdentity) CopyDirectoryPath(_ context.Context, _, _, _ string, _ string) (int, uint64, error) {
	panic("unexpected call: CopyDirectoryPath")
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestIdentityServer creates a ScaleIdentityServer backed by a minimal
// ScaleDriver. connmap is optional — pass nil to simulate no primary connection.
func newTestIdentityServer(nodeID string, conn connectors.SpectrumScaleConnector) *ScaleIdentityServer {
	d := &ScaleDriver{
		name:          "ibm-spectrum-scale-csi",
		vendorVersion: "2.13.0",
		nodeID:        nodeID,
		connmap:       map[string]connectors.SpectrumScaleConnector{},
		cmap:          settings.ScaleSettingsConfigMap{},
	}
	if conn != nil {
		d.connmap["primary"] = conn
	}
	return &ScaleIdentityServer{Driver: d}
}

// ---------------------------------------------------------------------------
// GetPluginCapabilities
// ---------------------------------------------------------------------------

// TestGetPluginCapabilities_AdvertisesControllerService verifies that the
// response contains exactly one capability: CONTROLLER_SERVICE.
// This is the single fixed response the method always returns — no branching.
func TestGetPluginCapabilities_AdvertisesControllerService(t *testing.T) {
	is := newTestIdentityServer("node-01", nil)
	resp, err := is.GetPluginCapabilities(context.Background(), &csi.GetPluginCapabilitiesRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Capabilities, 1,
		"expected exactly one plugin capability")

	svc := resp.Capabilities[0].GetService()
	require.NotNil(t, svc, "capability must be of Service type")
	assert.Equal(t,
		csi.PluginCapability_Service_CONTROLLER_SERVICE,
		svc.GetType(),
		"must advertise CONTROLLER_SERVICE")
}

// TestGetPluginCapabilities_NilRequestAccepted confirms the method does not
// panic when given a nil request (the gRPC framework may pass nil for empty
// proto messages).
func TestGetPluginCapabilities_NilRequestAccepted(t *testing.T) {
	is := newTestIdentityServer("node-01", nil)
	resp, err := is.GetPluginCapabilities(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Capabilities)
}

// ---------------------------------------------------------------------------
// Probe
// ---------------------------------------------------------------------------

// TestProbe_NoPrimaryConnection covers the path where connmap has no "primary"
// key. Probe must return Ready=true and no error — the driver cannot self-heal
// by restarting, so it stays ready.
func TestProbe_NoPrimaryConnection(t *testing.T) {
	is := newTestIdentityServer("node-01", nil) // no "primary" in connmap
	resp, err := is.Probe(context.Background(), &csi.ProbeRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Ready.Value,
		"Probe must report Ready=true even when primary connection is absent")
}

// TestProbe_NilPrimaryConnection covers the edge case where connmap["primary"]
// is explicitly set to nil (e.g. after a failed connector initialisation that
// stored nil). The !ok || conn == nil guard in Probe must catch this.
func TestProbe_NilPrimaryConnection(t *testing.T) {
	d := &ScaleDriver{
		name:          "ibm-spectrum-scale-csi",
		vendorVersion: "2.13.0",
		nodeID:        "node-01",
		connmap:       map[string]connectors.SpectrumScaleConnector{"primary": nil},
		cmap:          settings.ScaleSettingsConfigMap{},
	}
	is := &ScaleIdentityServer{Driver: d}

	resp, err := is.Probe(context.Background(), &csi.ProbeRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Ready.Value,
		"Probe must report Ready=true when primary connector is nil")
}

// TestProbe_HealthyNode exercises the happy path: primary connector exists and
// IsNodeComponentHealthy returns (true, nil). Probe must return Ready=true.
func TestProbe_HealthyNode(t *testing.T) {
	fake := &fakeConnectorForIdentity{healthy: true, err: nil}
	is := newTestIdentityServer("node-01", fake)

	resp, err := is.Probe(context.Background(), &csi.ProbeRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Ready.Value,
		"Probe must report Ready=true when GPFS is healthy")
}

// TestProbe_UnhealthyNode exercises the failure path: IsNodeComponentHealthy
// returns (false, err). Per the implementation's design, Probe still returns
// Ready=true (restarting the CSI pod would not fix a GPFS health problem).
func TestProbe_UnhealthyNode(t *testing.T) {
	fake := &fakeConnectorForIdentity{
		healthy: false,
		err:     fmt.Errorf("gpfs daemon not running"),
	}
	is := newTestIdentityServer("node-01", fake)

	resp, err := is.Probe(context.Background(), &csi.ProbeRequest{})

	require.NoError(t, err, "Probe must not surface the connector error to the caller")
	require.NotNil(t, resp)
	assert.True(t, resp.Ready.Value,
		"Probe must still report Ready=true even when GPFS is unhealthy")
}

// TestProbe_UnhealthyNode_NoError verifies the same Ready=true contract when
// IsNodeComponentHealthy returns (false, nil) — unhealthy state with no
// accompanying error object.
func TestProbe_UnhealthyNode_NoError(t *testing.T) {
	fake := &fakeConnectorForIdentity{healthy: false, err: nil}
	is := newTestIdentityServer("node-01", fake)

	resp, err := is.Probe(context.Background(), &csi.ProbeRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Ready.Value)
}

// TestProbe_NodeMappingApplied verifies that getNodeMapping() is exercised on
// the nodeID before the connector call. We set the SCALE_NODE_MAPPING_PREFIX
// env var and a mapped value so the translated name reaches the fake connector
// (which accepts any nodeName — this test just ensures no panic and the right
// Ready value).
func TestProbe_NodeMappingApplied(t *testing.T) {
	// Map k8s node "k8s-worker-0" → gpfs name "gpfs-node-0" via the prefix env var.
	t.Setenv("SCALE_NODE_MAPPING_PREFIX", "K8S_")
	t.Setenv("K8S_k8s-worker-0", "gpfs-node-0")

	fake := &fakeConnectorForIdentity{healthy: true}
	is := newTestIdentityServer("k8s-worker-0", fake)

	resp, err := is.Probe(context.Background(), &csi.ProbeRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Ready.Value)
}

// ---------------------------------------------------------------------------
// GetPluginInfo
// ---------------------------------------------------------------------------

// TestGetPluginInfo_ReturnsNameAndVersion is the happy-path test: a driver
// with both name and version set must return them verbatim.
func TestGetPluginInfo_ReturnsNameAndVersion(t *testing.T) {
	is := newTestIdentityServer("node-01", nil)
	// Override to known values.
	is.Driver.name = "ibm-spectrum-scale-csi"
	is.Driver.vendorVersion = "2.13.0"

	resp, err := is.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "ibm-spectrum-scale-csi", resp.Name)
	assert.Equal(t, "2.13.0", resp.VendorVersion)
}

// TestGetPluginInfo_EmptyName verifies that when Driver.name is the empty
// string GetPluginInfo returns a gRPC Unavailable error — the spec requires
// every CSI driver to have a non-empty name.
func TestGetPluginInfo_EmptyName(t *testing.T) {
	is := newTestIdentityServer("node-01", nil)
	is.Driver.name = "" // simulate mis-configured driver

	resp, err := is.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})

	require.Nil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "error must be a gRPC status error")
	assert.Equal(t, codes.Unavailable, st.Code())
	assert.Contains(t, st.Message(), "Driver name not configured")
}

// TestGetPluginInfo_EmptyVersion verifies that an empty vendorVersion is
// allowed (the spec only mandates a non-empty name).
func TestGetPluginInfo_EmptyVersion(t *testing.T) {
	is := newTestIdentityServer("node-01", nil)
	is.Driver.name = "ibm-spectrum-scale-csi"
	is.Driver.vendorVersion = ""

	resp, err := is.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	require.NoError(t, err)
	assert.Equal(t, "ibm-spectrum-scale-csi", resp.Name)
	assert.Equal(t, "", resp.VendorVersion)
}

// TestGetPluginInfo_TableDriven exercises multiple name/version combinations
// in a table to ensure consistent behaviour across edge cases.
func TestGetPluginInfo_TableDriven(t *testing.T) {
	cases := []struct {
		name          string
		driverName    string
		vendorVersion string
		wantErr       bool
		wantCode      codes.Code
	}{
		{
			name:          "normal name and version",
			driverName:    "ibm-spectrum-scale-csi",
			vendorVersion: "2.13.0",
			wantErr:       false,
		},
		{
			name:          "version with pre-release suffix",
			driverName:    "ibm-spectrum-scale-csi",
			vendorVersion: "2.13.0-alpha.1",
			wantErr:       false,
		},
		{
			name:          "empty driver name returns Unavailable",
			driverName:    "",
			vendorVersion: "2.13.0",
			wantErr:       true,
			wantCode:      codes.Unavailable,
		},
		{
			name:          "empty version is allowed",
			driverName:    "ibm-spectrum-scale-csi",
			vendorVersion: "",
			wantErr:       false,
		},
		{
			name:          "non-standard driver name accepted",
			driverName:    "custom.driver/name",
			vendorVersion: "1.0.0",
			wantErr:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			is := newTestIdentityServer("node-01", nil)
			is.Driver.name = tc.driverName
			is.Driver.vendorVersion = tc.vendorVersion

			resp, err := is.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})

			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tc.driverName, resp.Name)
				assert.Equal(t, tc.vendorVersion, resp.VendorVersion)
			}
		})
	}
}
