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

package bme680_test

import (
	"fmt"
	"math"
	"testing"

	"go.eqrx.net/mauzr/pkg/bme/bme680"
	"go.eqrx.net/mauzr/pkg/bme/common"
	"go.eqrx.net/mauzr/pkg/i2c"
)

//MeasurementMock fakes an BME680 device behind an I2C bus.
type MeasurementMock []byte

var (
	// measureMock contains a memory dump of a BME680 chip that is used for testing.
	measureMock = MeasurementMock{
		0x2d, 0xaa, 0x16, 0x4b, 0x13, 0x02, 0x54, 0x99, 0x00, 0x00, 0x01, 0x00, 0x02, 0x04, 0x02, 0xc8,
		0x10, 0x00, 0x40, 0x00, 0x80, 0x00, 0x20, 0x00, 0x1f, 0x7f, 0x1f, 0x10, 0x00, 0x00, 0x00, 0x66,
		0xa4, 0x40, 0x7b, 0x7b, 0xa0, 0x59, 0xdf, 0x80, 0x00, 0x00, 0xff, 0xe1, 0x00, 0x04, 0x00, 0x00,
		0x80, 0x00, 0x00, 0x80, 0x00, 0x00, 0x80, 0x00, 0x80, 0x00, 0x00, 0x00, 0x04, 0x00, 0x04, 0x00,
		0x00, 0x80, 0x00, 0x00, 0x80, 0x00, 0x00, 0x80, 0x00, 0x80, 0x00, 0x00, 0x00, 0x04, 0x00, 0x04,
		0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x73, 0x64, 0x65, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x10, 0x02, 0x04, 0x8c, 0x08, 0x00, 0x00, 0x0f, 0x04, 0xfe, 0x16, 0x9b, 0x08, 0x10, 0x00,
		0xa4, 0x6e, 0x89, 0x4b, 0x91, 0x4f, 0x09, 0x06, 0xb3, 0x00, 0x24, 0x66, 0x03, 0x0f, 0xab, 0x87,
		0x7b, 0xd7, 0x58, 0xff, 0xf3, 0x0f, 0x79, 0x00, 0x0c, 0x1e, 0x00, 0x00, 0xf6, 0x03, 0x00, 0xf0,
		0x1e, 0x01, 0x8c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0xc0,
		0x00, 0x54, 0x00, 0x00, 0x00, 0x00, 0x60, 0x02, 0x00, 0x01, 0x00, 0xc8, 0x1f, 0x60, 0x03, 0x00,
		0x04, 0x00, 0x8c, 0xff, 0x0f, 0x00, 0x00, 0x00, 0x02, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x10, 0x40, 0x00,
		0x00, 0x3f, 0xed, 0x2c, 0x00, 0x2d, 0x14, 0x78, 0x9c, 0x8f, 0x67, 0x10, 0xe1, 0xde, 0x12, 0xc8,
		0x00, 0x00, 0x02, 0x04, 0x8c, 0x08, 0x00, 0x66, 0xa4, 0x40, 0x7b, 0x7b, 0xa0, 0x59, 0xdf, 0x80,
	}
	// calibrationResult is the calibration data expected to be generated from the memory dump.
	calibrationResult = bme680.Calibrations{
		bme680.GasCalibration{-34, -7920, 18, 1, 1, 45},
		bme680.HumidityCalibration{717, 1022, 0, 45, 20, 120, -100},
		bme680.PressureCalibration{34731, -10373, 88, 4083, 121, 30, 12, 1014, -4096, 30},
		bme680.TemperatureCalibration{26511, 26148, 3},
	}
	// measurementResult is the measurment expected to be generated from the memory dump.
	measurementResult = common.Measurement{GasResistance: 2898707, Humidity: 63, Pressure: 101304.8, Temperature: 25.4}
)

func (m MeasurementMock) Write(data ...byte) func() error { return func() error { return nil } }
func (m MeasurementMock) Open() error                     { return nil }
func (m MeasurementMock) Close() error                    { return nil }

// WriteRead returns data from the given array.
func (m MeasurementMock) WriteRead(source []byte, destination []byte) func() error {
	return func() error {
		if len(source) != 1 {
			panic(fmt.Sprintf("Expected i2c.RegisterAddress to have length 1, was %v", len(source)))
		}
		copy(destination, m[source[0]:])
		return nil
	}
}

// TestCalibrationReadout tests if the driver reads BME680 calibration data correctly.
func TestCalibrationReadout(test *testing.T) {
	bus := ""
	address := uint16(0)
	i2c.New = func(bus string, address uint16) i2c.Device { return measureMock }
	model := bme680.New(bus, address)
	if err := model.Reset(); err == nil {
		cal := model.Calibrations()
		if cal != calibrationResult {
			test.Errorf("Reset(\"\", 0) provides calibration %v, expected %v", cal, calibrationResult)
		}
	} else {
		test.Errorf("Reset(\"\", 0) returned error %v, expected <nil>", err)
	}
}

// setupMeasurementTesting creates a fake measurement.
func setupMeasurementTesting() common.Measurement {
	bus := ""
	address := uint16(0)
	i2c.New = func(bus string, address uint16) i2c.Device {
		measureMock[0x1d] |= 0x80
		return measureMock
	}
	model := bme680.New(bus, address)

	err := model.Reset()
	if err != nil {
		panic(err)
	}

	m, err := model.Measure()
	if err != nil {
		panic(err)
	}
	return m
}

// TestGasResistanceMeasure tests if the BME680 driver handles the gas resistance readout correctly.
func TestGasResistanceMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Humidity-measurementResult.Humidity) > 0.5 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned gas resistance %v, expected %v", m.GasResistance, measurementResult.GasResistance)
	}
}

// TestHumidityMeasure tests if the BME680 driver handles the humidity readout correctly.
func TestHumidityMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Humidity-measurementResult.Humidity) > 0.5 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned humidity %v, expected %v", m.Humidity, measurementResult.Humidity)
	}
}

// TestPressureMeasure tests if the BME680 driver handles the pressure readout correctly.
func TestPressureMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Pressure-measurementResult.Pressure) > 0.5 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned pressure %v, expected %v", m.Pressure, measurementResult.Pressure)
	}
}

// TestTemperatureMeasure tests if the BME680 driver handles the temperature readout correctly.
func TestTemperatureMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Temperature-measurementResult.Temperature) > 0.5 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned temperature %v, expected %v", m.Temperature, measurementResult.Temperature)
	}
}
