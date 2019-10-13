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
)

func setupKeys(client, psk string) {
	mks := dtls.NewKeystoreInMemory()
	dtls.SetKeyStores([]dtls.Keystore{mks})
	mks.AddKey(client, []byte(psk))
}

func setupParams(client, gateway string) dtls.PeerParams {
	return dtls.PeerParams{
		Addr:             gateway + ":5684",
		Identity:         client,
		HandshakeTimeout: time.Second * 15,
	}
}

type light map[string]int

func (c *light) apply(params dtls.PeerParams, group string) error {
	rawChange, _ := json.Marshal(c)
	req := coap.Message{
		Type:    coap.Confirmable,
		Code:    coap.PUT,
		Payload: rawChange,
	}
	req.SetPath([]string{"15004", group})

	listener, err := dtls.NewUdpListener(":0", 15*time.Second)
	if err != nil {
		return err
	}

	peer, err := listener.AddPeerWithParams(&params)
	if err != nil {
		return err
	}

	peer.UseQueue(true)

	rawRequest, err := req.MarshalBinary()
	if err != nil {
		return err
	}

	if err = peer.Write(rawRequest); err != nil {
		return err
	}

	rawResponse, err := peer.Read(time.Second)
	if err != nil {
		return err
	}

	response, err := coap.ParseMessage(rawResponse)
	if err == nil && response.Code != coap.Changed {
		err = fmt.Errorf("Unexpected CoAP code: %s", response.Code)
	}

	return err
}

func (c light) setLevel(level float64) {
	dim := int(math.Max(0.0, math.Min(level, 1.0)) * 254.0)
	c["5851"] = dim
}

func (c light) setPower(on bool) {
	power := 0
	if on {
		power = 1
	}

	c["5850"] = power
}
