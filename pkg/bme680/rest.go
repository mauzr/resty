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

package bme680

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type measurementReply struct {
	Measurement
	tags map[string]string
}

type measurementHandler struct {
	log               *log.Logger
	bus               string
	device            uint16
	calibrations      chan Calibrations
	tags              map[string]string
	httpErrorCount    prometheus.Counter
	measureCount      prometheus.Counter
	measureErrorCount prometheus.Counter
	measureTime       prometheus.Histogram
}

// ServeHTTP serves BME680 measurement requests.
func (h measurementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	timer := prometheus.NewTimer(h.measureTime)
	defer timer.ObserveDuration()

	if r.Method != "GET" {
		h.httpErrorCount.Inc()
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
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

	h.measureCount.Inc()
	calibrations := <-h.calibrations
	var measurement Measurement
	if measurement, err = Measure(h.bus, h.device, calibrations); err == nil {
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(measurementReply{Measurement: measurement, tags: h.tags}); err != nil {
			return
		}
	} else {
		h.measureErrorCount.Inc()
	}

	h.httpErrorCount.Inc()
	h.log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func manageChip(ctx context.Context, logger *log.Logger, bus string, device uint16, calibrationsChannel chan Calibrations) {
	resetTimer, retryTimer := time.NewTimer(0), time.NewTimer(0)
	var calibrations Calibrations
	var err error
	for {
		for {
			calibrations, err = Reset(bus, device)
			if err == nil {
				break
			}
			logger.Println(err)
			retryTimer.Reset(10 * time.Second)

			select {
			case <-ctx.Done():
				return
			case <-retryTimer.C:
			}
		}

		resetTimer.Reset(time.Hour)
		for {
			select {
			case <-ctx.Done():
				return
			case calibrationsChannel <- calibrations:
				continue
			case <-resetTimer.C:
			}
			resetTimer.Reset(time.Hour)
			break
		}
	}
}

// RESTHandler creates a http.Handler that handles BME680 measurements.
func RESTHandler(ctx context.Context, bus string, device uint16, tags map[string]string) http.Handler {
	calibrations := make(chan Calibrations)
	logger := log.New(os.Stderr, "", 0)
	go manageChip(ctx, logger, bus, device, calibrations)

	return measurementHandler{log.New(os.Stderr, "", 0), bus, device, calibrations, tags,
		promauto.NewCounter(prometheus.CounterOpts{Name: "http_errors_total", Help: "Number of HTTP errors occurred"}),
		promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_total", Help: "Number of measurements executed"}),
		promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_errors", Help: "Number of measurements failed"}),
		promauto.NewHistogram(prometheus.HistogramOpts{Name: "measurements_execution_time", Help: "Duration of the measurement", Buckets: prometheus.LinearBuckets(0.01, 0.01, 10)}),
	}
}
