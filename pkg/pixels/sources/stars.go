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
	"math/rand"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

// Stars emulates that each managed pixel is an independent light source that flickers.
func Stars(theme color.RGBW) func(LoopSetting) {
	if theme == nil {
		panic("theme not set")
	}
	lower := color.Off().MixWith(0.1, theme)
	upper := theme
	return func(l LoopSetting) {
		for i := range l.Start {
			l.Start[i] = upper
		}
		go func() {
			defer close(l.Done)
			if len(l.Destination) == 0 {
				panic("zero length destination")
			}
			factors := make([]float64, len(l.Destination))
			changes := make([]float64, len(l.Destination))
			speed := 0.001 * float64(l.Framerate)
			for i := range factors {
				factors[i] = 1.0
				changes[i] = rand.Float64()*speed + 0.01
			}
			for {
				if _, ok := <-l.Tick; !ok {
					return
				}
				for i := range l.Destination {
					*l.Destination[i] = lower.MixWith(factors[i], upper)
				}
				l.Done <- nil

				for i := range factors {
					factors[i] += changes[i]
					switch {
					case factors[i] >= 1.0:
						factors[i] = 1.0
						changes[i] = -(rand.Float64()*speed + 0.01)
					case factors[i] <= 0.0:
						factors[i] = 0.0
						changes[i] = rand.Float64()*speed + 0.01
					}
				}
			}
		}()
	}
}
