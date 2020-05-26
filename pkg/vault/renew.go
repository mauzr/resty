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
	"time"
)

// AutoRenew automatically renews certificates.
func (c *Client) AutoRenew(ctx context.Context) <-chan error {
	errs := make(chan error)
	go func() {
		defer close(errs)
		timer := time.NewTimer(time.Until(c.expiry))
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
			case <-ctx.Done():
				return
			}

			if err := c.login(c); err != nil {
				errs <- err
				return
			}
			for _, c := range c.certificates {
				if err := c.refresh(); err != nil {
					errs <- err
					return
				}
			}
			timer.Reset(time.Until(c.expiry))
		}
	}()
	return errs
}
