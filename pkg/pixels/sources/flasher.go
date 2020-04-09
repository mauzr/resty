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
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

type flasher struct {
	length       int
	duration     time.Duration
	currentStep  int
	steps        []color.RGBW
	upper, lower color.RGBW
}

// NewFlasher returns a Loop that lets a given color flash (3/4 color, 1/4 black).
func NewFlasher(lower, upper color.RGBW, duration time.Duration) Loop {
	return &flasher{
		0,
		duration,
		0,
		nil,
		upper, lower,
	}
}

// Setup the loop for use. May be called only once.
func (f *flasher) Setup(length int, framerate int) {
	if f.length != 0 {
		panic(fmt.Errorf("reused source"))
	}
	if length == 0 {
		panic(fmt.Errorf("zero length"))
	}
	f.length = length
	stepLength := f.duration.Seconds() * float64(framerate)
	f.steps = make([]color.RGBW, int(stepLength))
	for i := range f.steps {
		if float64(i)/float64(len(f.steps)) < 0.75 {
			f.steps[i] = f.upper
		} else {
			f.steps[i] = f.lower
		}
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
