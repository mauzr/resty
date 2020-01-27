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

package bme

import (
	"context"
	"encoding/json"
	"time"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

// setupHandler creates a http.Handler that handles BME680 measurements.
func setupHandler(c rest.REST, manager Manager, tags map[string]string) {
	c.Endpoint("/measurement", "", func(query *rest.Request) {
		args := struct {
			MaxAge time.Duration `json:"maxAge,string"`
		}{}
		if err := query.Args(&args); err != nil {
			return
		}
		measureCtx, measureCtxCancel := context.WithTimeout(query.Ctx, 3*time.Second)
		defer measureCtxCancel()
		if measurement, err := manager.Measure(measureCtx, args.MaxAge); err != nil {
			query.InternalError = err
		} else {
			measurement.Tags = tags
			query.ResponseBody, query.InternalError = json.Marshal(measurement)
		}
	})
}
