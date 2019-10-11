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
	"log"
	"net/http"
	"time"

	"go.eqrx.net/mauzr/pkg/rest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type setHandler struct {
	strip          Strip
	log            *log.Logger
	httpErrorCount prometheus.Counter
	setCount       prometheus.Counter
	setErrorCount  prometheus.Counter
	setTime        prometheus.Histogram
}

// ServeHTTP handles sets to the SK6812 chain
func (h setHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rest.ServerHeader(w.Header())
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	setting := make([]uint8, 0)
	if err := json.NewDecoder(r.Body).Decode(&setting); err != nil {
		h.log.Println(err)
		http.Error(w, fmt.Errorf("Illegal setting data: %v", err).Error(), http.StatusBadRequest)
		return
	}

	var err error
	defer func() {
		if err != nil {
			h.httpErrorCount.Inc()
			h.log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()
	timer := prometheus.NewTimer(h.setTime)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err = h.strip.Set(ctx, setting)
	timer.ObserveDuration()
}

// RESTHandler provides a http.Handler that sets an SK6812 chain
func RESTHandler(logger *log.Logger, strip Strip) http.Handler {
	return setHandler{
		log:            logger,
		strip:          strip,
		httpErrorCount: promauto.NewCounter(prometheus.CounterOpts{Name: "http_errors_total", Help: "Number of HTTP errors occurred"}),
		setCount:       promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_total", Help: "Number of measurements executed"}),
		setErrorCount:  promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_errors", Help: "Number of measurements failed"}),
		setTime:        promauto.NewHistogram(prometheus.HistogramOpts{Name: "measurements_execution_time", Help: "Duration of the measurement", Buckets: prometheus.LinearBuckets(0.01, 0.01, 10)}),
	}
}
