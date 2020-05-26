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

// Package contact interfaces with contacts via GPI.
package contact

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.eqrx.net/mauzr/pkg/gpio"
	"go.eqrx.net/mauzr/pkg/log"
	"go.eqrx.net/mauzr/pkg/rest"
)

// ExposeSend exposes and sends the contact state.
func ExposeSend(ctx context.Context, c rest.Client, mux rest.Mux, input gpio.Input, path string, destinations ...string) error {
	var events <-chan gpio.InputEvent
	var closed bool
	if err := input.Current(&closed)(); err != nil {
		return fmt.Errorf("could not poll input for testing: %w", err)
	}
	ok := true
	fmt.Println(path)
	mux.Endpoint(path, func(query *rest.Request) {
		if !ok {
			query.Status = http.StatusInternalServerError
			query.ResponseBody = []byte(fmt.Sprintln("not ready"))
		} else {
			v := "closed"
			if !closed {
				v = "opened"
			}
			query.ResponseBody, query.InternalErr = json.Marshal(&v)
		}
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				ok = false
				return
			case e := <-events:
				closed = e.NewValue
				v := "closed"
				if !closed {
					v = "opened"
				}
				r := make([]rest.ClientRequest, len(destinations))
				for i, d := range destinations {
					r[i] = c.Request(context.Background(), d, http.MethodPut).StringBody(v)
				}
				rest.GoSendAll(http.StatusOK, log.Root.Warning, r...)
			}
		}
	}()
	return nil
}
