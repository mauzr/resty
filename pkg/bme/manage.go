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

package bme

import (
	"time"

	"go.eqrx.net/mauzr/pkg/bme/bme280"
	"go.eqrx.net/mauzr/pkg/bme/bme680"
	"go.eqrx.net/mauzr/pkg/bme/common"
)

// Measurement represents a taken measurement.
type Measurement = common.Measurement

// Chip represents a concrete model implementation.
type Chip interface {
	Measure() (Measurement, error)
	Reset() error
}

// Response to a query.
type Response struct {
	// Measurement is the resulting measurement.
	Measurement Measurement
	// Err is an error that was encountered or nil.
	Err error
}

// Request to produce a measurement.
type Request struct {
	// Response receives exactly one response and is closed afterwards.
	// This channel must have a buffer of at least one or the manager will panic.
	Response chan<- Response
	// MaxAge indicates how old the measurement may be to be considered valid for this request.
	MaxAge time.Time
}

func new(chip Chip, offsets Measurement, requests <-chan Request) {
	go func() {
		isReady := false
		var lastMeasurement *Measurement
		for {
			request, ok := <-requests
			switch {
			case !ok:
				return
			case cap(request.Response) < 1:
				close(request.Response)
				panic("received blocking channel for response")
			}

			if lastMeasurement != nil && lastMeasurement.Timestamp.After(request.MaxAge) {
				request.Response <- Response{*lastMeasurement, nil}
				close(request.Response)
				continue
			}

			if !isReady {
				if err := chip.Reset(); err != nil {
					request.Response <- Response{Measurement{}, err}
					close(request.Response)
					continue
				}
				isReady = true
			}

			if measurement, err := chip.Measure(); err == nil {
				measurement.Temperature += offsets.Temperature
				measurement.Humidity += offsets.Humidity
				measurement.GasResistance += offsets.GasResistance
				measurement.Pressure += offsets.Pressure
				lastMeasurement = &measurement
				request.Response <- Response{measurement, nil}
			} else {
				request.Response <- Response{Measurement{}, err}
				isReady = false
			}
			close(request.Response)
		}
	}()
}

// NewBME280 creates a new manager for a BME280 chip. Offset will be added to created measurements.
func NewBME280(bus string, address uint16, offsets Measurement, requests <-chan Request) {
	new(bme280.New(bus, address), offsets, requests)
}

// NewBME680 creates a new manager for a BME280 chip. Offset will be added to created measurements.
func NewBME680(bus string, address uint16, offsets Measurement, requests <-chan Request) {
	new(bme680.New(bus, address), offsets, requests)
}
