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

package strip

import (
	"reflect"
	"testing"

	"go.eqrx.net/mauzr/pkg/etc/pixels/color"

	"github.com/stretchr/testify/assert"
)

// TestSplitStrip: If splitting strips works.
func TestSplitStrip(t *testing.T) {
	colors := []color.RGBW{color.RandomRGBW(), color.RandomRGBW(), color.RandomRGBW()}
	expected := [][]color.RGBW{{colors[0]}, {colors[1], colors[2]}}
	whole := strip{"testInput", 3, make(chan []color.RGBW)}
	split0 := strip{"testOutput0", 1, make(chan []color.RGBW)}
	split1 := strip{"testOutput1", 2, make(chan []color.RGBW)}
	go split(whole, split0, split1)

	for i := 0; i < 3; i++ {
		whole.Set(colors)
		split0Value, split0Ok := split0.Get()
		split1Value, split1Ok := split1.Get()
		assert.True(t, split0Ok, "First part of the split was closed")
		assert.True(t, split1Ok, "Second part of the split was closed")
		actual := [][]color.RGBW{split0Value, split1Value}
		assert.Equalf(t, expected, actual, "Expect output %v, was %v", expected, actual)
	}

	whole.Close()
	_, split0Ok := split0.Get()
	assert.False(t, split0Ok, "Split did not close the first output channel")
	_, split1Ok := split1.Get()
	assert.False(t, split1Ok, "Split did not close the second output channel")
}

// BenchmarkSplitStrip: How fast is strip splitting?
func BenchmarkSplitStrip(b *testing.B) {
	colors := []color.RGBW{color.RandomRGBW(), color.RandomRGBW(), color.RandomRGBW()}
	whole := strip{"testInput", 3, make(chan []color.RGBW)}
	defer whole.Close()
	split0 := strip{"testOutput0", 1, make(chan []color.RGBW)}
	split1 := strip{"testOutput1", 2, make(chan []color.RGBW)}
	go split(whole, split0, split1)

	for i := 0; i < b.N; i++ {
		whole.Set(colors)
		split0.Get()
		split1.Get()
	}
}

// TestMergeStrip: If merging strips works.
func TestMergeStrip(t *testing.T) {
	expected := []color.RGBW{color.RandomRGBW(), color.RandomRGBW(), color.RandomRGBW()}
	inputColors := [][]color.RGBW{{expected[0]}, {expected[1], expected[2]}}

	split0 := strip{"testInput0", 1, make(chan []color.RGBW)}
	split1 := strip{"testInput1", 2, make(chan []color.RGBW)}
	whole := strip{"testOutput", 3, make(chan []color.RGBW)}
	go merge(whole, split0, split1)

	for i := 0; i < 3; i++ {
		split0.Set(inputColors[0])
		split1.Set(inputColors[1])
		actual, ok := whole.Get()
		assert.True(t, ok, "Expected channel not to be closed")
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expect output %v, was %v", expected, actual)
		}
	}

	split0.Close()
	split1.Close()
	_, ok := whole.Get()
	assert.False(t, ok, "Merge did not close output channel")
}

// TestMergeStripReplication: Test if merge handles shutdown of channels to merge correctly.
func TestMergeStripReplication(t *testing.T) {
	expected := []color.RGBW{color.RandomRGBW(), color.RandomRGBW(), color.RandomRGBW()}

	split0 := strip{"testInput0", 1, make(chan []color.RGBW)}
	split1 := strip{"testInput1", 2, make(chan []color.RGBW)}
	whole := strip{"testOutput", 3, make(chan []color.RGBW)}
	go func() {
		split0.Set([]color.RGBW{expected[0]})
		split0.Set([]color.RGBW{expected[0]})
		split0.Close()
	}()
	go func() {
		split1.Set([]color.RGBW{expected[1], expected[2]})
		split1.Close()
	}()

	go merge(whole, split0, split1)
	actual, ok := whole.Get()
	assert.True(t, ok, "Expected channel to be closed")
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expect output %v, was %v", expected, actual)
	}
	actual, ok = whole.Get()
	assert.True(t, ok, "Expected channel to be closed")
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expect output %v, was %v", expected, actual)
	}
	_, ok = whole.Get()
	assert.False(t, ok, "Merge did not close the channel")
}

// BenchmarkMergeStrip: How fast is strip merging?
func BenchmarkMergeStrip(b *testing.B) {
	inputColors := [][]color.RGBW{{color.RandomRGBW()}, {color.RandomRGBW(), color.RandomRGBW()}}

	split0 := strip{"testInput0", 1, make(chan []color.RGBW)}
	defer split0.Close()
	split1 := strip{"testInput1", 2, make(chan []color.RGBW)}
	defer split1.Close()
	whole := strip{"testOutput", 3, make(chan []color.RGBW)}
	go merge(whole, split0, split1)

	for i := 0; i < b.N; i++ {
		split0.Set(inputColors[0])
		split1.Set(inputColors[1])
		whole.Get()
	}
}

// TestSplitStripTooSmall: Is a split into strips too small handled correctly?
func TestSplitStripTooSmall(t *testing.T) {
	whole := strip{"testInput", 3, make(chan []color.RGBW)}
	split0 := strip{"testOutput0", 1, make(chan []color.RGBW)}
	split1 := strip{"testOutput1", 1, make(chan []color.RGBW)}
	defer whole.Close()
	assert.Panics(t, func() { split(whole, split0, split1) }, "Split accepted input that is too small")
}

// TestSplitStripTooLarge: Is a split into strips too large handled correctly?
func TestSplitStripTooLarge(t *testing.T) {
	whole := strip{"testInput", 3, make(chan []color.RGBW)}
	split0 := strip{"testOutput0", 2, make(chan []color.RGBW)}
	split1 := strip{"testOutput1", 2, make(chan []color.RGBW)}
	defer whole.Close()
	assert.Panics(t, func() { split(whole, split0, split1) }, "Split accepted input that is too large")
}

// TestMergeStripTooSmall: Is a merge into a strip too small handled correctly?
func TestMergeStripTooSmall(t *testing.T) {
	assert.Panics(t, func() {
		split0 := strip{"testInput0", 1, make(chan []color.RGBW)}
		split1 := strip{"testInput1", 2, make(chan []color.RGBW)}
		whole := strip{"testOutput", 7, make(chan []color.RGBW)}
		merge(whole, split0, split1)
	}, "Split accepted invalid strips")
}

// TestMergeStripTooLarge: Is a merge into a strip too large handled correctly?
func TestMergeStripTooLarge(t *testing.T) {
	assert.Panics(t, func() {
		split0 := strip{"testInput0", 1, make(chan []color.RGBW)}
		split1 := strip{"testInput1", 2, make(chan []color.RGBW)}
		whole := strip{"testOutput", 1, make(chan []color.RGBW)}
		merge(whole, split0, split1)
	}, "Split accepted invalid strips")
}
