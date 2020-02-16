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

type flasher struct {
	loopCommon
	currentStep int
	steps       []color.RGBW
}

// NewFlasher returns a Loop that lets a given color flash (3/4 color, 1/4 black).
func NewFlasher(lower, upper color.RGBW) Loop {
	return &flasher{
		loopCommon{},
		0,
		[]color.RGBW{
			upper,
			upper,
			upper,
			lower,
		},
	}
}

// Peer the next generated color (Next invocation will return the same color).
func (f *flasher) Peek() []color.RGBW {
	new := make([]color.RGBW, f.length)
	for i := range new {
		new[i] = f.steps[f.currentStep]
	}
	return new
}

// Pop the next generated color (Next invocation will return the next color).
func (f *flasher) Pop() []color.RGBW {
	new := f.Peek()
	f.currentStep = (f.currentStep + 1) % len(f.steps)
	return new
}
