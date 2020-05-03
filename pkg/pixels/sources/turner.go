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

type turner struct {
	length   int
	current  int
	theme    color.RGBW
	interval time.Duration
}

// NewTurner creates a new turner source.
func NewTurner(theme color.RGBW, interval time.Duration) Loop {
	return &turner{0, 0, theme, interval}
}

func (t *turner) Setup(length int, framerate int) {
	if t.length != 0 {
		panic("reused source")
	}
	if length == 0 {
		panic("zero length")
	}
	t.length = length
}

func (t *turner) Peek() []color.RGBW {
	new := make([]color.RGBW, t.length)
	new[(t.current+0)%t.length] = color.Off.MixWith(0.1, t.theme)
	new[(t.current+1)%t.length] = color.Off.MixWith(0.5, t.theme)
	new[(t.current+2)%t.length] = color.Off.MixWith(1.0, t.theme)
	new[(t.current+3)%t.length] = color.Off.MixWith(0.5, t.theme)
	new[(t.current+4)%t.length] = color.Off.MixWith(0.1, t.theme)

	return new
}

func (t *turner) Pop() []color.RGBW {
	new := t.Peek()
	t.current = (t.current + 1) % t.length
	return new
}
