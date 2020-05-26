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
	"net/http"
)

// GetSecret reads a secret into an interface.
func (c *Client) GetSecret(backend, path string, destination interface{}) error {
	if c.token == "" {
		return ErrNotLoggedIn
	}
	response := struct {
		Data struct {
			Data interface{} `json:"data"`
		} `json:"data"`
	}{}
	response.Data.Data = destination
	return c.http.Request(context.Background(), c.host+backend+"data/"+path, http.MethodGet).Header("X-Vault-Token", c.token).Send(http.StatusOK).JSONBody(&response).Check()
}

// UpdateSecret reads a secret from an interface into vault.
func (c *Client) UpdateSecret(backend, path string, source interface{}) error {
	if c.token == "" {
		return ErrNotLoggedIn
	}
	request := struct {
		Data interface{} `json:"data"`
	}{source}
	return c.http.Request(context.Background(), c.host+backend+"data/"+path, http.MethodPost).JSONBody(&request).Header("X-Vault-Token", c.token).Send(http.StatusOK).Check()
}
