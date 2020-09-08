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

// Package play contains logical handling of pixel color sources.
package play

import (
	"errors"
	"time"

	"go.eqrx.net/mauzr/pkg/pixels"
	"go.eqrx.net/mauzr/pkg/pixels/color"
	"go.eqrx.net/mauzr/pkg/pixels/sources"
)

const (
	defaultPartDuration          = 3 * time.Second
	transitionDuration           = 3 * time.Second
	minimumAlertLightLevelFactor = 0.1
)

// DefaultParts creates a map with the default part setup.
func DefaultParts() map[string]func(sources.LoopSetting) {
	return map[string]func(sources.LoopSetting){
		"off":     sources.Static(color.Off()),
		"bright":  sources.Static(color.Bright()),
		"alert":   sources.FadeLoop(defaultPartDuration, color.Off().MixWith(minimumAlertLightLevelFactor, color.Red()), color.Red()),
		"rainbow": sources.Rainbow(defaultPartDuration),
	}
}

// Request of a part change to the play manager.
type Request struct {
	// Response receives possible errors the occurred while processing it. Must have capacity greater 1 or the manager will panic.
	Response chan<- error
	// Part the play next.
	Part string
}

// ErrUnknownPart happens when a part was requested that is now known.
var ErrUnknownPart = errors.New("unknown part")

func handleRequest(parts map[string]func(sources.LoopSetting), request Request) string {
	defer close(request.Response)
	if cap(request.Response) < 1 {
		panic("received blocking channel for response")
	}

	if _, ok := parts[request.Part]; !ok {
		request.Response <- ErrUnknownPart

		return ""
	}

	return request.Part
}

func drainDone(c <-chan interface{}) {
	for ok := true; ok; {
		_, ok = <-c
	}
}

func managePart(parts map[string]func(sources.LoopSetting), currentPart string, manager pixels.SourceManager, requests <-chan Request) string {
	loopDone := make(chan interface{})
	defer drainDone(loopDone)
	transitionDone := make(chan interface{})
	defer drainDone(transitionDone)
	loopTick := make(chan interface{})
	defer close(loopTick)
	transitionTick := make(chan interface{})
	defer close(transitionTick)

	desired := make([]color.RGBW, len(manager.Destination()))
	l := sources.LoopSetting{Tick: loopTick, Done: loopDone, Destination: manager.Destination(), Framerate: manager.Framerate(), Start: desired}
	parts[currentPart](l)
	t := sources.TransitionSetting{Tick: transitionTick, Done: transitionDone, Destination: manager.Destination(), Desired: desired, Framerate: manager.Framerate()}
	sources.Fader(transitionDuration)(t)

	for ok := true; ok; {
		select {
		case <-manager.TickReceiveChan():
			transitionTick <- nil
			_, ok = <-transitionDone
			manager.DoneSendChan() <- nil
		case r, ok := <-requests:
			if !ok {
				return ""
			}
			if nextPart := handleRequest(parts, r); nextPart != "" && nextPart != currentPart {
				return nextPart
			}
		}
	}
	for {
		select {
		case <-manager.TickReceiveChan():
			loopTick <- nil
			if _, ok := <-loopDone; !ok {
				panic("loop stopped")
			}
			manager.DoneSendChan() <- nil
		case r, ok := <-requests:
			if !ok {
				return ""
			}
			if nextPart := handleRequest(parts, r); nextPart != "" {
				return nextPart
			}
		}
	}
}

// New creates a new play manager.
func New(parts map[string]func(sources.LoopSetting), manager pixels.SourceManager, requests <-chan Request) {
	destination := manager.Destination()
	shutdownDesired := make([]color.RGBW, len(destination))

	for i := range destination {
		*destination[i] = color.Unmanaged()
		shutdownDesired[i] = color.Unmanaged()
	}

	go func() {
		nextPart := "default"
		for nextPart != "" {
			nextPart = managePart(parts, nextPart, manager, requests)
		}

		t := sources.TransitionSetting{
			Tick:        manager.TickReceiveChan(),
			Done:        manager.DoneSendChan(),
			Destination: destination,
			Desired:     shutdownDesired,
			Framerate:   manager.Framerate(),
		}
		sources.Fader(transitionDuration)(t)
	}()
}
