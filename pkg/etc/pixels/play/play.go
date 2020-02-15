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
	"context"
	"fmt"

	"go.eqrx.net/mauzr/pkg/etc/pixels/color"
	"go.eqrx.net/mauzr/pkg/etc/pixels/sources"
	"go.eqrx.net/mauzr/pkg/etc/pixels/strip"
)

const (
	transitionSteps = 30
)

// DefaultParts creates a map with the default part setup.
func DefaultParts() map[string]sources.Loop {
	return map[string]sources.Loop{
		"off":     sources.NewStatic(color.Off),
		"bright":  sources.NewStatic(color.Bright),
		"alert":   sources.NewFadeLoop(color.Off.MixWith(0.1, color.Alert), color.Alert, transitionSteps),
		"rainbow": sources.NewRainbow(0.5),
	}
}

// Changer is a component that can change the part of a play.
type Changer interface {
	// ChangePart that is currently playing.
	ChangePart(ctx context.Context, mode string) error
}

type Play interface {
	// Run manages the play until canceled.
	Run(ctx context.Context)
	// ChangePart that is currently playing.
	ChangePart(ctx context.Context, part string) error
}

// play implements Play.
type play struct {
	parts         map[string]sources.Loop
	currentPart   string
	output        strip.Output
	currentColors []color.RGBW
	partChange    chan string
}

// New creates a new Play.
func New(parts map[string]sources.Loop, output strip.Output) Play {
	colors := make([]color.RGBW, output.Length())
	for i := range colors {
		colors[i] = color.Unmanaged
	}

	alreadySetup := make([]sources.Loop, 0)
	for _, value := range parts {
		setup := false
		for _, other := range alreadySetup {
			if other == value {
				setup = true
			}
		}
		if !setup {
			value.SetLength(output.Length())
			alreadySetup = append(alreadySetup, value)
		}
	}
	return &play{parts, "default", output, colors, make(chan string)}
}

// Run manages the play until canceled.
func (p *play) Run(ctx context.Context) {
	defer p.output.Close()
	var transition sources.Transition
	for {
		select {
		case <-ctx.Done():
			p.performShutdown()
			return
		case mode := <-p.partChange:
			p.currentPart = mode
			transition = sources.NewFader(transitionSteps)
			transition.SetBoundaries(p.currentColors, p.parts[p.currentPart].Peek())
		case p.output.SetChannel() <- p.currentColors:
			switch {
			case transition == nil:
				p.currentColors = p.parts[p.currentPart].Pop()
			case transition.HasNext():
				p.currentColors = transition.Pop()
			default:
				transition = nil
				p.currentColors = p.parts[p.currentPart].Pop()
			}
		}
	}
}

// performShutdown transits the managed strip into unmanged colors.
func (p *play) performShutdown() {
	target := make([]color.RGBW, len(p.currentColors))
	for i := range target {
		target[i] = color.Unmanaged
	}
	source := sources.NewFader(transitionSteps)
	source.SetBoundaries(p.currentColors, target)
	for source.HasNext() {
		p.output.Set(source.Pop())
	}
}

// ChangePart that is currently playing.
func (p *play) ChangePart(ctx context.Context, part string) error {
	if _, ok := p.parts[part]; !ok {
		return fmt.Errorf("unknown mode: %v", part)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.partChange <- part:
		return nil
	}
}
