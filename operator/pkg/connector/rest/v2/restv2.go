/**
Copyright 2022.

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

package rest

import (
	"fmt"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var csiLog = log.Log.WithName("csirest_connector")

func InitiateClusterAssessment() error {
	logger := csiLog.WithName("InitiateClusterAssessment")
	scaleConfig := settings.LoadScaleConfigSettings()

	for i := 0; i < len(scaleConfig.Clusters); i++ {
		cluster := scaleConfig.Clusters[i]

		sc, err := connectors.GetSpectrumScaleConnector(cluster)
		if err != nil {
			logger.Error(err, "Unable to initialize Spectrum Scale connector for cluster", "cluster", cluster.ID)
			return err
		}

		//clusterId validation
		clusterId, err := sc.GetClusterId()
		if err != nil {
			logger.Error(err, "Unable to get cluster ID from the cluster")
			return err
		}
		if cluster.ID != clusterId {
			logger.Error(err, "Cluster ID %s from scale config doesn't match the ID %s from cluster", cluster.ID, clusterId)
			return fmt.Errorf("Cluster ID doesn't match the cluster")
		}
		logger.Info("ClusterId from the cluster :=== ", clusterId, "clusterId from scale config :=== ", cluster.ID)

		//filesystem mount validation on GUI
		isMounted, err := sc.IsFilesystemMountedOnGUINode(cluster.Primary.GetPrimaryFs())
		if err != nil {
			logger.Error(err, "Unable to get filesystem mount details for %s on Primary cluster", cluster.Primary.GetPrimaryFs())
			return err
		}
		if !isMounted {
			logger.Error(err, "Primary filesystem %s is not mounted on GUI node of Primary cluster", cluster.Primary.GetPrimaryFs())
			return fmt.Errorf("Primary filesystem %s not mounted on GUI node Primary cluster", cluster.Primary.GetPrimaryFs())
		}
	}

	return nil
}
