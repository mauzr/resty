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

package rest

import (
	"context"
	"io"
	"net/http"
)

// BodyDummy mimics the http body.
type BodyDummy struct{}

// Close the dummy.
func (b *BodyDummy) Close() error { return nil }

// Read from the dummy.
func (b *BodyDummy) Read(p []byte) (n int, err error) {
	return 0, nil
}

// dummy mimics an actual REST interface.
type dummy struct {
}

// ServerNames that are being served by this interface.
func (d dummy) ServerNames() []string { return nil }

// AddDefaultResponseHeader to the given header.
func (d dummy) AddDefaultResponseHeader(http.Header) {}

// Serve blocks and runs the configured http servers.
func (d dummy) Serve(context.Context) error { return nil }

// Endpoint provides a server end point for a rest application. The given handler is called on each invoction.
func (d dummy) Endpoint(path, form string, queryHandler func(query *Request)) {}

// GetJSON from a remote site. It gets serialized into the given interface.
func (d dummy) GetJSON(context.Context, string, interface{}) Error { return nil }

// GetRaw response from a remote site.
func (d dummy) GetRaw(context.Context, string) (*http.Response, error) {
	return &http.Response{Body: &BodyDummy{}}, nil
}

// PostRaw from the given reader to a remote site.
func (d dummy) PostRaw(context.Context, string, io.Reader) (*http.Response, error) {
	return &http.Response{Body: &BodyDummy{}}, nil
}

// NewDummy creates a new dummy REST interface.
func NewDummy() REST {
	return &dummy{}
}