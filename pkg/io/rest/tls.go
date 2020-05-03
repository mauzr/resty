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
	"io/ioutil"
	"path/filepath"
)

// loadCA from a file.
func loadCA(caPath string) *x509.CertPool {
	pool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(filepath.Clean(caPath))
	if err != nil {
		panic(fmt.Errorf("failed to load CA file from %v: %w", caPath, err))
	}
	if !pool.AppendCertsFromPEM(ca) {
		panic(fmt.Sprintf("failed to parse CA file from %v", caPath))
	}
	return pool
}

// loadCertificate from files.
func loadCertificate(crtPath, keyPath string) tls.Certificate {
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to load TLS cert/key pair from %v & %v: %w", crtPath, keyPath, err))
	}
	return cert
}
