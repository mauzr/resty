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

package play

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.eqrx.net/mauzr/pkg/log"
	"go.eqrx.net/mauzr/pkg/rest"
)

const (
	// PartChangeForm ist the HTML5 form presented to the user when chaning parts.
	PartChangeForm = `
<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
	<form method="get">
    <input type="radio" name="stance" value="off" checked> Off<br>
    <input type="radio" name="stance" value="default"> Default<br>
    <input type="radio" name="stance" value="bright"> Bright<br>
		<input type="radio" name="stance" value="alert"> Alert<br>
		<input type="radio" name="stance" value="rainbow"> Rainbow<br>
		<input type="radio" name="stance" value="theme"> Theme<br>
		<br>
		<input type="submit" value="Submit">
	</form>
</body>
</html>
`
	updateTimeout = 3 * time.Second
)

// ExposeSend will listen for part change requests and gives out the current status.
func ExposeSend(m rest.Mux, c rest.Client, path string, receivers []string, changers ...chan<- Request) {
	current := "default"
	mutex := sync.Mutex{}

	m.Endpoint(path+"/status", func(query *rest.Request) {
		mutex.Lock()
		query.ResponseBody, query.InternalErr = json.Marshal(&current)
		mutex.Unlock()
	})
	m.Endpoint(path, func(query *rest.Request) {
		if !query.HasArgs {
			query.ResponseBody = []byte(PartChangeForm)

			return
		}
		args := struct {
			Stance string `json:"stance"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		stance := args.Stance
		mutex.Lock()
		updateAll(query, stance, changers)
		current = stance

		reqs := []rest.ClientRequest{}
		for _, receiver := range receivers {
			reqs = append(reqs, c.Request(context.Background(), receiver, http.MethodPut).JSONBody(&current))
		}
		rest.GoSendAll(http.StatusSeeOther, log.Root.Warning, reqs...)
		mutex.Unlock()
	})
}

func updateAll(query *rest.Request, stance string, changers []chan<- Request) {
	ctx, cancel := context.WithTimeout(query.Ctx, updateTimeout)
	defer cancel()
	for _, changer := range changers {
		response := make(chan error, 1)
		req := Request{response, stance}
		select {
		case <-ctx.Done():
			query.InternalErr = ctx.Err()

			return
		case changer <- req:
		}
		select {
		case <-ctx.Done():
			query.InternalErr = ctx.Err()

			return
		case err := <-response:
			if err != nil {
				query.RequestErr = err

				return
			}
		}
	}
}
