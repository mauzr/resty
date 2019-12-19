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
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/bme/common"
	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/i2c"
)

// calibrationInput contains variables that will be read out of the BME680 registers.
// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME680-DS001.pdf for details.
type calibrationInput struct {
	PAD0   uint8
	T2     int16
	T3     int8
	PAD1   uint8
	P1     uint16
	P2     int16
	P3     int8
	PAD2   uint8
	P4     int16
	P5     int16
	P7     int8
	P6     int8
	PAD3   uint16
	P8     int16
	P9     int16
	P10    uint8
	PAD    uint8
	H2     uint8
	MIDDLE uint8
	H1     uint8
	H3     int8
	H4     int8
	H5     int8
	H6     uint8
	H7     int8
	T1     uint16
	G2     int16
	G1     int8
	G3     int8
}

type Model struct {
	calibrations Calibrations
}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) Calibrations() Calibrations {
	return m.calibrations
}

// Reset resets the BME680 behind the given address and fetches the calibration.
func (m *Model) Reset(bus string, address uint16) error {
	device := i2c.NewDevice(bus, address)
	var data [42]byte
	var swError [1]byte
	actions := []io.Action{
		device.Open(),
		device.Write([]byte{0xe0, 0xb6}),
		io.Sleep(2 * time.Millisecond),
		device.WriteRead([]byte{0x89}, data[0:25]),
		device.WriteRead([]byte{0xe1}, data[25:41]),
		device.WriteRead([]byte{0x04}, swError[:]),
	}
	if err := io.Execute(actions, []io.Action{device.Close()}); err != nil {
		return fmt.Errorf("could not reset chip: %v", err)
	}

	var input calibrationInput
	if err := binary.Read(bytes.NewReader(data[:]), binary.LittleEndian, &input); err != nil {
		panic(err)
	}
	m.calibrations = Calibrations{
		GasCalibration{input.G1, input.G2, input.G3, swError[0] / 16},
		HumidityCalibration{uint16(input.H1)<<4 | (uint16(input.MIDDLE) & 0xf), uint16(input.H2)<<4 | uint16(input.MIDDLE)>>4, input.H3, input.H4, input.H5, input.H6, input.H7},
		PressureCalibration{input.P1, input.P2, input.P3, input.P4, input.P5, input.P6, input.P7, input.P8, input.P9, input.P10},
		TemperatureCalibration{input.T1, input.T2, input.T3},
	}
	return nil
}

// Measure creates a measurement with the given BME680 behind the given address.
func (m *Model) Measure(bus string, address uint16) (common.Measurement, error) {
	device := i2c.NewDevice(bus, address)
	var reading [15]byte
	actions := []io.Action{
		device.Write([]byte{0xe0, 0xb6}),
		io.Sleep(2 * time.Millisecond),
		device.Write([]byte{0x74, 0x01}),
		io.Sleep(500 * time.Millisecond),
		device.WriteRead([]byte{0x1d}, reading[:]),
		func() error {
			if reading[0]&0x80 == 0x00 {
				return fmt.Errorf("sensor was not ready on readout")
			}
			return nil
		},
	}
	if err := io.Execute(actions, []io.Action{device.Close()}); err != nil {
		return common.Measurement{}, fmt.Errorf("could not reset chip: %v", err)
	}

	pReading := uint32(reading[2])<<12 | uint32(reading[3])<<4 | uint32(reading[4])>>16
	tReading := uint32(reading[5])<<12 | uint32(reading[6])<<4 | uint32(reading[7])>>16
	hReading := uint16(reading[8])<<8 | uint16(reading[9])
	gresReading := uint16(reading[13])<<2 | uint16(reading[14])>>6
	granReading := reading[14] & 0x0f

	gasResistance, humidity, pressure, temperature := m.calibrations.Compensate(gresReading, granReading, hReading, pReading, tReading)
	measurement := common.Measurement{
		GasResistance: gasResistance,
		Humidity:      humidity,
		Pressure:      pressure,
		Temperature:   temperature,
		Timestamp:     time.Now(),
	}
	return measurement, nil
}
