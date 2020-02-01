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
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/bme/common"
	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/i2c"
)

// calibrationInput contains variables that will be read out of the BME280 registers.
// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280-DS002.pdf for details.
type calibrationInput struct {
	T1     uint16
	T2     int16
	T3     int16
	P1     uint16
	P2     int16
	P3     int16
	P4     int16
	P5     int16
	P6     int16
	P7     int16
	P8     int16
	P9     int16
	PAD    byte
	H1     uint8
	H2     int16
	H3     uint8
	Left   int8
	Middle uint8
	Right  int8
	H6     int8
}

// Model represents the specific BME280 model.
type Model struct {
	calibrations Calibrations
}

// New creates a new BME280 mode representation.
func New() *Model {
	return &Model{}
}

// Calibrations return the calibration data from the chip.
func (m *Model) Calibrations() Calibrations {
	return m.calibrations
}

// Reset resets the BME280 behind the given address and fetches the calibration.
func (m *Model) Reset(bus string, address uint16) error {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280-DS002.pdf on how this works
	device := i2c.New(bus, address)
	var data [36]byte
	actions := []io.Action{
		device.Open(),
		device.Write([]byte{0xe0, 0xb6}),
		io.Sleep(2 * time.Millisecond),
		device.WriteRead([]byte{0x88}, data[0:26]),
		device.WriteRead([]byte{0xe1}, data[26:35]),
	}
	if err := io.Execute(actions, []io.Action{device.Close()}); err != nil {
		return fmt.Errorf("could not reset chip: %v", err)
	}

	var i calibrationInput
	if err := binary.Read(bytes.NewReader(data[:]), binary.LittleEndian, &i); err != nil {
		panic(err)
	}
	m.calibrations = Calibrations{
		HumidityCalibration{i.H1, i.H2, i.H3, int16(i.Left)<<4 | int16(i.Middle&0xf), int16(i.Right<<4) | int16((i.Middle>>4)&0xf), i.H6},
		PressureCalibration{i.P1, i.P2, i.P3, i.P4, i.P5, i.P6, i.P7, i.P8, i.P9},
		TemperatureCalibration{i.T1, i.T2, i.T3},
	}
	return nil
}

// Measure creates a measurement with the given BME280 behind the given address.
func (m *Model) Measure(bus string, address uint16) (common.Measurement, error) {
	device := i2c.New(bus, address)
	var reading [8]byte
	actions := []io.Action{
		device.Open(),
		device.Write([]byte{0xf4, 0x3f}),
		device.Write([]byte{0xf2, 0x01}),
		device.Write([]byte{0xf4, 0x25}),
		device.WriteRead([]byte{0xf7}, reading[:]),
	}
	if err := io.Execute(actions, []io.Action{device.Close()}); err != nil {
		return common.Measurement{}, fmt.Errorf("could not read measurement from chip: %v", err)
	}

	pReading := (uint32(reading[0])<<16 | uint32(reading[1])<<8 | uint32(reading[2])) >> 4
	tReading := (uint32(reading[3])<<16 | uint32(reading[4])<<8 | uint32(reading[5])) >> 4
	hReading := uint32(reading[6])<<8 | uint32(reading[7])

	humidity, pressure, temperature := m.calibrations.Compensate(hReading, pReading, tReading)
	measurement := common.Measurement{
		Humidity:    humidity,
		Pressure:    pressure,
		Temperature: temperature,
		Timestamp:   time.Now(),
	}
	return measurement, nil
}
