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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

const loggerId = "logger_id"

func ReadFile(path string) ([]byte, error) {
	glog.V(6).Infof("utils ReadFile. path: %s", path)

	file, err := os.Open(path) // #nosec G304 This is valid path gererated internally. it is False positive
	if err != nil {
		glog.Errorf("Error in opening file %s: %v", path, err)
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			glog.Errorf("Error in closing file %s: %v", path, err)
		}
	}()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		glog.Errorf("Error in read file %s: %v", path, err)
		return nil, err
	}

	return bytes, nil
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

func Exists(ctx context.Context, path string) bool {
	logger.Trace(ctx, "utils Exists. path: %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func MkDir(ctx context.Context, path string) error {
	logger.Trace(ctx, "utils MkDir. path: %s", path)
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

func FsStatInfo(path string) (int64, int64, int64, int64, int64, int64, error) {
	statfs := &unix.Statfs_t{}
	err := unix.Statfs(path, statfs)

	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	available := int64(statfs.Bavail) * int64(statfs.Bsize)
	capacity := int64(statfs.Blocks) * int64(statfs.Bsize)
	usage := (int64(statfs.Blocks) - int64(statfs.Bfree)) * int64(statfs.Bsize)
	inodes := int64(statfs.Files)
	inodesFree := int64(statfs.Ffree)
	inodesUsed := inodes - inodesFree

	return available, capacity, usage, inodes, inodesFree, inodesUsed, nil
}

func SetLoggerId(ctx context.Context) context.Context {
	id := uuid.New().String()
	return context.WithValue(ctx, loggerId, id)
}

func GetLoggerId(ctx context.Context) string {
	logger, _ := ctx.Value(loggerId).(string)
	return logger
}

func GetExecutionTime() int64 {
	t := time.Now()
	timeinMilliSec := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
	return timeinMilliSec
}
