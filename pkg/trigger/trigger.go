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

// Package trigger controls a GPO pin that triggers something.
package trigger

import (
	"time"

	"go.eqrx.net/mauzr/pkg/errors"
	"go.eqrx.net/mauzr/pkg/gpio"
	"go.eqrx.net/mauzr/pkg/rest"
)

const (
	form = `
<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
	<form method="get">
	<input type="radio" name="trigger" value="true" checked> Toggle<br>
		<br>
		<input type="submit" value="Submit">
	</form>
</body>
</html>
`
)

// Expose a form for controlling the trigger.
func Expose(mux rest.Mux, path string, output gpio.Output) error {
	if err := errors.NewBatch(output.Open).Always(output.Close).Execute("testing trigger"); err != nil {
		return err
	}
	batch := errors.NewBatch(output.Open, output.Set(true), errors.BatchSleepAction(6*time.Second), output.Set(false)).Always(output.Close)
	mux.Endpoint(path, func(query *rest.Request) {
		if !query.HasArgs {
			query.ResponseBody = []byte(form)
			return
		}
		args := struct {
			Trigger bool `json:"trigger,string"`
		}{}

		if err := query.Args(&args); err == nil {
			query.InternalErr = batch.Execute("executing trigger")
		}
	})
	return nil
}
