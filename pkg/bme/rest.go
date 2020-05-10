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

package bme

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

// setupHandler creates a http handler that handles measurements with the given manager.
func setupHandler(c rest.REST, requests chan<- Request, tags map[string]string) {
	c.Endpoint("/measurement", func(query *rest.Request) {
		args := struct {
			MaxAge string `json:"maxAge"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		maxAge, err := time.ParseDuration(args.MaxAge)
		if err != nil {
			query.RequestError = err
			return
		}

		responses := make(chan Response, 1)
		request := Request{responses, time.Now().Add(-maxAge)}

		measureCtx, measureCtxCancel := context.WithTimeout(query.Ctx, 3*time.Second)
		defer measureCtxCancel()

		select {
		case <-measureCtx.Done():
			query.InternalError = measureCtx.Err()
			return
		case requests <- request:
		}
		select {
		case <-measureCtx.Done():
			fmt.Println("timeout")
			query.InternalError = measureCtx.Err()
		case response, ok := <-responses:
			switch {
			case !ok:
				panic("unknown internal error")
			case response.Err != nil:
				fmt.Println(response.Err)
				query.InternalError = response.Err
			default:
				response.Measurement.Tags = tags
				query.ResponseBody, query.InternalError = json.Marshal(response.Measurement)
			}
		}
	})
}
