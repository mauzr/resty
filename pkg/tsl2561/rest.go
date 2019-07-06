package tsl2561

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"mauzr.eqrx.net/go/pkg/i2c"
)

type measurementReply struct {
	Measurement
	tags map[string]string
}

type measurementHandler struct {
	tags              map[string]string
	bus               string
	device            i2c.DeviceAddress
	log               *log.Logger
	httpErrorCount    prometheus.Counter
	measureCount      prometheus.Counter
	measureErrorCount prometheus.Counter
	measureTime       prometheus.Histogram
}

// ServeHTTP handles measurment requests to an TSL2561
func (h measurementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	timer := prometheus.NewTimer(h.measureTime)
	defer timer.ObserveDuration()

	if r.Method != "GET" {
		h.httpErrorCount.Inc()
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	measurement, err := AttachAndMeasure(h.bus, h.device)
	h.measureCount.Inc()
	if err == nil {
		encoder := json.NewEncoder(w)
		err = encoder.Encode(measurementReply{Measurement: measurement, tags: h.tags})
		if err != nil {
			return
		}
	} else {
		h.measureErrorCount.Inc()
	}

	h.httpErrorCount.Inc()
	h.log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// RESTHandler returns a http.handler that handles measurment requests to an TSL2561 behind I2C
func RESTHandler(bus string, device i2c.DeviceAddress, tags map[string]string) http.Handler {
	return measurementHandler{
		log:               log.New(os.Stderr, "", 0),
		bus:               bus,
		httpErrorCount:    promauto.NewCounter(prometheus.CounterOpts{Name: "http_errors_total", Help: "Number of HTTP errors occured"}),
		measureCount:      promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_total", Help: "Number of measurements executed"}),
		measureErrorCount: promauto.NewCounter(prometheus.CounterOpts{Name: "measurements_errors", Help: "Number of measurements failed"}),
		measureTime:       promauto.NewHistogram(prometheus.HistogramOpts{Name: "measurements_execution_time", Help: "Duration of the measurement", Buckets: prometheus.LinearBuckets(0.01, 0.01, 10)}),
	}
}
