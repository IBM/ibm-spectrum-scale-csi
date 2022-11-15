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

package connectors

import (
	"net/http"
)

type RestConfig struct {
	user       string
	passwd     string
	httpClient *http.Client
}

func NewConnector() RestConfig {
	return RestConfig{}
}

/* type ConnectorClient struct {
	httpClient *http.Client
}

func (client ConnectorClient) NewConnector() RestConfig {
	return RestConfig{
		httpClient: client.httpClient,
	}
} */

/* func New(o Options) ConnectorClient {
	t := &http.Transport{
		TLSClientConfig: o.TLSClientConfig,
	}

	client := &http.Client{
		Transport: t,
		Timeout:   time.Second * 10,
	}

	return ConnectorClient{
		httpClient: client,
	}
}

type Options struct {
	TLSClientConfig *tls.Config
} */

/* type Endpointer interface {
	GetScheme() string
	GetHost() string
	GetCredentials() (user string, pass string)
} */

/* type RestV2Connector struct {
	Endpointer
	endPointIndex int
	httpClient	*http.Client
} */

/* type Endpoint struct {
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
}

func (end Endpoint) GetScheme() string {
	hostScheme, _ := end.getSchemeAndHost()
	if end.Scheme == "" {
		if hostScheme == "" {
			return "https"
		}
		return hostScheme
	}
	return end.Scheme
}

func (end Endpoint) GetHost() string {
	_, host := end.getSchemeAndHost()
	if end.Port == 0 {
		return host
	}
	return net.JoinHostPort(
		host,
		strconv.Itoa(end.Port),
	)
}

func (end Endpoint) getSchemeAndHost() (scheme string, host string) {
	split := strings.SplitN(end.Host, "://", 2)
	switch {
	case len(split) > 1:
		return split[0], split[1]
	case len(split) == 1:
		return "", split[0]
	default:
		return "", ""
	}
}

func (e Endpoint) GetCredentials() (user string, pass string) {
	return e.Username, e.Password
}
*/
