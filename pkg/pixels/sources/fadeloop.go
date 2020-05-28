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

// FadeLoop loops between two color values using fade as transition.
func FadeLoop(duration time.Duration, lower, upper color.RGBW) func(LoopSetting) {
	return func(l LoopSetting) {
		if len(l.Destination) == 0 || lower == nil || upper == nil {
			panic("invalid parameters")
		}
		for i := range l.Start {
			l.Start[i] = lower
		}
		go func() {
			defer close(l.Done)
			up := false
			desired := make([]color.RGBW, len(l.Destination))
			for {
				up = !up
				target := lower
				if up {
					target = upper
				}
				for i := range desired {
					desired[i] = target
				}
				done := make(chan interface{})
				t := TransitionSetting{l.Tick, done, l.Destination, desired, l.Framerate}
				handleFadeLoopStep(duration, done, t, l)
			}
		}()
	}
}

func handleFadeLoopStep(duration time.Duration, done <-chan interface{}, t TransitionSetting, l LoopSetting) {
	go Fader(duration / 2)(t)
	for ok := true; ok; {
		_, ok = <-done
		l.Done <- nil
	}
	select {
	case _, ok := <-l.Tick:
		if ok {
			panic("missed tick")
		} else {
			return
		}
	default:
	}
}
