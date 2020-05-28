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

// Flasher jumps between colors without a transition.
func Flasher(duration time.Duration, lower, upper color.RGBW) func(LoopSetting) {
	if lower == nil || upper == nil {
		panic("lower or upper not set")
	}
	return func(l LoopSetting) {
		stepLength := duration.Seconds() * float64(l.Framerate)
		for i := range l.Start {
			l.Start[i] = lower
		}
		go func() {
			defer close(l.Done)
			if len(l.Destination) == 0 {
				panic("zero length")
			}
			steps := make([]color.RGBW, int(stepLength))
			for i := range steps {
				if float64(i)/float64(len(steps)) < 0.75 {
					steps[i] = upper
				} else {
					steps[i] = lower
				}
			}
			currentStep := 0
			for {
				if _, ok := <-l.Tick; !ok {
					return
				}
				for i := range l.Destination {
					*l.Destination[i] = steps[currentStep]
				}
				l.Done <- nil
				currentStep = (currentStep + 1) % len(steps)
			}
		}()
	}
}
