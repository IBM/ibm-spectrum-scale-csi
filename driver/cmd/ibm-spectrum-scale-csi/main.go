/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	driver "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/natefinch/lumberjack"
	"k8s.io/klog/v2"
	"math/rand"
	"os"
	"path"
	"time"
)

// gitCommit that is injected via go build -ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
var (
	gitCommit string
)

var (
	endpoint       = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName     = flag.String("drivername", "spectrumscale.csi.ibm.com", "name of the driver")
	nodeID         = flag.String("nodeid", "", "node id")
	kubeletRootDir = flag.String("kubeletRootDirPath", "/var/lib/kubelet", "kubelet root directory path")
	vendorVersion  = "2.9.0"
)

const dirPath = "scalecsilogs"
const logFile = "ibm-spectrum-scale-csi.logs"
const logLevel = "LOGLEVEL"
const hostPath = "/host/var/log/"
const rotateSize = 1024

type LoggerLevel int

const (
	TRACE LoggerLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

func main() {
	klog.InitFlags(nil)
	level := getLevel()
	value := getVerboseLevel(level)
	flag.Set("logtostderr", "false")
	flag.Set("stderrthreshold", level)
	flag.Set("v", value)
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			klog.Infof("Recovered from panic: [%v]", r)
		}
	}()
	fpClose := InitFileLogger()
	defer fpClose()

	ctx := setContext()
	loggerId := utils.GetLoggerId(ctx)
	klog.V(0).Infof("[%s] Version Info: commit (%s)", loggerId, gitCommit)

	rand.Seed(time.Now().UnixNano())

	// PluginFolder defines the location of scaleplugin
	PluginFolder := path.Join(*kubeletRootDir, "plugins/spectrumscale.csi.ibm.com")
	OldPluginFolder := path.Join(*kubeletRootDir, "plugins/ibm-spectrum-scale-csi")

	if err := createPersistentStorage(path.Join(PluginFolder, "controller")); err != nil {
		klog.Errorf("[%s] failed to create persistent storage for controller %v", loggerId, err)
		os.Exit(1)
	}
	if err := createPersistentStorage(path.Join(PluginFolder, "node")); err != nil {
		klog.Errorf("[%s] failed to create persistent storage for node %v", loggerId, err)
		os.Exit(1)
	}

	if err := deleteStalePluginDir(OldPluginFolder); err != nil {
		klog.Errorf("[%s] failed to delete stale plugin folder %v, please delete manually. %v", loggerId, OldPluginFolder, err)
	}
	defer klog.Flush()
	handle(ctx)
	os.Exit(0)
}

func handle(ctx context.Context) {
	driver := driver.GetScaleDriver(ctx)
	err := driver.SetupScaleDriver(ctx, *driverName, vendorVersion, *nodeID)
	if err != nil {
		klog.V(0).Infof("[%s] Failed to initialize Scale CSI Driver: %v", utils.GetLoggerId(ctx), err)
	}
	driver.Run(ctx, *endpoint)
}

func createPersistentStorage(persistentStoragePath string) error {
	if _, err := os.Stat(persistentStoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(persistentStoragePath, os.FileMode(0644)); err != nil {
			return err
		}
	}
	return nil
}

func deleteStalePluginDir(stalePluginPath string) error {
	return os.RemoveAll(stalePluginPath)
}

func setContext() context.Context {
	newCtx := context.Background()
	ctx := utils.SetLoggerId(newCtx)
	return ctx
}

func getLevel() string {
	level := os.Getenv(logLevel)
	var logValue string
	klog.Infof("logValue: %s", level)
	if level == "" || level == DEBUG.String() || level == TRACE.String() {
		logValue = INFO.String()
	} else {
		logValue = level
	}
	return logValue
}

func (level LoggerLevel) String() string {
	switch level {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case INFO:
		return "INFO"
	default:
		return "INFO"
	}
}

func getVerboseLevel(level string) string {
	if level == DEBUG.String() {
		return "4"
	} else if level == TRACE.String() {
		return "6"
	} else {
		return "1"
	}
}

func InitFileLogger() func() {
	filePath := hostPath + dirPath + "/" + logFile
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fileDir, _ := path.Split(filePath)
		err := os.MkdirAll(fileDir, 0766)
		if err != nil {
			panic(fmt.Sprintf("failed to create log folder %v", err))
		}
	}

	logFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		panic(fmt.Sprintf("failed to init logger %v", err))
	}

	fileStat, err := logFile.Stat()
	if err != nil {
		panic(fmt.Sprintf("failed to stat logger file %v", err))
	}

	fileStatSize := int(fileStat.Size()) / 1024 / 1024

	// If log file size bigger than rotateSize, will use lunberjack to run the logrotate
	if fileStatSize < rotateSize {
		klog.SetOutput(logFile)
	} else {
		klog.SetOutput(&lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    rotateSize,
			MaxBackups: 5,
			MaxAge:     0,
			Compress:   true,
		})
	}

	return func() { logFile.Close() }
}
