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

package bme280_test

import (
	"fmt"
	"math"
	"testing"

	"go.eqrx.net/mauzr/pkg/bme280"
	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/i2c"
)

// MeasurementMock fakes an BME280 device behind an I2C bus.
type MeasurementMock []byte

var calibrationResult = bme280.Calibrations{
	bme280.HumidityCalibration{75, 354, 0, 339, 0, 30},
	bme280.PressureCalibration{36343, -10930, 3024, 7386, 103, -7, 9900, -10230, 4285},
	bme280.TemperatureCalibration{27603, 25947, 50},
}
var measurementResult = bme280.Measurement{Humidity: 60.5, Pressure: 100651.0, Temperature: 21.9}
var measureMock = MeasurementMock{
	0x8e, 0x6f, 0x89, 0x4f, 0xab, 0x52, 0xc9, 0x06, 0xd3, 0x6b, 0x5b, 0x65, 0x32, 0x00, 0xf7, 0x8d,
	0x4e, 0xd5, 0xd0, 0x0b, 0xda, 0x1c, 0x67, 0x00, 0xf9, 0xff, 0xac, 0x26, 0x0a, 0xd8, 0xbd, 0x10,
	0x00, 0x4b, 0x5b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0xc0,
	0x00, 0x54, 0x00, 0x00, 0x00, 0x00, 0x60, 0x02, 0x00, 0x01, 0xff, 0xff, 0x1f, 0x60, 0x03, 0x00,
	0x00, 0x00, 0x34, 0xff, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x62, 0x01, 0x00, 0x15, 0x03, 0x00, 0x1e, 0xd6, 0x41, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0x00, 0x01, 0x04, 0x34, 0x00, 0x00, 0x53, 0x76, 0x80, 0x7d, 0x2d, 0x00, 0x80, 0x24, 0x80,
}

func (m MeasurementMock) Write(data []byte) io.Action { return io.NoOperation }
func (m MeasurementMock) Open() io.Action             { return io.NoOperation }
func (m MeasurementMock) Close() io.Action            { return io.NoOperation }

// WriteRead returns data from the given array.
func (m MeasurementMock) WriteRead(source []byte, destination []byte) io.Action {
	return func() error {
		if len(source) != 1 {
			panic(fmt.Sprintf("Expected i2c.RegisterAddress to have length 1, was %v", len(source)))
		}
		copy(destination, m[source[0]+16*8:])
		return nil
	}
}

// TestCalibrationReadout tests if the driver reads BME280 calibration data correctly.
func TestCalibrationReadout(test *testing.T) {
	i2c.NewDevice = func(bus string, address uint16) i2c.Device { return measureMock }

	if cal, err := bme280.Reset("", 0); err == nil {
		if cal != calibrationResult {
			test.Errorf("Reset(\"\", 0) provides calibration %v, expected %v", cal, calibrationResult)
		}
	} else {
		test.Errorf("Reset(\"\", 0) returned error %v, expected <nil>", err)
	}
}

func setupMeasurementTesting() bme280.Measurement {
	i2c.NewDevice = func(bus string, address uint16) i2c.Device { return measureMock }

	if cal, err := bme280.Reset("", 0); err == nil {
		var m bme280.Measurement
		if m, err = bme280.Measure("", 0, cal); nil == err {
			return m
		}
		panic(err)
	} else {
		panic(err)
	}
}

// TestHumidityMeasure tests if the BME280 driver handles the humidity readout correctly.
func TestHumidityMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Humidity-measurementResult.Humidity) > 0.5 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned humidity %v, expected %v", m.Humidity, measurementResult.Humidity)
	}
}

// TestPressureMeasure tests if the BME280 driver handles the pressure readout correctly.
func TestPressureMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Pressure-measurementResult.Pressure) > 0.5 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned pressure %v, expected %v", m.Pressure, measurementResult.Pressure)
	}
}

// TestTemperatureMeasure tests if the BME280 driver handles the temperature readout correctly.
func TestTemperatureMeasure(test *testing.T) {
	m := setupMeasurementTesting()

	if math.Abs(m.Temperature-measurementResult.Temperature) > 0.09 {
		test.Errorf("AttachAndMeasure(\"\", 0) returned temperature %v, expected %v", m.Temperature, measurementResult.Temperature)
	}
}
