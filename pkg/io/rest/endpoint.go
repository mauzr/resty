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
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	RequestError, InternalError, GatewayError error
	Status                                    int
	ResponseBody                              []byte
	Ctx                                       context.Context
	URL                                       url.URL
}

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
	r.RequestError = json.Unmarshal(data, target)
	return r.RequestError
}

func (r *rest) AddDefaultResponseHeader(header http.Header) {
	header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}

func (r *rest) HandleFunc(path string, handler func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(path, handler)
}

func (r *rest) Endpoint(path, form string, queryHandler func(query *Request)) {
	r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		r.AddDefaultResponseHeader(w.Header())
		switch {
		case req.Method != http.MethodGet:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		case len(req.URL.Query()) == 0:
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(form))
			if err != nil {
				panic(err)
			}
		default:
			ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
			defer cancel()
			response := Request{nil, nil, nil, http.StatusOK, nil, ctx, *req.URL}
			queryHandler(&response)
			switch {
			case response.RequestError != nil:
				http.Error(w, response.RequestError.Error(), http.StatusBadRequest)
			case response.GatewayError != nil:
				http.Error(w, response.GatewayError.Error(), http.StatusBadGateway)
			case response.InternalError != nil:
				http.Error(w, response.InternalError.Error(), http.StatusInternalServerError)
			case response.ResponseBody != nil:
				w.WriteHeader(response.Status)
				_, err := w.Write(response.ResponseBody)
				if err != nil {
					panic(err)
				}
			default:
				http.Redirect(w, req, req.URL.Scheme+req.URL.Host+req.URL.Path, http.StatusSeeOther)
			}
		}
	})
}
