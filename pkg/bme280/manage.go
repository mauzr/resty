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
	"fmt"
	"sync"
	"time"
)

// Manager manages all functions of a chip
type manager struct {
	bus                     string
	device                  uint16
	calibrations            Calibrations
	measurement             Measurement
	latestMeasurement       chan Measurement
	requestedMeasurementAge chan time.Duration
}

// Chip represents a BME680.
type Chip interface {
	Manage(ctx context.Context, wg *sync.WaitGroup)
	Measure(ctx context.Context, maxAge time.Duration) (Measurement, error)
}

// NewChip creates a new BME680 representation.
func NewChip(bus string, device uint16) Chip {
	return &manager{bus, device, Calibrations{}, Measurement{}, make(chan Measurement), make(chan time.Duration)}
}

// Measure the current air state.
func (m *manager) Measure(ctx context.Context, maxAge time.Duration) (Measurement, error) {
	for {
		select {
		case measurement, more := <-m.latestMeasurement:
			switch {
			case !more:
				return Measurement{}, fmt.Errorf("management routine canceled")
			case time.Since(measurement.Time) < maxAge:
				return measurement, nil
			default:
				select {
				case m.requestedMeasurementAge <- maxAge:
				case <-ctx.Done():
					return Measurement{}, ctx.Err()
				}
			}
		case <-ctx.Done():
			return Measurement{}, ctx.Err()
		}
	}
}

func (m *manager) reset(ctx context.Context) {
	for {
		if calibrations, err := Reset(m.bus, m.device); err == nil {
			m.calibrations = calibrations
			break
		} else {
			fmt.Printf("Reset failed: %v\n", err)
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

func (m *manager) run(ctx context.Context) {
	var measurement Measurement

	for {
		select {
		case maxAge := <-m.requestedMeasurementAge:
			if time.Since(measurement.Time) >= maxAge {
				if newMeasurment, err := Measure(m.bus, m.device, m.calibrations); err == nil {
					m.measurement = newMeasurment
				} else {
					fmt.Printf("Measurement failed: %v\n", err)
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

// Manage the chip.
func (m *manager) Manage(ctx context.Context, wg *sync.WaitGroup) {
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
