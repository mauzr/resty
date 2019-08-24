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
	"strconv"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/rest/args"
)

type inputHandler struct {
	log *log.Logger
}

// ServeHTTP handles GPIO input requests
func (h inputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var identifier string
	var pull uint

	arguments := []args.Argument{
		args.String(&identifier, "identifier", false),
		args.Uint(&pull, "pull", true),
	}

	if err := args.Collect(r.URL, arguments); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pin := NewPin(identifier)
	value := false

	actions := []io.Action{pin.Export(), pin.Direction(Input), pin.Pull(pull), pin.Read(&value)}
	if err := io.Execute(actions, []io.Action{}); err != nil {
		h.log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(strconv.FormatBool(value))); err != nil {
			panic(err)
		}
	}

}

// InputHandler creates a http.Handler that handles GPIO input requests
func InputHandler() http.Handler {
	return inputHandler{log.New(os.Stderr, "", 0)}
}
