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
	"fmt"

	"go.eqrx.net/mauzr/pkg/etc/pixels/color"
)

// split a strip into multiple outputs.
func split(input Input, outputs ...Output) {
	outputLength := 0
	for _, output := range outputs {
		defer output.Close()
		outputLength += output.Length()
	}
	if outputLength != input.Length() {
		panic(fmt.Sprintf("inputs (%s) and output have different lengths: %v, %v", input.Name(), input.Length(), outputLength))
	}
	for {
		colors, ok := input.Get()
		if !ok {
			return
		}
		setCases := make([]SetCase, len(outputs))
		offset := 0
		for i, output := range outputs {
			newOffset := offset + output.Length()
			setCases[i] = output.SetCase(colors[offset:newOffset])
			offset = newOffset
		}

		for len(setCases) != 0 {
			_, setCases = SetAny(setCases...)
			// Printout for debugging
			//names := []string{}
			//for _, c := range setCases {
			//	names = append(names, c.output.Name())
			//}
			//fmt.Printf("%s->: waiting for %s\n", input.Name(), names)
		}
	}
}

// merge multiple input strips into one output.
func merge(output Output, inputs ...Input) {
	defer output.Close()
	offsets := map[string]int{}
	lastColors := make(map[string][]color.RGBW)
	getCases := make([]GetCase, len(inputs))
	inputLength := 0
	for i, input := range inputs {
		if _, ok := offsets[input.Name()]; ok {
			panic(fmt.Errorf("multiple inputs have the name :%s", input.Name()))
		}
		offsets[input.Name()] = inputLength
		inputLength += input.Length()
		getCases[i] = input.GetCase()
	}
	if inputLength != output.Length() {
		panic(fmt.Sprintf("inputs and output (%s) have different lengths: %v, %v", output.Name(), inputLength, output.Length()))
	}
	for {
		merged := make([]color.RGBW, output.Length())
		remaining := getCases
		anyOk := false
		for len(remaining) != 0 {
			input, colors, ok, newRemaining := GetAny(remaining...)
			remaining = newRemaining
			offset, offsetOk := offsets[input.Name()]
			if !offsetOk {
				panic(fmt.Errorf("no offset found for %v", input))
			}
			if ok {
				anyOk = true
				lastColors[input.Name()] = colors
			}
			for i, v := range lastColors[input.Name()] {
				merged[offset+i] = v
			}
			// Printout for debugging
			//names := []string{}
			//for _, c := range getCases {
			//	names = append(names, c.input.Name())
			//}
			//fmt.Printf("->%s: waiting for %s\n", output.Name(), names)
		}
		if !anyOk {
			return
		}
		output.Set(merged)
	}
}
