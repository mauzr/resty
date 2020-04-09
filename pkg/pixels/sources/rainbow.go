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
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

type rainbow struct {
	length        int
	cycleDuration time.Duration
	offset, speed float64
}

// NewRainbow returns a Loop that lets the pixels shine in a full color rainbow.
func NewRainbow(cycleDuration time.Duration) Loop {
	return &rainbow{0, cycleDuration, 0, 0}
}

// Setup the loop for use. May be called only once.
func (r *rainbow) Setup(length int, framerate int) {
	if r.length != 0 {
		panic(fmt.Errorf("reused source"))
	}
	if length == 0 {
		panic(fmt.Errorf("zero length"))
	}
	r.length = length
	r.speed = time.Second.Seconds() / (r.cycleDuration.Seconds() * float64(framerate))
}

// Peer the next generated color (Next invocation will return the same color).
func (r *rainbow) Peek() []color.RGBW {
	new := make([]color.RGBW, r.length)
	current := 0.0
	step := 1.0 / float64(r.length)
	for i := int(r.offset) % r.length; i < r.length; i++ {
		new[i] = color.HSV{Hue: current - math.Floor(current), Saturation: 1, Value: 1}.RGBW()
		current += step
	}
	for i := 0; i < int(r.offset)%r.length; i++ {
		new[i] = color.HSV{Hue: current - math.Floor(current), Saturation: 1, Value: 1}.RGBW()
		current += step
	}

	return new
}

// Pop the next generated color (Next invocation will return the next color).
func (r *rainbow) Pop() []color.RGBW {
	new := r.Peek()
	r.offset += r.speed
	return new
}
