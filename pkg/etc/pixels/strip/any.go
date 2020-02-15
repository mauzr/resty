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

package strip

import (
	"reflect"

	"go.eqrx.net/mauzr/pkg/etc/pixels/color"
)

// GetCase for channel receive reflect.SelectCase.
type GetCase struct {
	input      Input
	selectCase reflect.SelectCase
}

// GetCase for channel send reflect.SelectCase.
type SetCase struct {
	output     Output
	selectCase reflect.SelectCase
}

// SetAny tries to execute any of the given SetCases and return when done so. Returns remaining cases.
func SetAny(outputs ...SetCase) (Output, []SetCase) {
	cases := make([]reflect.SelectCase, len(outputs))
	for i, output := range outputs {
		cases[i] = output.selectCase
	}
	i, _, _ := reflect.Select(cases)
	done := outputs[i]
	remaining := make([]SetCase, 0, len(outputs)-1)
	for _, output := range outputs {
		if output.output != done.output {
			remaining = append(remaining, output)
		}
	}
	return done.output, remaining
}

// GetAny tries to execute any of the given SetCases and return when done so. Returns result and remaining cases.
func GetAny(inputs ...GetCase) (Input, []color.RGBW, bool, []GetCase) {
	cases := make([]reflect.SelectCase, len(inputs))
	for i, input := range inputs {
		cases[i] = input.selectCase
	}
	i, value, ok := reflect.Select(cases)
	done := inputs[i]
	remaining := make([]GetCase, 0, len(inputs)-1)
	for _, input := range inputs {
		if input.input != done.input {
			remaining = append(remaining, input)
		}
	}
	return inputs[i].input, value.Interface().([]color.RGBW), ok, remaining
}

// GetCase returns a GetCase for the strip.
func (s strip) GetCase() GetCase {
	return GetCase{
		&s,
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(s.channel),
		},
	}
}

// SetCase returns a SetCase for the given strip with the value set.
func (s strip) SetCase(colors []color.RGBW) SetCase {
	return SetCase{
		&s,
		reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(s.channel),
			Send: reflect.ValueOf(colors),
		},
	}
}
