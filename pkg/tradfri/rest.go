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
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/bocajim/dtls"
	coap "github.com/dustin/go-coap"
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

func handleLamp(c rest.REST, name, group string, peer *dtls.Peer) {
	c.Endpoint("/"+name, form, func(query *rest.Request) {
		args := struct {
			Power *bool    `json:"power,string"`
			Level *float64 `json:"level,string"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		if args.Power == nil && args.Level == nil {
			query.RequestError = fmt.Errorf("nothing requested")
		}

		request := map[string]int{}
		switch {
		case args.Power == nil:
		case *args.Power:
			request["5850"] = 1
		case !*args.Power:
			request["5850"] = 0
		}

		switch {
		case args.Level == nil:
		default:
			request["5851"] = int(math.Max(0.0, math.Min(*args.Level, 1.0)) * 254.0)
		}

		rawChange, err := json.Marshal(&request)
		if err != nil {
			panic(err)
		}
		req := coap.Message{
			Type:      coap.Confirmable,
			Code:      coap.PUT,
			Payload:   rawChange,
			MessageID: 1,
		}
		req.SetPath([]string{"15004", group})
		rawRequest, err := req.MarshalBinary()
		if err != nil {
			panic(err)
		}

		if err = peer.Write(rawRequest); err != nil {
			query.GatewayError = err
			return
		}

		rawResponse, err := peer.Read(10 * time.Second)
		if err != nil {
			query.GatewayError = err
			return
		}

		response, err := coap.ParseMessage(rawResponse)
		if err == nil && response.Code != coap.Changed {
			query.GatewayError = fmt.Errorf("unexpected CoAP code: %s", response.Code)
			return
		}
	})
}
