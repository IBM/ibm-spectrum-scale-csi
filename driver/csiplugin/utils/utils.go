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
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"golang.org/x/sys/unix"
)

const (
        // Stat type
        RESTAPI_STATS = 0
        CSIOP_STATS   = 1
)

type OpStats struct {
        statName string
        statTotalCount int64
        statErrCount   int64
	statTotalTime  float64
        avgTime        float64
        minTime        float64
        maxTime        float64
	statMtx        sync.Mutex
}

var statMap sync.Map

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

func SumStats(opType int, opName string, opTotalTime int64, err error) {

	statsType := "RESTAPIStats"
        if opType != RESTAPI_STATS {
                statsType = "CSIOPStats"
        }

	success := true
	if err != nil {
		success = false
	}

//	glog.V(4).Infof("%v START SAVING STATS statMap: %p opType: %v opName: %v opTotalTime: %v success: %v", statsType, &statMap, opType, opName, opTotalTime, success)
	totalTime := float64(opTotalTime)
	opDetails, found := statMap.Load(opName)
	if found {
//		glog.V(4).Infof("%v found stats already for time: %v, opName: %v", statsType, totalTime, opName)
		StoreStats(statsType, opName, opDetails.(OpStats), totalTime, success)
	} else {
		newStats := OpStats{}
		newStats.statName = opName
		opDetails, loaded := statMap.LoadOrStore(opName, newStats)
//		glog.V(4).Infof("%v didn't find stats for time: %v opName: %v", statsType, totalTime, opName)
		if loaded {
//			glog.V(4).Infof("%v didn't find stats earlier but now found stats for time: %v opName: %v", statsType, totalTime, opName)
                } else {
//			glog.V(4).Infof("%v Saved stats for first time for statsMap: %v time:%v opName:%v", statsType, statMap, totalTime, opName)
			opDetails, _ = statMap.Load(opName)
                }
		StoreStats(statsType, opName, opDetails.(OpStats), totalTime, success)
	}
}

func StoreStats(statsType string, opName string, oopDetails OpStats, totalTime float64, success bool) {
//	glog.V(4).Infof("%v waiting for lock for opName: %v", statsType, opName)
	oopDetails.statMtx.Lock()
//	glog.V(4).Infof("%v got lock for opName: %v", statsType, opName)
	opDetails, _ := statMap.Load(opName)
		newStats := OpStats{}
		newStats.statName = opDetails.(OpStats).statName
                newStats.statTotalCount = opDetails.(OpStats).statTotalCount + 1
                newStats.statErrCount = opDetails.(OpStats).statErrCount
                if !success {
                        newStats.statErrCount = opDetails.(OpStats).statErrCount + 1
                }
                newStats.statTotalTime = opDetails.(OpStats).statTotalTime + totalTime
                newStats.avgTime = newStats.statTotalTime / float64(newStats.statTotalCount)
                newStats.minTime = opDetails.(OpStats).minTime
                newStats.maxTime = opDetails.(OpStats).maxTime
                if opDetails.(OpStats).minTime > totalTime || opDetails.(OpStats).minTime == 0 {
                        newStats.minTime = totalTime
                }
                if opDetails.(OpStats).maxTime < totalTime {
                        newStats.maxTime = totalTime
                }
                statMap.Store(opName, newStats)
//		glog.V(4).Infof("%v Saved stats for time:%v opName:%v", statsType, totalTime, opName)
	oopDetails.statMtx.Unlock()
//	glog.V(4).Infof("%v released lock for time:%v opName:%v", statsType, totalTime, opName)
glog.V(4).Infof("STATS: type: %v opName: %v totalCount: %v errorCount: %v avgTime: %v minTime: %v maxTime: %v", statsType, newStats.statName, newStats.statTotalCount, newStats.statErrCount, newStats.avgTime, newStats.minTime, newStats.maxTime)
}
