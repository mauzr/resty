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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Request contains information from a request and how to handle it.
type Request struct {
	Ctx                                       context.Context
	URL                                       url.URL
	RequestBody                               []byte
	ResponseBody                              []byte
	Status                                    int
	RequestError, InternalError, GatewayError error
	Redirect                                  string
}

var ErrRequest = errors.New("invalid request")

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
		r.RequestError = fmt.Errorf("%w: %s", ErrRequest, err)
	}
	return r.RequestError
}

// AddDefaultResponseHeader to the given header.
func (r *rest) AddDefaultResponseHeader(header http.Header) {
	header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}

// Endpoint provides a server end point for a rest application. The given handler is called on each invoction.
func (r *rest) Endpoint(path, form string, queryHandler func(query *Request)) {
	r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		r.AddDefaultResponseHeader(w.Header())
		switch {
		case req.Method == http.MethodGet && len(req.URL.Query()) == 0 && form != "":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(form))
			if err != nil {
				panic(err)
			}
		default:
			requestBody, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			response := Request{req.Context(), *req.URL, requestBody, nil, http.StatusOK, nil, nil, nil, req.URL.Scheme + req.URL.Host + req.URL.Path}
			queryHandler(&response)
			switch {
			case response.RequestError != nil:
				http.Error(w, response.RequestError.Error(), http.StatusBadRequest)
			case response.GatewayError != nil:
				http.Error(w, response.GatewayError.Error(), http.StatusBadGateway)
			case response.InternalError != nil:
				http.Error(w, response.InternalError.Error(), http.StatusInternalServerError)
			case req.Method != http.MethodGet && response.ResponseBody != nil:
				panic("response body only allowed for get method")
			case response.ResponseBody != nil:
				w.WriteHeader(response.Status)
				_, err := w.Write(response.ResponseBody)
				if err != nil {
					panic(err)
				}
			default:
				http.Redirect(w, req, response.Redirect, http.StatusSeeOther)
			}
		}
	})
}
