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

// Package vault interfaces with the vault.
package vault

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"go.eqrx.net/mauzr/pkg/rest"
)

type certificate struct {
	vault            *Client
	path, role, name string
	destination      *tls.Certificate
	caPool           *x509.CertPool
}

// Client for vault operations.
type Client struct {
	http         rest.Client
	host         string
	token        string
	expiry       time.Time
	certificates []*certificate
	login        func(*Client) error
}

// New creates a new vault client.
func New(h rest.Client, host string, login func(*Client) error) *Client {
	return &Client{h, host + "v1/", "", time.Time{}, []*certificate{}, login}
}
