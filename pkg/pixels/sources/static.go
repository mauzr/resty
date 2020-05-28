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
	"go.eqrx.net/mauzr/pkg/pixels/color"
)

// Static just puts a given value to the destination.
func Static(target color.RGBW) func(LoopSetting) {
	if target == nil {
		panic("target not set")
	}
	return func(l LoopSetting) {
		for i := range l.Start {
			l.Start[i] = target
		}
		go func() {
			defer close(l.Done)
			if len(l.Destination) == 0 {
				panic("zero length destination")
			}
			for {
				if _, ok := <-l.Tick; !ok {
					return
				}
				for i := range l.Destination {
					*l.Destination[i] = target
				}
				l.Done <- nil
			}
		}()
	}
}
