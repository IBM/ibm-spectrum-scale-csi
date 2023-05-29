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

package connectors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"errors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const errConnectionRefused string = "connection refused"
const errNoSuchHost string = "no such host"

var GetLoggerId = utils.GetLoggerId

type SpectrumRestV2 struct {
	HTTPclient    *http.Client
	Endpoint      []string
	EndPointIndex int
	User          string
	Password      string
}

func (s *SpectrumRestV2) isStatusOK(statusCode int) bool {
	klog.V(4).Infof("rest_v2 isStatusOK. statusCode: %d", statusCode)

	if (statusCode == http.StatusOK) ||
		(statusCode == http.StatusCreated) ||
		(statusCode == http.StatusAccepted) {
		return true
	}
	return false
}

func (s *SpectrumRestV2) checkAsynchronousJob(statusCode int) bool {
	klog.V(4).Infof("rest_v2 checkAsynchronousJob. statusCode: %d", statusCode)
	if (statusCode == http.StatusAccepted) ||
		(statusCode == http.StatusCreated) {
		return true
	}
	return false
}

func (s *SpectrumRestV2) isRequestAccepted(ctx context.Context, response GenericResponse, url string) error {
	klog.V(4).Infof("[%s] rest_v2 isRequestAccepted. url: %s, response: %v", utils.GetLoggerId(ctx), url, response)

	if !s.isStatusOK(response.Status.Code) {
		return fmt.Errorf("error %v for url %v", response, url)
	}

	if len(response.Jobs) == 0 {
		return fmt.Errorf("unable to get Job details for %s request: %v", url, response)
	}
	return nil
}

func (s *SpectrumRestV2) WaitForJobCompletion(ctx context.Context, statusCode int, jobID uint64) error {
	klog.V(4).Infof("[%s] rest_v2 waitForJobCompletion. jobID: %d, statusCode: %d", utils.GetLoggerId(ctx), jobID, statusCode)

	if s.checkAsynchronousJob(statusCode) {
		jobURL := fmt.Sprintf("scalemgmt/v2/jobs/%d?fields=:all:", jobID)
		_, err := s.AsyncJobCompletion(ctx, jobURL)
		if err != nil {
			klog.Errorf("[%s] error in waiting for job completion %v, %v", utils.GetLoggerId(ctx), jobID, err)
			return err
		}
	}
	return nil
}

func (s *SpectrumRestV2) WaitForJobCompletionWithResp(ctx context.Context, statusCode int, jobID uint64) (GenericResponse, error) {
	klog.V(4).Infof("[%s] rest_v2 WaitForJobCompletionWithResp. jobID: %d, statusCode: %d", utils.GetLoggerId(ctx), jobID, statusCode)

	if s.checkAsynchronousJob(statusCode) {
		response := GenericResponse{}
		jobURL := fmt.Sprintf("scalemgmt/v2/jobs/%d?fields=:all:", jobID)
		response, err := s.AsyncJobCompletion(ctx, jobURL)
		if err != nil {
			return GenericResponse{}, err
		}
		return response, nil
	}
	return GenericResponse{}, nil
}

func (s *SpectrumRestV2) AsyncJobCompletion(ctx context.Context, jobURL string) (GenericResponse, error) {
	klog.V(4).Infof("[%s] rest_v2 AsyncJobCompletion. jobURL: %s", utils.GetLoggerId(ctx), jobURL)

	jobQueryResponse := GenericResponse{}
	var waitTime time.Duration = 2
	for {
		err := s.doHTTP(ctx, jobURL, "GET", &jobQueryResponse, nil)
		if err != nil {
			return GenericResponse{}, err
		}
		if len(jobQueryResponse.Jobs) == 0 {
			return GenericResponse{}, fmt.Errorf("unable to get Job details for %s: %v", jobURL, jobQueryResponse)
		}

		if jobQueryResponse.Jobs[0].Status == "RUNNING" {
			time.Sleep(waitTime * time.Second)
			if waitTime < 16 {
				waitTime = waitTime * 2
			}
			continue
		}
		break
	}
	if jobQueryResponse.Jobs[0].Status == "COMPLETED" || jobQueryResponse.Jobs[0].Status == "UNKNOWN" {
		return jobQueryResponse, nil
	} else {
		klog.Errorf("[%s] Async Job failed: %v", utils.GetLoggerId(ctx), jobQueryResponse)
		return GenericResponse{}, fmt.Errorf("%v", jobQueryResponse.Jobs[0].Result.Stderr)
	}
}

func NewSpectrumRestV2(ctx context.Context, scaleConfig settings.Clusters) (SpectrumScaleConnector, error) {
	klog.V(4).Infof("[%s] rest_v2 NewSpectrumRestV2.", utils.GetLoggerId(ctx))

	guiUser := scaleConfig.MgmtUsername
	guiPwd := scaleConfig.MgmtPassword
	var rest *SpectrumRestV2
	var tr *http.Transport

	if scaleConfig.SecureSslMode {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(scaleConfig.CacertValue); !ok {
			return &SpectrumRestV2{}, fmt.Errorf("parsing CA cert %v failed", scaleConfig.Cacert)
		}
		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool, MinVersion: tls.VersionTLS12}}
		klog.V(4).Infof("[%s] created IBM Storage Scale connector with SSL mode for guiHost(s)", utils.GetLoggerId(ctx))
	} else {
		//#nosec G402 InsecureSkipVerify was requested by user.
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}} //nolint:gosec
		klog.V(4).Infof("[%s] created IBM Storage Scale connector without SSL mode for guiHost(s)", utils.GetLoggerId(ctx))
	}

	rest = &SpectrumRestV2{
		HTTPclient: &http.Client{
			Transport: tr,
			Timeout:   time.Second * 60,
		},
		User:          guiUser,
		Password:      guiPwd,
		EndPointIndex: 0, //Use first GUI as primary by default
	}

	for i := range scaleConfig.RestAPI {
		guiHost := scaleConfig.RestAPI[i].GuiHost
		guiPort := scaleConfig.RestAPI[i].GuiPort
		if guiPort == 0 {
			guiPort = settings.DefaultGuiPort
		}
		endpoint := fmt.Sprintf("%s://%s:%d/", settings.GuiProtocol, guiHost, guiPort)
		rest.Endpoint = append(rest.Endpoint, endpoint)
	}
	return rest, nil
}

func (s *SpectrumRestV2) GetClusterId(ctx context.Context) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetClusterId", utils.GetLoggerId(ctx))

	getClusterURL := "scalemgmt/v2/cluster"
	getClusterResponse := GetClusterResponse{}

	err := s.doHTTP(ctx, getClusterURL, "GET", &getClusterResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get cluster ID: %v", utils.GetLoggerId(ctx), err)
		return "", err
	}
	cidStr := fmt.Sprintf("%v", getClusterResponse.Cluster.ClusterSummary.ClusterID)
	return cidStr, nil
}

// GetClusterSummary function returns the information details of the cluster.
func (s *SpectrumRestV2) GetClusterSummary(ctx context.Context) (ClusterSummary, error) {
	klog.V(4).Infof("[%s] rest_v2 GetClusterSummary", utils.GetLoggerId(ctx))

	getClusterURL := "scalemgmt/v2/cluster"
	getClusterResponse := GetClusterResponse{}

	err := s.doHTTP(ctx, getClusterURL, "GET", &getClusterResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get cluster summary: %v", utils.GetLoggerId(ctx), err)
		return ClusterSummary{}, err
	}
	return getClusterResponse.Cluster.ClusterSummary, nil
}

func (s *SpectrumRestV2) GetTimeZoneOffset(ctx context.Context) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetTimeZoneOffset", utils.GetLoggerId(ctx))

	getConfigURL := "scalemgmt/v2/config"
	getConfigResponse := GetConfigResponse{}

	err := s.doHTTP(ctx, getConfigURL, "GET", &getConfigResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get cluster configuration: %v", utils.GetLoggerId(ctx), err)
		return "", err
	}
	timezone := fmt.Sprintf("%v", getConfigResponse.Config.ClusterConfig.TimeZoneOffset)
	return timezone, nil
}

func (s *SpectrumRestV2) GetScaleVersion(ctx context.Context) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetScaleVersion", utils.GetLoggerId(ctx))

	getVersionURL := "scalemgmt/v2/info"
	getVersionResponse := GetInfoResponse_v2{}

	err := s.doHTTP(ctx, getVersionURL, "GET", &getVersionResponse, nil)
	if err != nil {
		klog.Errorf("[%s] unable to get IBM Storage Scale version: [%v]", utils.GetLoggerId(ctx), err)
		return "", err
	}

	if len(getVersionResponse.Info.ServerVersion) == 0 {
		return "", fmt.Errorf("unable to get IBM Storage Scale version")
	}

	return getVersionResponse.Info.ServerVersion, nil
}

func (s *SpectrumRestV2) GetFilesystemMountDetails(ctx context.Context, filesystemName string) (MountInfo, error) {
	klog.V(4).Infof("[%s] rest_v2 GetFilesystemMountDetails. filesystemName: %s", utils.GetLoggerId(ctx), filesystemName)

	getFilesystemURL := fmt.Sprintf("%s%s", "scalemgmt/v2/filesystems/", filesystemName)
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get filesystem details for %s: %v", utils.GetLoggerId(ctx), filesystemName, err)
		return MountInfo{}, err
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount, nil
	} else {
		return MountInfo{}, fmt.Errorf("unable to fetch mount details for %s", filesystemName)
	}
}

func (s *SpectrumRestV2) IsFilesystemMountedOnGUINode(ctx context.Context, filesystemName string) (bool, error) {
	klog.V(4).Infof("[%s] rest_v2 IsFilesystemMountedOnGUINode. filesystemName: %s", utils.GetLoggerId(ctx), filesystemName)

	mountURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName)
	mountResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, mountURL, "GET", &mountResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in getting filesystem mount details for %s: %v", utils.GetLoggerId(ctx), filesystemName, err)
		return false, err
	}

	if len(mountResponse.FileSystems) > 0 {
		klog.V(4).Infof("[%s] filesystem [%s] is [%v] on GUI node", utils.GetLoggerId(ctx), filesystemName, mountResponse.FileSystems[0].Mount.Status)
		if mountResponse.FileSystems[0].Mount.Status == "mounted" {
			return true, nil
		} else if mountResponse.FileSystems[0].Mount.Status == "not mounted" {
			return false, nil
		}
		return false, fmt.Errorf("unable to determine mount status of filesystem %s", filesystemName)
	} else {
		return false, fmt.Errorf("unable to fetch mount details for %s", filesystemName)
	}
}

func (s *SpectrumRestV2) ListFilesystems(ctx context.Context) ([]string, error) {
	klog.V(4).Infof("[%s] rest_v2 ListFilesystems", utils.GetLoggerId(ctx))

	listFilesystemsURL := "scalemgmt/v2/filesystems"
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, listFilesystemsURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in listing filesystems: %v", utils.GetLoggerId(ctx), err)
		return nil, err
	}
	fsNumber := len(getFilesystemResponse.FileSystems)
	filesystems := make([]string, fsNumber)
	for i := 0; i < fsNumber; i++ {
		filesystems[i] = getFilesystemResponse.FileSystems[i].Name
	}
	return filesystems, nil
}

func (s *SpectrumRestV2) GetFilesystemMountpoint(ctx context.Context, filesystemName string) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetFilesystemMountpoint. filesystemName: %s", utils.GetLoggerId(ctx), filesystemName)

	getFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName)
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in getting filesystem details for %s: %v", utils.GetLoggerId(ctx), filesystemName, err)
		return "", err
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount.MountPoint, nil
	} else {
		return "", fmt.Errorf("unable to fetch mount point for %s", filesystemName)
	}
}

func (s *SpectrumRestV2) CopyFsetSnapshotPath(ctx context.Context, filesystemName string, filesetName string, snapshotName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error) {
	klog.V(4).Infof("[%s] rest_v2 CopyFsetSnapshotPath. filesystem: %s, fileset: %s, snapshot: %s, srcPath: %s, targetPath: %s, nodeclass: %s", utils.GetLoggerId(ctx), filesystemName, filesetName, snapshotName, srcPath, targetPath, nodeclass)

	copySnapReq := CopySnapshotRequest{}
	copySnapReq.TargetPath = targetPath

	if nodeclass != "" {
		copySnapReq.NodeClass = nodeclass
	}

	formattedSrcPath := strings.ReplaceAll(srcPath, "/", "%2F")
	copySnapURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshotCopy/%s/path/%s", filesystemName, filesetName, snapshotName, formattedSrcPath)
	copySnapResp := GenericResponse{}

	err := s.doHTTP(ctx, copySnapURL, "PUT", &copySnapResp, copySnapReq)
	if err != nil {
		klog.Errorf("[%s] Error in copy snapshot request: %v", utils.GetLoggerId(ctx), err)
		return 0, 0, err
	}

	err = s.isRequestAccepted(ctx, copySnapResp, copySnapURL)
	if err != nil {
		klog.Errorf("[%s] request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return 0, 0, err
	}

	return copySnapResp.Status.Code, copySnapResp.Jobs[0].JobID, nil
}

func (s *SpectrumRestV2) CopyFilesetPath(ctx context.Context, filesystemName string, filesetName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error) {
	klog.V(4).Infof("[%s] rest_v2 CopyFilesetPath. filesystem: %s, fileset: %s, srcPath: %s, targetPath: %s, nodeclass: %s", utils.GetLoggerId(ctx), filesystemName, filesetName, srcPath, targetPath, nodeclass)

	copyVolReq := CopyVolumeRequest{}
	copyVolReq.TargetPath = targetPath

	if nodeclass != "" {
		copyVolReq.NodeClass = nodeclass
	}

	formattedSrcPath := strings.ReplaceAll(srcPath, "/", "%2F")
	copyVolURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/directoryCopy/%s", filesystemName, filesetName, formattedSrcPath)
	copyVolResp := GenericResponse{}

	err := s.doHTTP(ctx, copyVolURL, "PUT", &copyVolResp, copyVolReq)
	if err != nil {
		klog.Errorf("[%s] Error in copy volume request: %v", utils.GetLoggerId(ctx), err)
		return 0, 0, err
	}

	err = s.isRequestAccepted(ctx, copyVolResp, copyVolURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return 0, 0, err
	}

	return copyVolResp.Status.Code, copyVolResp.Jobs[0].JobID, nil
}

func (s *SpectrumRestV2) CopyDirectoryPath(ctx context.Context, filesystemName string, srcPath string, targetPath string, nodeclass string) (int, uint64, error) {
	klog.V(4).Infof("[%s] rest_v2 CopyDirectoryPath. filesystem: %s, srcPath: %s, targetPath: %s, nodeclass: %s", utils.GetLoggerId(ctx), filesystemName, srcPath, targetPath, nodeclass)

	copyVolReq := CopyVolumeRequest{}
	copyVolReq.TargetPath = targetPath

	if nodeclass != "" {
		copyVolReq.NodeClass = nodeclass
	}

	formattedSrcPath := strings.ReplaceAll(srcPath, "/", "%2F")
	copyVolURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/directoryCopy/%s", filesystemName, formattedSrcPath)
	copyVolResp := GenericResponse{}

	err := s.doHTTP(ctx, copyVolURL, "PUT", &copyVolResp, copyVolReq)
	if err != nil {
		klog.Errorf("[%s] Error in copy volume request: %v", utils.GetLoggerId(ctx), err)
		return 0, 0, err
	}

	err = s.isRequestAccepted(ctx, copyVolResp, copyVolURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return 0, 0, err
	}

	return copyVolResp.Status.Code, copyVolResp.Jobs[0].JobID, nil
}

func (s *SpectrumRestV2) CreateSnapshot(ctx context.Context, filesystemName string, filesetName string, snapshotName string) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 CreateSnapshot. filesystem: %s, fileset: %s, snapshot: %v", loggerId, filesystemName, filesetName, snapshotName)

	snapshotreq := CreateSnapshotRequest{}
	snapshotreq.SnapshotName = snapshotName

	createSnapshotURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots", filesystemName, filesetName)
	createSnapshotResponse := GenericResponse{}

	err := s.doHTTP(ctx, createSnapshotURL, "POST", &createSnapshotResponse, snapshotreq)
	if err != nil {
		klog.Errorf("[%s] error in create snapshot request: %v", loggerId, err)
		return err
	}

	err = s.isRequestAccepted(ctx, createSnapshotResponse, createSnapshotURL)
	if err != nil {
		klog.Errorf("[%s] request not accepted for processing: %v", loggerId, err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, createSnapshotResponse.Status.Code, createSnapshotResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSP1102C") { // job failed as snapshot already exists
			fmt.Println(err)
			return nil
		}
		klog.Errorf("[%s] unable to create snapshot %s: %v", loggerId, snapshotName, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) DeleteSnapshot(ctx context.Context, filesystemName string, filesetName string, snapshotName string) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 DeleteSnapshot. filesystem: %s, fileset: %s, snapshot: %v", loggerId, filesystemName, filesetName, snapshotName)

	deleteSnapshotURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots/%s", filesystemName, filesetName, snapshotName)
	deleteSnapshotResponse := GenericResponse{}

	err := s.doHTTP(ctx, deleteSnapshotURL, "DELETE", &deleteSnapshotResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in delete snapshot request: %v", loggerId, err)
		return err
	}

	err = s.isRequestAccepted(ctx, deleteSnapshotResponse, deleteSnapshotURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", loggerId, err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, deleteSnapshotResponse.Status.Code, deleteSnapshotResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] Unable to delete snapshot %s: %v", loggerId, snapshotName, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) GetLatestFilesetSnapshots(ctx context.Context, filesystemName string, filesetName string) ([]Snapshot_v2, error) {
	klog.V(4).Infof("[%s] rest_v2 GetLatestFilesetSnapshots. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	getLatestFilesetSnapshotsURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots/latest", filesystemName, filesetName)
	getLatestFilesetSnapshotsResponse := GetSnapshotResponse_v2{}

	err := s.doHTTP(ctx, getLatestFilesetSnapshotsURL, "GET", &getLatestFilesetSnapshotsResponse, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest list of snapshots for fileset [%v]. Error [%v]", filesetName, err)
	}

	return getLatestFilesetSnapshotsResponse.Snapshots, nil
}

func (s *SpectrumRestV2) UpdateFileset(ctx context.Context, filesystemName string, filesetName string, opts map[string]interface{}) error {
	klog.V(4).Infof("[%s] rest_v2 UpdateFileset. filesystem: %s, fileset: %s, opts: %v", utils.GetLoggerId(ctx), filesystemName, filesetName, opts)
	filesetreq := CreateFilesetRequest{}
	inodeLimit, inodeLimitSpecified := opts[UserSpecifiedInodeLimit]
	if inodeLimitSpecified {
		filesetreq.MaxNumInodes = inodeLimit.(string)
		//filesetreq.AllocInodes = "1024"
	}
	updateFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	updateFilesetResponse := GenericResponse{}
	err := s.doHTTP(ctx, updateFilesetURL, "PUT", &updateFilesetResponse, filesetreq)
	if err != nil {
		klog.Errorf("[%s] error in update fileset request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, updateFilesetResponse, updateFilesetURL)
	if err != nil {
		klog.Errorf("[%s] request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, updateFilesetResponse.Status.Code, updateFilesetResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] unable to update fileset %s: %v", utils.GetLoggerId(ctx), filesetName, err)
		return err
	}
	return nil
}

func (s *SpectrumRestV2) CreateFileset(ctx context.Context, filesystemName string, filesetName string, opts map[string]interface{}) error {
	klog.V(4).Infof("[%s] rest_v2 CreateFileset. filesystem: %s, fileset: %s, opts: %v", utils.GetLoggerId(ctx), filesystemName, filesetName, opts)

	filesetreq := CreateFilesetRequest{}
	filesetreq.FilesetName = filesetName
	filesetreq.Comment = FilesetComment

	filesetType, filesetTypeSpecified := opts[UserSpecifiedFilesetType]
	inodeLimit, inodeLimitSpecified := opts[UserSpecifiedInodeLimit]

	if !filesetTypeSpecified {
		filesetType, filesetTypeSpecified = opts[UserSpecifiedFilesetTypeDep]
	}

	if !inodeLimitSpecified {
		inodeLimit, inodeLimitSpecified = opts[UserSpecifiedInodeLimitDep]
	}

	if filesetTypeSpecified && filesetType.(string) == "dependent" {
		/* Add fileset for dependent fileset-name: */
		parentFileSetName, parentFileSetNameSpecified := opts[UserSpecifiedParentFset]
		if parentFileSetNameSpecified {
			filesetreq.InodeSpace = parentFileSetName.(string)
		} else {
			filesetreq.InodeSpace = "root"
		}
	} else {
		filesetreq.InodeSpace = "new"
		if inodeLimitSpecified {
			filesetreq.MaxNumInodes = inodeLimit.(string)
			filesetreq.AllocInodes = "1024"
		}
	}

	uid, uidSpecified := opts[UserSpecifiedUID]
	gid, gidSpecified := opts[UserSpecifiedGID]
	permissions, permissionsSpecified := opts[UserSpecifiedPermissions]

	if uidSpecified && gidSpecified {
		filesetreq.Owner = fmt.Sprintf("%s:%s", uid, gid)
	} else if uidSpecified {
		filesetreq.Owner = fmt.Sprintf("%s", uid)
	}
	if permissionsSpecified {
		filesetreq.Permissions = fmt.Sprintf("%s", permissions)
	}

	createFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName)
	createFilesetResponse := GenericResponse{}

	err := s.doHTTP(ctx, createFilesetURL, "POST", &createFilesetResponse, filesetreq)
	if err != nil {
		klog.Errorf("[%s] Error in create fileset request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, createFilesetResponse, createFilesetURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, createFilesetResponse.Status.Code, createFilesetResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSP1102C") { // job failed as fileset already exists
			fmt.Println(err)
			return nil
		}
		klog.Errorf("[%s] Unable to create fileset %s: %v", utils.GetLoggerId(ctx), filesetName, err)
		return err
	}
	return nil
}

func (s *SpectrumRestV2) DeleteFileset(ctx context.Context, filesystemName string, filesetName string) error {
	klog.V(4).Infof("[%s] rest_v2 DeleteFileset. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	deleteFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	deleteFilesetResponse := GenericResponse{}

	err := s.doHTTP(ctx, deleteFilesetURL, "DELETE", &deleteFilesetResponse, nil)
	if err != nil {
		if strings.Contains(deleteFilesetResponse.Status.Message, "Invalid value in 'fsetName'") { // job failed as dir already exists
			klog.V(6).Infof("[%s] Fileset would have been deleted. So returning success %v", utils.GetLoggerId(ctx), err)
			return nil
		}

		klog.Errorf("[%s] Error in delete fileset request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, deleteFilesetResponse, deleteFilesetURL)
	if err != nil {
		klog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, deleteFilesetResponse.Status.Code, deleteFilesetResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("Unable to delete fileset %s: %v", filesetName, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) LinkFileset(ctx context.Context, filesystemName string, filesetName string, linkpath string) error {
	klog.V(4).Infof("[%s] rest_v2 LinkFileset. filesystem: %s, fileset: %s, linkpath: %s", utils.GetLoggerId(ctx), filesystemName, filesetName, linkpath)

	linkReq := LinkFilesetRequest{}
	linkReq.Path = linkpath
	linkFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link", filesystemName, filesetName)
	linkFilesetResponse := GenericResponse{}

	err := s.doHTTP(ctx, linkFilesetURL, "POST", &linkFilesetResponse, linkReq)
	if err != nil {
		klog.Errorf("[%s] Error in link fileset request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, linkFilesetResponse, linkFilesetURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, linkFilesetResponse.Status.Code, linkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] Error in linking fileset %s: %v", utils.GetLoggerId(ctx), filesetName, err)
		return err
	}
	return nil
}

func (s *SpectrumRestV2) UnlinkFileset(ctx context.Context, filesystemName string, filesetName string) error {
	klog.V(4).Infof("[%s] rest_v2 UnlinkFileset. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	unlinkFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link?force=True", filesystemName, filesetName)
	unlinkFilesetResponse := GenericResponse{}

	err := s.doHTTP(ctx, unlinkFilesetURL, "DELETE", &unlinkFilesetResponse, nil)

	if err != nil {
		klog.Errorf("[%s] Error in unlink fileset request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, unlinkFilesetResponse, unlinkFilesetURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, unlinkFilesetResponse.Status.Code, unlinkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] Error in unlink fileset %s: %v", utils.GetLoggerId(ctx), filesetName, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) ListFileset(ctx context.Context, filesystemName string, filesetName string) (Fileset_v2, error) {
	klog.V(4).Infof("[%s] rest_v2 ListFileset. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	getFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(ctx, getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in list fileset request: %v", utils.GetLoggerId(ctx), err)
		return Fileset_v2{}, err
	}

	if len(getFilesetResponse.Filesets) == 0 {
		klog.Errorf("[%s] No fileset returned for %s", utils.GetLoggerId(ctx), filesetName)
		return Fileset_v2{}, fmt.Errorf("no fileset returned for %s", filesetName)
	}

	return getFilesetResponse.Filesets[0], nil
}

func (s *SpectrumRestV2) GetFilesetsInodeSpace(ctx context.Context, filesystemName string, inodeSpace int) ([]Fileset_v2, error) {
	klog.V(4).Infof("[%s] rest_v2 ListAllFilesets. filesystem: %s", utils.GetLoggerId(ctx), filesystemName)

	getFilesetsURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets?filter=config.inodeSpace=%d", filesystemName, inodeSpace)
	getFilesetsResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(ctx, getFilesetsURL, "GET", &getFilesetsResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in list filesets request: %v", utils.GetLoggerId(ctx), err)
		return nil, err
	}

	return getFilesetsResponse.Filesets, nil
}

func (s *SpectrumRestV2) IsFilesetLinked(ctx context.Context, filesystemName string, filesetName string) (bool, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 IsFilesetLinked. filesystem: %s, fileset: %s", loggerId, filesystemName, filesetName)

	fileset, err := s.ListFileset(ctx, filesystemName, filesetName)
	if err != nil {
		return false, err
	}

	if (fileset.Config.Path == "") ||
		(fileset.Config.Path == "--") {
		return false, nil
	}
	return true, nil
}

func (s *SpectrumRestV2) FilesetRefreshTask(ctx context.Context) error {
	klog.V(4).Infof("[%s] rest_v2 FilesetRefreshTask", utils.GetLoggerId(ctx))

	filesetRefreshURL := "scalemgmt/v2/refreshTask/enqueue?taskId=FILESETS&maxDelay=0"
	filesetRefreshResponse := GenericResponse{}

	err := s.doHTTP(ctx, filesetRefreshURL, "POST", &filesetRefreshResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in fileset refresh task: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) MakeDirectory(ctx context.Context, filesystemName string, relativePath string, uid string, gid string) error {
	klog.V(4).Infof("[%s] rest_v2 MakeDirectory. filesystem: %s, path: %s, uid: %s, gid: %s", utils.GetLoggerId(ctx), filesystemName, relativePath, uid, gid)

	dirreq := CreateMakeDirRequest{}

	if uid != "" {
		_, err := strconv.Atoi(uid)
		if err != nil {
			dirreq.USER = uid
		} else {
			dirreq.UID = uid
		}
	} else {
		dirreq.UID = "0"
	}

	if gid != "" {
		_, err := strconv.Atoi(gid)
		if err != nil {
			dirreq.GROUP = gid
		} else {
			dirreq.GID = gid
		}
	} else {
		dirreq.GID = "0"
	}

	formattedPath := strings.ReplaceAll(relativePath, "/", "%2F")
	makeDirURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, formattedPath)

	makeDirResponse := GenericResponse{}

	err := s.doHTTP(ctx, makeDirURL, "POST", &makeDirResponse, dirreq)

	if err != nil {
		klog.Errorf("[%s] Error in make directory request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, makeDirResponse, makeDirURL)
	if err != nil {
		klog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, makeDirResponse.Status.Code, makeDirResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0762C") { // job failed as dir already exists
			klog.V(6).Infof("[%s] Directory exists. %v", utils.GetLoggerId(ctx), err)
			return nil
		}

		klog.Errorf("[%s] Unable to make directory %s: %v.", utils.GetLoggerId(ctx), relativePath, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) MakeDirectoryV2(ctx context.Context, filesystemName string, relativePath string, uid string, gid string, permissions string) error {
	klog.V(4).Infof("[%s] rest_v2 MakeDirectoryV2. filesystem: %s, path: %s, uid: %s, gid: %s, permissions: %s", utils.GetLoggerId(ctx), filesystemName, relativePath, uid, gid, permissions)

	dirreq := CreateMakeDirRequest{}

	if uid != "" {
		_, err := strconv.Atoi(uid)
		if err != nil {
			dirreq.USER = uid
		} else {
			dirreq.UID = uid
		}
	} else {
		dirreq.UID = "0"
	}

	if gid != "" {
		_, err := strconv.Atoi(gid)
		if err != nil {
			dirreq.GROUP = gid
		} else {
			dirreq.GID = gid
		}
	} else {
		dirreq.GID = "0"
	}

	dirreq.PERMISSIONS = permissions

	formattedPath := strings.ReplaceAll(relativePath, "/", "%2F")
	makeDirURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, formattedPath)

	makeDirResponse := GenericResponse{}

	err := s.doHTTP(ctx, makeDirURL, "POST", &makeDirResponse, dirreq)

	if err != nil {
		klog.Errorf("[%s] Error in make directory request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, makeDirResponse, makeDirURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, makeDirResponse.Status.Code, makeDirResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0762C") { // job failed as dir already exists
			klog.V(6).Infof("[%s] Directory exists. %v", utils.GetLoggerId(ctx), err)
			return nil
		}

		klog.Errorf("[%s] Unable to make directory %s: %v.", utils.GetLoggerId(ctx), relativePath, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) SetFilesetQuota(ctx context.Context, filesystemName string, filesetName string, quota string) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 SetFilesetQuota. filesystem: %s, fileset: %s, quota: %s", loggerId, filesystemName, filesetName, quota)

	setQuotaURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName)
	quotaRequest := SetQuotaRequest_v2{}

	quotaRequest.BlockHardLimit = quota
	quotaRequest.BlockSoftLimit = quota
	quotaRequest.OperationType = "setQuota"
	quotaRequest.QuotaType = "fileset"
	quotaRequest.ObjectName = filesetName

	setQuotaResponse := GenericResponse{}

	err := s.doHTTP(ctx, setQuotaURL, "POST", &setQuotaResponse, quotaRequest)
	if err != nil {
		klog.Errorf("[%s] Error in set fileset quota request: %v", loggerId, err)
		return err
	}

	err = s.isRequestAccepted(ctx, setQuotaResponse, setQuotaURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", loggerId, err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, setQuotaResponse.Status.Code, setQuotaResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] Unable to set quota for fileset %s: %v", loggerId, filesetName, err)
		return err
	}
	return nil
}

func (s *SpectrumRestV2) CheckIfFSQuotaEnabled(ctx context.Context, filesystemName string) error {
	klog.V(4).Infof("[%s] rest_v2 CheckIfFSQuotaEnabled. filesystem: %s", utils.GetLoggerId(ctx), filesystemName)

	checkQuotaURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName)
	QuotaResponse := GetQuotaResponse_v2{}

	err := s.doHTTP(ctx, checkQuotaURL, "GET", &QuotaResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Error in check quota: %v", utils.GetLoggerId(ctx), err)
		return err
	}
	return nil
}

func (s *SpectrumRestV2) IsValidNodeclass(ctx context.Context, nodeclass string) (bool, error) {
	klog.V(4).Infof("[%s] rest_v2 IsValidNodeclass. nodeclass: %s", utils.GetLoggerId(ctx), nodeclass)

	checkNodeclassURL := fmt.Sprintf("scalemgmt/v2/nodeclasses/%s", nodeclass)
	nodeclassResponse := GenericResponse{}

	err := s.doHTTP(ctx, checkNodeclassURL, "GET", &nodeclassResponse, nil)
	if err != nil {
		if strings.Contains(nodeclassResponse.Status.Message, "Invalid value in nodeclassName") {
			// nodeclass is not present
			return false, nil
		}
		return false, fmt.Errorf("unable to get nodeclass details")
	}
	return true, nil
}

func (s *SpectrumRestV2) IsSnapshotSupported(ctx context.Context) (bool, error) {
	klog.V(4).Infof("[%s] rest_v2 IsSnapshotSupported", utils.GetLoggerId(ctx))

	getVersionURL := "scalemgmt/v2/info"
	getVersionResponse := GetInfoResponse_v2{}

	err := s.doHTTP(ctx, getVersionURL, "GET", &getVersionResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get cluster information: [%v]", utils.GetLoggerId(ctx), err)
		return false, err
	}

	if len(getVersionResponse.Info.Paths.SnapCopyOp) == 0 {
		return false, nil
	}

	return true, nil
}

func (s *SpectrumRestV2) GetFilesetQuotaDetails(ctx context.Context, filesystemName string, filesetName string) (Quota_v2, error) {
	klog.V(4).Infof("[%s] rest_v2 GetFilesetQuotaDetails. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	listQuotaURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas?filter=objectName=%s", filesystemName, filesetName)
	listQuotaResponse := GetQuotaResponse_v2{}

	err := s.doHTTP(ctx, listQuotaURL, "GET", &listQuotaResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to fetch quota information for fileset %s:%s: [%v]", utils.GetLoggerId(ctx), filesystemName, filesetName, err)
		return Quota_v2{}, err
	}

	if len(listQuotaResponse.Quotas) == 0 {
		klog.Errorf("[%s] No quota information found for fileset %s:%s ", utils.GetLoggerId(ctx), filesystemName, filesetName)
		return Quota_v2{}, nil
	}

	return listQuotaResponse.Quotas[0], nil
}

func (s *SpectrumRestV2) ListFilesetQuota(ctx context.Context, filesystemName string, filesetName string) (string, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 ListFilesetQuota. filesystem: %s, fileset: %s", loggerId, filesystemName, filesetName)

	listQuotaResponse, err := s.GetFilesetQuotaDetails(ctx, filesystemName, filesetName)

	if err != nil {
		return "", err
	}

	if listQuotaResponse.BlockLimit > 0 {
		return fmt.Sprintf("%dK", listQuotaResponse.BlockLimit), nil
	} else {
		klog.Errorf("[%s] No quota information found for fileset %s", loggerId, filesetName)
		return "", nil
	}
}

func (s *SpectrumRestV2) doHTTP(ctx context.Context, urlSuffix string, method string, responseObject interface{}, param interface{}) error {
	klog.V(4).Infof("[%s] rest_v2 doHTTP: urlSuffix: %s, method: %s, param: %v", utils.GetLoggerId(ctx), urlSuffix, method, param)
	endpoint := s.Endpoint[s.EndPointIndex]
	klog.V(4).Infof("[%s] rest_v2 doHTTP: endpoint: %s", utils.GetLoggerId(ctx), endpoint)
	response, err := utils.HttpExecuteUserAuth(ctx, s.HTTPclient, method, endpoint+urlSuffix, s.User, s.Password, param)

	activeEndpointFound := false
	if err != nil {
		if strings.Contains(err.Error(), errConnectionRefused) || strings.Contains(err.Error(), errNoSuchHost) || errors.Is(err, context.DeadlineExceeded) {
			klog.Errorf("[%s] rest_v2 doHTTP: Error in connecting to GUI endpoint %s: %v, checking next endpoint", utils.GetLoggerId(ctx), endpoint, err)
			// Out of n endpoints, one has failed already, so loop over the
			// remaining n-1 endpoints till we get an active GUI endpoint.
			n := len(s.Endpoint)
			for i := 0; i < n-1; i++ {
				endpoint = s.getNextEndpoint(ctx)
				response, err = utils.HttpExecuteUserAuth(ctx, s.HTTPclient, method, endpoint+urlSuffix, s.User, s.Password, param)
				if err == nil {
					activeEndpointFound = true
					break
				} else {
					if strings.Contains(err.Error(), errConnectionRefused) || strings.Contains(err.Error(), errNoSuchHost) || errors.Is(err, context.DeadlineExceeded) {
						klog.Errorf("[%s] rest_v2 doHTTP: Error in connecting to GUI endpoint %s: %v, checking next endpoint", utils.GetLoggerId(ctx), endpoint, err)
					} else {
						klog.Errorf("[%s] rest_v2 doHTTP: Error in authentication request on endpoint %s: %v", utils.GetLoggerId(ctx), endpoint, err)
					}
				}
			}
		} else {
			klog.Errorf("[%s] rest_v2 doHTTP: Error in authentication request on endpoint %s: %v", utils.GetLoggerId(ctx), endpoint, err)
			return status.Error(codes.Internal, fmt.Sprintf("Error in authentication: %s request %v%v, user: %v, param: %v, response status: %v", method, endpoint, urlSuffix, s.User, param, response.Status))
		}
	} else {
		activeEndpointFound = true
	}
	if !activeEndpointFound {
		klog.Errorf("[%s] rest_v2 doHTTP: Could not find any active GUI endpoint: %v", utils.GetLoggerId(ctx), err)
		return status.Error(codes.Internal, fmt.Sprintf("Could not find any active GUI endpoint: %s request %v%v, user: %v, param: %v, response status: %v", method, endpoint, urlSuffix, s.User, param, response.Status))
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return status.Error(codes.Unauthenticated, fmt.Sprintf("%v: Unauthorized %s request: %v%v, user: %v, param: %v, response status: %v", http.StatusUnauthorized, method, endpoint, urlSuffix, s.User, param, response.Status))
	} else if response.StatusCode == http.StatusForbidden {
		return status.Error(codes.Internal, fmt.Sprintf("%v: Forbidden %s request %v%v, user: %v, param: %v, response status: %v", http.StatusForbidden, method, endpoint, urlSuffix, s.User, param, response.Status))
	}

	err = utils.UnmarshalResponse(ctx, response, responseObject)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Response unmarshal failed: %s request %v%v, user: %v, param: %v, response status: %v, error %v", method, endpoint, urlSuffix, s.User, param, response.Status, err))
	}

	if !s.isStatusOK(response.StatusCode) {
		return status.Error(codes.Internal, fmt.Sprintf("remote call failed with response %v: %s request %v%v, user: %v, param: %v, response status: %v", responseObject, method, endpoint, urlSuffix, s.User, param, response.Status))
	}

	return nil
}

func (s *SpectrumRestV2) MountFilesystem(ctx context.Context, filesystemName string, nodeName string) error { //nolint:dupl
	klog.V(4).Infof("[%s] rest_v2 MountFilesystem. filesystem: %s, node: %s", utils.GetLoggerId(ctx), filesystemName, nodeName)

	mountreq := MountFilesystemRequest{}
	mountreq.Nodes = append(mountreq.Nodes, nodeName)

	mountFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/mount", filesystemName)
	mountFilesystemResponse := GenericResponse{}

	err := s.doHTTP(ctx, mountFilesystemURL, "PUT", &mountFilesystemResponse, mountreq)
	if err != nil {
		klog.Errorf("[%s] Error in mount filesystem request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, mountFilesystemResponse, mountFilesystemURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, mountFilesystemResponse.Status.Code, mountFilesystemResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] Unable to Mount filesystem %s on node %s: %v", utils.GetLoggerId(ctx), filesystemName, nodeName, err)
		return err
	}
	return nil
}

func (s *SpectrumRestV2) UnmountFilesystem(ctx context.Context, filesystemName string, nodeName string) error { //nolint:dupl
	klog.V(4).Infof("[%s] rest_v2 UnmountFilesystem. filesystem: %s, node: %s", utils.GetLoggerId(ctx), filesystemName, nodeName)

	unmountreq := UnmountFilesystemRequest{}
	unmountreq.Nodes = append(unmountreq.Nodes, nodeName)

	unmountFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/unmount", filesystemName)
	unmountFilesystemResponse := GenericResponse{}

	err := s.doHTTP(ctx, unmountFilesystemURL, "PUT", &unmountFilesystemResponse, unmountreq)
	if err != nil {
		klog.Errorf("[%s] Error in unmount filesystem request: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.isRequestAccepted(ctx, unmountFilesystemResponse, unmountFilesystemURL)
	if err != nil {
		klog.Errorf("[%s] Request not accepted for processing: %v", utils.GetLoggerId(ctx), err)
		return err
	}

	err = s.WaitForJobCompletion(ctx, unmountFilesystemResponse.Status.Code, unmountFilesystemResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("Unable to unmount filesystem %s on node %s: %v", filesystemName, nodeName, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) GetFilesystemName(ctx context.Context, filesystemUUID string) (string, error) { //nolint:dupl
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 GetFilesystemName. UUID: %s", loggerId, filesystemUUID)

	getFilesystemNameURL := fmt.Sprintf("scalemgmt/v2/filesystems?filter=uuid=%s", filesystemUUID)
	getFilesystemNameURLResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, getFilesystemNameURL, "GET", &getFilesystemNameURLResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get filesystem name for uuid %s: %v", loggerId, filesystemUUID, err)
		return "", err
	}

	if len(getFilesystemNameURLResponse.FileSystems) == 0 {
		klog.Errorf("[%s] Unable to fetch filesystem name details for %s", loggerId, filesystemUUID)
		return "", fmt.Errorf("unable to fetch filesystem name details for %s", filesystemUUID)
	}
	return getFilesystemNameURLResponse.FileSystems[0].Name, nil
}

func (s *SpectrumRestV2) GetFilesystemDetails(ctx context.Context, filesystemName string) (FileSystem_v2, error) { //nolint:dupl
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 GetFilesystemDetails. Name: %s", loggerId, filesystemName)

	getFilesystemDetailsURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName)
	getFilesystemDetailsURLResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, getFilesystemDetailsURL, "GET", &getFilesystemDetailsURLResponse, nil)
	if err != nil {
		klog.Errorf("[%s] Unable to get filesystem details for filesystem %s: %v", loggerId, filesystemName, err)
		return FileSystem_v2{}, err
	}

	if len(getFilesystemDetailsURLResponse.FileSystems) == 0 {
		klog.Errorf("[%s] Unable to fetch filesystem details for %s", loggerId, filesystemName)
		return FileSystem_v2{}, fmt.Errorf("unable to fetch filesystem details for %s", filesystemName)
	}

	return getFilesystemDetailsURLResponse.FileSystems[0], nil
}

func (s *SpectrumRestV2) GetFsUid(ctx context.Context, filesystemName string) (string, error) {
	klog.V(4).Infof("rest_v2 GetFsUid. filesystem: %s", filesystemName)

	getFilesystemURL := fmt.Sprintf("%s%s", "scalemgmt/v2/filesystems/", filesystemName)
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(ctx, getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		return "", fmt.Errorf("unable to get filesystem details for %s", filesystemName)
	}

	fmt.Println(getFilesystemResponse)
	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].UUID, nil
	} else {
		return "", fmt.Errorf("unable to fetch mount details for %s", filesystemName)
	}
}

func (s *SpectrumRestV2) DeleteSymLnk(ctx context.Context, filesystemName string, lnkName string) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 DeleteSymLnk. filesystem: %s, link: %s", loggerId, filesystemName, lnkName)

	lnkName = strings.ReplaceAll(lnkName, "/", "%2F")
	deleteLnkURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/symlink/%s", filesystemName, lnkName)
	deleteLnkResponse := GenericResponse{}

	err := s.doHTTP(ctx, deleteLnkURL, "DELETE", &deleteLnkResponse, nil)
	if err != nil {
		return fmt.Errorf("unable to delete Symlink %v", lnkName)
	}

	err = s.isRequestAccepted(ctx, deleteLnkResponse, deleteLnkURL)
	if err != nil {
		return err
	}

	err = s.WaitForJobCompletion(ctx, deleteLnkResponse.Status.Code, deleteLnkResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG2006C") {
			klog.V(4).Infof("[%s] Since slink %v was already deleted, so returning success", loggerId, lnkName)
			return nil
		}
		return fmt.Errorf("unable to delete symLnk %v:%v", lnkName, err)
	}

	return nil
}

func (s *SpectrumRestV2) DeleteDirectory(ctx context.Context, filesystemName string, dirName string, safe bool) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 DeleteDirectory. filesystem: %s, dir: %s, safe: %v", loggerId, filesystemName, dirName, safe)

	NdirName := strings.ReplaceAll(dirName, "/", "%2F")
	deleteDirURL := ""
	if safe {
		deleteDirURL = fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s?safe=True", filesystemName, NdirName)
	} else {
		deleteDirURL = fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, NdirName)
	}
	deleteDirResponse := GenericResponse{}

	err := s.doHTTP(ctx, deleteDirURL, "DELETE", &deleteDirResponse, nil)
	if err != nil {
		return fmt.Errorf("unable to delete dir %v", dirName)
	}

	err = s.isRequestAccepted(ctx, deleteDirResponse, deleteDirURL)
	if err != nil {
		return err
	}

	err = s.WaitForJobCompletion(ctx, deleteDirResponse.Status.Code, deleteDirResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("unable to delete dir %v:%v", dirName, err)
	}

	return nil
}

func (s *SpectrumRestV2) StatDirectory(ctx context.Context, filesystemName string, dirName string) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 StatDirectory. filesystem: %s, dir: %s", utils.GetLoggerId(ctx), filesystemName, dirName)

	fmtDirName := strings.ReplaceAll(dirName, "/", "%2F")
	statDirURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, fmtDirName)
	statDirResponse := GenericResponse{}

	err := s.doHTTP(ctx, statDirURL, "GET", &statDirResponse, nil)
	if err != nil {
		return "", fmt.Errorf("unable to stat dir %v", dirName)
	}

	err = s.isRequestAccepted(ctx, statDirResponse, statDirURL)
	if err != nil {
		return "", err
	}

	jobResp, err := s.WaitForJobCompletionWithResp(ctx, statDirResponse.Status.Code, statDirResponse.Jobs[0].JobID)
	if err != nil {
		return "", fmt.Errorf("unable to stat dir %v:%v", dirName, err)
	}

	statInfo := jobResp.Jobs[0].Result.Stdout[0]

	return statInfo, nil
}

func (s *SpectrumRestV2) GetFileSetUid(ctx context.Context, filesystemName string, filesetName string) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetFileSetUid. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	filesetResponse, err := s.GetFileSetResponseFromName(ctx, filesystemName, filesetName)
	if err != nil {
		return "", fmt.Errorf("fileset response not found for fileset %v:%v", filesystemName, filesetName)
	}

	return fmt.Sprintf("%d", filesetResponse.Config.Id), nil
}

func (s *SpectrumRestV2) GetFileSetResponseFromName(ctx context.Context, filesystemName string, filesetName string) (Fileset_v2, error) {
	klog.V(4).Infof("[%s] rest_v2 GetFileSetResponseFromName. filesystem: %s, fileset: %s", utils.GetLoggerId(ctx), filesystemName, filesetName)

	getFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(ctx, getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		return Fileset_v2{}, fmt.Errorf("unable to list fileset %v", filesetName)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return Fileset_v2{}, fmt.Errorf("unable to list fileset %v", filesetName)
	}

	return getFilesetResponse.Filesets[0], nil
}

// CheckIfFilesetExist Checking if fileset exist in filesystem
func (s *SpectrumRestV2) CheckIfFilesetExist(ctx context.Context, filesystemName string, filesetName string) (bool, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 CheckIfFilesetExist. filesystem: %s, fileset: %s", loggerId, filesystemName, filesetName)

	checkFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(ctx, checkFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		if strings.Contains(getFilesetResponse.Status.Message, "Invalid value in 'filesetName'") {
			// snapshot is not present
			return false, nil
		}
		return false, fmt.Errorf("unable to get fileset details for filesystem: %v, fileset: %v", filesystemName, filesetName)
	}
	return true, nil
}

func (s *SpectrumRestV2) GetFileSetNameFromId(ctx context.Context, filesystemName string, Id string) (string, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 GetFileSetNameFromId. filesystem: %s, fileset id: %s", loggerId, filesystemName, Id)

	filesetResponse, err := s.GetFileSetResponseFromId(ctx, filesystemName, Id)
	if err != nil {
		return "", fmt.Errorf("fileset response not found for fileset Id %v:%v", filesystemName, Id)
	}
	return filesetResponse.FilesetName, nil
}

func (s *SpectrumRestV2) GetFileSetResponseFromId(ctx context.Context, filesystemName string, Id string) (Fileset_v2, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 GetFileSetResponseFromId. filesystem: %s, fileset id: %s", loggerId, filesystemName, Id)

	getFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets?filter=config.id=%s", filesystemName, Id)
	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(ctx, getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		return Fileset_v2{}, fmt.Errorf("unable to get name for fileset Id %v:%v", filesystemName, Id)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return Fileset_v2{}, fmt.Errorf("no filesets found for Id %v:%v", filesystemName, Id)
	}

	return getFilesetResponse.Filesets[0], nil
}

//nolint:dupl
func (s *SpectrumRestV2) GetSnapshotCreateTimestamp(ctx context.Context, filesystemName string, filesetName string, snapName string) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetSnapshotCreateTimestamp. filesystem: %s, fileset: %s, snapshot: %s ", utils.GetLoggerId(ctx), filesystemName, filesetName, snapName)

	getSnapshotURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots/%s", filesystemName, filesetName, snapName)
	getSnapshotResponse := GetSnapshotResponse_v2{}

	err := s.doHTTP(ctx, getSnapshotURL, "GET", &getSnapshotResponse, nil)
	if err != nil {
		return "", fmt.Errorf("unable to list snapshot %v", snapName)
	}

	if len(getSnapshotResponse.Snapshots) == 0 {
		return "", fmt.Errorf("unable to list snapshot %v", snapName)
	}

	return fmt.Sprintf(getSnapshotResponse.Snapshots[0].Created), nil
}

//nolint:dupl
func (s *SpectrumRestV2) GetSnapshotUid(ctx context.Context, filesystemName string, filesetName string, snapName string) (string, error) {
	klog.V(4).Infof("[%s] rest_v2 GetSnapshotUid. filesystem: %s, fileset: %s, snapshot: %s ", utils.GetLoggerId(ctx), filesystemName, filesetName, snapName)

	getSnapshotURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots/%s", filesystemName, filesetName, snapName)
	getSnapshotResponse := GetSnapshotResponse_v2{}

	err := s.doHTTP(ctx, getSnapshotURL, "GET", &getSnapshotResponse, nil)
	if err != nil {
		return "", fmt.Errorf("unable to list snapshot %v", snapName)
	}

	if len(getSnapshotResponse.Snapshots) == 0 {
		return "", fmt.Errorf("unable to list snapshot %v", snapName)
	}

	return fmt.Sprintf("%d", getSnapshotResponse.Snapshots[0].SnapID), nil
}

// CheckIfSnapshotExist Checking if snapshot exist in fileset
func (s *SpectrumRestV2) CheckIfSnapshotExist(ctx context.Context, filesystemName string, filesetName string, snapshotName string) (bool, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 CheckIfSnapshotExist. filesystem: %s, fileset: %s, snapshot: %s ", loggerId, filesystemName, filesetName, snapshotName)

	getSnapshotURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots/%s", filesystemName, filesetName, snapshotName)
	getSnapshotResponse := GetSnapshotResponse_v2{}

	err := s.doHTTP(ctx, getSnapshotURL, "GET", &getSnapshotResponse, nil)
	if err != nil {
		if strings.Contains(getSnapshotResponse.Status.Message, "Invalid value in 'snapshotName'") && len(getSnapshotResponse.Snapshots) == 0 {
			// snapshot is not present
			return false, nil
		}
		return false, fmt.Errorf("unable to get snapshot details for filesystem: %v, fileset: %v and snapshot: %v", filesystemName, filesetName, snapshotName)
	}
	return true, nil
}

// ListFilesetSnapshots Return list of snapshot under fileset, true if snapshots present
func (s *SpectrumRestV2) ListFilesetSnapshots(ctx context.Context, filesystemName string, filesetName string) ([]Snapshot_v2, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 ListFilesetSnapshots. filesystem: %s, fileset: %s", loggerId, filesystemName, filesetName)

	listFilesetSnapshotURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/snapshots", filesystemName, filesetName)
	listFilesetSnapshotResponse := GetSnapshotResponse_v2{}

	err := s.doHTTP(ctx, listFilesetSnapshotURL, "GET", &listFilesetSnapshotResponse, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to list snapshots for fileset %v. Error [%v]", filesetName, err)
	}

	return listFilesetSnapshotResponse.Snapshots, nil
}

func (s *SpectrumRestV2) CheckIfFileDirPresent(ctx context.Context, filesystemName string, relPath string) (bool, error) {
	klog.V(4).Infof("[%s] rest_v2 CheckIfFileDirPresent. filesystem: %s, path: %s", utils.GetLoggerId(ctx), filesystemName, relPath)

	RelPath := strings.ReplaceAll(relPath, "/", "%2F")
	checkFilDirUrl := fmt.Sprintf("scalemgmt/v2/filesystems/%s/owner/%s", filesystemName, RelPath)
	ownerResp := OwnerResp_v2{}

	err := s.doHTTP(ctx, checkFilDirUrl, "GET", &ownerResp, nil)
	if err != nil {
		if strings.Contains(ownerResp.Status.Message, "File not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *SpectrumRestV2) CreateSymLink(ctx context.Context, SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error {
	klog.V(4).Infof("[%s] rest_v2 CreateSymLink. SlnkfilesystemName: %s, TargetFs: %s, relativePath: %s, LnkPath: %s", utils.GetLoggerId(ctx), SlnkfilesystemName, TargetFs, relativePath, LnkPath)

	symLnkReq := SymLnkRequest{}
	symLnkReq.FilesystemName = TargetFs
	symLnkReq.RelativePath = relativePath

	LnkPath = strings.ReplaceAll(LnkPath, "/", "%2F")

	symLnkUrl := fmt.Sprintf("scalemgmt/v2/filesystems/%s/symlink/%s", SlnkfilesystemName, LnkPath)

	makeSlnkResp := GenericResponse{}

	err := s.doHTTP(ctx, symLnkUrl, "POST", &makeSlnkResp, symLnkReq)

	if err != nil {
		return err
	}

	err = s.isRequestAccepted(ctx, makeSlnkResp, symLnkUrl)
	if err != nil {
		return err
	}

	err = s.WaitForJobCompletion(ctx, makeSlnkResp.Status.Code, makeSlnkResp.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0762C") { // job failed as dir already exists
			return nil
		}
	}
	return err
}

func (s *SpectrumRestV2) IsNodeComponentHealthy(ctx context.Context, nodeName string, component string) (bool, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 GetNodeHealthStates, nodeName: %s, component: %s", loggerId, nodeName, component)

	getNodeHealthStatesURL := fmt.Sprintf("scalemgmt/v2/nodes/%s/health/states?filter=state=HEALTHY,entityType=NODE,component=%s", nodeName, component)
	getNodeHealthStatesResponse := GetNodeHealthStatesResponse_v2{}

	err := s.doHTTP(ctx, getNodeHealthStatesURL, "GET", &getNodeHealthStatesResponse, nil)
	if err != nil {
		return false, fmt.Errorf("unable to get health states for nodename %v", nodeName)
	}

	if len(getNodeHealthStatesResponse.States) == 0 {
		return false, nil
	}

	return true, nil
}

func (s *SpectrumRestV2) SetFilesystemPolicy(ctx context.Context, policy *Policy, filesystemName string) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 setFilesystemPolicy for filesystem %s", loggerId, filesystemName)

	setPolicyURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/policies", filesystemName)
	setPolicyResponse := GenericResponse{}

	err := s.doHTTP(ctx, setPolicyURL, "PUT", &setPolicyResponse, policy)
	if err != nil {
		klog.Errorf("[%s] unable to send filesystem policy: %v", loggerId, setPolicyResponse.Status.Message)
		return err
	}

	err = s.WaitForJobCompletion(ctx, setPolicyResponse.Status.Code, setPolicyResponse.Jobs[0].JobID)
	if err != nil {
		klog.Errorf("[%s] setting policy rule %s for filesystem %s failed with error %v", loggerId, policy.Policy, filesystemName, err)
		return err
	}

	return nil
}

func (s *SpectrumRestV2) DoesTierExist(ctx context.Context, tierName string, filesystemName string) error {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 DoesTierExist. name %s, filesystem %s", loggerId, tierName, filesystemName)

	_, err := s.GetTierInfoFromName(ctx, tierName, filesystemName)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid value in 'storagePool'") {
			return fmt.Errorf("invalid tier '%s' specified for filesystem %s", tierName, filesystemName)
		}
		return err
	}

	return nil
}

func (s *SpectrumRestV2) GetTierInfoFromName(ctx context.Context, tierName string, filesystemName string) (*StorageTier, error) {
	klog.V(4).Infof("[%s] rest_v2 GetTierInfoFromName. name %s, filesystem %s", utils.GetLoggerId(ctx), tierName, filesystemName)

	tierUrl := fmt.Sprintf("scalemgmt/v2/filesystems/%s/pools/%s", filesystemName, tierName)
	getTierResponse := &StorageTiers{}

	err := s.doHTTP(ctx, tierUrl, "GET", getTierResponse, nil)
	if err != nil {
		klog.Errorf("Unable to get tier: %s err: %v", tierName, err)
		return nil, err
	}

	if len(getTierResponse.StorageTiers) > 0 {
		return &getTierResponse.StorageTiers[0], nil
	} else {
		return nil, fmt.Errorf("unable to fetch storage tiers for %s", filesystemName)
	}
}

func (s *SpectrumRestV2) CheckIfDefaultPolicyPartitionExists(ctx context.Context, partitionName string, filesystemName string) bool {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 CheckIfDefaultPolicyPartitionExists. name %s, filesystem %s", loggerId, partitionName, filesystemName)

	partitionURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/partition/%s", filesystemName, partitionName)
	getPartitionResponse := GenericResponse{}

	// If it does or doesn't exist and we get an error we will default to just setting it again as an override
	err := s.doHTTP(ctx, partitionURL, "GET", &getPartitionResponse, nil)
	return err == nil
}

func (s *SpectrumRestV2) GetFirstDataTier(ctx context.Context, filesystemName string) (string, error) {
	loggerId := GetLoggerId(ctx)
	klog.V(4).Infof("[%s] rest_v2 GetFirstDataTier. filesystem %s", loggerId, filesystemName)

	tiersURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/pools", filesystemName)
	getTierResponse := &StorageTiers{}

	err := s.doHTTP(ctx, tiersURL, "GET", getTierResponse, nil)
	if err != nil {
		return "", err
	}

	for _, tier := range getTierResponse.StorageTiers {
		if tier.StorageTierName == "system" {
			continue
		}

		tierInfo, err := s.GetTierInfoFromName(ctx, tier.StorageTierName, tier.FilesystemName)
		if err != nil {
			return "", err
		}
		if tierInfo.TotalDataInKB > 0 {
			klog.Infof("[%s] GetFirstDataTier: Setting default tier to %s", loggerId, tierInfo.StorageTierName)
			return tierInfo.StorageTierName, nil
		}
	}

	klog.V(6).Infof("[%s] GetFirstDataTier: Defaulting to system tier", loggerId)
	return "system", nil
}

// getNextEndpoint returns the next endpoint to be used for
// GUI REST calls. This function gets called when current
// endpoint is not active.
func (s *SpectrumRestV2) getNextEndpoint(ctx context.Context) string {
	len := len(s.Endpoint)
	s.EndPointIndex++
	if s.EndPointIndex >= len {
		s.EndPointIndex = s.EndPointIndex % len
	}
	endpoint := s.Endpoint[s.EndPointIndex]
	klog.V(6).Infof("[%s] getNextEndpoint: returning next endpoint: %s", utils.GetLoggerId(ctx), endpoint)
	return endpoint
}
