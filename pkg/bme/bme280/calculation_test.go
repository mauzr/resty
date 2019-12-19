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
	"math"
	"testing"

	"go.eqrx.net/mauzr/pkg/bme/bme280"
)

// TestTemperatureCompensation tests the BME280 temperature compensation isolated.
func TestTemperatureCompensation(t *testing.T) {
	tc := bme280.TemperatureCalibration{27603, 25947, 50}
	var reading uint32 = 508272
	var tFineTarget float64 = 105523
	var temperatureTarget float64 = 20.6

	tFine, temperature := tc.Compensate(reading)

	if math.Abs(tFine-tFineTarget) > 1.0 {
		t.Errorf("temperatureCalibration.Compensate(%v) returned tfine %v want %v", reading, tFine, tFineTarget)
	}

	if math.Abs(temperature-temperatureTarget) > 0.05 {
		t.Errorf("temperatureCalibration.Compensate(%v) returned temperature %v want %v", reading, temperature, temperatureTarget)
	}
}

// TestHumidityCompensation tests the BME280 humidity compensation isolated.
func TestHumidityCompensation(t *testing.T) {
	hc := bme280.HumidityCalibration{75, 354, 0, 339, 0, 30}
	var tFine float64 = 105752
	var reading uint32 = 30405
	var target float64 = 47.3

	humidity := hc.Compensate(reading, tFine)

	if math.Abs(humidity-target) > 0.05 {
		t.Errorf("humidityCalibration.Compensate(%v, %v) = %v; want %v", reading, tFine, humidity, target)
	}
}

// TestPressureCompensation tests the BME280 pressure compensation isolated.
func TestPressureCompensation(t *testing.T) {
	pc := bme280.PressureCalibration{36343, -10930, 3024, 7386, 103, -7, 990, -10230, 4285}
	var tFine float64 = 105980
	var reading uint32 = 337382
	var target float64 = 100658.0

	pressure := pc.Compensate(reading, tFine)

	if math.Abs(pressure-target) > 0.5 {
		t.Errorf("pressureCalibration.Compensate(%v, %v) = %v; want %v", reading, tFine, pressure, target)
	}
}
