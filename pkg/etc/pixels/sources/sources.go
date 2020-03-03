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

import "go.eqrx.net/mauzr/pkg/etc/pixels/color"

// Loop is a color generator that can generate endlessly.
type Loop interface {
	// Pop the next generated color (Next invocation will return the next color).
	Pop() []color.RGBW
	// Peer the next generated color (Next invocation will return the same color).
	Peek() []color.RGBW
	// SetLength of the target pixel strip. May be called only once.
	SetLength(length int)
}

// Transition is a color generator that converts the current pixel settings to an other.
type Transition interface {
	// Pop the next generated color (Next invocation will return the next color).
	Pop() []color.RGBW
	// Peer the next generated color (Next invocation will return the same color).
	Peek() []color.RGBW
	// HasNext returns true if the transition contains more colors.
	HasNext() bool
	// SetBoundaries sets start and desired end state of the pixel strip. May be called only once.
	SetBoundaries(start, stop []color.RGBW)
}

// loopCommon contains common stuff for loop implementations.
type loopCommon struct {
	length int
}

// SetLength of the target pixel strip. May be called only once.
func (c *loopCommon) SetLength(length int) {
	switch {
	case c.length != 0:
		panic("reused source")
	case length == 0:
		panic("zero length")
	default:
		c.length = length
	}
}
