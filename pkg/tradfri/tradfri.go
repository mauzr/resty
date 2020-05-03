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
	"math/rand"
	"time"

	"github.com/bocajim/dtls"
	coap "github.com/dustin/go-coap"
)

const identityLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type CoAPError struct {
	StatusCode coap.COAPCode
	Cause      error
}

func (c CoAPError) Error() string {
	switch {
	case c.Cause != nil:
		return c.Cause.Error()
	case c.StatusCode != 0:
		return c.StatusCode.String()
	default:
		panic("empty error")
	}
}

func (c CoAPError) Unwrap() error {
	return c.Cause
}

func login(address, identity, key string) (psk string, err error) {
	listener, err := dtls.NewUdpListener(":0", time.Second*900)
	if err != nil {
		return
	}
	defer func() {
		if err := listener.Shutdown(); err != nil {
			panic(err)
		}
	}()
	loginIdentity := "Client_identity"
	mks := dtls.NewKeystoreInMemory()
	dtls.SetKeyStores([]dtls.Keystore{mks})
	mks.AddKey(loginIdentity, []byte(key))
	loginPeer, err := listener.AddPeerWithParams(&dtls.PeerParams{Addr: address, Identity: loginIdentity, HandshakeTimeout: time.Second * 15})
	if err != nil {
		return
	}
	loginPeer.UseQueue(true)

	pskPayload := struct {
		Ident string `json:"9090"`
	}{Ident: identity}
	pskData, err := json.Marshal(pskPayload)
	if err != nil {
		return
	}
	pskRequest := coap.Message{Type: coap.Confirmable, Code: coap.POST, MessageID: 0, Payload: pskData}
	pskRequest.SetPath([]string{"15011", "9063"})
	pskBinary, err := pskRequest.MarshalBinary()
	if err != nil {
		return
	}
	if err = loginPeer.Write(pskBinary); err != nil {
		return
	}

	pskResponseBinary, err := loginPeer.Read(time.Second)
	if err != nil {
		return
	}
	pskMessage, err := coap.ParseMessage(pskResponseBinary)
	if err != nil {
		return
	}
	if pskMessage.Code != coap.Created {
		err = &CoAPError{StatusCode: pskMessage.Code}
		return
	}
	pskResp := struct {
		PSK string `json:"9091"`
	}{}
	err = json.Unmarshal(pskMessage.Payload, &pskResp)
	if err != nil {
		panic(err)
	}
	psk = pskResp.PSK

	return psk, err
}

func connect(address, key string) (peer *dtls.Peer, err error) {
	rand.Seed(time.Now().UnixNano())
	identity := ""
	for i := 0; i < 8; i++ {
		identity += string(identityLetters[rand.Intn(len(identityLetters))])
	}
	psk, err := login(address, identity, key)
	if err != nil {
		return
	}

	mks := dtls.NewKeystoreInMemory()
	dtls.SetKeyStores([]dtls.Keystore{mks})
	mks.AddKey(identity, []byte(psk))

	listener, err := dtls.NewUdpListener(":0", time.Second*900)
	if err != nil {
		panic(err)
	}

	peerParams := &dtls.PeerParams{Addr: address, Identity: identity, HandshakeTimeout: time.Second * 15}
	peer, err = listener.AddPeerWithParams(peerParams)
	if err != nil {
		return
	}
	peer.UseQueue(true)
	return
}
