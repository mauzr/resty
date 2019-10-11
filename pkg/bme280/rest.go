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

package bme280

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.eqrx.net/mauzr/pkg/rest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type measurementHandler struct {
	log               *log.Logger
	chip              Chip
	tags              map[string]string
	httpErrorCount    prometheus.Counter
	measureCount      prometheus.Counter
	measureErrorCount prometheus.Counter
	measureTime       prometheus.Histogram
}

// ServeHTTP serves BME280 measurement requests.
func (h measurementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rest.ServerHeader(w.Header())
	timer := prometheus.NewTimer(h.measureTime)
	defer timer.ObserveDuration()

	if r.Method != http.MethodGet {
		h.httpErrorCount.Inc()
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var maxAge time.Duration
	if err := rest.CollectArguments(r.URL, []rest.Argument{rest.DurationArgument(&maxAge, "maxAge", false)}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	measureCtx, measureCtxCancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer measureCtxCancel()

	h.measureCount.Inc()
	var measurement Measurement
	if measurement, err = h.chip.Measure(measureCtx, maxAge); err == nil {
		encoder := json.NewEncoder(w)
		reply := make(map[string]interface{})
		reply["temperature"] = measurement.Temperature
		reply["pressure"] = measurement.Pressure
		reply["humidity"] = measurement.Humidity
		reply["timestamp"] = measurement.Time.Unix()
		for k, v := range h.tags {
			reply[k] = v
		}
		if err = encoder.Encode(reply); err == nil {
			return
		}
	}
}

// RESTHandler creates a http.Handler that handles BME280 measurements.
func RESTHandler(logger *log.Logger, chip Chip, tags map[string]string) http.Handler {
	return measurementHandler{logger, chip, tags,
		promauto.NewCounter(prometheus.CounterOpts{Name: "http_errors_total", Help: "Number of HTTP errors occurred"}),
		promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_total", Help: "Number of measurements executed"}),
		promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_errors", Help: "Number of measurements failed"}),
		promauto.NewHistogram(prometheus.HistogramOpts{Name: "measurements_execution_time", Help: "Duration of the measurement", Buckets: prometheus.LinearBuckets(0.01, 0.01, 10)}),
	}
}
