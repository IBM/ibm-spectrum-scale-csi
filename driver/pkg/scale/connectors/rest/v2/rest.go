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

package rest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/connectors/api/v2"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

type httpMethod int

const (
	GET = iota
	POST
	PUT
	DELETE
)

func (m httpMethod) String() string {
	return [...]string{"GET", "POST", "PUT", "DELETE"}[m]
}

type Connector struct {
	connectors.HasClusterID
	HttpClient SpectrumRestV2
}

type SpectrumRestV2 struct {
	httpClient *http.Client
	configMap  *settings.ConfigMap
}

func NewSpectrumV2(scaleConfig *settings.ConfigMap) SpectrumRestV2 {
	klog.V(4).Infof("rest_v2 NewSpectrumRestV2.")

	httpTrans := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            scaleConfig.RootCAs,
			InsecureSkipVerify: scaleConfig.InsecureSkipTLSVerify,
		},
	}
	httpClient := &http.Client{
		Transport: httpTrans,
		Timeout:   time.Second * 10,
	}

	return SpectrumRestV2{
		httpClient: httpClient,
		configMap:  scaleConfig,
	}
}

func isStatusOK(statusCode int) bool {
	klog.V(4).Infof("rest_v2 isStatusOK. statusCode: %d", statusCode)

	if (statusCode == http.StatusOK) ||
		(statusCode == http.StatusCreated) ||
		(statusCode == http.StatusAccepted) {
		return true
	}
	return false
}

func isRequestAccepted(response api.GenericResponse, url string) error {
	klog.V(4).Infof("rest_v2 isRequestAccepted. url: %s, response: %v", url, response)

	if !isStatusOK(response.Status.Code) {
		return fmt.Errorf("error %v for url %v", response, url)
	}

	if len(response.Jobs) == 0 {
		return fmt.Errorf("Unable to get Job details for %s request: %v", url, response)
	}
	return nil
}

func checkAsynchronousJob(statusCode int) bool {
	klog.V(4).Infof("rest_v2 checkAsynchronousJob. statusCode: %d", statusCode)
	if (statusCode == http.StatusAccepted) ||
		(statusCode == http.StatusCreated) {
		return true
	}
	return false
}

func (r SpectrumRestV2) NewConnector(in connectors.HasClusterID) connectors.Connector {
	return &Connector{
		HasClusterID: in,
		HttpClient:   r,
	}
}

func (s *Connector) waitForJobCompletion(statusCode int, jobID uint64) error {
	klog.V(4).Infof("rest_v2 waitForJobCompletion. jobID: %d, statusCode: %d", jobID, statusCode)

	if checkAsynchronousJob(statusCode) {
		jobURL := fmt.Sprintf("scalemgmt/v2/jobs/%d?fields=:all:", jobID)
		err := s.asyncJobCompletion(jobURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Connector) asyncJobCompletion(jobURL string) error {
	klog.V(4).Infof("rest_v2 AsyncJobCompletion. jobURL: %s", jobURL)

	jobQueryResponse := api.GenericResponse{}
	for {
		err := s.doHTTP(GET, jobURL, nil, &jobQueryResponse)
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
		klog.Errorf("Async Job failed: %v", jobQueryResponse)
		return fmt.Errorf("%v", jobQueryResponse.Jobs[0].Result.Stderr)
	}
}

func newRequest(method httpMethod, relPath string, body interface{}) (*http.Request, error) {
	klog.V(5).Infof(`request method: %v, relPath: %s`, method, relPath)
	klog.V(6).Infof(`request body: %v`, body)

	encBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(`json.Marshal of '%#v' had error: %w`, body, err)
	}

	return http.NewRequest(method.String(), relPath, bytes.NewBuffer(encBody))
}

func (s *Connector) doHTTP(method httpMethod, urlPath string, requestParam interface{}, responseObject interface{}) error {
	//we for loop instead of map here because we expect clusters to be 1 or 2, max maybe 10
	for _, cluster := range s.HttpClient.configMap.Clusters {
		if cluster.ID == s.ClusterID() {

			request, err := newRequest(method, "/"+urlPath, requestParam)
			if err != nil {
				return fmt.Errorf("could not create request for %v: %#v", urlPath, err)
			}
			request.URL.Scheme = settings.GuiProtocol
			request.URL.Host = net.JoinHostPort(
				cluster.RestAPI[0].GuiHost,
				strconv.Itoa(cluster.RestAPI[0].GuiPort),
			)
			request.Host = request.URL.Host //Host header

			request.Header.Add("Content-Type", "application/json")
			request.Header.Add("Accept", "application/json")

			klog.V(4).Infof("doHTTP method: %v, requestURL: %v, requestParam: %v", method, request.URL, requestParam)
			klog.V(6).Infof("doHTTP request: %+v", request)

			request.SetBasicAuth(cluster.MgmtUsername, cluster.MgmtPassword)

			response, err := s.HttpClient.httpClient.Do(request)
			if err != nil {
				return err
			}
			defer response.Body.Close()

			if response.StatusCode == http.StatusUnauthorized {
				return status.Error(codes.Unauthenticated, fmt.Sprintf("%v %v request to %v", response.Status, method, request.URL))
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("ioutil.ReadAll failed %v", err)
			}

			err = json.Unmarshal(body, responseObject)
			if err != nil {
				//json.InvalidUnmarshalError
				return fmt.Errorf("json.Unmarshal failed %v", err)
			}

			//not the best way to do this, as status is not error, and status has meaning...
			//but existing code relies on it.
			if !isStatusOK(response.StatusCode) {
				return fmt.Errorf("Remote call completed with error [%v]", response.Status)
			}

			return nil
		}
	}
	return fmt.Errorf("could not find matching cluster ID in config map")
}

func (s *Connector) GetClusterId() (string, error) {
	glog.V(4).Infof("rest_v2 GetClusterId")

	getClusterURL := "scalemgmt/v2/cluster"
	getClusterResponse := api.GetClusterResponse{}

	err := s.doHTTP(GET, getClusterURL, nil, &getClusterResponse)
	if err != nil {
		glog.Errorf("Unable to get cluster ID: %v", err)
		return "", err
	}
	cid_str := fmt.Sprintf("%v", getClusterResponse.Cluster.ClusterSummary.ClusterID)
	return cid_str, nil
}

func (s *Connector) GetFilesystemMountDetails(filesystemName string) (api.MountInfo, error) {
	glog.V(4).Infof("rest_v2 GetFilesystemMountDetails. filesystemName: %s", filesystemName)

	getFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName)
	getFilesystemResponse := api.GetFilesystemResponse_v2{}

	err := s.doHTTP(GET, getFilesystemURL, nil, &getFilesystemResponse)
	if err != nil {
		glog.Errorf("Unable to get filesystem details for %s: %v", filesystemName, err)
		return api.MountInfo{}, err
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount, nil
	} else {
		return api.MountInfo{}, fmt.Errorf("Unable to fetch mount details for %s", filesystemName)
	}
}

func (s *Connector) IsFilesystemMounted(filesystemName string) (bool, error) {
	glog.V(4).Infof("rest_v2 IsFilesystemMounted. filesystemName: %s", filesystemName)

	ownerURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/owner/%s", filesystemName, url.QueryEscape("/"))
	ownerResponse := api.OwnerResp_v2{}

	err := s.doHTTP(GET, ownerURL, nil, &ownerResponse)
	if err != nil {
		glog.Errorf("Error in getting owner info for filesystem %s: %v", filesystemName, err)
		return false, err
	}
	return true, nil
}

func (s *Connector) ListFilesystems() ([]string, error) {
	glog.V(4).Infof("rest_v2 ListFilesystems")

	listFilesystemsURL := "scalemgmt/v2/filesystems"
	getFilesystemResponse := api.GetFilesystemResponse_v2{}

	err := s.doHTTP(GET, listFilesystemsURL, nil, &getFilesystemResponse)
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

func (s *Connector) GetFilesystemMountpoint(filesystemName string) (string, error) {
	glog.V(4).Infof("rest_v2 GetFilesystemMountpoint. filesystemName: %s", filesystemName)

	getFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName)
	getFilesystemResponse := api.GetFilesystemResponse_v2{}

	err := s.doHTTP(GET, getFilesystemURL, nil, &getFilesystemResponse)
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

func (s *Connector) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
	glog.V(4).Infof("rest_v2 CreateFileset. filesystem: %s, fileset: %s, opts: %v", filesystemName, filesetName, opts)

	filesetreq := api.CreateFilesetRequest{}
	filesetreq.FilesetName = filesetName
	filesetreq.Comment = "Fileset created by IBM Container Storage Interface driver"

	filesetType, filesetTypeSpecified := opts[settings.FilesetType]
	inodeLimit, inodeLimitSpecified := opts[settings.InodeLimit]

	if !filesetTypeSpecified {
		filesetType, filesetTypeSpecified = opts[settings.FilesetTypeDep]
	}

	if !inodeLimitSpecified {
		inodeLimit, inodeLimitSpecified = opts[settings.InodeLimitDep]
	}

	if filesetTypeSpecified && filesetType.(string) == "dependent" {
		/* Add fileset for dependent fileset-name: */
		parentFileSetName, parentFileSetNameSpecified := opts[settings.ParentFset]
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

	uid, uidSpecified := opts[settings.Uid]
	gid, gidSpecified := opts[settings.Gid]

	if uidSpecified && gidSpecified {
		filesetreq.Owner = fmt.Sprintf("%s:%s", uid, gid)
	} else if uidSpecified {
		filesetreq.Owner = fmt.Sprintf("%s", uid)
	}

	createFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName)
	createFilesetResponse := api.GenericResponse{}

	err := s.doHTTP(POST, createFilesetURL, filesetreq, &createFilesetResponse)
	if err != nil {
		glog.Errorf("Error in create fileset request: %v", err)
		return err
	}

	err = isRequestAccepted(createFilesetResponse, createFilesetURL)
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

func (s *Connector) DeleteFileset(filesystemName string, filesetName string) error {
	glog.V(4).Infof("rest_v2 DeleteFileset. filesystem: %s, fileset: %s", filesystemName, filesetName)

	deleteFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	deleteFilesetResponse := api.GenericResponse{}

	err := s.doHTTP(DELETE, deleteFilesetURL, nil, &deleteFilesetResponse)
	if err != nil {
		if strings.Contains(deleteFilesetResponse.Status.Message, "Invalid value in 'fsetName'") { // job failed as dir already exists
			glog.Infof("Fileset would have been deleted. So returning success %v", err)
			return nil
		}

		glog.Errorf("Error in delete fileset request: %v", err)
		return err
	}

	err = isRequestAccepted(deleteFilesetResponse, deleteFilesetURL)
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

func (s *Connector) LinkFileset(filesystemName string, filesetName string, linkpath string) error {
	glog.V(4).Infof("rest_v2 LinkFileset. filesystem: %s, fileset: %s, linkpath: %s", filesystemName, filesetName, linkpath)

	linkReq := api.LinkFilesetRequest{}
	linkReq.Path = linkpath
	linkFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link", filesystemName, filesetName)
	linkFilesetResponse := api.GenericResponse{}

	err := s.doHTTP(POST, linkFilesetURL, linkReq, &linkFilesetResponse)
	if err != nil {
		glog.Errorf("Error in link fileset request: %v", err)
		return err
	}

	err = isRequestAccepted(linkFilesetResponse, linkFilesetURL)
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

func (s *Connector) UnlinkFileset(filesystemName string, filesetName string) error {
	glog.V(4).Infof("rest_v2 UnlinkFileset. filesystem: %s, fileset: %s", filesystemName, filesetName)

	unlinkFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link?force=True", filesystemName, filesetName)
	unlinkFilesetResponse := api.GenericResponse{}

	err := s.doHTTP(DELETE, unlinkFilesetURL, nil, &unlinkFilesetResponse)

	if err != nil {
		glog.Errorf("Error in unlink fileset request: %v", err)
		return err
	}

	err = isRequestAccepted(unlinkFilesetResponse, unlinkFilesetURL)
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

func (s *Connector) ListFileset(filesystemName string, filesetName string) (api.Fileset_v2, error) {
	glog.V(4).Infof("rest_v2 ListFileset. filesystem: %s, fileset: %s", filesystemName, filesetName)

	getFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	getFilesetResponse := api.GetFilesetResponse_v2{}

	err := s.doHTTP(GET, getFilesetURL, nil, &getFilesetResponse)
	if err != nil {
		glog.Errorf("Error in list fileset request: %v", err)
		return api.Fileset_v2{}, err
	}

	if len(getFilesetResponse.Filesets) == 0 {
		glog.Errorf("No fileset returned for %s", filesetName)
		return api.Fileset_v2{}, fmt.Errorf("No fileset returned for %s", filesetName)
	}

	return getFilesetResponse.Filesets[0], nil
}

func (s *Connector) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
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

func (s *Connector) MakeDirectory(filesystemName string, relativePath string, uid string, gid string) error {
	glog.V(4).Infof("rest_v2 MakeDirectory. filesystem: %s, path: %s, uid: %s, gid: %s", filesystemName, relativePath, uid, gid)

	dirreq := api.CreateMakeDirRequest{}

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

	makeDirResponse := api.GenericResponse{}

	err := s.doHTTP(POST, makeDirURL, dirreq, &makeDirResponse)

	if err != nil {
		glog.Errorf("Error in make directory request: %v", err)
		return err
	}

	err = isRequestAccepted(makeDirResponse, makeDirURL)
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

func (s *Connector) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	glog.V(4).Infof("rest_v2 SetFilesetQuota. filesystem: %s, fileset: %s, quota: %s", filesystemName, filesetName, quota)

	setQuotaURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName)
	quotaRequest := api.SetQuotaRequest_v2{}

	quotaRequest.BlockHardLimit = quota
	quotaRequest.BlockSoftLimit = quota
	quotaRequest.OperationType = "setQuota"
	quotaRequest.QuotaType = "fileset"
	quotaRequest.ObjectName = filesetName

	setQuotaResponse := api.GenericResponse{}

	err := s.doHTTP(POST, setQuotaURL, quotaRequest, &setQuotaResponse)
	if err != nil {
		glog.Errorf("Error in set fileset quota request: %v", err)
		return err
	}

	err = isRequestAccepted(setQuotaResponse, setQuotaURL)
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

func (s *Connector) CheckIfFSQuotaEnabled(filesystemName string) error {
	glog.V(4).Infof("rest_v2 CheckIfFSQuotaEnabled. filesystem: %s", filesystemName)

	checkQuotaURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName)
	QuotaResponse := api.GetQuotaResponse_v2{}

	err := s.doHTTP(GET, checkQuotaURL, nil, &QuotaResponse)
	if err != nil {
		glog.Errorf("Error in check quota: %v", err)
		return err
	}
	return nil
}

func (s *Connector) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	glog.V(4).Infof("rest_v2 ListFilesetQuota. filesystem: %s, fileset: %s", filesystemName, filesetName)

	listQuotaURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas?filter=objectName=%s", filesystemName, filesetName)
	listQuotaResponse := api.GetQuotaResponse_v2{}

	err := s.doHTTP(GET, listQuotaURL, nil, &listQuotaResponse)
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

func (s *Connector) MountFilesystem(filesystemName string, nodeName string) error { //nolint:dupl
	glog.V(4).Infof("rest_v2 MountFilesystem. filesystem: %s, node: %s", filesystemName, nodeName)

	mountreq := api.MountFilesystemRequest{}
	mountreq.Nodes = append(mountreq.Nodes, nodeName)

	mountFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/mount", filesystemName)
	mountFilesystemResponse := api.GenericResponse{}

	err := s.doHTTP(PUT, mountFilesystemURL, mountreq, &mountFilesystemResponse)
	if err != nil {
		glog.Errorf("Error in mount filesystem request: %v", err)
		return err
	}

	err = isRequestAccepted(mountFilesystemResponse, mountFilesystemURL)
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

func (s *Connector) UnmountFilesystem(filesystemName string, nodeName string) error { //nolint:dupl
	glog.V(4).Infof("rest_v2 UnmountFilesystem. filesystem: %s, node: %s", filesystemName, nodeName)

	unmountreq := api.UnmountFilesystemRequest{}
	unmountreq.Nodes = append(unmountreq.Nodes, nodeName)

	unmountFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/unmount", filesystemName)
	unmountFilesystemResponse := api.GenericResponse{}

	err := s.doHTTP(PUT, unmountFilesystemURL, unmountreq, &unmountFilesystemResponse)
	if err != nil {
		glog.Errorf("Error in unmount filesystem request: %v", err)
		return err
	}

	err = isRequestAccepted(unmountFilesystemResponse, unmountFilesystemURL)
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

func (s *Connector) GetFilesystemName(filesystemUUID string) (string, error) {
	glog.V(4).Infof("rest_v2 GetFilesystemName. UUID: %s", filesystemUUID)

	getFilesystemNameURL := fmt.Sprintf("scalemgmt/v2/filesystems?filter=uuid=%s", filesystemUUID)
	getFilesystemNameURLResponse := api.GetFilesystemResponse_v2{}

	err := s.doHTTP(GET, getFilesystemNameURL, nil, &getFilesystemNameURLResponse)
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

func (s *Connector) GetFsUid(filesystemName string) (string, error) {
	getFilesystemURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName)
	getFilesystemResponse := api.GetFilesystemResponse_v2{}

	err := s.doHTTP(GET, getFilesystemURL, nil, &getFilesystemResponse)
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

func (s *Connector) DeleteSymLnk(filesystemName string, LnkName string) error {
	LnkName = strings.ReplaceAll(LnkName, "/", "%2F")
	deleteLnkURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/symlink/%s", filesystemName, LnkName)
	deleteLnkResponse := api.GenericResponse{}

	err := s.doHTTP(DELETE, deleteLnkURL, nil, &deleteLnkResponse)
	if err != nil {
		return fmt.Errorf("Unable to delete Symlink %v.", LnkName)
	}

	err = isRequestAccepted(deleteLnkResponse, deleteLnkURL)
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

func (s *Connector) DeleteDirectory(filesystemName string, dirName string) error {
	NdirName := strings.ReplaceAll(dirName, "/", "%2F")
	deleteDirURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/directory/%s", filesystemName, NdirName)
	deleteDirResponse := api.GenericResponse{}

	err := s.doHTTP(DELETE, deleteDirURL, nil, &deleteDirResponse)
	if err != nil {
		return fmt.Errorf("Unable to delete dir %v.", dirName)
	}

	err = isRequestAccepted(deleteDirResponse, deleteDirURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(deleteDirResponse.Status.Code, deleteDirResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to delete dir %v:%v", dirName, err)
	}

	return nil
}

func (s *Connector) GetFileSetUid(filesystemName string, filesetName string) (string, error) {
	getFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName)
	getFilesetResponse := api.GetFilesetResponse_v2{}

	err := s.doHTTP(GET, getFilesetURL, nil, &getFilesetResponse)
	if err != nil {
		return "", fmt.Errorf("Unable to list fileset %v.", filesetName)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return "", fmt.Errorf("Unable to list fileset %v.", filesetName)
	}

	return fmt.Sprintf("%d", getFilesetResponse.Filesets[0].Config.Id), nil
}

func (s *Connector) GetFileSetNameFromId(filesystemName string, Id string) (string, error) {
	getFilesetURL := fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets?filter=config.id=%s", filesystemName, Id)

	getFilesetResponse := api.GetFilesetResponse_v2{}

	err := s.doHTTP(GET, getFilesetURL, nil, &getFilesetResponse)
	if err != nil {
		return "", fmt.Errorf("Unable to get name for fileset Id %v:%v.", filesystemName, Id)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return "", nil
	}

	return getFilesetResponse.Filesets[0].FilesetName, nil
}

func (s *Connector) CheckIfFileDirPresent(filesystemName string, relPath string) (bool, error) {
	RelPath := strings.ReplaceAll(relPath, "/", "%2F")
	checkFilDirUrl := fmt.Sprintf("scalemgmt/v2/filesystems/%s/owner/%s", filesystemName, RelPath)
	ownerResp := api.OwnerResp_v2{}
	err := s.doHTTP(GET, checkFilDirUrl, nil, &ownerResp)
	if err != nil {
		if strings.Contains(ownerResp.Status.Message, "File not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Connector) CreateSymLink(SlnkfilesystemName string, TargetFs string, relativePath string, LnkPath string) error {
	symLnkReq := api.SymLnkRequest{}
	symLnkReq.FilesystemName = TargetFs
	symLnkReq.RelativePath = relativePath

	LnkPath = strings.ReplaceAll(LnkPath, "/", "%2F")

	symLnkUrl := fmt.Sprintf("scalemgmt/v2/filesystems/%s/symlink/%s", SlnkfilesystemName, LnkPath)

	makeSlnkResp := api.GenericResponse{}

	err := s.doHTTP(POST, symLnkUrl, symLnkReq, &makeSlnkResp)

	if err != nil {
		return err
	}

	err = isRequestAccepted(makeSlnkResp, symLnkUrl)
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
