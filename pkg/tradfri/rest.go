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
	"net/http"
	"sync"
	"time"

	coap "github.com/go-ocf/go-coap"
	"github.com/go-ocf/go-coap/codes"
	"go.eqrx.net/mauzr/pkg/log"
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
    <input type="radio" name="power" value="false" checked> Off<br>
    <input type="radio" name="power" value="true"> On<br>
		<input type="submit" value="Submit">
	</form>
</body>
</html>
`
)

// LampSetting describe the state of a lamp.
type LampSetting struct {
	Power bool    `json:"power,string"`
	Level float64 `json:"level,string"`
}

// HandleLamp handles a tradfri lamp.
func HandleLamp(c rest.Client, m rest.Mux, name, group string, connection *coap.ClientConn, notifies ...string) {
	mutex := sync.Mutex{}
	lastSetting := LampSetting{}
	m.Endpoint(fmt.Sprintf("/%s/status", name), func(query *rest.Request) {
		mutex.Lock()
		query.ResponseBody, query.InternalErr = json.Marshal(&lastSetting)
		mutex.Unlock()
	})
	m.Endpoint(fmt.Sprintf("/%s", name), func(query *rest.Request) {
		if !query.HasArgs {
			query.ResponseBody = []byte(form)
			return
		}
		args := LampSetting{}
		if err := query.Args(&args); err != nil {
			return
		}

		request := map[string]int{"5850": 0}
		if args.Power {
			request["5850"] = 1
		}
		request["5851"] = int(math.Max(0.0, math.Min(args.Level, 1.0)) * 254.0)

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
			query.GatewayErr = err
		case message.Code() != codes.Changed:
			query.GatewayErr = &CoAPError{StatusCode: message.Code()}
		default:
			notifyRequests := []rest.ClientRequest{}
			for _, notify := range notifies {
				notifyRequests = append(notifyRequests, c.Request(context.Background(), notify, http.MethodPut).JSONBody(&args))
			}
			rest.GoSendAll(http.StatusOK, log.Root.Warning, notifyRequests...)

			mutex.Lock()
			lastSetting = args
			mutex.Unlock()
		}
	})
}
