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

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

// split a strip into multiple outputs.
func split(input Input, outputs ...Output) {
	outputLength := 0
	outputChannels := make([]chan []color.RGBW, len(outputs))
	for i, output := range outputs {
		outputLength += output.Length()
		outputChannels[i] = make(chan []color.RGBW)
		defer close(outputChannels[i])
		go func(v <-chan []color.RGBW, o Output) {
			defer o.Close()
			for {
				c, ok := <-v
				if !ok {
					return
				}
				co := make([]color.RGBW, len(c))
				copy(co, c)
				o.Set(co)
			}
		}(outputChannels[i], output)
	}
	if outputLength != input.Length() {
		panic(fmt.Errorf("inputs (%s) and output have different lengths: %v, %v", input.Name(), input.Length(), outputLength))
	}
	for {
		colors, ok := input.Get()
		if !ok {
			return
		}
		offset := int(0)
		for i, output := range outputs {
			newOffset := offset + output.Length()
			outputChannels[i] <- colors[offset:newOffset]
			offset = newOffset
		}
	}
}

// merge multiple input strips into one output.
func merge(output Output, inputs ...Input) {
	defer output.Close()
	offsets := make([]int, len(inputs))
	lastColors := make([][]color.RGBW, len(inputs))
	inputChannels := make([]chan []color.RGBW, len(inputs))
	inputLength := 0
	inputNameSet := make(map[string]bool)
	for i, input := range inputs {
		lastColors[i] = make([]color.RGBW, input.Length())
		for j := 0; j < input.Length(); j++ {
			lastColors[i][j] = color.Unmanaged
		}
		if inputNameSet[input.Name()] {
			panic(fmt.Errorf("multiple inputs have the name :%s", input.Name()))
		}
		inputNameSet[input.Name()] = true
		offsets[i] = inputLength
		inputLength += input.Length()
		inputChannels[i] = make(chan []color.RGBW)
		go func(v chan<- []color.RGBW, i Input) {
			defer close(v)
			for {
				c, ok := i.Get()
				if !ok {
					return
				}
				v <- c
			}
		}(inputChannels[i], input)
	}
	if inputLength != output.Length() {
		panic(fmt.Errorf("inputs and output (%s) have different lengths: %v, %v", output.Name(), inputLength, output.Length()))
	}
	for {
		anyOk := false
		for i, c := range inputChannels {
			v, ok := <-c
			if ok {
				anyOk = true
				lastColors[i] = v
			}
		}
		if !anyOk {
			return
		}
		merged := make([]color.RGBW, output.Length())

		for input, colors := range lastColors {
			offset := offsets[input]
			copy(merged[offset:], colors)
		}
		output.Set(merged)
	}
}
