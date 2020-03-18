/**
 * Copyright 2019 IBM Corp.
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
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/IBM/ibm-spectrum-scale-csi/driver/pkg/scale/utils"
	"github.com/golang/glog"
	"k8s.io/klog"
)

/*ConfigMap layout and json structure of CSI for Scale plugin.
 */
type ConfigMap struct {
	Clusters []Cluster `json:"clusters"`

	/* References Primary found in parsed Clusters

	note: this is kind of *not ideal*, but it is the way the original ConfigMap is structured,
	see `Validate()` which should error if multiple clusters are designated as primary, or differing primaries are designated
	*/
	Primary *Primary

	RootCAs *x509.CertPool

	/*InsecureSkipTLSVerify
	 */
	InsecureSkipTLSVerify bool
}

/*Primary layout and json structure for primary (i.e. local) scale cluster config
 */
type Primary struct {
	PrimaryFSDep  string `json:"primaryFS"` // Deprecated
	PrimaryFs     string `json:"primaryFs"`
	PrimaryFset   string `json:"primaryFset"`
	PrimaryCid    string `json:"primaryCid"`
	InodeLimitDep string `json:"inode-limit"` // Deprecated
	InodeLimits   string `json:"inodeLimit"`
	RemoteCluster string `json:"remoteCluster"`
	RemoteFSDep   string `json:"remoteFS"` // Deprecated
	RemoteFs      string `json:"remoteFs"`

	PrimaryFSMount      string
	PrimaryFsetLink     string
	SymlinkAbsolutePath string
	SymlinkRelativePath string
}

/*To support backwards compatibility if the PrimaryFs field is not defined then
 *use the previous version of the field.
 */
func (primary Primary) GetPrimaryFs() string {
	if primary.PrimaryFs == "" {
		return primary.PrimaryFSDep
	}
	return primary.PrimaryFs
}

/*To support backwards compatibility if the RemoteFs field is not defined then
 *use the previous version of the field.
 */
func (primary Primary) GetRemoteFs() string {
	if primary.RemoteFs == "" {
		return primary.RemoteFSDep
	}
	return primary.RemoteFs
}

/*To support backwards compatibility if the InodeLimit field is not defined then
 *use the previous version of the field.
 */
func (primary Primary) GetInodeLimit() string {
	if primary.InodeLimits == "" {
		return primary.InodeLimitDep
	}
	return primary.InodeLimits
}

/*Cluster structure within CSI for Scale plugin's ConfigMap.
 */
type Cluster struct {
	ID            string    `json:"id"`
	Primary       Primary   `json:"primary,omitempty"`
	SecureSslMode bool      `json:"secureSslMode"`
	Cacert        string    `json:"cacert"`
	Secrets       string    `json:"secrets"`
	RestAPI       []RestAPI `json:"restApi"`

	MgmtUsername string
	MgmtPassword string
}

func (c *Cluster) ClusterID() string {
	return c.ID
}

func (p *Primary) ClusterID() string {
	return p.PrimaryCid
}

/*RestAPI structure of element within Cluster.
 */
type RestAPI struct {
	GuiHost string `json:"guiHost"`
	GuiPort int    `json:"guiPort"`
}

/*Default constants related to CSI for Scale plugin's configuration.
 */
const (
	DefaultGuiPort  int    = 443
	GuiProtocol     string = "https"
	ConfigMapFile   string = "/var/lib/ibm/config/spectrum-scale-config.json"
	SecretBasePath  string = "/var/lib/ibm/" //nolint:gosec
	CertificatePath string = "/var/lib/ibm/ssl/public"
)

/*LoadScaleConfig from ConfigMap for CSI for Scale plugin.
 */
func LoadScaleConfig() (*ConfigMap, error) {
	klog.V(5).Infof("configMap LoadScaleConfig")

	file, err := ioutil.ReadFile(ConfigMapFile) // TODO
	if err != nil {
		return nil, fmt.Errorf("Spectrum Scale configuration not found: %v", err)
	}
	cmsj := &ConfigMap{}
	err = json.Unmarshal(file, cmsj)
	if err != nil {
		return nil, fmt.Errorf("Error in unmarshalling Spectrum Scale configuration json: %v", err)
	}

	err = handleSecretsAndCerts(cmsj)
	if err != nil {
		return nil, fmt.Errorf("Error in secrets or certificates: %v", err)
	}

	//fail fast!; note: this also sets the Primary reference
	if err := cmsj.Validate(); err != nil {
		return nil, fmt.Errorf("could not validate Spectrum Scale configuration: %v", err)
	}

	return cmsj, nil
}

func handleSecretsAndCerts(config *ConfigMap) error {
	klog.V(5).Infof("configMap handleSecretsAndCerts")

	rootCAs := x509.NewCertPool()

	for i := range config.Clusters {
		//refernce it, as we will be modifying
		cluster := &config.Clusters[i]

		if cluster.Secrets != "" {
			userPath := filepath.Join(SecretBasePath, cluster.Secrets, "username")
			user, err := ioutil.ReadFile(userPath)
			if err != nil {
				return fmt.Errorf(`Spectrum Scale secret not found: %w`, err)
			}
			cluster.MgmtUsername = strings.TrimSpace(string(user))

			passPath := path.Join(SecretBasePath, cluster.Secrets, "password")
			pass, err := ioutil.ReadFile(passPath)
			if err != nil {
				return fmt.Errorf(`Spectrum Scale secret not found: %w`, err)
			}
			cluster.MgmtPassword = strings.TrimSpace(string(pass))
		}

		if cluster.Cacert != "" {
			certPath := path.Join(CertificatePath, cluster.Cacert)
			cert, err := ioutil.ReadFile(certPath)
			if err != nil {
				return fmt.Errorf(`Spectrum Scale CA certificate not found: %w`, err)
			}
			ok := rootCAs.AppendCertsFromPEM(cert)
			if !ok {
				return fmt.Errorf(`Parsing CA cert %v failed`, cert)
			}
			config.RootCAs = rootCAs
		}

		if !cluster.SecureSslMode {
			klog.V(4).Infof("Spectrum Scale REST connections are InsecureSkipTLSVerify")
			config.InsecureSkipTLSVerify = true
		}
	}

	return nil
}

/*Validate Spectrum Scale configuration ConfigMap
 */
func (config *ConfigMap) Validate() error {
	klog.V(4).Infof("gpfs ConfigMap Validate.")
	if len(config.Clusters) == 0 {
		return fmt.Errorf("Missing cluster information in Spectrum Scale configuration")
	}

	primaryClusterFound := false
	rClusterForPrimaryFS := ""
	var cl []string = make([]string, len(config.Clusters))
	issueFound := false

	for i, cluster := range config.Clusters {

		if cluster.ID == "" {
			issueFound = true
			glog.Errorf("Mandatory parameter 'id' is not specified")
		}
		if len(cluster.RestAPI) == 0 {
			issueFound = true
			glog.Errorf("Mandatory section 'restApi' is not specified for cluster %v", cluster.ID)
		}
		if len(cluster.RestAPI) != 0 && cluster.RestAPI[0].GuiHost == "" {
			issueFound = true
			glog.Errorf("Mandatory parameter 'guiHost' is not specified for cluster %v", cluster.ID)
		}
		if cluster.RestAPI[0].GuiPort == 0 {
			cluster.RestAPI[0].GuiPort = DefaultGuiPort
		}

		if cluster.Primary != (Primary{}) {
			if primaryClusterFound {
				issueFound = true
				glog.Errorf("More than one primary clusters specified")
			}

			primaryClusterFound = true
			config.Primary = &cluster.Primary

			if cluster.Primary.GetPrimaryFs() == "" {
				issueFound = true
				glog.Errorf("Mandatory parameter 'primaryFs' is not specified for primary cluster %v", cluster.ID)
			}
			if cluster.Primary.PrimaryFset == "" {
				issueFound = true
				glog.Errorf("Mandatory parameter 'primaryFset' is not specified for primary cluster %v", cluster.ID)
			}

			rClusterForPrimaryFS = cluster.Primary.RemoteCluster
		} else {
			cl[i] = cluster.ID
		}

		if cluster.Secrets == "" {
			return fmt.Errorf("Invalid secret specified for cluster %v", cluster.ID)
		}

		if cluster.Secrets == "" {
			issueFound = true
			glog.Errorf("Mandatory parameter 'secrets' is not specified for cluster %v", cluster.ID)
		}

		if cluster.SecureSslMode && cluster.Cacert == "" {
			issueFound = true
			glog.Errorf("CA certificate not specified in secure SSL mode for cluster %v", cluster.ID)

		}
	}

	if !primaryClusterFound {
		issueFound = true
		glog.Errorf("No primary clusters specified")
	}

	if rClusterForPrimaryFS != "" && !utils.StringInSlice(rClusterForPrimaryFS, cl) {
		issueFound = true
		glog.Errorf("Remote cluster specified for primary filesystem: %s, but no definition found for it in config", rClusterForPrimaryFS)
	}

	if issueFound {
		return fmt.Errorf("one or more issue found in Spectrum scale csi driver configuration, check Spectrum Scale csi driver logs")
	}

	return nil
}
