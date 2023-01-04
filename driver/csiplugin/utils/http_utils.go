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

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
)

var logger *CsiLogger

/*
	func ExtractErrorResponse(response *http.Response) error {
		errorResponse := connectors.GenericResponse{}
		err := UnmarshalResponse(response, &errorResponse)
		if err != nil {
			return fmt.Errorf("json.Unmarshal failed %v", err)
		}
		return fmt.Errorf(errorResponse.Err)
	}
*/

func UnmarshalResponse(ctx context.Context, r *http.Response, object interface{}) error {
	logger.Trace(ctx, "http_utils UnmarshalResponse. response: %v", r.Body)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll failed %v", err)
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return fmt.Errorf("json.Unmarshal failed %v", err)
	}

	return nil
}

func HttpExecuteUserAuth(ctx context.Context, httpClient *http.Client, requestType string, requestURL string, user string, password string, rawPayload interface{}) (*http.Response, error) {
	glog.V(4).Infof("[%s] http_utils HttpExecuteUserAuth. type: %s, url: %s, user: %s", GetLoggerId(ctx), requestType, requestURL, user)
	logger.Trace(ctx, "http_utils HttpExecuteUserAuth. request payload: %v", rawPayload)

	payload, err := json.MarshalIndent(rawPayload, "", " ")
	if err != nil {
		err = fmt.Errorf("Internal error marshaling params. url: %s: %#v", requestURL, err)
		return nil, fmt.Errorf("failed %v", err)
	}

	if user == "" {
		return nil, fmt.Errorf("Empty UserName passed")
	}

	request, err := http.NewRequest(requestType, requestURL, bytes.NewBuffer(payload))
	if err != nil {
		err = fmt.Errorf("Error in creating request. url: %s: %#v", requestURL, err)
		return nil, fmt.Errorf("failed %v", err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	request.SetBasicAuth(user, password)
	logger.Trace(ctx, "http_utils HttpExecuteUserAuth request: %+v", request)

	return httpClient.Do(request)
}

func WriteResponse(w http.ResponseWriter, code int, object interface{}) {
	glog.V(4).Infof("http_utils WriteResponse. code: %d, object: %v", code, object)

	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	fmt.Fprint(w, string(data))
}

func Unmarshal(r *http.Request, object interface{}) error {
	glog.V(6).Infof("http_utils Unmarshal. request: %v, object: %v", r, object)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalDataFromRequest(r *http.Request, object interface{}) error {
	glog.V(6).Infof("http_utils UnmarshalDataFromRequest. request: %v, object: %v", r, object)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}
