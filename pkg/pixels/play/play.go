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

package play

import (
	"errors"
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/color"
	"go.eqrx.net/mauzr/pkg/pixels/sources"
	"go.eqrx.net/mauzr/pkg/pixels/strip"
)

// DefaultParts creates a map with the default part setup.
func DefaultParts() map[string]sources.Loop {
	return map[string]sources.Loop{
		"off":     sources.NewStatic(color.Off),
		"bright":  sources.NewStatic(color.Bright),
		"alert":   sources.NewFadeLoop(color.Off.MixWith(0.1, color.RGBW{Red: 1.0}), color.RGBW{Red: 1.0}, 3*time.Second),
		"rainbow": sources.NewRainbow(3 * time.Second),
	}
}

type Request struct {
	Response chan<- error
	Part     string
}

func setupParts(parts map[string]sources.Loop, output strip.Output, framerate int) {
	alreadySetup := make([]sources.Loop, 0)
	for _, value := range parts {
		setup := false
		for _, other := range alreadySetup {
			if other == value {
				setup = true
			}
		}
		if !setup {
			value.Setup(output.Length(), framerate)
			alreadySetup = append(alreadySetup, value)
		}
	}
}

func shutdown(colors []color.RGBW, output strip.Output, framerate int) {
	target := make([]color.RGBW, len(colors))
	for i := range target {
		target[i] = color.Unmanaged
	}
	source := sources.NewFader(3 * time.Second)
	source.Setup(colors, target, framerate)
	for source.HasNext() {
		output.Set(source.Pop())
	}
}

var ErrUnknownPart = errors.New("unknown part")

func New(parts map[string]sources.Loop, output strip.Output, framerate int, requests <-chan Request) <-chan string {
	if framerate < 0 {
		panic("framerate must be > 0")
	}
	colors := make([]color.RGBW, output.Length())
	for i := range colors {
		colors[i] = color.Unmanaged
	}

	current := make(chan string)
	setupParts(parts, output, framerate)
	go func() {
		currentPart := "default"
		var transition sources.Transition
		defer output.Close()
		defer shutdown(colors, output, framerate)
		defer close(current)

		for {
			select {
			case current <- currentPart:
			case request, ok := <-requests:
				switch {
				case !ok:
					return
				case cap(request.Response) < 1:
					close(request.Response)
					panic("received blocking channel for response")
				default:
					if _, ok := parts[request.Part]; !ok {
						request.Response <- ErrUnknownPart
						close(request.Response)
					} else {
						currentPart = request.Part
						request.Response <- nil
						close(request.Response)
						transition = sources.NewFader(3 * time.Second)
						transition.Setup(colors, parts[currentPart].Peek(), framerate)
					}
				}
			case output.SetChannel() <- colors:
				switch {
				case transition == nil:
					colors = parts[currentPart].Pop()
				case transition.HasNext():
					colors = transition.Pop()
				default:
					transition = nil
					colors = parts[currentPart].Pop()
				}
			}
		}
	}()
	return current
}
