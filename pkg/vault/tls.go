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

package vault

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"
)

func (c *certificate) refresh() error {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)})
	req := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: c.name},
		DNSNames: []string{c.name},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, req, private)
	if err != nil {
		return err
	}
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: csr})

	outData := &struct {
		CSR    string `json:"csr"`
		Common string `json:"common_name"`
		TTL    string `json:"ttl"`
	}{string(csrPEM), c.name, fmt.Sprintf("%v", int(time.Until(c.vault.expiry).Seconds()))}
	inData := &struct {
		Data struct {
			PublicPEM string `json:"certificate"`
			CAPEM     string `json:"issuing_ca"`
		} `json:"data"`
	}{}

	if err := c.vault.http.Request(context.Background(), c.vault.host+c.path+"sign/"+c.role, http.MethodPost).Header("X-Vault-Token", c.vault.token).JSONBody(outData).Send(http.StatusOK).JSONBody(inData).Check(); err != nil {
		return err
	}
	if !c.caPool.AppendCertsFromPEM([]byte(inData.Data.CAPEM)) {
		panic("invalid PEM received")
	}

	crt, err := tls.X509KeyPair([]byte(inData.Data.PublicPEM), privatePEM)
	if err != nil {
		return err
	}
	*c.destination = crt
	return nil
}

func (c *Client) certificate(path, role, name string, destination *tls.Certificate, caPool *x509.CertPool) *certificate {
	r := &certificate{c, path, role, name, destination, caPool}
	c.certificates = append(c.certificates, r)
	return r
}
