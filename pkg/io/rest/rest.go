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

// Package rest contains tooling for server and client side REST interaction.
package rest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"time"
)

//go:generate esc -o static.go --pkg rest --prefix=../../../web/ ../../../web/

// REST provides the interface to REST io.
type REST interface {
	// GetJSON from a remote site. It gets serialized into the given interface.
	GetJSON(context.Context, string, interface{}) error
	// GetRaw response from a remote site.
	GetRaw(context.Context, string) (*http.Response, error)
	// PostRaw from the given reader to a remote site.
	PostRaw(context.Context, string, io.Reader) (*http.Response, error)
	// Endpoint provides a server end point for a rest application. The given handler is called on each invoction.
	Endpoint(path, form string, queryHandler func(query *Request))
	// Serve blocks and runs the configured http servers.
	Serve() []<-chan error
	// AddDefaultResponseHeader to the given header.
	AddDefaultResponseHeader(http.Header)
	// ServerNames that are being served by this interface.
	ServerNames() []string

	WebserverContext() context.Context

	Mux() *http.ServeMux
	Client() *http.Client
}

// rest is the implementation of the REST interface.
type rest struct {
	webserverContext context.Context
	mux              *http.ServeMux
	client           http.Client
	listeners        []net.Listener
	servers          []http.Server
	serverNames      []string
}

func (r *rest) Client() *http.Client {
	return &r.client
}

func (r *rest) Mux() *http.ServeMux {
	return r.mux
}

// New creates a new REST interface.
func New(ctx context.Context, serviceName string, listeners []net.Listener) REST {
	rest := rest{
		ctx,
		http.NewServeMux(),
		http.Client{},
		listeners,
		make([]http.Server, len(listeners)),
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
					loadCertificate("/etc/ssl/certs/"+serviceName+"-client.crt", "/etc/ssl/private/"+serviceName+"-client.key"),
				},
			},
		},
	}
	serverCertificate := loadCertificate("/etc/ssl/certs/"+serviceName+".crt", "/etc/ssl/private/"+serviceName+".key")
	c, err := x509.ParseCertificate(serverCertificate.Certificate[0])
	if err != nil {
		panic(err)
	}
	rest.serverNames = c.DNSNames
	tlsConfig := tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		ClientCAs:                loadCA("/etc/ssl/certs/" + serviceName + "-ca.crt"),
		Certificates:             []tls.Certificate{serverCertificate},
		NextProtos:               []string{"h2"},
	}
	for i := range listeners {
		rest.servers[i] = http.Server{
			TLSConfig:         &tlsConfig,
			ReadHeaderTimeout: 3 * time.Second,
			IdleTimeout:       120 * time.Second,
			Handler:           rest.mux,
		}
	}
	rest.mux.Handle("/favicon.ico", http.FileServer(FS(false)))
	rest.mux.HandleFunc("/debug/pprof/", pprof.Index)
	rest.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	rest.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	rest.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	rest.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	rest.Endpoint("/health", "I am alive!", func(r *Request) { r.RequestError = fmt.Errorf("%w: no arguments supported", ErrRequest) })

	return &rest
}

// ServerNames that are being served by this interface.
func (r *rest) ServerNames() []string {
	return r.serverNames
}
