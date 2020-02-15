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

package sk6812

import (
	"fmt"

	"go.eqrx.net/mauzr/pkg/etc/pixels/strip"
)

// channelToByte converts a single channel to its byte representation.
func channelToByte(channel float64) uint8 {
	if channel < 0.0 || channel > 1.0 {
		panic(fmt.Errorf("illegal channel value: %f", channel))
	}
	return uint8(255.0 * channel)
}

// StripToGRBWBytes converts a strip input to byte representation arrays of the set colors.
func StripToGRBWBytes(input strip.Input, output chan<- []uint8) {
	defer close(output)
	channels := make([]uint8, input.Length()*4)
	for {
		colors, ok := input.Get()
		if !ok {
			return
		}
		for i, color := range colors {
			channels[i*4+0] = channelToByte(color.Green)
			channels[i*4+1] = channelToByte(color.Red)
			channels[i*4+2] = channelToByte(color.Blue)
			channels[i*4+3] = channelToByte(color.White)
		}
		output <- channels
	}
}
