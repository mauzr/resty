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
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

func scanDoorLUT(lower, upper color.RGBW, speed time.Duration, framerate int) [][11]color.RGBW {
	destinationPositions := [11]float64{0, 1 / 4, 2 / 4, 3 / 4, 1, 1, 1, 3 / 4, 2 / 4, 1 / 4, 0}
	lut := make([][11]color.RGBW, int(speed.Seconds()*float64(framerate/2)))
	for i := range lut {
		for position := range lut[i] {
			relPosition := float64(position) / float64(len(lut[i]))
			distance := math.Abs(destinationPositions[position] - relPosition)
			lut[i][position] = lower.MixWith(distance, upper)
		}
	}
	return lut
}

// ScanDoor does some specific things with the door pixels in my home.
func ScanDoor(theme color.RGBW, speed time.Duration) func(LoopSetting) {
	if theme == nil {
		panic("theme not set")
	}
	return func(l LoopSetting) {
		lut := scanDoorLUT(color.Off(), theme, speed, l.Framerate)

		for i := range l.Start {
			l.Start[i] = lut[0][i]
		}
		go func() {
			defer close(l.Done)
			if len(l.Destination) == 0 {
				panic("zero length destination")
			}

			if len(l.Destination) != 11 {
				panic("strip length must be 11")
			}
			position := 0
			up := true
			for {
				if _, ok := <-l.Tick; !ok {
					return
				}
				for i := range l.Destination {
					*l.Destination[i] = lut[position][i]
				}
				l.Done <- nil

				if up {
					position++
					up = position != len(lut)-1
				} else {
					position--
					up = position == 0
				}
			}
		}()
	}
}
