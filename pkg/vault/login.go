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
	"time"
)

// TokenLogin just uses the given token for authentication with vault.
func TokenLogin(token string) func(*Client) error {
	return func(c *Client) error {
		response := struct {
			ExpireTime time.Time `json:"expire_time"`
		}{}
		c.token = token
		err := c.http.Request(context.Background(), c.host+"auth/token/lookup-self", http.MethodGet).Header("X-Vault-Token", c.token).Send(http.StatusOK).JSONBody(&response).Check()
		if err != nil {
			c.token = ""
			return err
		}
		c.expiry = response.ExpireTime
		return nil
	}
}

// AppRoleLogin logs in with the given app role.
func AppRoleLogin(role, secret string) func(*Client) error {
	return func(c *Client) error {
		request := struct {
			Role   string `json:"role_id"`
			Secret string `json:"secret_id"`
		}{role, secret}
		response := struct {
			Auth struct {
				Duration int    `json:"lease_duration"`
				Token    string `json:"client_token"`
			} `json:"auth"`
		}{}
		err := c.http.Request(context.Background(), c.host+"auth/approle/login", http.MethodPost).JSONBody(&request).Send(http.StatusOK).JSONBody(&response).Check()
		if err != nil {
			return err
		}

		c.expiry = time.Now().Add(time.Duration(response.Auth.Duration) * time.Second)
		c.token = response.Auth.Token
		return nil
	}
}

// Login once.
func (c *Client) Login() error {
	return c.login(c)
}
