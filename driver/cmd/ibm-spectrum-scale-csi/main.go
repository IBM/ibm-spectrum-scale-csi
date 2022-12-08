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
	"flag"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"math/rand"
	"os"
	"path"
	"time"

	driver "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
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
	vendorVersion  = "2.8.0"
)

var logger *utils.CsiLogger

func main() {
	utils.InitLogger()

	logger.Infof("Version Info: commit (%s)", gitCommit)

	rand.Seed(time.Now().UnixNano())

	// PluginFolder defines the location of scaleplugin
	PluginFolder := path.Join(*kubeletRootDir, "plugins/spectrumscale.csi.ibm.com")
	OldPluginFolder := path.Join(*kubeletRootDir, "plugins/ibm-spectrum-scale-csi")

	if err := createPersistentStorage(path.Join(PluginFolder, "controller")); err != nil {
		logger.Errorf("failed to create persistent storage for controller %v", err)
		os.Exit(1)
	}
	if err := createPersistentStorage(path.Join(PluginFolder, "node")); err != nil {
		logger.Errorf("failed to create persistent storage for node %v", err)
		os.Exit(1)
	}

	if err := deleteStalePluginDir(OldPluginFolder); err != nil {
		logger.Errorf("failed to delete stale plugin folder %v, please delete manually. %v", OldPluginFolder, err)
	}

	handle()
	logger.Flush()
	os.Exit(0)
}

func handle() {
	driver := driver.GetScaleDriver()
	err := driver.SetupScaleDriver(*driverName, vendorVersion, *nodeID)
	if err != nil {
		logger.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
	}
	driver.Run(*endpoint)
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
