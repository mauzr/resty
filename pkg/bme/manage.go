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
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/bme/bme280"
	"go.eqrx.net/mauzr/pkg/bme/bme680"
	"go.eqrx.net/mauzr/pkg/bme/common"
)

// Measurments represents a taken measurement.
type Measurement = common.Measurement

type Chip interface {
	Measure() (Measurement, error)
	Reset() error
}

type Response struct {
	Measurement Measurement
	Err         error
}

type Request struct {
	Response chan<- Response
	MaxAge   time.Time
}

func New(chip Chip, offsets Measurement, requests <-chan Request) {
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
				panic(fmt.Errorf("received blocking channel for response"))
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

func NewBME280(bus string, address uint16, offsets Measurement, requests <-chan Request) {
	New(bme280.New(bus, address), offsets, requests)
}

func NewBME680(bus string, address uint16, offsets Measurement, requests <-chan Request) {
	New(bme680.New(bus, address), offsets, requests)
}
