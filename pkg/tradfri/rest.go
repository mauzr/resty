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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	coap "github.com/go-ocf/go-coap"
	"github.com/go-ocf/go-coap/codes"
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

func handleLamp(c rest.REST, name, group string, connection *coap.ClientConn) {
	mutex := sync.Mutex{}
	lastPower := false
	c.Endpoint(fmt.Sprintf("/%s/status", name), func(query *rest.Request) {
		mutex.Lock()
		query.ResponseBody = []byte("off")
		if lastPower {
			query.ResponseBody = []byte("on")
		}
		mutex.Unlock()
	})
	c.Endpoint(fmt.Sprintf("/%s", name), func(query *rest.Request) {
		if !query.HasArgs {
			query.ResponseBody = []byte(form)
			return
		}
		args := struct {
			Power bool     `json:"power,string"`
			Level *float64 `json:"level,string"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		if args.Level != nil {
			*args.Level = math.Max(0.0, math.Min(*args.Level, 1.0)) * 254.0
		}

		request := map[string]int{}
		if args.Power {
			request["5850"] = 1
		} else {
			request["5850"] = 0
		}

		switch {
		case args.Level == nil:
		default:
			request["5851"] = int(*args.Level)
		}

		ctx, cancel := context.WithTimeout(query.Ctx, 15*time.Second)
		rawChange, err := json.Marshal(request)
		if err != nil {
			panic(err)
		}
		buf := bytes.NewBuffer(rawChange)
		message, err := connection.PutWithContext(ctx, fmt.Sprintf("/15004/%v", group), coap.TextPlain, buf)
		cancel()

		switch {
		case err != nil:
			query.GatewayError = err
		case message.Code() != codes.Changed:
			query.GatewayError = &CoAPError{StatusCode: message.Code()}
		default:
			mutex.Lock()
			lastPower = args.Power
			mutex.Unlock()
		}
	})
}
