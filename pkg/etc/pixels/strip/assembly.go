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

// Assembly is a set of pixel strips. Strips of an assembly are on the same hardware line and all result in the same end.
type Assembly interface {
	// New creates a new strip with the given name and length.
	New(name string, length int) Strip
	// Merge multiple input strips into one output.
	Merge(output Output, inputs ...Input)
	// Split a strip into multiple outputs.
	Split(input Input, outputs ...Output)
	// Check if the strips are correctly wired and that the given input is the end of the string.
	Check(end Input)
}

// assembly implements Assembly.
type assembly struct {
	links map[string][]string
}

// New creates a new Assembly that is a set of pixel strips. Strips of an assembly are on the same hardware line and all result in the same end.
func New() Assembly {
	return &assembly{map[string][]string{}}
}

// Check if the strips are correctly wired and that the given input is the end of the string.
func (a *assembly) Check(end Input) {
	loose := []string{}
	for input, outputs := range a.links {
		if len(outputs) == 0 {
			loose = append(loose, input)
		}
	}
	if len(loose) != 1 {
		panic(fmt.Errorf("strips %s goes into nothing", loose))
	}
	if loose[0] != end.Name() {
		panic(fmt.Errorf("assembly should end in %s, but ends in %s", loose[0], end.Name()))
	}
}

// New creates a new strip with the given name and length.
func (a *assembly) New(name string, length int) Strip {
	strip := strip{name, length, make(chan []color.RGBW)}
	if _, ok := a.links[strip.Name()]; ok {
		panic(fmt.Errorf("duplicate strip name: %s", strip.Name()))
	}
	a.links[strip.Name()] = []string{}
	return &strip
}

// Merge multiple input strips into one output.
func (a *assembly) Merge(output Output, inputs ...Input) {
	for _, input := range inputs {
		if _, ok := a.links[input.Name()]; !ok {
			panic(fmt.Errorf("input unknown: %s", input.Name()))
		}
		for _, other := range a.links[input.Name()] {
			if output.Name() == other {
				panic(fmt.Errorf("%s receives colors from %s multiple times", other, input.Name()))
			}
		}
		a.links[input.Name()] = append(a.links[input.Name()], output.Name())
	}
	go merge(output, inputs...)
}

// Split a strip into multiple outputs.
func (a *assembly) Split(input Input, outputs ...Output) {
	if _, ok := a.links[input.Name()]; !ok {
		panic(fmt.Errorf("input unknown: %s", input.Name()))
	}

	for _, output := range outputs {
		for _, other := range a.links[input.Name()] {
			if output.Name() == other {
				panic(fmt.Errorf("%s receives colors from %s multiple times", other, input.Name()))
			}
		}
		a.links[input.Name()] = append(a.links[input.Name()], output.Name())
	}
	go split(input, outputs...)
}
