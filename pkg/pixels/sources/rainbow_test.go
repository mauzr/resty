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

package sources_test

import (
	"testing"
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
	"go.eqrx.net/mauzr/pkg/pixels/sources"
)

// BenchmarkRainbow benchmarks the rainbow loop.
func BenchmarkRainbow(b *testing.B) {
	tick := make(chan interface{})
	done := make(chan interface{})
	destination := make([]*color.RGBW, benchmarkStripLength)
	for i := range destination {
		v := color.Off()
		destination[i] = &v
	}
	c := sources.LoopSetting{
		Tick:        tick,
		Done:        done,
		Destination: destination,
		Start:       make([]color.RGBW, benchmarkStripLength),
		Framerate:   4,
	}
	sources.Rainbow(time.Second)(c)
	for i := 0; i < b.N; i++ {
		tick <- nil
		<-done
	}
}
