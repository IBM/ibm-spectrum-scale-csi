/**
 * Copyright 2016, 2019 IBM Corp.
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

func ReadAndUnmarshal(object interface{}, dir string, fileName string) error {
	glog.V(6).Infof("utils ReadAndUnmarshal. object: %v, dir: %s, fileName: %s", object, dir, fileName)

	path := dir + string(os.PathSeparator) + fileName

	bytes, err := ReadFile(path)
	if err != nil {
		glog.Errorf("Error in reading file %s: %v", path, err)
		return err
	}

	err = json.Unmarshal(bytes, object)
	if err != nil {
		glog.Errorf("Error in unmarshalling file %s: %v", path, err)
		return err
	}

	return nil
}

func MarshalAndRecord(object interface{}, dir string, fileName string) error {
	glog.V(6).Infof("utils MarshalAndRecord. object: %v, dir: %s, fileName: %s", object, dir, fileName)

	_ = MkDir(dir)
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := json.MarshalIndent(object, "", " ")
	if err != nil {
		glog.Errorf("Error in MarshalIndent %v: %v", object, err)
		return err
	}

	return WriteFile(path, bytes)
}

func ReadFile(path string) ([]byte, error) {
	glog.V(6).Infof("utils ReadFile. path: %s", path)

	file, err := os.Open(path)
	if err != nil {
		glog.Errorf("Error in opening file %s: %v", path, err)
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		glog.Errorf("Error in read file %s: %v", path, err)
		return nil, err
	}

	return bytes, nil
}

func WriteFile(path string, content []byte) error {
	glog.V(6).Infof("utils WriteFile. path: %s", path)

	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		glog.Errorf("Error in write file %s: %v", path, err)
		return err
	}

	return nil
}

func GetPath(paths []string) string {
	glog.V(6).Infof("utils GetPath. paths: %v", paths)

	workDirectory, _ := os.Getwd()

	if len(paths) == 0 {
		return workDirectory
	}

	resultPath := workDirectory

	for _, path := range paths {
		resultPath += string(os.PathSeparator)
		resultPath += path
	}

	return resultPath
}

func Exists(path string) bool {
	glog.V(6).Infof("utils Exists. path: %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func MkDir(path string) error {
	glog.V(6).Infof("utils MkDir. path: %s", path)
	var err error
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			glog.Errorf("Error in creating dir %s: %v", path, err)
			return err
		}
	}

	return err
}

func StringInSlice(a string, list []string) bool {
	glog.V(6).Infof("utils StringInSlice. string: %s, slice: %v", a, list)
	for _, b := range list {
		if strings.EqualFold(b, a) {
			return true
		}
	}
	return false
}

func ConvertToBytes(inputStr string) (uint64, error) {
	glog.V(6).Infof("utils ConvertToBytes. string: %s", inputStr)
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
		return 0, fmt.Errorf("Invalid number specified %v", inputStr)
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
		return 0, fmt.Errorf("Invalid Unit %v supplied with %v", unit, inputStr)
	}

	if retValue > uintMax64 {
		return 0, fmt.Errorf("Overflow detected %v", inputStr)
	}

	return retValue, nil
}

func GetEnv(envName string, defaultValue string) string {
	glog.V(6).Infof("utils GetEnv. envName: %s", envName)
	envValue := os.Getenv(envName)
	if envValue == "" {
		envValue = defaultValue
	}
	return envValue
}
