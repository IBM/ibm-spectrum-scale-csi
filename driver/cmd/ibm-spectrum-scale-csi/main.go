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
	"os"
	"path"
	"strings"

	driver "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"github.com/natefinch/lumberjack"
	"k8s.io/klog/v2"
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
	vendorVersion  = "2.10.0"
)

const dirPath = "scalecsilogs"
const logFile = "ibm-spectrum-scale-csi.logs"
const logLevel = "LOGLEVEL"
const persistentLog = "PERSISTENT_LOG"
const hostPath = "/host/var/adm/ras/"
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
	level, persistentLogEnabled := getLogEnv()
	logValue, isIncorrectLogLevel := getLogLevel(level)
	value := getVerboseLevel(level)
	err1 := flag.Set("stderrthreshold", logValue)
	err2 := flag.Set("v", value)

	defer func() {
		if r := recover(); r != nil {
			klog.Infof("Recovered from panic: [%v]", r)
		}
	}()
	if persistentLogEnabled == "ENABLED" {
		errf := flag.Set("logtostderr", "false")
		flag.Parse()
		if errf != nil {
			klog.Errorf("Failed to set logtostderr value to false")
		}
		fpClose := InitFileLogger()
		defer fpClose()
	} else {
		errf := flag.Set("logtostderr", "true")
		flag.Parse()
		if errf != nil {
			klog.Errorf("Failed to set logtostderr value to true")
		}
	}

	ctx := setContext()
	loggerId := utils.GetLoggerId(ctx)
	if err1 != nil || err2 != nil {
		klog.Errorf("[%s] Failed to set flag value", loggerId)
	}

	if isIncorrectLogLevel {
		klog.Infof("[%s] logger level is empty or incorrect. Defaulting logValue to INFO", loggerId)
	} else {
		klog.Infof("[%s] logValue: %s", loggerId, level)
	}
	klog.V(0).Infof("[%s] Version Info: commit (%s)", loggerId, gitCommit)

	// PluginFolder defines the location of scaleplugin
	PluginFolder := path.Join(*kubeletRootDir, "plugins/spectrumscale.csi.ibm.com")

	if err := createPersistentStorage(path.Join(PluginFolder, "controller")); err != nil {
		klog.Errorf("[%s] failed to create persistent storage for controller %v", loggerId, err)
		os.Exit(1)
	}
	if err := createPersistentStorage(path.Join(PluginFolder, "node")); err != nil {
		klog.Errorf("[%s] failed to create persistent storage for node %v", loggerId, err)
		os.Exit(1)
	}

	defer klog.Flush()
	handle(ctx)
	os.Exit(0)
}

func handle(ctx context.Context) {
	driver := driver.GetScaleDriver(ctx)
	err := driver.SetupScaleDriver(ctx, *driverName, vendorVersion, *nodeID)
	if err != nil {
		klog.Fatalf("[%s] Failed to initialize Scale CSI Driver: %v", utils.GetLoggerId(ctx), err)
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

func setContext() context.Context {
	newCtx := context.Background()
	ctx := utils.SetLoggerId(newCtx)
	return ctx
}

func getLogEnv() (string, string) {
	level := os.Getenv(logLevel)
	persistentLogEnabled := os.Getenv(persistentLog)
	if strings.ToUpper(persistentLogEnabled) != "ENABLED" {
		persistentLogEnabled = "DISABLED"
	}
	return strings.ToUpper(level), strings.ToUpper(persistentLogEnabled)
}

func getLogLevel(level string) (string, bool) {
	var logValue string
	isIncorrectLogLevel := false

	if !(level == TRACE.String() || level == DEBUG.String() || level == INFO.String() || level == WARNING.String() || level == ERROR.String() || level == FATAL.String()) {
		isIncorrectLogLevel = true
	}
	if level == DEBUG.String() || level == TRACE.String() || isIncorrectLogLevel {
		logValue = INFO.String()
	} else {
		logValue = level
	}
	return logValue, isIncorrectLogLevel
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
		/* #nosec G301 -- false positive */
		err := os.MkdirAll(fileDir, 0755)
		if err != nil {
			panic(fmt.Sprintf("failed to create log folder %v", err))
		}
	}

	/* #nosec G302 -- false positive */
	logFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		panic(fmt.Sprintf("failed to init logger %v", err))
	}

	l := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    rotateSize,
		MaxBackups: 5,
		MaxAge:     0,
		Compress:   true,
	}
	klog.SetOutput(l)

	closeFn := func() {
		err := logFile.Close()
		if err != nil {
			panic(fmt.Sprintf("failed to close log file %v", err))
		}
	}
	return closeFn
}
