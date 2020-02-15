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
	"math"

	"go.eqrx.net/mauzr/pkg/etc/pixels/color"
)

type fader struct {
	loopCommon
	stepsLeft      int
	current, steps []color.RGBW
}

// NewFlasher returns a transition from the later set start and end in the given amount of steps.
func NewFader(steps int) Transition {
	return &fader{loopCommon{}, steps, nil, nil}
}

// SetBoundaries sets start and desired end state of the pixel strip. May be called only once.
func (f *fader) SetBoundaries(start, end []color.RGBW) {
	if f.current != nil {
		panic("reused fader")
	}
	if len(start) != len(end) {
		panic("start and end have different sizes")
	}
	f.current = make([]color.RGBW, len(start))
	copy(f.current, start)
	f.steps = make([]color.RGBW, len(start))
	for i := range f.steps {
		f.steps[i] = color.RGBW{
			Red:   (end[i].Red - start[i].Red) / float64(f.stepsLeft-1),
			Green: (end[i].Green - start[i].Green) / float64(f.stepsLeft-1),
			Blue:  (end[i].Blue - start[i].Blue) / float64(f.stepsLeft-1),
			White: (end[i].White - start[i].White) / float64(f.stepsLeft-1),
		}
	}
}

// HasNext returns true if the transition contains more colors.
func (f *fader) HasNext() bool {
	return f.stepsLeft != 0
}

// Peer the next generated color (Next invocation will return the same color).
func (f *fader) Peek() []color.RGBW {
	new := make([]color.RGBW, len(f.current))
	copy(new, f.current)
	return new
}

// Pop the next generated color (Next invocation will return the next color).
func (f *fader) Pop() []color.RGBW {
	new := f.Peek()
	if !f.HasNext() {
		panic("no values left")
	}
	for i := range f.current {
		f.current[i].Red = math.Max(0.0, math.Min(f.current[i].Red+f.steps[i].Red, 1.0))
		f.current[i].Green = math.Max(0.0, math.Min(f.current[i].Green+f.steps[i].Green, 1.0))
		f.current[i].Blue = math.Max(0.0, math.Min(f.current[i].Blue+f.steps[i].Blue, 1.0))
		f.current[i].White = math.Max(0.0, math.Min(f.current[i].White+f.steps[i].White, 1.0))
	}
	f.stepsLeft--
	return new
}
