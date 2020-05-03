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
)

type static struct {
	length int
	target color.RGBW
}

// NewStatic returns a Loop that always sets the given color.
func NewStatic(target color.RGBW) Loop {
	return &static{0, target}
}

// Setup the loop for use. May be called only once.
func (s *static) Setup(length int, _ int) {
	if s.length != 0 {
		panic("reused source")
	}
	if length == 0 {
		panic("zero length")
	}
	s.length = length
}

// Peer the next generated color (Next invocation will return the same color).
func (s *static) Peek() []color.RGBW {
	new := make([]color.RGBW, s.length)
	for i := range new {
		new[i] = s.target
	}
	return new
}

// Pop the next generated color (Next invocation will return the next color).
func (s *static) Pop() []color.RGBW {
	return s.Peek()
}
