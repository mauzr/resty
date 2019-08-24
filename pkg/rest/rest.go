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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func tlsConfig(caPath, crtPath, keyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load TLS cert/key pair from %v & %v: %v", crtPath, keyPath, err)
	}

	pem, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load CA file from %v: %v", caPath, err)
	}

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("Failed to parse CA file from %v", caPath)
	}
	config := &tls.Config{ // Make things "a little" incompatible but secure. Basics taken from https://cipherli.st .
		Certificates:             []tls.Certificate{cert},
		ClientCAs:                certpool,
		Rand:                     rand.Reader,
		ClientAuth:               tls.RequireAndVerifyClientCert,
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}
	return config, nil
}

// Serve creates a REST server secured by TLS with client cert authentication.
func Serve(listen string, caPath string, crtPath string, keyPath string, mux *http.ServeMux) error {
	tlsConfig, err := tlsConfig(caPath, crtPath, keyPath)
	if err != nil {
		return nil
	}
	server := http.Server{
		Addr:              listen,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           mux,
		// Disable HTTP v2.0 since that would require 128 bit ciphers.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return server.ListenAndServeTLS("", "")
}
