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
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

type fadeLoop struct {
	length       int
	steps        int
	framerate    int
	duration     time.Duration
	up           bool
	lower, upper color.RGBW
	current      []color.RGBW
	currentFader Transition
}

func NewFadeLoop(lower, upper color.RGBW, duration time.Duration) Loop {
	return &fadeLoop{
		0,
		0,
		0,
		duration,
		true,
		lower, upper,
		nil,
		nil,
	}
}

// Setup the loop for use. May be called only once.
func (f *fadeLoop) Setup(length int, framerate int) {
	if f.current != nil {
		panic("reused source")
	}
	if length == 0 {
		panic("zero length")
	}
	f.framerate = framerate
	f.length = length
	f.current = make([]color.RGBW, length)
	for i := range f.current {
		f.current[i] = f.lower
	}
}

// Peer the next generated color (Next invocation will return the same color).
func (f *fadeLoop) Peek() []color.RGBW {
	current := make([]color.RGBW, f.length)
	copy(current, f.current)
	return current
}

// Pop the next generated color (Next invocation will return the next color).
func (f *fadeLoop) Pop() []color.RGBW {
	if f.currentFader == nil {
		f.up = !f.up
		target := f.lower
		if f.up {
			target = f.upper
		}
		stop := make([]color.RGBW, f.length)
		for i := range stop {
			stop[i] = target
		}
		f.currentFader = NewFader(f.duration / 2)
		f.currentFader.Setup(f.current, stop, f.framerate)
	}
	f.current = f.currentFader.Pop()
	if !f.currentFader.HasNext() {
		f.currentFader = nil
	}
	return f.Peek()
}
