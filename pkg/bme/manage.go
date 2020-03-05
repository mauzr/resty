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
	"context"
	"fmt"
	"sync"
	"time"

	"go.eqrx.net/mauzr/pkg/bme/bme280"
	"go.eqrx.net/mauzr/pkg/bme/bme680"
	"go.eqrx.net/mauzr/pkg/bme/common"
)

// Measurments represents a taken measurement.
type Measurement = common.Measurement

// Model represents a specific BME model
type Model interface {
	Measure(bus string, device uint16) (Measurement, error)
	Reset(bus string, device uint16) error
}

// Manager manages all functions of a chip.
type Manager struct {
	model                   Model
	bus                     string
	device                  uint16
	measurement             Measurement
	latestMeasurement       chan Measurement
	requestedMeasurementAge chan time.Duration
}

// NewBME280Manager creates a manager for a BME280.
func NewBME280Manager(bus string, address uint16) Manager {
	return Manager{
		bme280.New(),
		bus,
		address,
		Measurement{},
		make(chan Measurement),
		make(chan time.Duration),
	}
}

// NewBME280Manager creates a manager for a BME680.
func NewBME680Manager(bus string, address uint16) Manager {
	return Manager{
		bme680.New(),
		bus,
		address,
		Measurement{},
		make(chan Measurement),
		make(chan time.Duration),
	}
}

// Measure return a measurement that is not older than the given maximum age.
func (m *Manager) Measure(ctx context.Context, maxAge time.Duration) (Measurement, error) {
	if maxAge == 0 {
		panic(fmt.Errorf("maxAge may not be 0"))
	}
	for {
		select {
		case measurement, ok := <-m.latestMeasurement:
			switch {
			case !ok:
				return Measurement{}, fmt.Errorf("management routine canceled")
			case time.Since(measurement.Timestamp) < maxAge:
				return measurement, nil
			default:
				select {
				case <-ctx.Done():
					return Measurement{}, ctx.Err()
				case m.requestedMeasurementAge <- maxAge:
					continue
				}
			}
		case <-ctx.Done():
			return Measurement{}, ctx.Err()
		}
	}
}

// reset resets the chip.
func (m *Manager) reset(ctx context.Context) {
	for {
		if err := m.model.Reset(m.bus, m.device); err == nil {
			break
		} else {
			fmt.Printf("reset failed: %v\n", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.NewTimer(10 * time.Second).C:
			continue
		}
	}
	select {
	case <-ctx.Done():
		return
	case <-time.NewTimer(3 * time.Second).C:
	}
}

// run serves calls to the manager.
func (m *Manager) run(ctx context.Context) {
	var measurement Measurement

	for {
		select {
		case maxAge := <-m.requestedMeasurementAge:
			if time.Since(measurement.Timestamp) >= maxAge {
				if newMeasurment, err := m.model.Measure(m.bus, m.device); err == nil {
					m.measurement = newMeasurment
				} else {
					fmt.Printf("measurement failed: %v\n", err)
					return
				}
			}
		case m.latestMeasurement <- m.measurement:
			continue
		case <-ctx.Done():
			return
		}
	}
}

// Manage blocks and manages calls to the manager.
func (m *Manager) Manage(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	defer close(m.latestMeasurement)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m.reset(ctx)
			m.run(ctx)
		}
	}
}
