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

package sources

import (
	"go.eqrx.net/mauzr/pkg/pixels/color"
	"go.eqrx.net/mauzr/pkg/pixels/strip"
)

type input struct {
	length  int
	input   strip.Input
	current []color.RGBW
	last    []color.RGBW
}

// Setup the loop for use. May be called only once.
func (i *input) Setup(length int, _ int) {
	if i.length != 0 {
		panic("reused source")
	}
	if length == 0 {
		panic("zero length")
	}
	if length != i.input.Length() {
		panic("length does not match input")
	}
	i.length = length
}

// Peer the next generated color (Next invocation will return the same color).
func (i *input) Peek() []color.RGBW {
	if i.current != nil {
		return i.current
	}
	next, hasNext := i.input.Get()
	if !hasNext {
		return i.last
	}
	i.last = next
	i.current = next
	return i.current
}

// Pop the next generated color (Next invocation will return the next color).
func (i *input) Pop() []color.RGBW {
	next := i.Peek()
	i.current = nil
	return next
}

func FromInput(i strip.Input) Loop {
	return &input{0, i, nil, nil}
}
