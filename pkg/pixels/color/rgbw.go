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
type RGBW interface {
	Red() float64
	Green() float64
	Blue() float64
	White() float64
	Channels() [4]float64
	MixWith(amount float64, other RGBW) RGBW
}

type rgbw struct {
	channels [4]float64
}

// NewRGBW creates a new RGBW representation with the given channels.
func NewRGBW(red, green, blue, white float64) RGBW {
	return &rgbw{[4]float64{red, green, blue, white}}
}

// RandomRGBW for testing.
func RandomRGBW() RGBW {
	return &rgbw{[4]float64{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()}}
}

func (r rgbw) Red() float64 {
	return r.channels[0]
}
func (r rgbw) Green() float64 {
	return r.channels[1]
}
func (r rgbw) Blue() float64 {
	return r.channels[2]
}
func (r rgbw) White() float64 {
	return r.channels[3]
}
func (r rgbw) Channels() [4]float64 {
	return r.channels
}

// mixChannels with each other with the desired preference.
func mixChannels(self, amount, other float64) float64 {
	if self < other {
		return self + (other-self)*amount
	}
	return other + (self-other)*(1-amount)
}

// MixWith some other color and the given tendency to one or the other (0: just take the left color, 1: just take the right color).
func (r rgbw) MixWith(amount float64, other RGBW) RGBW {
	if amount < 0.0 || amount > 1.0 {
		panic(fmt.Sprintf("illegal amount: %v", amount))
	}
	return NewRGBW(
		mixChannels(r.channels[0], amount, other.Red()),
		mixChannels(r.channels[1], amount, other.Green()),
		mixChannels(r.channels[2], amount, other.Blue()),
		mixChannels(r.channels[3], amount, other.White()),
	)
}
