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

// CreateSubToken from the current token with the given policy only.
func (c *Client) CreateSubToken(policy string) (string, error) {
	if c.token == "" {
		return "", ErrNotLoggedIn
	}
	request := struct {
		Policies        string `json:"policies"`
		NoDefaultPolicy bool   `json:"no_default_policy"`
		Renewable       bool   `json:"renewable"`
	}{policy, true, false}
	response := struct {
		Auth struct {
			Token string `json:"client_token"`
		} `json:"auth"`
	}{}
	err := c.http.Request(context.Background(), c.host+"auth/token/create", http.MethodPost).JSONBody(&request).Header("X-Vault-Token", c.token).Send(http.StatusOK).JSONBody(&response).Check()
	return response.Auth.Token, err
}
