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

package gpio

import (
	"log"
	"net/http"
	"os"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/rest"
)

type outputHandler struct {
	log *log.Logger
}

// ServeHTTP handles GPIO output requests
func (h outputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rest.ServerHeader(w.Header())
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var identifier string
	var value bool

	arguments := []rest.Argument{
		rest.StringArgument(&identifier, "identifier", false),
		rest.BoolArgument(&value, "value", false),
	}

	if err := rest.CollectArguments(r.URL, arguments); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pin := NewPin(identifier)
	actions := []io.Action{pin.Export(), pin.Direction(Output), pin.Write(value)}
	if err := io.Execute(actions, []io.Action{}); err != nil {
		h.log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte("Set")); err == nil {
			panic(err)
		}
	}
}

// OutputHandler creates a http.Handler that handles GPIO output requests
func OutputHandler() http.Handler {
	return outputHandler{log.New(os.Stderr, "", 0)}
}
