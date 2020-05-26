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

package main

import (
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/bme"
)

func main() {
	dut := make(chan bme.Request)
	com := make(chan bme.Request)
	bme.NewBME680("/dev/i2c-1", 0x76, bme.Measurement{}, map[string]string{}, dut)
	bme.NewBME680("/dev/i2c-1", 0x77, bme.Measurement{}, map[string]string{}, com)
	ticker := time.NewTicker(30 * time.Second)
	for {
		dutResponses := make(chan bme.Response, 1)
		comResponses := make(chan bme.Response, 1)
		maxAge := time.Now().Add(-3 * time.Second)
		dutRequest := bme.Request{Response: dutResponses, MaxAge: maxAge}
		comRequest := bme.Request{Response: comResponses, MaxAge: maxAge}
		dut <- dutRequest
		com <- comRequest

		dutResponse := <-dutResponses
		if dutResponse.Err != nil {
			panic(dutResponse.Err)
		}
		comResponse := <-comResponses
		if comResponse.Err != nil {
			panic(dutResponse.Err)
		}

		comMeasurement := comResponse.Measurement
		dutMeasurement := dutResponse.Measurement

		fmt.Printf("gas_resistance: %v, humidity: %v, pressure: %v, temperature: %v\n",
			comMeasurement.GasResistance-dutMeasurement.GasResistance,
			comMeasurement.Humidity-dutMeasurement.Humidity,
			comMeasurement.Pressure-dutMeasurement.Pressure,
			comMeasurement.Temperature-dutMeasurement.Temperature,
		)

		<-ticker.C
	}
}
