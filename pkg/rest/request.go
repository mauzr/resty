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
	"encoding/json"
	"fmt"
	"net/url"
)

// Request contains information from a request and how to handle it.
type Request struct {
	Ctx                                 context.Context
	URL                                 url.URL
	RequestBody                         []byte
	ResponseBody                        []byte
	Status                              int
	RequestErr, InternalErr, GatewayErr error
	HasArgs                             bool
}

// Args are parsed from the url into the given struct.
func (r *Request) Args(target interface{}) error {
	args := r.URL.Query()
	buffer := map[string]interface{}{}
	for key, value := range args {
		if len(value) > 1 {
			buffer[key] = value
		} else {
			buffer[key] = value[0]
		}
	}
	data, err := json.Marshal(buffer)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		r.RequestErr = fmt.Errorf("%w: %s", ErrRequest, err)
	}
	return r.RequestErr
}
