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

package rest

import (
	"encoding/json"
)

func UnmarshalArguments(args map[string][]string, destination interface{}) error {
	buffer := map[string]interface{}{}
	for key, value := range args {
		if len(value) > 1 {
			buffer[key] = value
		} else {
			buffer[key] = value[0]
		}
	}
	data, err := json.Marshal(buffer)
	if err != nil {
		panic(err)
	}
	return json.Unmarshal(data, destination)
}
