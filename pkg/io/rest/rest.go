/*
Copyright 2019 Alexander Sowitzki.

GNU Affero General Public License version 3 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://opensource.org/licenses/AGPL-3.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

type REST interface {
	GetJSON(context.Context, string, interface{}) Error
	GetRaw(context.Context, string) (*http.Response, error)
	PostRaw(context.Context, string, io.Reader) (*http.Response, error)
	Endpoint(path, form string, queryHandler func(query *Request))
	Serve(context.Context) error
	AddDefaultResponseHeader(http.Header)
	ServerNames() []string
}

type rest struct {
	mux          *http.ServeMux
	client       http.Client
	servers      []http.Server
	serverErrors chan error
	serverNames  []string
}

func New(hostname string, listenAddresses []string) REST {
	rest := rest{
		http.NewServeMux(),
		http.Client{},
		make([]http.Server, len(listenAddresses)),
		make(chan error),
		nil,
	}
	rest.client = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ClientAuth:               tls.RequireAndVerifyClientCert,
				MinVersion:               tls.VersionTLS13,
				PreferServerCipherSuites: true,
				Certificates: []tls.Certificate{
					loadCertificate("/etc/ssl/certs/"+hostname+"-client.crt", "/etc/ssl/private/"+hostname+"-client.key"),
				},
			},
		},
	}
	serverCertificate := loadCertificate("/etc/ssl/certs/"+hostname+".crt", "/etc/ssl/private/"+hostname+".key")
	rest.serverNames = serverCertificate.Leaf.DNSNames
	tlsConfig := tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		ClientCAs:                loadCA("/etc/ssl/certs/" + hostname + "-ca.crt"),
		Certificates:             []tls.Certificate{serverCertificate},
	}
	for i, listenAddress := range listenAddresses {
		rest.servers[i] = http.Server{
			Addr:              listenAddress,
			TLSConfig:         &tlsConfig,
			ReadHeaderTimeout: 3 * time.Second,
			IdleTimeout:       120 * time.Second,
			Handler:           rest.mux,
		}
	}
	return &rest
}
