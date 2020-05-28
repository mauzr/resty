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
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

func applyTurner(theme color.RGBW, i int, destination []color.RGBW) {
	for j := range destination {
		destination[j] = color.Off()
	}
	length := len(destination)
	destination[(i+0)%length] = color.Off().MixWith(0.1, theme)
	destination[(i+1)%length] = color.Off().MixWith(0.5, theme)
	destination[(i+2)%length] = color.Off().MixWith(1.0, theme)
	destination[(i+3)%length] = color.Off().MixWith(0.5, theme)
	destination[(i+4)%length] = color.Off().MixWith(0.1, theme)
}

// Turner puts a rotating light on circular positioned pixels.
func Turner(theme color.RGBW, interval time.Duration) func(LoopSetting) {
	return func(l LoopSetting) {
		applyTurner(theme, 0, l.Start)
		go func() {
			defer close(l.Done)
			if len(l.Destination) == 0 {
				panic("zero length destination")
			}
			length := len(l.Destination)
			current := 0
			for {
				if _, ok := <-l.Tick; !ok {
					return
				}
				for i := range l.Destination {
					*l.Destination[i] = color.Off()
				}
				*l.Destination[(current+0)%length] = color.Off().MixWith(0.1, theme)
				*l.Destination[(current+1)%length] = color.Off().MixWith(0.5, theme)
				*l.Destination[(current+2)%length] = color.Off().MixWith(1.0, theme)
				*l.Destination[(current+3)%length] = color.Off().MixWith(0.5, theme)
				*l.Destination[(current+4)%length] = color.Off().MixWith(0.1, theme)

				l.Done <- nil

				current = (current + 1) % length
			}
		}()
	}
}
