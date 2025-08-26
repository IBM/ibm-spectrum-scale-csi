/**
 * Copyright 2019,2024 IBM Corp.
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

package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"k8s.io/klog/v2"
)

type ScaleSettingsConfigMap struct {
	LocalScaleCluster string `json:"localScaleCluster"`
	Clusters          []Clusters
}

type Primary struct {
	PrimaryFSDep  string `json:"primaryFS"`   // Deprecated
	PrimaryFs     string `json:"primaryFs"`   //Deprecated
	PrimaryFset   string `json:"primaryFset"` //Deprecated
	PrimaryCid    string `json:"primaryCid"`
	InodeLimitDep string `json:"inode-limit"`   // Deprecated
	InodeLimits   string `json:"inodeLimit"`    //Deprecated
	RemoteCluster string `json:"remoteCluster"` //Deprecated

	PrimaryFSMount      string
	PrimaryFsetLink     string
	SymlinkAbsolutePath string
	SymlinkRelativePath string
}

const (
	secretFileSuffix = "-secret" // #nosec G101 false positive
	cacertFileSuffix = "-cacert"
)
const (
	DirPath               = "scalecsilogs"
	LogFile               = "ibm-spectrum-scale-csi.logs"
	PersistentLog         = "PERSISTENT_LOG"
	NodePublishMethod     = "NODEPUBLISH_METHOD"
	VolumeStatsCapability = "VOLUME_STATS_CAPABILITY"
	HostPath              = "/host/var/adm/ras/"
	RotateSize            = 1024
	DiscoverCGFileset     = "DISCOVER_CG_FILESET"
)

type RestAPI struct {
	GuiHost string `json:"guiHost"`
	GuiPort int    `json:"guiPort"`
}

type Clusters struct {
	ID             string    `json:"id"`
	Primary        Primary   `json:"primary,omitempty"`
	SecureSslMode  bool      `json:"secureSslMode"`
	Cacert         string    `json:"cacert"`
	Secrets        string    `json:"secrets"`
	RestAPI        []RestAPI `json:"restApi"`
	PrimaryCluster string    `json:"primaryCluster"`

	MgmtUsername string
	MgmtPassword string
	CacertValue  []byte
}

const (
	DefaultGuiPort int    = 443
	GuiProtocol    string = "https"
	ConfigMapFile  string = "/var/lib/ibm/config/spectrum-scale-config.json"
	// #nosec G101
	SecretBasePath  string = "/var/lib/ibm/" //nolint:gosec
	CertificatePath string = "/var/lib/ibm/ssl/public"
)

func LoadScaleConfigSettings(ctx context.Context) ScaleSettingsConfigMap {
	klog.V(6).Infof("[%s] scale_config LoadScaleConfigSettings", utils.GetLoggerId(ctx))

	file, e := os.ReadFile(ConfigMapFile) // TODO
	if e != nil {
		klog.Errorf("[%s] IBM Storage Scale configuration not found: %v", utils.GetLoggerId(ctx), e)
		return ScaleSettingsConfigMap{}
	}
	cmsj := &ScaleSettingsConfigMap{}
	e = json.Unmarshal(file, cmsj)
	if e != nil {
		klog.Errorf("[%s] error in unmarshalling IBM Storage Scale configuration json: %v", utils.GetLoggerId(ctx), e)
		return ScaleSettingsConfigMap{}
	}

	e = HandleSecretsAndCerts(ctx, cmsj)
	if e != nil {
		klog.Errorf("[%s] error in secrets or certificates: %v", utils.GetLoggerId(ctx), e)
		return ScaleSettingsConfigMap{}
	}
	return *cmsj
}

func HandleSecretsAndCerts(ctx context.Context, cmap *ScaleSettingsConfigMap) error {
	klog.V(6).Infof("[%s] scale_config HandleSecrets", utils.GetLoggerId(ctx))
	for i := 0; i < len(cmap.Clusters); i++ {
		if cmap.Clusters[i].Secrets != "" {
			unamePath := path.Join(SecretBasePath, cmap.Clusters[i].ID+secretFileSuffix, "username")
			file, e := os.ReadFile(unamePath) // #nosec G304 Valid Path is generated internally
			if e != nil {
				return fmt.Errorf("the IBM Storage Scale secret not found: %v", e)
			}
			file_s := string(file)
			file_s = strings.TrimSpace(file_s)
			file_s = strings.TrimSuffix(file_s, "\n")
			cmap.Clusters[i].MgmtUsername = file_s

			pwdPath := path.Join(SecretBasePath, cmap.Clusters[i].ID+secretFileSuffix, "password")
			file, e = os.ReadFile(pwdPath) // #nosec G304 Valid Path is generated internally
			if e != nil {
				return fmt.Errorf("the IBM Storage Scale secret not found: %v", e)
			}
			file_s = string(file)
			file_s = strings.TrimSpace(file_s)
			file_s = strings.TrimSuffix(file_s, "\n")
			cmap.Clusters[i].MgmtPassword = file_s
		}

		if cmap.Clusters[i].SecureSslMode && cmap.Clusters[i].Cacert != "" {
			certPath := path.Join(CertificatePath, cmap.Clusters[i].ID+cacertFileSuffix)
			certFile, err := os.ReadDir(certPath)
			if err != nil {
				return fmt.Errorf("failed to read the IBM Storage Scale secret: %v", err)
			}

			// the certificate directory i.e. cacert configMap should contain exactly one cert file with the dynamic name
			if len(certFile) != 1 {
				return fmt.Errorf("no cert files found in secret mount path or multiple cert files found in the same cert cm: %v", certFile)
			}
			// only one data file should present in the cert cm, above check
			certPath = path.Join(certPath, certFile[0].Name())
			// certPath = path.Join(certPath, cmap.Clusters[i].Cacert)
			file, e := os.ReadFile(certPath) // #nosec G304 Valid Path is generated internally
			if e != nil {
				return fmt.Errorf("the IBM Storage Scale CA certificate not found: %v", e)
			}
			cmap.Clusters[i].CacertValue = file
		}
	}
	return nil
}
