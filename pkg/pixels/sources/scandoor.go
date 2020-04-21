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
	"math"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

type scanDoor struct {
	lower, upper                color.RGBW
	position, change, maxHeight float64
	positions                   [11][2]float64
}

// NewScanDoor returns a Loop that lets the pixels around doors do something fancy.
func NewScanDoor(theme color.RGBW) Loop {
	return &scanDoor{
		color.Off, theme,
		0.0, 0, 5.0,
		[11][2]float64{
			{0, 0}, {0, 1}, {0, 2}, {0, 3},
			{0, 4}, {1, 4}, {2, 4},
			{2, 3}, {2, 2}, {2, 1}, {2, 0},
		},
	}
}

// Setup the loop for use. May be called only once.
func (s *scanDoor) Setup(length int, framerate int) {
	if s.change != 0 {
		panic("reused source")
	}
	if length != 11 {
		panic(fmt.Errorf("strip length must be 11"))
	}
	s.change = 0.3 / float64(framerate)
}

// Peer the next generated color (Next invocation will return the same color).
func (s *scanDoor) Peek() []color.RGBW {
	new := make([]color.RGBW, 11)
	for i, pixel := range s.positions {
		relativedistance := math.Abs(pixel[1]-s.position) / s.maxHeight
		new[i] = s.lower.MixWith(math.Max(0.0, math.Min(relativedistance, 1.0)), s.upper)
	}

	return new
}

// Pop the next generated color (Next invocation will return the next color).
func (s *scanDoor) Pop() []color.RGBW {
	new := s.Peek()
	s.position += s.change
	if s.position >= s.maxHeight || s.position <= 0.0 {
		s.change = -s.change
	}
	return new
}
