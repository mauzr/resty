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
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

func TLSConfig(crtPath, keyPath string) *tls.Config {
	config := tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		ClientCAs:                x509.NewCertPool(),
	}

	if cert, err := tls.LoadX509KeyPair(crtPath, keyPath); err != nil {
		panic(fmt.Errorf("failed to load TLS cert/key pair from %v & %v: %v", crtPath, keyPath, err))
	} else {
		config.Certificates = []tls.Certificate{cert}
		return &config
	}
}
