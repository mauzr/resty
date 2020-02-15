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

type rainbow struct {
	loopCommon
	offset, speed float64
}

// NewRainbow returns a Loop that lets the pixels shine in a full color rainbow.
func NewRainbow(speed float64) Loop {
	return &rainbow{loopCommon{}, 0, speed}
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
