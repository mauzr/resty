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
func TestFadingUp(t *testing.T) {
	assert := assert.New(t)
	tick := make(chan interface{})
	done := make(chan interface{})
	v := color.Off()
	c := sources.TransitionSetting{
		Tick:        tick,
		Done:        done,
		Destination: []*color.RGBW{&v},
		Desired:     []color.RGBW{color.Bright()},
		Framerate:   4,
	}

	sources.Fader(1 * time.Second)(c)
	assert.Equal(color.NewRGBW(0, 0, 0, 0.00).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	assert.Equal(color.NewRGBW(0, 0, 0, 0.25).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	assert.Equal(color.NewRGBW(0, 0, 0, 0.50).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	assert.Equal(color.NewRGBW(0, 0, 0, 0.75).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	_, ok := <-done
	assert.False(ok, "fader did not stop")
	assert.Equal(color.NewRGBW(0, 0, 0, 1.00).Channels(), v.Channels(), "unexpected color array")
}

// TestFadingDown: If fading downwards works.
func TestFadingDown(t *testing.T) {
	assert := assert.New(t)
	tick := make(chan interface{})
	done := make(chan interface{})
	v := color.Bright()
	c := sources.TransitionSetting{
		Tick:        tick,
		Done:        done,
		Destination: []*color.RGBW{&v},
		Desired:     []color.RGBW{color.Off()},
		Framerate:   4,
	}
	sources.Fader(1 * time.Second)(c)
	assert.Equal(color.NewRGBW(0, 0, 0, 1.00).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	assert.Equal(color.NewRGBW(0, 0, 0, 0.75).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	assert.Equal(color.NewRGBW(0, 0, 0, 0.50).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	assert.Equal(color.NewRGBW(0, 0, 0, 0.25).Channels(), v.Channels(), "unexpected color array")
	tick <- nil
	<-done
	_, ok := <-done
	assert.False(ok, "fader did not stop")
	assert.Equal(color.NewRGBW(0, 0, 0, 0.00), v, "unexpected color array")
}
