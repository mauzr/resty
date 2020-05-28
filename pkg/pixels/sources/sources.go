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

// Package sources contains colors sources for pixels.
package sources

import (
	"go.eqrx.net/mauzr/pkg/pixels/color"
)

// LoopSetting is the value interface with a loop generator.
type LoopSetting struct {
	Tick        <-chan interface{}
	Done        chan<- interface{}
	Destination []*color.RGBW
	Start       []color.RGBW
	Framerate   int
}

// TransitionSetting is the value interface with a transition generator.
type TransitionSetting struct {
	Tick        <-chan interface{}
	Done        chan<- interface{}
	Destination []*color.RGBW
	Desired     []color.RGBW
	Framerate   int
}
