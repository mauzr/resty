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
	"fmt"
	"net/http"

	"github.com/bocajim/dtls"
	"go.eqrx.net/mauzr/pkg/rest"
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
    <input type="radio" name="power" value="0" checked> Off<br>
    <input type="radio" name="power" value="1"> On<br>
		<input type="submit" value="Submit">
	</form>
</body>
</html>
`
)

func setupMapping(mux *http.ServeMux, name, group string, params dtls.PeerParams) {
	rest.Endpoint(mux, "/"+name, form, func(query *rest.Query) {
		var level float64
		var power, powerSet, levelSet bool

		arguments := []rest.Argument{
			rest.BoolArgument("power", &power, &powerSet),
			rest.FloatArgument("level", &level, &levelSet),
		}

		if query.QueryError = query.CollectArguments(arguments); query.QueryError == nil {
			fmt.Println(query.QueryError, powerSet, power)
			change := light{}
			if powerSet {
				change.setPower(power)
			}
			if levelSet {
				change.setLevel(level)
			}
			query.InternalError = change.apply(params, group)
		}
	})
}
