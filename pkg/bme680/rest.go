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
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.eqrx.net/mauzr/pkg/rest"
)

// setupHandler creates a http.Handler that handles BME680 measurements.
func setupHandler(mux *http.ServeMux, chip Chip, tags map[string]string) {
	rest.Endpoint(mux, "/measurement", "", func(query *rest.Query) {
		args := struct {
			MaxAge time.Duration `json:"maxAge,string"`
		}{}
		if err := query.UnmarshalArguments(&args); err != nil {
			return
		}
		measureCtx, measureCtxCancel := context.WithTimeout(query.Ctx, 3*time.Second)
		defer measureCtxCancel()
		if measurement, err := chip.Measure(measureCtx, args.MaxAge); err != nil {
			query.InternalError = err
		} else {
			reply := map[string]interface{}{
				"temperature":    measurement.Temperature,
				"pressure":       measurement.Pressure,
				"humidity":       measurement.Humidity,
				"gas_resistance": measurement.GasResistance,
				"timestamp":      measurement.Time.Unix(),
			}
			for k, v := range tags {
				reply[k] = v
			}
			query.Body, query.InternalError = json.Marshal(reply)
		}
	})
}
