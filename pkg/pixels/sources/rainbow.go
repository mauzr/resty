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

func rainbowLUT(destinationLength int, duration time.Duration, framerate int) [][]color.RGBW {
	lut := make([][]color.RGBW, int(duration.Seconds()*float64(framerate)))
	stepDelta := 1 / float64(len(lut))
	step := 0.0
	for i := range lut {
		lut[i] = make([]color.RGBW, destinationLength)
		for j := range lut[i] {
			lut[i][j] = color.HSV{Hue: math.Mod(float64(i)*step, 1.0), Saturation: 1, Value: 1}.RGBW()
		}
		step += stepDelta
	}

	return lut
}

// Rainbow fills the controlled pixels with a rainbow that has a moving offset.
func Rainbow(duration time.Duration) func(LoopSetting) {
	return func(l LoopSetting) {
		lut := rainbowLUT(len(l.Destination), duration, l.Framerate)
		for i := range l.Start {
			l.Start[i] = lut[0][i]
		}
		go func() {
			defer close(l.Done)
			if len(l.Destination) == 0 {
				panic("zero length destination")
			}

			offset := 0
			for {
				if _, ok := <-l.Tick; !ok {
					return
				}
				for i := range l.Destination {
					*l.Destination[i] = lut[offset][i]
				}
				l.Done <- nil
				offset = (offset + 1) % len(lut)
			}
		}()
	}
}
