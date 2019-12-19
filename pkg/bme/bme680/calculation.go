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
	"math"
)

var (
	k1Range = []float64{0.0, 0.0, 0.0, 0.0, 0.0, -1.0, 0.0, -0.8, 0.0, 0.0, -0.2, -0.5, 0.0, -1.0, 0.0, 0.0}
	k2Range = []float64{0.0, 0.0, 0.0, 0.0, 0.1, 0.7, 0.0, -0.8, -0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
)

// TemperatureCalibration is used to compensate temperature readings of the BME680.
type TemperatureCalibration struct {
	T1 uint16
	T2 int16
	T3 int8
}

// PressureCalibration is used to compensate pressure readings of the BME680.
type PressureCalibration struct {
	P1  uint16
	P2  int16
	P3  int8
	P4  int16
	P5  int16
	P6  int8
	P7  int8
	P8  int16
	P9  int16
	P10 uint8
}

// HumidityCalibration is used to compensate humidity readings of the BME680.
type HumidityCalibration struct {
	H1 uint16
	H2 uint16
	H3 int8
	H4 int8
	H5 int8
	H6 uint8
	H7 int8
}

// GasCalibration is used to compensate gas resistance readings of the BME680.
type GasCalibration struct {
	G1      int8  // 25
	G2      int16 // 24
	G3      int8  // 26
	SWError uint8
}

// Calibrations contains all required calibrations for the chip.
type Calibrations struct {
	Gas         GasCalibration
	Humidity    HumidityCalibration
	Pressure    PressureCalibration
	Temperature TemperatureCalibration
}

// Compensate compensates all readings of the BME680.
func (c Calibrations) Compensate(gresReading uint16, granReading uint8, humidityReading uint16, pressureReading uint32, temperatureReading uint32) (gasResistance float64, humidity float64, pressure float64, temperature float64) {
	var tFine float64
	tFine, temperature = c.Temperature.Compensate(temperatureReading)
	pressure = c.Pressure.Compensate(pressureReading, tFine)
	humidity = c.Humidity.Compensate(humidityReading, tFine)
	gasResistance = c.Gas.Compensate(gresReading, granReading)
	return
}

// Compensate compensates the temperature reading of the BME680.
func (c TemperatureCalibration) Compensate(reading uint32) (tFine float64, temperature float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME680-DS001.pdf for the math.

	var1 := float64(reading)/8 - float64(c.T1)*2
	tFine = var1*float64(c.T2)/2048 + math.Pow(var1/2, 2)*float64(c.T3)*16/67108864
	temperature = (tFine*5 + 128) / 25600
	return
}

// Compensate compensates the pressure reading of the BME680.
func (c PressureCalibration) Compensate(reading uint32, tFine float64) (pressure float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME680-DS001.pdf for the math.

	var1 := tFine/2 - 64000
	var2 := math.Pow(var1/4, 2)/2048*float64(c.P6)/4 + var1*float64(c.P5)*2/4 + float64(c.P4)*65536
	var1 = (32768 + (math.Pow(var1/4, 2)/8192*float64(c.P3)*32/8+float64(c.P2)*var1/2)/262144) * float64(c.P1) / 32768
	pressure = (float64(1048576) - float64(reading) - var2/4096) / var1 * 6250
	pressure += (float64(c.P9)*math.Pow(pressure/8, 2)/33554432 + pressure/4*float64(c.P8)/8192 + math.Pow(pressure/256, 3)*float64(c.P10)/131072 + float64(c.P7)*128) / 16
	return
}

// Compensate compensates the humidity reading of the BME680.
func (c HumidityCalibration) Compensate(reading uint16, tFine float64) (humidity float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME680-DS001.pdf for the math.

	scaled := (tFine*5 + 128) / 256
	var1 := (float64(reading) - float64(c.H1)*16 - scaled*float64(c.H3)/200) * float64(c.H2) * (scaled*float64(c.H4)/100 + scaled*scaled*float64(c.H5)/640000 + 16384) / 1024
	humidity = (var1 + (float64(c.H6)*8+scaled*float64(c.H7)/1600)*math.Pow(var1/16384, 2)/2048) / 4194304
	humidity = math.Max(0, math.Min(humidity, 100))
	return
}

// Compensate compensates the gas resistance reading of the BME680.
func (c GasCalibration) Compensate(reading uint16, gasRange uint8) (gasResistance float64) {
	// See https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME680-DS001.pdf for the math.

	var1 := (1340 + 5*float64(c.SWError)) * (1 + k1Range[gasRange]/100)
	gasResistance = 1 / ((1 + k2Range[gasRange]/100) * 0.000000125 * float64(int(1<<gasRange)) * (((float64(reading) - 512) / var1) + 1))

	return
}
