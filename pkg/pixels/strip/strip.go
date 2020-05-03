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

// Package strip manages physically connected pixels.
package strip

import (
	"fmt"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

// Strip is a chain of pixels.
type Strip interface {
	// Name of the strip.
	Name() string
	// Length of the strip.
	Length() int
	// Get the next value.
	Get() (colors []color.RGBW, ok bool)
	// GetChannel returns the channel to receive strip values by.
	GetChannel() <-chan []color.RGBW
	// Set colors of the strip.
	Set(colors []color.RGBW)
	// SetChannel returns the channel to send strip values by.
	SetChannel() chan<- []color.RGBW
	// Close the strips data channel.
	Close()
}

// Input is a source for strip values.
type Input interface {
	// Name of the input.
	Name() string
	// Length of the input.
	Length() int
	// Get the next value.
	Get() (colors []color.RGBW, ok bool)
	// GetChannel returns the channel to receive strip values by.
	GetChannel() <-chan []color.RGBW
}

// Output is a sink for strip values.
type Output interface {
	// Name of the output.
	Name() string
	// Length of the output.
	Length() int
	// Set colors of the strip.
	Set(colors []color.RGBW)
	// SetChannel returns the channel to send output values by.
	SetChannel() chan<- []color.RGBW
	// Close the outputs data channel.
	Close()
}

// strip implements Strip.
type strip struct {
	name    string
	length  int
	channel chan []color.RGBW
}

// Length of the strip.
func (s strip) Length() int {
	return s.length
}

// Name of the strip.
func (s strip) Name() string {
	return s.name
}

// Get the next value.
func (s strip) Get() ([]color.RGBW, bool) {
	colors, ok := <-s.channel
	return colors, ok
}

// GetChannel returns the channel to receive strip values by.
func (s strip) GetChannel() <-chan []color.RGBW {
	return s.channel
}

// Set colors of the strip.
func (s strip) Set(colors []color.RGBW) {
	if len(colors) != s.length {
		panic(fmt.Sprintf("illegal colors length (%d) received for %s. Expected %d", len(colors), s.name, s.length))
	}
	s.channel <- colors
}

// SetChannel returns the channel to send strip values by.
func (s strip) SetChannel() chan<- []color.RGBW {
	return s.channel
}

// Close the strips data channel.
func (s strip) Close() {
	close(s.channel)
}
