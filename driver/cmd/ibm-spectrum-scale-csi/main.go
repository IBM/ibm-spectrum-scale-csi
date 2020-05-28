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
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/golang/glog"

	driver "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
)

var (
	endpoint      = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName    = flag.String("drivername", "spectrumscale.csi.ibm.com", "name of the driver")
	nodeID        = flag.String("nodeid", "", "node id")
	vendorVersion = "2.0.0"
)

func main() {
	_ = flag.Set("logtostderr", "true")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	if err := createPersistentStorage(path.Join(driver.PluginFolder, "controller")); err != nil {
		glog.Errorf("failed to create persistent storage for controller %v", err)
		os.Exit(1)
	}
	if err := createPersistentStorage(path.Join(driver.PluginFolder, "node")); err != nil {
		glog.Errorf("failed to create persistent storage for node %v", err)
		os.Exit(1)
	}

	if err := deleteStalePluginDir(driver.OldPluginFolder); err != nil {
		glog.Errorf("failed to delete stale plugin folder %v, please delete manually. %v", driver.OldPluginFolder, err)
	}

	handle()
	os.Exit(0)
}

func handle() {
	driver := driver.GetScaleDriver()
	err := driver.SetupScaleDriver(*driverName, vendorVersion, *nodeID)
	if err != nil {
		glog.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
	}
	driver.Run(*endpoint)
}

func createPersistentStorage(persistentStoragePath string) error {
	if _, err := os.Stat(persistentStoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(persistentStoragePath, os.FileMode(0755)); err != nil {
			return err
		}
	}
	return nil
}

func deleteStalePluginDir(stalePluginPath string) error {
	return os.RemoveAll(stalePluginPath)
}
