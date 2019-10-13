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
	"net/http"
	"net/url"
	"time"
)

type Query struct {
	QueryError, InternalError, GatewayError error
	Status                                  int
	Body                                    []byte
	Ctx                                     context.Context
	URL                                     url.URL
}

func (r *Query) CollectArguments(arguments []Argument) error {
	r.QueryError = CollectArguments(r.URL, arguments)
	return r.QueryError
}

func Endpoint(mux *http.ServeMux, path, form string, queryHandler func(query *Query)) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		DefaultResponseHeader(w.Header())
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if len(r.URL.Query()) == 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(form))
		} else {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			response := Query{nil, nil, nil, http.StatusOK, nil, ctx, *r.URL}
			queryHandler(&response)
			switch {
			case response.QueryError != nil:
				http.Error(w, response.QueryError.Error(), http.StatusBadRequest)
			case response.GatewayError != nil:
				http.Error(w, response.GatewayError.Error(), http.StatusBadGateway)
			case response.InternalError != nil:
				http.Error(w, response.InternalError.Error(), http.StatusInternalServerError)
			case response.Body != nil:
				w.WriteHeader(response.Status)
				_, _ = w.Write(response.Body)
			default:
				http.Redirect(w, r, r.URL.Scheme+r.URL.Host+r.URL.Path, http.StatusSeeOther)
			}
		}
	})
}
