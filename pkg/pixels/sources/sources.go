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

import "go.eqrx.net/mauzr/pkg/pixels/color"

// Loop is a color generator that can generate endlessly.
type Loop interface {
	// Pop the next generated color (Next invocation will return the next color).
	Pop() []color.RGBW
	// Peer the next generated color (Next invocation will return the same color).
	Peek() []color.RGBW
	// Setup the loop for use. May be called only once.
	Setup(length int, framerate int)
}

// Transition is a color generator that converts the current pixel settings to an other.
type Transition interface {
	// Pop the next generated color (Next invocation will return the next color).
	Pop() []color.RGBW
	// Peer the next generated color (Next invocation will return the same color).
	Peek() []color.RGBW
	// HasNext returns true if the transition contains more colors.
	HasNext() bool
	// Setup the transition for use. May be called only once.
	Setup(start, stop []color.RGBW, framerate int)
}
