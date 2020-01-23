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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type spectrumRestV2 struct {
	httpClient *http.Client
	endpoint   string
	user       string
	password   string
}

func (s *spectrumRestV2) isStatusOK(statusCode int) bool {
	glog.V(4).Infof("rest_v2 isStatusOK. statusCode: %d", statusCode)

	if (statusCode == http.StatusOK) ||
		(statusCode == http.StatusCreated) ||
		(statusCode == http.StatusAccepted) {
		return true
	}
	return false
}

func (s *spectrumRestV2) checkAsynchronousJob(statusCode int) bool {
	glog.V(4).Infof("rest_v2 checkAsynchronousJob. statusCode: %d", statusCode)
	if (statusCode == http.StatusAccepted) ||
		(statusCode == http.StatusCreated) {
		return true
	}
	return false
}

func (s *spectrumRestV2) isRequestAccepted(response GenericResponse, url string) error {
	glog.V(4).Infof("rest_v2 isRequestAccepted. url: %s, response: %v", url, response)

	if !s.isStatusOK(response.Status.Code) {
		return fmt.Errorf("error %v for url %v", response, url)
	}

	if len(response.Jobs) == 0 {
		return fmt.Errorf("Unable to get Job details for %s request: %v", url, response)
	}
	return nil
}

func (s *spectrumRestV2) waitForJobCompletion(statusCode int, jobID uint64) error {
	glog.V(4).Infof("rest_v2 waitForJobCompletion. jobID: %d, statusCode: %d", jobID, statusCode)

	if s.checkAsynchronousJob(statusCode) {
		jobURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/jobs/%d?fields=:all:", jobID))
		err := s.AsyncJobCompletion(jobURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *spectrumRestV2) AsyncJobCompletion(jobURL string) error {
	glog.V(4).Infof("rest_v2 AsyncJobCompletion. jobURL: %s", jobURL)

	jobQueryResponse := GenericResponse{}
	for {
		err := s.doHTTP(jobURL, "GET", &jobQueryResponse, nil)
		if err != nil {
			return err
		}
		if len(jobQueryResponse.Jobs) == 0 {
			return fmt.Errorf("Unable to get Job details for %s: %v", jobURL, jobQueryResponse)
		}

		if jobQueryResponse.Jobs[0].Status == "RUNNING" {
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		break
	}
	if jobQueryResponse.Jobs[0].Status == "COMPLETED" {
		return nil
	} else {
		glog.Errorf("Async Job failed: %v", jobQueryResponse)
		return fmt.Errorf("%v", jobQueryResponse.Jobs[0].Result.Stderr)
	}
}

func NewSpectrumRestV2(scaleConfig settings.Clusters) (SpectrumScaleConnector, error) {
	glog.V(4).Infof("rest_v2 NewSpectrumRestV2.")

	guiHost := scaleConfig.RestAPI[0].GuiHost
	guiUser := scaleConfig.MgmtUsername
	guiPwd := scaleConfig.MgmtPassword
	guiPort := scaleConfig.RestAPI[0].GuiPort
	if guiPort == 0 {
		guiPort = settings.DefaultGuiPort
	}

	var tr *http.Transport
	endpoint := fmt.Sprintf("%s://%s:%d/", settings.GuiProtocol, guiHost, guiPort)

	if scaleConfig.SecureSslMode {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(scaleConfig.CacertValue); !ok {
			return &spectrumRestV2{}, fmt.Errorf("Parsing CA cert %v failed", scaleConfig.Cacert)
		}
		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}
		glog.V(4).Infof("Created Spectrum Scale connector with SSL mode for %v", guiHost)
	} else {
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} //nolint:gosec //InsecureSkipVerify was requested by user.
		glog.V(4).Infof("Created Spectrum Scale connector without SSL mode for %v", guiHost)
	}

	return &spectrumRestV2{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   time.Second * 10,
		},
		endpoint: endpoint,
		user:     guiUser,
		password: guiPwd,
	}, nil
}

func (s *spectrumRestV2) GetClusterId() (string, error) {
	glog.V(4).Infof("rest_v2 GetClusterId")

	getClusterURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/cluster")
	getClusterResponse := GetClusterResponse{}

	err := s.doHTTP(getClusterURL, "GET", &getClusterResponse, nil)
	if err != nil {
		glog.Errorf("Unable to get cluster ID: %v", err)
		return "", err
	}
	cid_str := fmt.Sprintf("%v", getClusterResponse.Cluster.ClusterSummary.ClusterID)
	return cid_str, nil
}

func (s *spectrumRestV2) GetFilesystemMountDetails(filesystemName string) (MountInfo, error) {
	glog.V(4).Infof("rest_v2 GetFilesystemMountDetails. filesystemName: %s", filesystemName)

	getFilesystemURL := fmt.Sprintf("%s%s%s", s.endpoint, "scalemgmt/v2/filesystems/", filesystemName)
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		glog.Errorf("Unable to get filesystem details for %s: %v", filesystemName, err)
		return MountInfo{}, err
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount, nil
	} else {
		return MountInfo{}, fmt.Errorf("Unable to fetch mount details for %s", filesystemName)
	}
}

func (s *spectrumRestV2) IsFilesystemMounted(filesystemName string) (bool, error) {
	glog.V(4).Infof("rest_v2 IsFilesystemMounted. filesystemName: %s", filesystemName)

	ownerResp := OwnerResp_v2{}
	ownerUrl := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/owner/%s", filesystemName, url.QueryEscape("/")))
	err := s.doHTTP(ownerUrl, "GET", &ownerResp, nil)
	if err != nil {
		glog.Errorf("Error in getting owner info for filesystem %s: %v", filesystemName, err)
		return false, err
	}
	return true, nil
}

func (s *spectrumRestV2) ListFilesystems() ([]string, error) {
	glog.V(4).Infof("rest_v2 ListFilesystems")

	listFilesystemsURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/filesystems")
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(listFilesystemsURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		glog.Errorf("Error in listing filesystems: %v", err)
		return nil, err
	}
	fsNumber := len(getFilesystemResponse.FileSystems)
	filesystems := make([]string, fsNumber)
	for i := 0; i < fsNumber; i++ {
		filesystems[i] = getFilesystemResponse.FileSystems[i].Name
	}
	return filesystems, nil
}

func (s *spectrumRestV2) GetFilesystemMountpoint(filesystemName string) (string, error) {
	glog.V(4).Infof("rest_v2 GetFilesystemMountpoint. filesystemName: %s", filesystemName)

	getFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName))
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		glog.Errorf("Error in getting filesystem details for %s: %v", filesystemName, err)
		return "", err
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount.MountPoint, nil
	} else {
		return "", fmt.Errorf("Unable to fetch mount point for %s.", filesystemName)
	}
}

func (s *spectrumRestV2) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
	glog.V(4).Infof("rest_v2 CreateFileset. filesystem: %s, fileset: %s, opts: %v", filesystemName, filesetName, opts)

	filesetreq := CreateFilesetRequest{}
	filesetreq.FilesetName = filesetName
	filesetreq.Comment = "Fileset created by IBM Container Storage Interface driver"

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
			filesetreq.AllocInodes = inodeLimit.(string)
		}
	}

	uid, uidSpecified := opts[UserSpecifiedUID]
	gid, gidSpecified := opts[UserSpecifiedGID]

	if uidSpecified && gidSpecified {
		filesetreq.Owner = fmt.Sprintf("%s:%s", uid, gid)
	} else if uidSpecified {
		filesetreq.Owner = fmt.Sprintf("%s", uid)
	}

	createFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName))
	createFilesetResponse := GenericResponse{}

	err := s.doHTTP(createFilesetURL, "POST", &createFilesetResponse, filesetreq)
	if err != nil {
		glog.Errorf("Error in create fileset request: %v", err)
		return err
	}

	err = s.isRequestAccepted(createFilesetResponse, createFilesetURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(createFilesetResponse.Status.Code, createFilesetResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSP1102C") { // job failed as fileset already exists
			fmt.Println(err)
			return nil
		}
		glog.Errorf("Unable to create fileset %s: %v", filesetName, err)
		return err
	}
	return nil
}

func (s *spectrumRestV2) DeleteFileset(filesystemName string, filesetName string) error {
	glog.V(4).Infof("rest_v2 DeleteFileset. filesystem: %s, fileset: %s", filesystemName, filesetName)

	deleteFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	deleteFilesetResponse := GenericResponse{}

	err := s.doHTTP(deleteFilesetURL, "DELETE", &deleteFilesetResponse, nil)
	if err != nil {
		if strings.Contains(deleteFilesetResponse.Status.Message, "Invalid value in 'fsetName'") { // job failed as dir already exists
			glog.Infof("Fileset would have been deleted. So returning success %v", err)
			return nil
		}

		glog.Errorf("Error in delete fileset request: %v", err)
		return err
	}

	err = s.isRequestAccepted(deleteFilesetResponse, deleteFilesetURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(deleteFilesetResponse.Status.Code, deleteFilesetResponse.Jobs[0].JobID)
	if err != nil {
		glog.Errorf("Unable to delete fileset %s: %v", filesetName, err)
		return err
	}

	return nil
}

func (s *spectrumRestV2) LinkFileset(filesystemName string, filesetName string, linkpath string) error {
	glog.V(4).Infof("rest_v2 LinkFileset. filesystem: %s, fileset: %s, linkpath: %s", filesystemName, filesetName, linkpath)

	linkReq := LinkFilesetRequest{}
	linkReq.Path = linkpath
	linkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link", filesystemName, filesetName))
	linkFilesetResponse := GenericResponse{}

	err := s.doHTTP(linkFilesetURL, "POST", &linkFilesetResponse, linkReq)
	if err != nil {
		glog.Errorf("Error in link fileset request: %v", err)
		return err
	}

	err = s.isRequestAccepted(linkFilesetResponse, linkFilesetURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(linkFilesetResponse.Status.Code, linkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		glog.Errorf("Error in linking fileset %s: %v", filesetName, err)
		return err
	}
	return nil
}

func (s *spectrumRestV2) UnlinkFileset(filesystemName string, filesetName string) error {
	glog.V(4).Infof("rest_v2 UnlinkFileset. filesystem: %s, fileset: %s", filesystemName, filesetName)

	unlinkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link?force=True", filesystemName, filesetName))
	unlinkFilesetResponse := GenericResponse{}

	err := s.doHTTP(unlinkFilesetURL, "DELETE", &unlinkFilesetResponse, nil)

	if err != nil {
		glog.Errorf("Error in unlink fileset request: %v", err)
		return err
	}

	err = s.isRequestAccepted(unlinkFilesetResponse, unlinkFilesetURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(unlinkFilesetResponse.Status.Code, unlinkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		glog.Errorf("Error in unlink fileset %s: %v", filesetName, err)
		return err
	}

	return nil
}

func (s *spectrumRestV2) ListFileset(filesystemName string, filesetName string) (Fileset_v2, error) {
	glog.V(4).Infof("rest_v2 ListFileset. filesystem: %s, fileset: %s", filesystemName, filesetName)

	getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		glog.Errorf("Error in list fileset request: %v", err)
		return Fileset_v2{}, err
	}

	if len(getFilesetResponse.Filesets) == 0 {
		glog.Errorf("No fileset returned for %s", filesetName)
		return Fileset_v2{}, fmt.Errorf("No fileset returned for %s", filesetName)
	}

	return getFilesetResponse.Filesets[0], nil
}

func (s *spectrumRestV2) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	glog.V(4).Infof("rest_v2 IsFilesetLinked. filesystem: %s, fileset: %s", filesystemName, filesetName)

	fileset, err := s.ListFileset(filesystemName, filesetName)
	if err != nil {
		return false, err
	}

	if (fileset.Config.Path == "") ||
		(fileset.Config.Path == "--") {
		return false, nil
	}
	return true, nil
}

func (s *spectrumRestV2) MakeDirectory(filesystemName string, relativePath string, uid string, gid string) error {
	glog.V(4).Infof("rest_v2 MakeDirectory. filesystem: %s, path: %s, uid: %s, gid: %s", filesystemName, relativePath, uid, gid)

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
	makeDirURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, formattedPath))

	makeDirResponse := GenericResponse{}

	err := s.doHTTP(makeDirURL, "POST", &makeDirResponse, dirreq)

	if err != nil {
		glog.Errorf("Error in make directory request: %v", err)
		return err
	}

	err = s.isRequestAccepted(makeDirResponse, makeDirURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(makeDirResponse.Status.Code, makeDirResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0762C") { // job failed as dir already exists
			glog.Infof("Directory exists. %v", err)
			return nil
		}

		glog.Errorf("Unable to make directory %s: %v.", relativePath, err)
		return err
	}

	return nil
}

func (s *spectrumRestV2) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	glog.V(4).Infof("rest_v2 SetFilesetQuota. filesystem: %s, fileset: %s, quota: %s", filesystemName, filesetName, quota)

	setQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName))
	quotaRequest := SetQuotaRequest_v2{}

	quotaRequest.BlockHardLimit = quota
	quotaRequest.BlockSoftLimit = quota
	quotaRequest.OperationType = "setQuota"
	quotaRequest.QuotaType = "fileset"
	quotaRequest.ObjectName = filesetName

	setQuotaResponse := GenericResponse{}

	err := s.doHTTP(setQuotaURL, "POST", &setQuotaResponse, quotaRequest)
	if err != nil {
		glog.Errorf("Error in set fileset quota request: %v", err)
		return err
	}

	err = s.isRequestAccepted(setQuotaResponse, setQuotaURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(setQuotaResponse.Status.Code, setQuotaResponse.Jobs[0].JobID)
	if err != nil {
		glog.Errorf("Unable to set quota for fileset %s: %v", filesetName, err)
		return err
	}
	return nil
}

func (s *spectrumRestV2) CheckIfFSQuotaEnabled(filesystemName string) error {
	glog.V(4).Infof("rest_v2 CheckIfFSQuotaEnabled. filesystem: %s", filesystemName)

	checkQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName))
	QuotaResponse := GetQuotaResponse_v2{}

	err := s.doHTTP(checkQuotaURL, "GET", &QuotaResponse, nil)
	if err != nil {
		glog.Errorf("Error in check quota: %v", err)
		return err
	}
	return nil
}

func (s *spectrumRestV2) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	glog.V(4).Infof("rest_v2 ListFilesetQuota. filesystem: %s, fileset: %s", filesystemName, filesetName)

	listQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas?filter=objectName=%s", filesystemName, filesetName))
	listQuotaResponse := GetQuotaResponse_v2{}

	err := s.doHTTP(listQuotaURL, "GET", &listQuotaResponse, nil)
	if err != nil {
		glog.Errorf("Unable to fetch quota information: %v", err)
		return "", err
	}

	//TODO check which quota in quotas[] and which attribute
	if len(listQuotaResponse.Quotas) > 0 {
		return fmt.Sprintf("%dK", listQuotaResponse.Quotas[0].BlockLimit), nil
	} else {
		glog.Errorf("No quota information found for fileset %s: %s", filesetName, err)
		return "", err
	}
}

func (s *spectrumRestV2) doHTTP(endpoint string, method string, responseObject interface{}, param interface{}) error {
	glog.V(4).Infof("rest_v2 doHTTP. endpoint: %s, method: %s, param: %v", endpoint, method, param)

	response, err := utils.HttpExecuteUserAuth(s.httpClient, method, endpoint, s.user, s.password, param)
	if err != nil {
		glog.Errorf("Error in authentication request: %v", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return status.Error(codes.Unauthenticated, fmt.Sprintf("Unauthorized %s request to %v: %v", method, endpoint, response.Status))
	}

	err = utils.UnmarshalResponse(response, responseObject)
	if err != nil {
		return err
	}

	if !s.isStatusOK(response.StatusCode) {
		return fmt.Errorf("Remote call completed with error [%v]", response.Status)
	}

	return nil
}

func (s *spectrumRestV2) MountFilesystem(filesystemName string, nodeName string) error { //nolint:dupl
	glog.V(4).Infof("rest_v2 MountFilesystem. filesystem: %s, node: %s", filesystemName, nodeName)

	mountreq := MountFilesystemRequest{}
	mountreq.Nodes = append(mountreq.Nodes, nodeName)

	mountFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/mount", filesystemName))
	mountFilesystemResponse := GenericResponse{}

	err := s.doHTTP(mountFilesystemURL, "PUT", &mountFilesystemResponse, mountreq)
	if err != nil {
		glog.Errorf("Error in mount filesystem request: %v", err)
		return err
	}

	err = s.isRequestAccepted(mountFilesystemResponse, mountFilesystemURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(mountFilesystemResponse.Status.Code, mountFilesystemResponse.Jobs[0].JobID)
	if err != nil {
		glog.Errorf("Unable to Mount filesystem %s on node %s: %v", filesystemName, nodeName, err)
		return err
	}
	return nil
}

func (s *spectrumRestV2) UnmountFilesystem(filesystemName string, nodeName string) error { //nolint:dupl
	glog.V(4).Infof("rest_v2 UnmountFilesystem. filesystem: %s, node: %s", filesystemName, nodeName)

	unmountreq := UnmountFilesystemRequest{}
	unmountreq.Nodes = append(unmountreq.Nodes, nodeName)

	unmountFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/unmount", filesystemName))
	unmountFilesystemResponse := GenericResponse{}

	err := s.doHTTP(unmountFilesystemURL, "PUT", &unmountFilesystemResponse, unmountreq)
	if err != nil {
		glog.Errorf("Error in unmount filesystem request: %v", err)
		return err
	}

	err = s.isRequestAccepted(unmountFilesystemResponse, unmountFilesystemURL)
	if err != nil {
		glog.Errorf("Request not accepted for processing: %v", err)
		return err
	}

	err = s.waitForJobCompletion(unmountFilesystemResponse.Status.Code, unmountFilesystemResponse.Jobs[0].JobID)
	if err != nil {
		glog.Errorf("Unable to unmount filesystem %s on node %s: %v", filesystemName, nodeName, err)
		return err
	}

	return nil
}

func (s *spectrumRestV2) GetFilesystemName(filesystemUUID string) (string, error) {
	glog.V(4).Infof("rest_v2 GetFilesystemName. UUID: %s", filesystemUUID)

	getFilesystemNameURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems?filter=uuid=%s", filesystemUUID))
	getFilesystemNameURLResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(getFilesystemNameURL, "GET", &getFilesystemNameURLResponse, nil)
	if err != nil {
		glog.Errorf("Unable to get filesystem name for uuid %s: %v", filesystemUUID, err)
		return "", err
	}

	if len(getFilesystemNameURLResponse.FileSystems) == 0 {
		glog.Errorf("Unable to fetch filesystem name details for %s", filesystemUUID)
		return "", fmt.Errorf("Unable to fetch filesystem name details for %s", filesystemUUID)
	}
	return getFilesystemNameURLResponse.FileSystems[0].Name, nil
}

func (s *spectrumRestV2) GetFsUid(filesystemName string) (string, error) {
	getFilesystemURL := fmt.Sprintf("%s%s%s", s.endpoint, "scalemgmt/v2/filesystems/", filesystemName)
	getFilesystemResponse := GetFilesystemResponse_v2{}

	err := s.doHTTP(getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		return "", fmt.Errorf("Unable to get filesystem details for %s", filesystemName)
	}

	fmt.Println(getFilesystemResponse)
	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].UUID, nil
	} else {
		return "", fmt.Errorf("Unable to fetch mount details for %s", filesystemName)
	}
}

func (s *spectrumRestV2) DeleteSymLnk(filesystemName string, LnkName string) error {
	LnkName = strings.ReplaceAll(LnkName, "/", "%2F")
	deleteLnkURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/symlink/%s", filesystemName, LnkName))
	deleteLnkResponse := GenericResponse{}

	err := s.doHTTP(deleteLnkURL, "DELETE", &deleteLnkResponse, nil)
	if err != nil {
		return fmt.Errorf("Unable to delete Symlink %v.", LnkName)
	}

	err = s.isRequestAccepted(deleteLnkResponse, deleteLnkURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(deleteLnkResponse.Status.Code, deleteLnkResponse.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG2006C") {
			glog.V(4).Infof("Since slink %v was already deleted, so returning success", LnkName)
			return nil
		}
		return fmt.Errorf("Unable to delete symLnk %v:%v.", LnkName, err)
	}

	return nil
}

func (s *spectrumRestV2) DeleteDirectory(filesystemName string, dirName string) error {
	NdirName := strings.ReplaceAll(dirName, "/", "%2F")
	deleteDirURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, NdirName))
	deleteDirResponse := GenericResponse{}

	err := s.doHTTP(deleteDirURL, "DELETE", &deleteDirResponse, nil)
	if err != nil {
		return fmt.Errorf("Unable to delete dir %v.", dirName)
	}

	err = s.isRequestAccepted(deleteDirResponse, deleteDirURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(deleteDirResponse.Status.Code, deleteDirResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to delete dir %v:%v", dirName, err)
	}

	return nil
}

func (s *spectrumRestV2) GetFileSetUid(filesystemName string, filesetName string) (string, error) {
	getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		return "", fmt.Errorf("Unable to list fileset %v.", filesetName)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return "", fmt.Errorf("Unable to list fileset %v.", filesetName)
	}

	return fmt.Sprintf("%d", getFilesetResponse.Filesets[0].Config.Id), nil
}

func (s *spectrumRestV2) GetFileSetNameFromId(filesystemName string, Id string) (string, error) {
	getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets?filter=config.id=%s", filesystemName, Id))

	getFilesetResponse := GetFilesetResponse_v2{}

	err := s.doHTTP(getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		return "", fmt.Errorf("Unable to get name for fileset Id %v:%v.", filesystemName, Id)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return "", nil
	}

	return getFilesetResponse.Filesets[0].FilesetName, nil
}

func (s *spectrumRestV2) CheckIfFileDirPresent(filesystemName string, relPath string) (bool, error) {
	RelPath := strings.ReplaceAll(relPath, "/", "%2F")
	checkFilDirUrl := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/owner/%s", filesystemName, RelPath))
	ownerResp := OwnerResp_v2{}
	err := s.doHTTP(checkFilDirUrl, "GET", &ownerResp, nil)
	if err != nil {
		if strings.Contains(ownerResp.Status.Message, "File not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *spectrumRestV2) CreateSymLink(SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error {
	symLnkReq := SymLnkRequest{}
	symLnkReq.FilesystemName = TargetFs
	symLnkReq.RelativePath = relativePath

	LnkPath = strings.ReplaceAll(LnkPath, "/", "%2F")

	symLnkUrl := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/symlink/%s", SlnkfilesystemName, LnkPath))

	makeSlnkResp := GenericResponse{}

	err := s.doHTTP(symLnkUrl, "POST", &makeSlnkResp, symLnkReq)

	if err != nil {
		return err
	}

	err = s.isRequestAccepted(makeSlnkResp, symLnkUrl)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(makeSlnkResp.Status.Code, makeSlnkResp.Jobs[0].JobID)
	if err != nil {
		if strings.Contains(err.Error(), "EFSSG0762C") { // job failed as dir already exists
			return nil
		}
	}
	return err
}
