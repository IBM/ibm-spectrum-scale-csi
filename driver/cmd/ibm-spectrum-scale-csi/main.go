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
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
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
	vendorVersion  = "2.11.0"
)

func main() {
	klog.InitFlags(nil)
	if val, ok := os.LookupEnv(settings.LogLevel); ok {
		klog.Infof("[%s] found in the env : %s", settings.LogLevel, val)
	}
	if val, ok := os.LookupEnv(settings.PersistentLog); ok {
		klog.Infof("[%s] found in the env : %s", settings.PersistentLog, val)
	}
	if val, ok := os.LookupEnv(settings.NodePublishMethod); ok {
		klog.Infof("[%s] found in the env : %s", settings.NodePublishMethod, val)
	}
	if val, ok := os.LookupEnv(settings.VolumeStatsCapability); ok {
		klog.Infof("[%s] found in the env : %s", settings.VolumeStatsCapability, val)
	}
	level, persistentLogEnabled := getLogEnv()
	logValue := getLogLevel(level)
	value := getVerboseLevel(level)
	err := flag.Set("logtostderr", "false")
	err1 := flag.Set("stderrthreshold", logValue)
	err2 := flag.Set("v", value)
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			klog.Infof("Recovered from panic: [%v]", r)
		}
	}()
	if persistentLogEnabled == "ENABLED" {
		fpClose := InitFileLogger()
		defer fpClose()
	}

	ctx := setContext()
	loggerId := utils.GetLoggerId(ctx)
	if err != nil || err1 != nil || err2 != nil {
		klog.Errorf("[%s] Failed to set flag value", loggerId)
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
	loggerId := utils.GetLoggerId(ctx)
	driver := driver.GetScaleDriver(ctx)
	err := driver.SetupScaleDriver(ctx, *driverName, vendorVersion, *nodeID)
	if err != nil {
		klog.Fatalf("[%s] Failed to initialize Scale CSI Driver: %v", loggerId, err)
	}
	klog.Infof("[%s] Scale CSI Driver Initialized: [%+v]", loggerId, driver)
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
	level := os.Getenv(settings.LogLevel)
	persistentLogEnabled := os.Getenv(settings.PersistentLog)
	return strings.ToUpper(level), strings.ToUpper(persistentLogEnabled)
}

func getLogLevel(level string) string {
	var logValue string
	if level == settings.DEBUG.String() || level == settings.TRACE.String() {
		logValue = settings.INFO.String()
	} else {
		logValue = level
	}
	return logValue
}

func getVerboseLevel(level string) string {
	if level == settings.DEBUG.String() {
		return "4"
	} else if level == settings.TRACE.String() {
		return "6"
	} else {
		return "1"
	}
}

func InitFileLogger() func() {
	filePath := settings.HostPath + settings.DirPath + "/" + settings.LogFile
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
		MaxSize:    settings.RotateSize,
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
