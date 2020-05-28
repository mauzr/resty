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

package pixels

import (
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
)

// SourceSet is a helper for the creation of multiple sources.
type SourceSet struct {
	Framerate int
	Sources   []Source
}

// Add a new source pair. The Source will be stored in the list inside the struct, the manager is returned.
func (s *SourceSet) Add(destination []*color.RGBW) SourceManager {
	manager, source := NewSourcePair(s.Framerate, destination)
	s.Sources = append(s.Sources, source)
	return manager
}

// SourceManager is the representation of the source manager for the source.
type SourceManager interface {
	TickReceiveChan() <-chan interface{}
	DoneSendChan() chan<- interface{}
	Framerate() int
	Destination() []*color.RGBW
}

// Source is the representation of the source for the manager.
type Source interface {
	AwaitDone(<-chan time.Time) bool
	SendTick(<-chan time.Time)
}

func (s *sourceConfig) Destination() []*color.RGBW {
	return s.destination
}

func (s *sourceConfig) Framerate() int {
	return s.framerate
}

func (s *sourceConfig) TickReceiveChan() <-chan interface{} {
	return s.tick
}

func (s *sourceConfig) DoneSendChan() chan<- interface{} {
	return s.done
}

type sourceConfig struct {
	tick        chan interface{}
	done        chan interface{}
	tickClosed  bool
	framerate   int
	destination []*color.RGBW
}

func (s *sourceConfig) AwaitDone(c <-chan time.Time) bool {
	_, ok := <-s.done
	if !ok && !s.tickClosed {
		close(s.tick)
		s.tickClosed = true
	}
	return !s.tickClosed
}

func (s *sourceConfig) SendTick(c <-chan time.Time) {
	if s.tickClosed {
		return
	}
	s.tick <- nil
}

func (s *sourceConfig) IsTickClosed() bool {
	return s.tickClosed
}

// NewSourcePair creates a new Source and SourceManager pair.
func NewSourcePair(framerate int, destination []*color.RGBW) (SourceManager, Source) {
	if framerate < 0 {
		panic("framerate must be > 0")
	}
	if len(destination) == 0 {
		panic("destination has no elements")
	}
	s := sourceConfig{make(chan interface{}), make(chan interface{}), false, framerate, destination}
	return &s, &s
}
