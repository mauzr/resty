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

package sk6812_test

import (
	"reflect"
	"testing"

	"go.eqrx.net/mauzr/pkg/etc/pixels/color"
	"go.eqrx.net/mauzr/pkg/etc/pixels/strip"
	"go.eqrx.net/mauzr/pkg/sk6812"
)

// TestConvertStripTo8bitChannels: Test if conversion from RGBW to a byte channel array works.
func TestConvertStripTo8bitChannels(t *testing.T) {
	assembly := strip.New()
	inputColors := []color.RGBW{color.RandomRGBW(), color.RandomRGBW(), color.RandomRGBW()}
	expected := []uint8{
		uint8(inputColors[0].Green * 255.0), uint8(inputColors[0].Red * 255.0), uint8(inputColors[0].Blue * 255.0), uint8(inputColors[0].White * 255.0),
		uint8(inputColors[1].Green * 255.0), uint8(inputColors[1].Red * 255.0), uint8(inputColors[1].Blue * 255.0), uint8(inputColors[1].White * 255.0),
		uint8(inputColors[2].Green * 255.0), uint8(inputColors[2].Red * 255.0), uint8(inputColors[2].Blue * 255.0), uint8(inputColors[2].White * 255.0),
	}
	testStrip := assembly.New("test", 3)
	output := make(chan []uint8)
	defer close(output)
	go sk6812.StripToGRBWBytes(testStrip, output)
	testStrip.Set(inputColors)
	actual := <-output
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expect output %v, was %v", expected, actual)
	}
}

// BenchmarkConvertStripTo8bitChannels: Test how fast the conversion from RGBW to a byte channel array is.
func BenchmarkConvertStripTo8bitChannels(b *testing.B) {
	assembly := strip.New()
	inputColors := []color.RGBW{color.RandomRGBW(), color.RandomRGBW(), color.RandomRGBW()}
	testStrip := assembly.New("test", 3)
	output := make(chan []uint8)
	defer close(output)
	go sk6812.StripToGRBWBytes(testStrip, output)
	for i := 0; i < b.N; i++ {
		testStrip.Set(inputColors)
		<-output
	}
}
