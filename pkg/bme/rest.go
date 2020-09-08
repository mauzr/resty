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
	"net/http"
	"time"

	"go.eqrx.net/mauzr/pkg/log"
	"go.eqrx.net/mauzr/pkg/rest"
)

const (
	measureTimeout = 3 * time.Second
)

// Send a measurement to remote sites.
func Send(ctx context.Context, c rest.Client, requests chan<- Request, interval time.Duration, destinations ...string) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			resps := make(chan Response, 1)
			select {
			case <-ctx.Done():
				return
			case requests <- Request{resps, time.Now().Add(interval)}:
			}

			var resp Response
			select {
			case <-ctx.Done():
				return
			case resp = <-resps:
			}

			if resp.Err != nil {
				log.Root.Warning("could not fetch measurement: %v", resp.Err)
			}

			reqs := make([]rest.ClientRequest, len(destinations))
			for i, d := range destinations {
				reqs[i] = c.Request(context.Background(), d, http.MethodPut).JSONBody(&resp.Measurement)
			}
			rest.GoSendAll(http.StatusOK, log.Root.Warning, reqs...)
		}
	}()
}

// Expose creates a http handler that handles measurements with the given manager.
func Expose(mux rest.Mux, path string, requests chan<- Request) {
	mux.Endpoint(path, func(query *rest.Request) {
		args := struct {
			MaxAge string `json:"maxAge"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		maxAge, err := time.ParseDuration(args.MaxAge)
		if err != nil {
			query.RequestErr = err

			return
		}

		responses := make(chan Response, 1)
		request := Request{responses, time.Now().Add(-maxAge)}

		measureCtx, measureCtxCancel := context.WithTimeout(query.Ctx, measureTimeout)
		defer measureCtxCancel()

		select {
		case <-measureCtx.Done():
			query.InternalErr = measureCtx.Err()

			return
		case requests <- request:
		}
		select {
		case <-measureCtx.Done():
			query.InternalErr = measureCtx.Err()
		case response, ok := <-responses:
			switch {
			case !ok:
				panic("unknown internal error")
			case response.Err != nil:
				query.InternalErr = response.Err
			default:
				query.ResponseBody, query.InternalErr = json.Marshal(response.Measurement)
			}
		}
	})
}
