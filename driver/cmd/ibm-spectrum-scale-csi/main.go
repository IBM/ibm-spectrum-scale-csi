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
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/settings"
	"k8s.io/klog"
)

var (
	driverName    = flag.String("drivername", "spectrumscale.csi.ibm.com", "name of the driver")
	endpoint      = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID        = flag.String("nodeid", "", "node id")
	vendorVersion = "1.1.0"
	version       = flag.Bool("version", false, "Print the version and exit.")
)

func main() {
	_ = flag.Set("logtostderr", "true")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	if *version {
		fmt.Printf("%s %s\n", *driverName, vendorVersion)
		os.Exit(0)
	}

	configMap, err := settings.LoadScaleConfig()
	if err != nil {
		klog.Fatalf(`Could not load CSI for Scale ConfigMap: %v`, err)
	}

	driver := scale.NewDriver(*driverName, vendorVersion, *nodeID, configMap)

	err = driver.PluginInitialize(configMap)
	if err != nil {
		klog.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
	}

	err = driver.Run(*endpoint)
	if err != nil {
		klog.Fatalf("Failed running Scale CSI Driver: %v", err)
	}
}
