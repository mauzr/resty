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

	"github.com/stretchr/testify/assert"
	"go.eqrx.net/mauzr/pkg/pixels/color"
	"go.eqrx.net/mauzr/pkg/pixels/sources"
)

// TestFlasher: If the flasher actually flashes.
func TestFlasher(t *testing.T) { //nolint
	fader := sources.NewFlasher(color.Off, color.Bright, 1*time.Second)
	fader.Setup(1, 4)
	assert.Equal(t, []color.RGBW{color.Bright}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Bright}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Bright}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Off}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Bright}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Bright}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Bright}, fader.Pop())
	assert.Equal(t, []color.RGBW{color.Off}, fader.Pop())
}

func BenchmarkFlasher(b *testing.B) {
	fader := sources.NewFlasher(color.Off, color.Bright, 1*time.Second)
	fader.Setup(benchmarkStripLength, 4)
	for i := 0; i < b.N; i++ {
		fader.Pop()
	}
}