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

package sk6812

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.eqrx.net/mauzr/pkg/rest"
)

// setupHandler provides a http.Handler that sets an SK6812 chain
func setupHandler(mux *http.ServeMux, strip Strip) {
	mux.HandleFunc("/color", func(w http.ResponseWriter, r *http.Request) {
		rest.DefaultResponseHeader(w.Header())
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		setting := make([]uint8, 0)
		if err := json.NewDecoder(r.Body).Decode(&setting); err != nil {
			http.Error(w, fmt.Errorf("illegal setting data: %v", err).Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := strip.Set(ctx, setting); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}
