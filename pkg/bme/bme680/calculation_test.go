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
	"math"
	"testing"

	"go.eqrx.net/mauzr/pkg/bme/bme680"
)

// TestTemperatureCompensation tests the BME680 temperature compensation isolated.
func TestTemperatureCompensation(t *testing.T) {
	tc := bme680.TemperatureCalibration{26511, 26148, 3}
	var reading uint32 = 496030
	var tFineTarget float64 = 114689
	var temperatureTarget float64 = 22.4

	tFine, temperature := tc.Compensate(reading)

	if math.Abs(tFine-tFineTarget) > 1.0 {
		t.Errorf("temperatureCalibration.Compensate(%v) returned tfine %v want %v",
			reading, tFine, tFineTarget)
	}

	if math.Abs(temperature-temperatureTarget) > 0.05 {
		t.Errorf("temperatureCalibration.Compensate(%v) returned temperature %v want %v", reading, temperature, temperatureTarget)
	}
}

// TestHumidityCompensation tests the BME680 humidity compensation isolated.
func TestHumidityCompensation(t *testing.T) {
	hc := bme680.HumidityCalibration{718, 1021, 0, 45, 20, 120, -100}
	var tFine float64 = 114689
	var reading uint16 = 22903
	var target float64 = 61.7

	humidity := hc.Compensate(reading, tFine)

	if math.Abs(humidity-target) > 0.11 {
		t.Errorf("humidityCalibration.Compensate(%v, %v) = %v; want %v", reading, tFine, humidity, target)
	}
}

// TestPressureCompensation tests the BME680 pressure compensation isolated.
func TestPressureCompensation(t *testing.T) {
	pc := bme680.PressureCalibration{34731, -10373, 88, 4083, 121, 30, 12, 1040, -4096, 30}
	var tFine float64 = 114689
	var reading uint32 = 421304
	var target float64 = 100688.5

	pressure := pc.Compensate(reading, tFine)

	if math.Abs(pressure-target) > 6 {
		t.Errorf("pressureCalibration.Compensate(%v, %v) = %v; want %v", reading, tFine, pressure, target)
	}
}

// TestGasCompensation tests the BME680 gas resistance compensation isolated.
func TestGasCompensation(t *testing.T) {
	gc := bme680.GasCalibration{-34, -7920, 18, 1}
	var target float64 = 243687
	var reading uint16 = 537
	var gasRange uint8 = 5

	gas := gc.Compensate(reading, gasRange)

	if math.Abs(gas-target) > 0.1 {
		t.Errorf("gasCalibration.Compensate(%v, %v) = %v; want %v", reading, gasRange, gas, target)
	}
}
