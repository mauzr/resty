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

// Package color contains color types.
package color

const (
	logicalColorsDampendingFactor = 0.1
)

// Yellow returns the the color yellow.
func Yellow() RGBW {
	return &rgbw{[4]float64{1.0, 0.9, 0, 0}}
}

// Red returns  is the the color red.
func Red() RGBW {
	return &rgbw{[4]float64{1.0, 0, 0, 0}}
}

// White returns is the the color white.
func White() RGBW {
	return &rgbw{[4]float64{0, 0, 0, 1.0}}
}

// Green returns is the the color green.
func Green() RGBW {
	return &rgbw{[4]float64{0, 1.0, 0, 0}}
}

// Off turns the pixel off.
func Off() RGBW {
	return rgbw{[4]float64{0, 0, 0, 0}}
}

// Bright sets the pixel to as bright as possible.
func Bright() RGBW {
	return White()
}

// Unmanaged indicates that the pixel is not actively managed.
func Unmanaged() RGBW {
	return rgbw{[4]float64{1.0, 0, 0, 0}}
}

// Error indicates that something is wrong.
func Error() RGBW {
	return Off().MixWith(logicalColorsDampendingFactor, Red())
}

// Warning indicates that something requires attention.
func Warning() RGBW {
	return Off().MixWith(logicalColorsDampendingFactor, Yellow())
}

// Good indicated that everything is fine.
func Good() RGBW {
	return Off().MixWith(logicalColorsDampendingFactor, Green())
}
