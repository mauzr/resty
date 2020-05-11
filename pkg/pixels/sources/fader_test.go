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
	"go.eqrx.net/mauzr/pkg/testing/assert"
)

// TestFadingUp: If fading upwards works.
func TestFadingUp(t *testing.T) { //nolint
	assert := assert.New(t)
	fader := sources.NewFader(1 * time.Second)
	start, end := make([]color.RGBW, 3), make([]color.RGBW, 3)
	for i := range start {
		start[i] = color.Off
		end[i] = color.Bright
	}
	fader.Setup(start, end, 5)
	assert.Equal([]color.RGBW{color.Off, color.Off, color.Off}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{{White: 0.25}, {White: 0.25}, {White: 0.25}}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{{White: 0.5}, {White: 0.5}, {White: 0.5}}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{{White: 0.75}, {White: 0.75}, {White: 0.75}}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{color.Bright, color.Bright, color.Bright}, fader.Pop(), "unexpected color array")
	assert.False(fader.HasNext(), "unexpected color array")
}

// TestFadingUp: If fading downwards works.
func TestFadingDown(t *testing.T) { //nolint
	assert := assert.New(t)
	fader := sources.NewFader(1 * time.Second)
	start, end := make([]color.RGBW, 3), make([]color.RGBW, 3)
	for i := range start {
		start[i] = color.Bright
		end[i] = color.Off
	}
	fader.Setup(start, end, 5)
	assert.Equal([]color.RGBW{color.Bright, color.Bright, color.Bright}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{{White: 0.75}, {White: 0.75}, {White: 0.75}}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{{White: 0.5}, {White: 0.5}, {White: 0.5}}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{{White: 0.25}, {White: 0.25}, {White: 0.25}}, fader.Pop(), "unexpected color array")
	assert.Equal([]color.RGBW{color.Off, color.Off, color.Off}, fader.Pop(), "unexpected color array")
	assert.False(fader.HasNext(), "unexpected color array")
}
