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

package sk6812

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

// setupHandler provides a REST handler that sets an SK6812 chain.
func setupHandler(c rest.REST, strip Manager) {
	c.Endpoint("/color", "", func(r *rest.Request) {
		setting := make([]uint8, 0)
		if err := json.Unmarshal(r.RequestBody, &setting); err != nil {
			r.RequestError = fmt.Errorf("illegal setting data: %v", err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Ctx, 1*time.Second)
		defer cancel()
		if err := strip.Set(ctx, setting); err != nil {
			r.InternalError = err
		}
	})
}
