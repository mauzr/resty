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
	"time"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

const (
	partChangeForm = `
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
)

// AddPartChangerEndpoint that will listen for part change requests.
func AddPartChangerEndpoint(c rest.REST, path string, status <-chan string, changers ...chan<- Request) {
	if status != nil {
		c.Endpoint(path+"/status", "", func(query *rest.Request) {
			s := <-status
			query.ResponseBody = []byte(s)
		})
	}
	c.Endpoint(path, partChangeForm, func(query *rest.Request) {
		args := struct {
			Stance string `json:"stance"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(query.Ctx, 3*time.Second)
		defer cancel()
		for _, changer := range changers {
			response := make(chan error, 1)
			req := Request{response, args.Stance}
			select {
			case <-ctx.Done():
				query.InternalError = ctx.Err()
				return
			case changer <- req:
			}
			select {
			case <-ctx.Done():
				query.InternalError = ctx.Err()
				return
			case err, ok := <-response:
				if !ok {
					panic("closed response channel")
				}
				if err != nil {
					query.RequestError = err
					return
				}
			}
		}
	})
}
