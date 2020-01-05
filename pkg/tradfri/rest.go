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

package tradfri

import (
	"github.com/bocajim/dtls"
	"go.eqrx.net/mauzr/pkg/io/rest"
)

const (
	form = `
<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
	<form method="get">
    <input type="radio" name="power" value="false" checked> Off<br>
    <input type="radio" name="power" value="true"> On<br>
		<input type="submit" value="Submit">
	</form>
</body>
</html>
`
)

func setupMapping(c rest.REST, name, group string, params dtls.PeerParams) {
	// TODO get states.
	c.GetEndpoint("/"+name, form, func(query *rest.Request) {
		args := struct {
			Power *bool    `json:"power,string"`
			Level *float64 `json:"level,string"`
		}{}

		if err := query.Args(&args); err != nil {
			return
		}
		change := light{}
		if args.Power != nil {
			change.setPower(*args.Power)
		}
		if args.Level != nil {
			change.setLevel(*args.Level)
		}
		query.InternalError = change.apply(params, group)
	})
}
