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
	"fmt"
	"math/rand"
)

// RGBW is a color represented with four channels: red, green, blue and white.
type RGBW struct {
	Red, Green, Blue, White float64
}

// RandomeRGBW for testing.
func RandomRGBW() RGBW {
	return RGBW{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()}
}

// mixChannels with each other with the desired preference.
func (r RGBW) mixChannels(self, amount, other float64) float64 {
	if self < other {
		return self + (other-self)*amount
	}
	return other + (self-other)*(1-amount)
}

// MixWith some other color and the given tendency to one or the other (0: just take the left color, 1: just take the right color).
func (r RGBW) MixWith(amount float64, other RGBW) RGBW {
	if amount < 0.0 || amount > 1.0 {
		panic(fmt.Sprintf("illegal amount: %v", amount))
	}
	return RGBW{
		Red:   r.mixChannels(r.Red, amount, other.Red),
		Green: r.mixChannels(r.Green, amount, other.Green),
		Blue:  r.mixChannels(r.Blue, amount, other.Blue),
		White: r.mixChannels(r.White, amount, other.White),
	}
}
