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

import "time"
import "context"
import "log"
import "fmt"

type Manager struct {
	bus                     string
	device                  uint16
	calibrations            Calibrations
	measurement             Measurement
	latestMeasurement       chan Measurement
	requestedMeasurementAge chan time.Duration
}

// Chip represents a BME280.
type Chip interface {
	Manage(ctx context.Context, logger *log.Logger)
	Measure(ctx context.Context, maxAge time.Duration) (Measurement, error)
}

// NewChip creates a new BME280 representation.
func NewChip(bus string, device uint16) Chip {
	return &Manager{bus, device, Calibrations{}, Measurement{}, make(chan Measurement), make(chan time.Duration)}
}

// Measure the current air state.
func (m *Manager) Measure(ctx context.Context, maxAge time.Duration) (Measurement, error) {
	for {
		select {
		case measurement, more := <-m.latestMeasurement:
			if !more {
				return Measurement{}, fmt.Errorf("management routine canceled")
			} else if time.Since(measurement.Time) < maxAge {
				return measurement, nil
			} else {
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

func (m *Manager) reset(ctx context.Context, logger *log.Logger) {
	for {
		if calibrations, err := Reset(m.bus, m.device); err == nil {
			m.calibrations = calibrations
			break
		} else {
			logger.Printf("Reset failed: %v\n", err)
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

func (m *Manager) run(ctx context.Context, logger *log.Logger) {
	var measurement Measurement
	resetTimer := time.NewTimer(1 * time.Hour)
	for {
		select {
		case maxAge := <-m.requestedMeasurementAge:
			if time.Since(measurement.Time) >= maxAge {
				if newMeasurment, err := Measure(m.bus, m.device, m.calibrations); err == nil {
					m.measurement = newMeasurment
				} else {
					logger.Printf("Measurement failed: %v\n", err)
					return
				}
			}
		case m.latestMeasurement <- m.measurement:
			continue
		case <-ctx.Done():
			return
		case <-resetTimer.C:
			return
		}
	}
}

// Manage the chip.
func (m *Manager) Manage(ctx context.Context, logger *log.Logger) {
	defer close(m.latestMeasurement)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m.reset(ctx, logger)
			m.run(ctx, logger)
		}
	}
}
