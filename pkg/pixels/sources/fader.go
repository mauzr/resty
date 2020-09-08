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

// Fader is a transition that moves to a target color.
func Fader(duration time.Duration) func(TransitionSetting) {
	return func(t TransitionSetting) {
		for i := range t.Destination {
			if t.Destination[i] == nil {
				panic("invalid destination")
			}
			if t.Desired[i] == nil {
				panic("invalid desired")
			}
		}
		if len(t.Destination) == 0 {
			panic("zero length destination")
		}
		if len(t.Destination) != len(t.Desired) {
			panic("desired length is not equal to destination legth")
		}
		go func() {
			defer close(t.Done)
			stepsLeft := int(duration.Seconds() / time.Second.Seconds() * float64(t.Framerate))
			steps := make([]color.RGBW, len(t.Destination))
			for i := range steps {
				desiredChannels := t.Desired[i].Channels()
				destinationChannels := (*t.Destination[i]).Channels()
				steps[i] = color.NewRGBW(
					(desiredChannels[0]-destinationChannels[0])/float64(stepsLeft),
					(desiredChannels[1]-destinationChannels[1])/float64(stepsLeft),
					(desiredChannels[2]-destinationChannels[2])/float64(stepsLeft),
					(desiredChannels[3]-destinationChannels[3])/float64(stepsLeft),
				)
			}
			for {
				if _, ok := <-t.Tick; !ok {
					return
				}
				for i := range t.Destination {
					destinationChannels := (*t.Destination[i]).Channels()
					stepChannels := steps[i].Channels()
					//nolint:gomnd // 0.0 is minimum, 1.0 is maximum
					c := color.NewRGBW(
						math.Max(0.0, math.Min(destinationChannels[0]+stepChannels[0], 1.0)),
						math.Max(0.0, math.Min(destinationChannels[1]+stepChannels[1], 1.0)),
						math.Max(0.0, math.Min(destinationChannels[2]+stepChannels[2], 1.0)),
						math.Max(0.0, math.Min(destinationChannels[3]+stepChannels[3], 1.0)),
					)
					*t.Destination[i] = c
				}
				stepsLeft--
				if stepsLeft == 0 {
					return
				}
				t.Done <- nil
			}
		}()
	}
}
