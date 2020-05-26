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

// Package tradfri interfaces with a tradfri gateway and controls lights and stuff.
package tradfri

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	coap "github.com/go-ocf/go-coap"
	"github.com/go-ocf/go-coap/codes"
	dtls "github.com/pion/dtls/v2"
)

const identityLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Connect to the tradfri gateway.
func Connect(address, master, identity, key string) (*coap.ClientConn, string, string, error) {
	mainConfig := dtls.Config{
		PSK:             func(hint []byte) ([]byte, error) { return []byte(key), nil },
		PSKIdentityHint: []byte(identity),
		CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_CCM_8},
	}
	sessionConnection, err := coap.DialDTLSWithTimeout("udp-dtls", net.JoinHostPort(address, "5684"), &mainConfig, 15*time.Second)
	if err == nil {
		return sessionConnection, identity, key, nil
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	identity = ""
	for i := 0; i < 8; i++ {
		identity += string(identityLetters[rand.Intn(len(identityLetters))])
	}
	loginConfig := dtls.Config{
		PSK:             func(hint []byte) ([]byte, error) { return []byte(master), nil },
		PSKIdentityHint: []byte("Client_identity"),
		CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_CCM_8},
	}
	loginConnection, err := coap.DialDTLSWithTimeout("udp-dtls", net.JoinHostPort(address, "5684"), &loginConfig, 15*time.Second)
	if err != nil {
		return nil, "", "", err
	}

	ctx, cancel := context.WithCancel(context.Background())
	loginMessage, err := loginConnection.PostWithContext(ctx, "15011/9063", coap.TextPlain, strings.NewReader(fmt.Sprintf("{\"9090\": \"%s\"}", identity)))
	cancel()

	loginMessage.Code()
	if err != nil {
		return nil, "", "", err
	}

	if loginMessage.Code() != codes.Created {
		return nil, "", "", &CoAPError{StatusCode: loginMessage.Code()}
	}
	_ = loginConnection.Close()

	key = strings.Split(string(loginMessage.Payload()), "\"")[3]
	mainConfig.PSKIdentityHint = []byte(identity)
	mainConfig.PSK = func(hint []byte) ([]byte, error) { return []byte(key), nil }

	sessionConnection, err = coap.DialDTLSWithTimeout("udp-dtls", net.JoinHostPort(address, "5684"), &mainConfig, 15*time.Second)
	return sessionConnection, identity, key, err
}
