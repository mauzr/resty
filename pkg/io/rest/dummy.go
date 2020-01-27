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

type dummy struct {
}

func (d dummy) GetRaw(context.Context, string) (*http.Response, error)             { return nil, nil }
func (d dummy) PostRaw(context.Context, string, io.Reader) (*http.Response, error) { return nil, nil }
func (d dummy) GetJSON(context.Context, string, interface{}) Error                 { return nil }
func (d dummy) Endpoint(path, form string, queryHandler func(query *Request))      {}
func (d dummy) Serve(context.Context) error                                        { return nil }
func (d dummy) AddDefaultResponseHeader(http.Header)                               {}
func (d dummy) ServerNames() []string                                              { return nil }

func NewDummy() REST {
	return &dummy{}
}
