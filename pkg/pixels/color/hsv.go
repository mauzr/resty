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

package color

import (
	"math"
)

// HSV is a color represented as HSV value.
type HSV struct {
	Hue, Saturation, Value float64
}

// RGBW converts the color to RGBW (with white set to 0).
func (h HSV) RGBW() RGBW {
	i := math.Floor(h.Hue * 6)
	f := h.Hue*6 - i
	p := h.Value * (1 - h.Saturation)
	q := h.Value * (1 - f*h.Saturation)
	t := h.Value * (1 - (1-f)*h.Saturation)

	switch int(i) % 6 {
	case 0:
		return RGBW{h.Value, t, p, 0}
	case 1:
		return RGBW{q, h.Value, p, 0}
	case 2:
		return RGBW{p, h.Value, t, 0}
	case 3:
		return RGBW{p, q, h.Value, 0}
	case 4:
		return RGBW{t, p, h.Value, 0}
	case 5:
		return RGBW{h.Value, p, q, 0}
	}
	panic("Invalid calculation")
}
