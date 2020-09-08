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
	"math"
)

const (
	minimumTemperature float64 = -40
	maximumTemperature float64 = 85
	minimumPressure    float64 = 30000
	maximumPressure    float64 = 110000
	minimumHumidity    float64 = 0
	maximumHumidity    float64 = 100
)

// TemperatureCalibration is used to compensate temperature readings of the BME280.
type TemperatureCalibration struct {
	T1 uint16
	T2 int16
	T3 int16
}

// PressureCalibration is used to compensate pressure readings of the BME280.
type PressureCalibration struct {
	P1 uint16
	P2 int16
	P3 int16
	P4 int16
	P5 int16
	P6 int16
	P7 int16
	P8 int16
	P9 int16
}

// HumidityCalibration is used to compensate humidity readings of the BME280.
type HumidityCalibration struct {
	H1 uint8
	H2 int16
	H3 uint8
	H4 int16
	H5 int16
	H6 int8
}

// Calibrations contains all required calibrations for the chip.
type Calibrations struct {
	Humidity    HumidityCalibration
	Pressure    PressureCalibration
	Temperature TemperatureCalibration
}

// Compensate compensates all readings of the BME280.
func (c Calibrations) Compensate(humidityReading uint32, pressureReading uint32, temperatureReading uint32) (humidity float64, pressure float64, temperature float64) {
	var tFine float64
	tFine, temperature = c.Temperature.Compensate(temperatureReading)
	pressure = c.Pressure.Compensate(pressureReading, tFine)
	humidity = c.Humidity.Compensate(humidityReading, tFine)

	return
}

// Compensate compensates the temperature reading of the BME280.
//nolint:gomnd // Hardware interfacing.
func (c TemperatureCalibration) Compensate(reading uint32) (tFine float64, temperature float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280-DS002.pdf for explanation of the math.

	var1 := (float64(reading)/16384.0 - float64(c.T1)/1024.0) * float64(c.T2)
	var2 := math.Pow(float64(reading)/131072.0-float64(c.T1)/8192.0, 2) * float64(c.T3)
	tFine = var1 + var2
	temperature = (var1 + var2) / 5120.0

	temperature = math.Max(minimumTemperature, math.Min(temperature, maximumTemperature))

	return
}

// Compensate compensates the pressure reading of the BME280.
//nolint:gomnd // Hardware interfacing.
func (c PressureCalibration) Compensate(reading uint32, tFine float64) (pressure float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280-DS002.pdf for explanation of the math.

	var1 := tFine/2.0 - 64000.0
	var2 := ((math.Pow(var1, 2)*float64(c.P6)/32768.0 + var1*float64(c.P5)*2.0) / 4.0) + float64(c.P4)*65536.0
	var1 = (1.0 + (float64(c.P3)*math.Pow(var1, 2)/524288.0+float64(c.P2)*var1)/524288.0/32768.0) * float64(c.P1)

	if var1 != 0 {
		pressure = (1048576.0 - float64(reading) - var2/4096.0) * 6250.0 / var1
		pressure += (float64(c.P9)*math.Pow(pressure, 2)/2147483648.0 + pressure*float64(c.P8)/32768.0 + float64(c.P7)) / 16.0
	}
	pressure = math.Max(minimumPressure, math.Min(pressure, maximumPressure))

	return
}

// Compensate compensates the humidity reading of the BME280.
//nolint:gomnd // Hardware interfacing.
func (c HumidityCalibration) Compensate(reading uint32, tFine float64) (humidity float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280-DS002.pdf for explanation of the math.

	var1 := tFine - 76800.0
	var2 := (1.0 + (float64(c.H3)/67108864.0)*var1)
	var3 := var2 * (1.0 + (float64(c.H6)/67108864.0)*var1*var2) * (float64(reading) - (float64(c.H4)*64.0 + (float64(c.H5)/16384.0)*var1)) * float64(c.H2) / 65536.0
	humidity = var3 * (1.0 - float64(c.H1)*var3/524288.0)
	humidity = math.Max(minimumHumidity, math.Min(humidity, maximumHumidity))

	return
}
