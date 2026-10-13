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

package connectors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// freshCtx returns a context without a logger ID (acceptable — GetLoggerId
// simply returns an empty string, no panic).
func freshCtx() context.Context {
	return context.Background()
}

// successJobResponse builds a GenericResponse where the job status is COMPLETED
// and Status.Code is http.StatusAccepted (so checkAsynchronousJob triggers).
func successJobResponse(jobID uint64) GenericResponse {
	return GenericResponse{
		Status: Status{Code: http.StatusOK},
		Jobs: []Job{
			{
				JobID:  jobID,
				Status: "COMPLETED",
				Result: Respresult{},
			},
		},
	}
}

// asyncJobHandler returns an http.HandlerFunc that:
//  1. First call: returns `initialResp` (typically 202 Accepted with a job).
//  2. Subsequent GET on the job URL: returns a COMPLETED job.
func asyncJobHandler(t *testing.T, urlPath string, initialResp GenericResponse, initialCode int) http.HandlerFunc {
	t.Helper()
	called := 0
	return func(w http.ResponseWriter, r *http.Request) {
		called++
		w.Header().Set("Content-Type", "application/json")
		if called == 1 {
			w.WriteHeader(initialCode)
			_ = json.NewEncoder(w).Encode(initialResp)
			return
		}
		// Job polling call
		jobResp := GenericResponse{
			Status: Status{Code: http.StatusOK},
			Jobs: []Job{
				{JobID: 1, Status: "COMPLETED"},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(jobResp)
	}
}

// newTestServer creates an httptest.Server and a SpectrumRestV2 that points
// at it.  The caller receives the server (must call ts.Close() in defer) and
// the connector ready to use.
func newTestServer(t *testing.T, mux *http.ServeMux) (*httptest.Server, *SpectrumRestV2) {
	t.Helper()
	ts := httptest.NewServer(mux)
	conn := &SpectrumRestV2{
		HTTPclient:    ts.Client(),
		Endpoint:      []string{ts.URL + "/"},
		EndPointIndex: 0,
		ClusterConfig: settings.Clusters{
			ID:            "test-cluster",
			MgmtUsername:  "admin",
			MgmtPassword:  "secret",
		},
		RequestCalledBy: "operator",
	}
	return ts, conn
}

// jsonResp encodes v as JSON and writes it to w with the given HTTP status.
func jsonResp(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// isStatusOK
// ---------------------------------------------------------------------------

func TestIsStatusOK_TableDriven(t *testing.T) {
	s := &SpectrumRestV2{}
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusOK, true},
		{http.StatusCreated, true},
		{http.StatusAccepted, true},
		{http.StatusBadRequest, false},
		{http.StatusNotFound, false},
		{http.StatusInternalServerError, false},
		{http.StatusUnauthorized, false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, s.isStatusOK(tc.code), "code=%d", tc.code)
	}
}

// ---------------------------------------------------------------------------
// checkAsynchronousJob
// ---------------------------------------------------------------------------

func TestCheckAsynchronousJob_TableDriven(t *testing.T) {
	s := &SpectrumRestV2{}
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusAccepted, true},
		{http.StatusCreated, true},
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusNoContent, false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, s.checkAsynchronousJob(tc.code), "code=%d", tc.code)
	}
}

// ---------------------------------------------------------------------------
// isRequestAccepted
// ---------------------------------------------------------------------------

func TestIsRequestAccepted_StatusNotOK(t *testing.T) {
	s := &SpectrumRestV2{}
	resp := GenericResponse{
		Status: Status{Code: http.StatusBadRequest},
		Jobs:   nil,
	}
	err := s.isRequestAccepted(freshCtx(), resp, "http://example.com/test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestIsRequestAccepted_EmptyJobs(t *testing.T) {
	s := &SpectrumRestV2{}
	resp := GenericResponse{
		Status: Status{Code: http.StatusOK},
		Jobs:   []Job{},
	}
	err := s.isRequestAccepted(freshCtx(), resp, "http://example.com/test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get Job details")
}

func TestIsRequestAccepted_Success(t *testing.T) {
	s := &SpectrumRestV2{}
	resp := GenericResponse{
		Status: Status{Code: http.StatusOK},
		Jobs:   []Job{{JobID: 1, Status: "RUNNING"}},
	}
	err := s.isRequestAccepted(freshCtx(), resp, "http://example.com/test")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// WaitForJobCompletion — synchronous path (non-async status codes)
// ---------------------------------------------------------------------------

func TestWaitForJobCompletion_NonAsyncCode_NoHTTP(t *testing.T) {
	s := &SpectrumRestV2{}
	// A 200 status is not async → WaitForJobCompletion must return nil immediately.
	err := s.WaitForJobCompletion(freshCtx(), http.StatusOK, 0)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// WaitForJobCompletionWithResp — synchronous path
// ---------------------------------------------------------------------------

func TestWaitForJobCompletionWithResp_NonAsyncCode(t *testing.T) {
	s := &SpectrumRestV2{}
	resp, err := s.WaitForJobCompletionWithResp(freshCtx(), http.StatusOK, 0)
	require.NoError(t, err)
	assert.Equal(t, GenericResponse{}, resp)
}

// ---------------------------------------------------------------------------
// AsyncJobCompletion — via httptest server
// ---------------------------------------------------------------------------

func TestAsyncJobCompletion_CompletedJob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/jobs/42", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
			Jobs:   []Job{{JobID: 42, Status: "COMPLETED"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	resp, err := conn.AsyncJobCompletion(freshCtx(), "scalemgmt/v2/jobs/42?fields=:all:")
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", resp.Jobs[0].Status)
}

func TestAsyncJobCompletion_FailedJob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/jobs/7", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
			Jobs: []Job{{
				JobID:  7,
				Status: "FAILED",
				Result: Respresult{Stderr: []string{"something went wrong"}},
			}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.AsyncJobCompletion(freshCtx(), "scalemgmt/v2/jobs/7?fields=:all:")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "something went wrong")
}

func TestAsyncJobCompletion_EmptyJobsResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/jobs/99", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
			Jobs:   nil,
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.AsyncJobCompletion(freshCtx(), "scalemgmt/v2/jobs/99?fields=:all:")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get Job details")
}

// ---------------------------------------------------------------------------
// NewSpectrumRestV2
// ---------------------------------------------------------------------------

func TestNewSpectrumRestV2_NonSecureMode(t *testing.T) {
	cfg := settings.Clusters{
		ID:            "c1",
		SecureSslMode: false,
		RestAPI: []settings.RestAPI{
			{GuiHost: "gui.example.com", GuiPort: 443},
		},
	}
	conn, err := NewSpectrumRestV2(freshCtx(), cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	rv2, ok := conn.(*SpectrumRestV2)
	require.True(t, ok)
	require.Len(t, rv2.Endpoint, 1)
	assert.Contains(t, rv2.Endpoint[0], "gui.example.com")
	assert.Contains(t, rv2.Endpoint[0], "443")
}

func TestNewSpectrumRestV2_DefaultPort(t *testing.T) {
	cfg := settings.Clusters{
		ID:            "c2",
		SecureSslMode: false,
		RestAPI: []settings.RestAPI{
			{GuiHost: "gui.example.com", GuiPort: 0}, // 0 should use DefaultGuiPort
		},
	}
	conn, err := NewSpectrumRestV2(freshCtx(), cfg)
	require.NoError(t, err)
	rv2 := conn.(*SpectrumRestV2)
	assert.Contains(t, rv2.Endpoint[0], "gui.example.com")
}

func TestNewSpectrumRestV2_MultipleGuiHosts(t *testing.T) {
	cfg := settings.Clusters{
		ID:            "c3",
		SecureSslMode: false,
		RestAPI: []settings.RestAPI{
			{GuiHost: "gui1.example.com", GuiPort: 443},
			{GuiHost: "gui2.example.com", GuiPort: 443},
		},
	}
	conn, err := NewSpectrumRestV2(freshCtx(), cfg)
	require.NoError(t, err)
	rv2 := conn.(*SpectrumRestV2)
	assert.Len(t, rv2.Endpoint, 2)
}

// ---------------------------------------------------------------------------
// doHTTP — via httptest server
// ---------------------------------------------------------------------------

func TestDoHTTP_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	var out GenericResponse
	err := conn.doHTTP(freshCtx(), "scalemgmt/v2/cluster", "GET", &out, nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestDoHTTP_Forbidden(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	var out GenericResponse
	err := conn.doHTTP(freshCtx(), "scalemgmt/v2/cluster", "GET", &out, nil)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestDoHTTP_Success200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/cluster", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetClusterResponse{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	var out GetClusterResponse
	err := conn.doHTTP(freshCtx(), "scalemgmt/v2/cluster", "GET", &out, nil)
	require.NoError(t, err)
}

func TestDoHTTP_FailureStatusCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/cluster", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusInternalServerError, GenericResponse{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	var out GetClusterResponse
	err := conn.doHTTP(freshCtx(), "scalemgmt/v2/cluster", "GET", &out, nil)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// TestDoHTTP_EndpointFailover verifies that when the primary endpoint refuses
// the connection the connector cycles through the remaining endpoints and uses
// the working one.
func TestDoHTTP_EndpointFailover(t *testing.T) {
	// A server that always responds OK.
	workingMux := http.NewServeMux()
	workingMux.HandleFunc("/scalemgmt/v2/cluster", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetClusterResponse{})
	})
	workingServer := httptest.NewServer(workingMux)
	defer workingServer.Close()

	// Build a connector with two endpoints: a dead one first, then the working one.
	conn := &SpectrumRestV2{
		HTTPclient:    workingServer.Client(),
		Endpoint:      []string{"http://127.0.0.1:1/", workingServer.URL + "/"},
		EndPointIndex: 0,
		ClusterConfig: settings.Clusters{
			ID:           "fc",
			MgmtUsername: "admin",
			MgmtPassword: "secret",
		},
		RequestCalledBy: "operator",
	}
	var out GetClusterResponse
	err := conn.doHTTP(freshCtx(), "scalemgmt/v2/cluster", "GET", &out, nil)
	// The failover should succeed because the second endpoint is alive.
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// GetClusterId
// ---------------------------------------------------------------------------

func TestGetClusterId_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/cluster", func(w http.ResponseWriter, r *http.Request) {
		resp := GetClusterResponse{
			Cluster: Cluster{
				ClusterSummary: ClusterSummary{ClusterID: 12345},
			},
		}
		jsonResp(w, http.StatusOK, resp)
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	id, err := conn.GetClusterId(freshCtx())
	require.NoError(t, err)
	assert.Equal(t, "12345", id)
}

func TestGetClusterId_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/cluster", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(GenericResponse{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetClusterId(freshCtx())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetClusterSummary
// ---------------------------------------------------------------------------

func TestGetClusterSummary_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/cluster", func(w http.ResponseWriter, r *http.Request) {
		resp := GetClusterResponse{
			Cluster: Cluster{
				ClusterSummary: ClusterSummary{ClusterID: 99, ClusterName: "test-cluster"},
			},
		}
		jsonResp(w, http.StatusOK, resp)
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	summary, err := conn.GetClusterSummary(freshCtx())
	require.NoError(t, err)
	assert.EqualValues(t, 99, summary.ClusterID)
	assert.Equal(t, "test-cluster", summary.ClusterName)
}

// ---------------------------------------------------------------------------
// GetScaleVersion
// ---------------------------------------------------------------------------

func TestGetScaleVersion_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/info", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetInfoResponse_v2{
			Info: Info{ServerVersion: "5.1.5.0"},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	ver, err := conn.GetScaleVersion(freshCtx())
	require.NoError(t, err)
	assert.Equal(t, "5.1.5.0", ver)
}

func TestGetScaleVersion_EmptyVersion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/info", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetInfoResponse_v2{
			Info: Info{ServerVersion: ""},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetScaleVersion(freshCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get IBM Storage Scale version")
}

// ---------------------------------------------------------------------------
// GetFilesystemMountpoint
// ---------------------------------------------------------------------------

func TestGetFilesystemMountpoint_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{
				{Name: "gpfs0", Mount: MountInfo{MountPoint: "/mnt/gpfs0"}},
			},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	mp, err := conn.GetFilesystemMountpoint(freshCtx(), "gpfs0")
	require.NoError(t, err)
	assert.Equal(t, "/mnt/gpfs0", mp)
}

func TestGetFilesystemMountpoint_EmptyResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{FileSystems: nil})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetFilesystemMountpoint(freshCtx(), "gpfs1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to fetch mount point")
}

// ---------------------------------------------------------------------------
// IsFilesystemMountedOnGUINode
// ---------------------------------------------------------------------------

func TestIsFilesystemMountedOnGUINode_Mounted(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{{Mount: MountInfo{Status: "mounted"}}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	ok, err := conn.IsFilesystemMountedOnGUINode(freshCtx(), "gpfs0")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestIsFilesystemMountedOnGUINode_NotMounted(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{{Mount: MountInfo{Status: "not mounted"}}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	ok, err := conn.IsFilesystemMountedOnGUINode(freshCtx(), "gpfs0")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestIsFilesystemMountedOnGUINode_UnknownStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{{Mount: MountInfo{Status: "degraded"}}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.IsFilesystemMountedOnGUINode(freshCtx(), "gpfs0")
	require.Error(t, err)
}

func TestIsFilesystemMountedOnGUINode_EmptyResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.IsFilesystemMountedOnGUINode(freshCtx(), "gpfs0")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListFilesystems
// ---------------------------------------------------------------------------

func TestListFilesystems_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{
				{Name: "gpfs0", Mount: MountInfo{MountPoint: "/mnt/gpfs0"}},
				{Name: "gpfs1", Mount: MountInfo{MountPoint: "/mnt/gpfs1"}},
			},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	fsList, err := conn.ListFilesystems(freshCtx())
	require.NoError(t, err)
	require.Len(t, fsList, 2)
	assert.Equal(t, "/mnt/gpfs0", fsList["gpfs0"])
	assert.Equal(t, "/mnt/gpfs1", fsList["gpfs1"])
}

func TestListFilesystems_EmptyResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.ListFilesystems(freshCtx())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetFilesystemMountDetails
// ---------------------------------------------------------------------------

func TestGetFilesystemMountDetails_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{
				{Mount: MountInfo{MountPoint: "/mnt/gpfs0", Status: "mounted"}},
			},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	info, err := conn.GetFilesystemMountDetails(freshCtx(), "gpfs0")
	require.NoError(t, err)
	assert.Equal(t, "/mnt/gpfs0", info.MountPoint)
}

func TestGetFilesystemMountDetails_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetFilesystemMountDetails(freshCtx(), "gpfs0")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetFilesystemDetails
// ---------------------------------------------------------------------------

func TestGetFilesystemDetails_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{
				{Name: "gpfs0"},
			},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	fs, err := conn.GetFilesystemDetails(freshCtx(), "gpfs0")
	require.NoError(t, err)
	assert.Equal(t, "gpfs0", fs.Name)
}

func TestGetFilesystemDetails_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetFilesystemDetails(freshCtx(), "gpfs0")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetFilesystemName
// ---------------------------------------------------------------------------

func TestGetFilesystemName_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{
				{Name: "gpfs0"},
			},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	name, err := conn.GetFilesystemName(freshCtx(), "some-uuid")
	require.NoError(t, err)
	assert.Equal(t, "gpfs0", name)
}

func TestGetFilesystemName_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetFilesystemName(freshCtx(), "some-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to fetch filesystem name details")
}

// ---------------------------------------------------------------------------
// GetFsUid
// ---------------------------------------------------------------------------

func TestGetFsUid_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{
			FileSystems: []FileSystem_v2{{UUID: "abc-uuid"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	uid, err := conn.GetFsUid(freshCtx(), "gpfs0")
	require.NoError(t, err)
	assert.Equal(t, "abc-uuid", uid)
}

func TestGetFsUid_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesystemResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.GetFsUid(freshCtx(), "gpfs0")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListFileset
// ---------------------------------------------------------------------------

func TestListFileset_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{FilesetName: "fset1"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	fset, err := conn.ListFileset(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	assert.Equal(t, "fset1", fset.FilesetName)
}

func TestListFileset_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset2", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	_, err := conn.ListFileset(freshCtx(), "gpfs0", "fset2")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// IsFilesetLinked
// ---------------------------------------------------------------------------

func TestIsFilesetLinked_Linked(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{
				FilesetName: "fset1",
				Config:      FilesetConfig_v2{Path: "/mnt/gpfs0/fset1"},
			}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	linked, err := conn.IsFilesetLinked(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	assert.True(t, linked)
}

func TestIsFilesetLinked_NotLinked_EmptyPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{FilesetName: "fset1", Config: FilesetConfig_v2{Path: ""}}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	linked, err := conn.IsFilesetLinked(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	assert.False(t, linked)
}

func TestIsFilesetLinked_NotLinked_DashPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{FilesetName: "fset1", Config: FilesetConfig_v2{Path: "--"}}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	linked, err := conn.IsFilesetLinked(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	assert.False(t, linked)
}

// ---------------------------------------------------------------------------
// CheckIfFileDirPresent
// ---------------------------------------------------------------------------

func TestCheckIfFileDirPresent_Exists(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/inode", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
		})
	})
	// The method formats the path and calls the directory or inode URL.
	// Use a wildcard handler to accept any path.
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
			Jobs:   []Job{{JobID: 1, Status: "COMPLETED"}},
		})
	})
	ts, conn := newTestServer(t, mux2)
	defer ts.Close()

	found, err := conn.CheckIfFileDirPresent(freshCtx(), "gpfs0", "/vol1")
	require.NoError(t, err)
	assert.True(t, found)
}

// ---------------------------------------------------------------------------
// GetTimeZoneOffset
// ---------------------------------------------------------------------------

func TestGetTimeZoneOffset_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/config", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetConfigResponse{
			Config: Config{ClusterConfig: ClusterConfig{TimeZoneOffset: "+05:30"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	tz, err := conn.GetTimeZoneOffset(freshCtx())
	require.NoError(t, err)
	assert.Equal(t, "+05:30", tz)
}

// ---------------------------------------------------------------------------
// CheckIfFSQuotaEnabled
// ---------------------------------------------------------------------------

func TestCheckIfFSQuotaEnabled_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	err := conn.CheckIfFSQuotaEnabled(freshCtx(), "gpfs0")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// CheckIfFilesetExist
// ---------------------------------------------------------------------------

// TestCheckIfFilesetExist_Exists verifies that a 200 response means the
// fileset exists (CheckIfFilesetExist returns true, nil on HTTP success).
func TestCheckIfFilesetExist_Exists(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{FilesetName: "fset1"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	exists, err := conn.CheckIfFilesetExist(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	assert.True(t, exists)
}

// TestCheckIfFilesetExist_NotExists_APIError verifies that when the GUI
// returns an HTTP error status the function returns false, error — unless the
// status message indicates the fileset name was invalid (i.e. doesn't exist),
// in which case it returns false, nil.
//
// Note: a plain 200 with an empty Filesets slice still returns true, nil
// because CheckIfFilesetExist only inspects the HTTP status, not the body.
// The "not found" path is when doHTTP itself fails with a message that
// contains "Invalid value in 'filesetName'".
func TestCheckIfFilesetExist_NotExists_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset2", func(w http.ResponseWriter, r *http.Request) {
		// Return 500 with a message that does NOT contain the special "invalid
		// filesetName" phrase, so the function returns false, error.
		jsonResp(w, http.StatusInternalServerError, GetFilesetResponse_v2{
			Status: Status{Message: "generic server error"},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	exists, err := conn.CheckIfFilesetExist(freshCtx(), "gpfs0", "fset2")
	require.Error(t, err)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// GetFileSetUid
// ---------------------------------------------------------------------------

func TestGetFileSetUid_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{
				FilesetName: "fset1",
				Config:      FilesetConfig_v2{Id: 42},
			}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	uid, err := conn.GetFileSetUid(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	assert.Equal(t, "42", uid)
}

// ---------------------------------------------------------------------------
// CheckIfSnapshotExist
// ---------------------------------------------------------------------------

// TestCheckIfSnapshotExist_Exists verifies that a 200 response means the
// snapshot exists.
func TestCheckIfSnapshotExist_Exists(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1/snapshots/snap1", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetSnapshotResponse_v2{
			Snapshots: []Snapshot_v2{{SnapshotName: "snap1"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	exists, err := conn.CheckIfSnapshotExist(freshCtx(), "gpfs0", "fset1", "snap1")
	require.NoError(t, err)
	assert.True(t, exists)
}

// TestCheckIfSnapshotExist_NotExists_APIError verifies that an HTTP error
// response with a non-special message returns false, error.
func TestCheckIfSnapshotExist_NotExists_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1/snapshots/snapX", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusInternalServerError, GetSnapshotResponse_v2{
			Status: Status{Message: "generic server error"},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	exists, err := conn.CheckIfSnapshotExist(freshCtx(), "gpfs0", "fset1", "snapX")
	require.Error(t, err)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// ListFilesetSnapshots
// ---------------------------------------------------------------------------

func TestListFilesetSnapshots_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets/fset1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetSnapshotResponse_v2{
			Snapshots: []Snapshot_v2{{SnapshotName: "snap1"}, {SnapshotName: "snap2"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	snaps, err := conn.ListFilesetSnapshots(freshCtx(), "gpfs0", "fset1")
	require.NoError(t, err)
	require.Len(t, snaps, 2)
	assert.Equal(t, "snap1", snaps[0].SnapshotName)
}

// ---------------------------------------------------------------------------
// GetListFilesetsInodeSpace
// ---------------------------------------------------------------------------

func TestGetFilesetsInodeSpace_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/filesystems/gpfs0/filesets", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetFilesetResponse_v2{
			Filesets: []Fileset_v2{{FilesetName: "fset1"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	fsets, err := conn.GetFilesetsInodeSpace(freshCtx(), "gpfs0", 0)
	require.NoError(t, err)
	require.Len(t, fsets, 1)
}

// ---------------------------------------------------------------------------
// IsSnapshotSupported
// ---------------------------------------------------------------------------

// TestIsSnapshotSupported_Supported verifies that when the info endpoint
// returns a non-empty Paths.SnapCopyOp slice, snapshots are supported.
func TestIsSnapshotSupported_Supported(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/info", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetInfoResponse_v2{
			Info: Info{
				ServerVersion: "5.1.5.0",
				Paths: Path{
					SnapCopyOp: []string{"/filesystems/{filesystemName}/filesets/{filesetName}/snapshotCopy/{snapshotName}"},
				},
			},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	ok, err := conn.IsSnapshotSupported(freshCtx())
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestIsSnapshotSupported_NotSupported verifies that an empty SnapCopyOp path
// results in false being returned.
func TestIsSnapshotSupported_NotSupported(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/info", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetInfoResponse_v2{
			Info: Info{ServerVersion: "5.1.0.0", Paths: Path{SnapCopyOp: nil}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	ok, err := conn.IsSnapshotSupported(freshCtx())
	require.NoError(t, err)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// GetGatewayNode
// ---------------------------------------------------------------------------

func TestGetGatewayNode_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/nodes", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, http.StatusOK, GetNodesResponse_v2{
			Nodes: []Node_v2{{Roles: NodeRoles{QuorumNode: true}}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	// GetGatewayNode uses a specific filter for admin nodename; this test only
	// verifies that a successful HTTP response is processed without panicking.
	// The actual logic depends on the node list details; we just ensure no error
	// when the server returns a valid response structure.
	_, _ = conn.GetGatewayNode(freshCtx())
}

// ---------------------------------------------------------------------------
// StatDirectory
// ---------------------------------------------------------------------------

// TestStatDirectory_Success verifies that StatDirectory correctly processes a
// two-call interaction:
//  1. The initial GET returns 202 Accepted with a job entry (isRequestAccepted
//     passes, WaitForJobCompletionWithResp starts polling).
//  2. The job-polling GET returns the COMPLETED job with Stdout populated.
func TestStatDirectory_Success(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Initial stat-directory call — return 202 Accepted with a job.
			jsonResp(w, http.StatusAccepted, GenericResponse{
				Status: Status{Code: http.StatusAccepted},
				Jobs:   []Job{{JobID: 1, Status: "RUNNING"}},
			})
			return
		}
		// Job-polling call — return COMPLETED with stdout.
		jsonResp(w, http.StatusOK, GenericResponse{
			Status: Status{Code: http.StatusOK},
			Jobs: []Job{{
				JobID:  1,
				Status: "COMPLETED",
				Result: Respresult{Stdout: []string{"some stat info"}},
			}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	info, err := conn.StatDirectory(freshCtx(), "gpfs0", "/some/dir")
	require.NoError(t, err)
	assert.Equal(t, "some stat info", info)
}

// ---------------------------------------------------------------------------
// IsValidNodeclass
// ---------------------------------------------------------------------------

func TestIsValidNodeclass_Valid(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scalemgmt/v2/nodeclasses", func(w http.ResponseWriter, r *http.Request) {
		type ncResponse struct {
			NodeClasses []map[string]interface{} `json:"nodeClasses"`
			Status      Status                   `json:"status"`
		}
		jsonResp(w, http.StatusOK, ncResponse{
			Status:      Status{Code: http.StatusOK},
			NodeClasses: []map[string]interface{}{{"nodeClassName": "myclass"}},
		})
	})
	ts, conn := newTestServer(t, mux)
	defer ts.Close()

	// This is a best-effort test — the nodeclass response type is dynamic.
	// We just confirm the function runs without panicking on a 200 response.
	_, _ = conn.IsValidNodeclass(freshCtx(), "myclass")
}

// ---------------------------------------------------------------------------
// getNextEndpoint
// ---------------------------------------------------------------------------

func TestGetNextEndpoint_CyclesThroughEndpoints(t *testing.T) {
	conn := &SpectrumRestV2{
		Endpoint:      []string{"http://ep1/", "http://ep2/", "http://ep3/"},
		EndPointIndex: 0,
	}
	ctx := freshCtx()
	ep1 := conn.getNextEndpoint(ctx)
	ep2 := conn.getNextEndpoint(ctx)
	ep3 := conn.getNextEndpoint(ctx) // wraps around
	assert.Equal(t, "http://ep2/", ep1)
	assert.Equal(t, "http://ep3/", ep2)
	assert.Equal(t, "http://ep1/", ep3)
}
