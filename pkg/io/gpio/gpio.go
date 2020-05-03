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

package gpio

import (
	"fmt"
	"os"
	"strconv"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

const (
	// Input means the pin is an input
	Input = true
	// Output means the pin is an output
	Output = false
	// PullUp means that the pin is pulled to VCC
	PullUp = 1
	// PullDown means that the pin is pulled to ground
	PullDown = 1
	// PullNone means that the pin is not pulled
	PullNone = 2
)

// Pin represents a GPIO pin.
type Pin interface {
	// Pull sets the pull up/down setting of the pin.
	Pull(direction uint) func() error
	// Direction sets the direction of the pin.
	Direction(input bool) func() error
	// Exports the the pin.
	Export() func() error
	// Read returns the current value of the pin.
	Read(destination *bool) func() error
	// Write sets the pin value.
	Write(source bool) func() error
	// Identifier returns the identifier of the pin.
	Identifier() string
}

// pin represents a gpio pin.
type pin struct {
	identifier    string
	exportFile    file.File
	directionFile file.File
	valueFile     file.File
	gpioMemFile   file.MemoryMap
}

// New creates a new pin.
// See /sys/kernel/debug/gpio for the pin number when using mainline.
func New(identifier string) Pin {
	return &pin{
		identifier:    identifier,
		exportFile:    file.New("/sys/class/gpio/export"),
		directionFile: file.New(fmt.Sprintf("/sys/class/gpio/gpio%v/direction", identifier)),
		valueFile:     file.New(fmt.Sprintf("/sys/class/gpio/gpio%v/value", identifier)),
		gpioMemFile:   file.NewMemoryMap("/dev/gpiomem"),
	}
}

// Identifier returns the identifier of the pin.
func (p *pin) Identifier() string {
	return p.identifier
}

// Exports the the pin.
func (p *pin) Export() func() error {
	return func() error {
		if _, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%v/direction", p.identifier)); os.IsNotExist(err) {
			f := p.exportFile

			actions := []io.Action{f.Open(os.O_WRONLY, 0660), f.WriteString(p.identifier)}
			if err := io.Execute(fmt.Sprintf("could not export linux GPIO pin %v", p.identifier), actions, []io.Action{f.Close()}); err != nil {
				return err
			}
		}
		return nil
	}
}

// Direction sets the direction of the pin.
func (p *pin) Direction(direction bool) func() error {
	return func() error {
		d := "out"
		if direction {
			d = "in"
		}
		f := p.directionFile

		actions := []io.Action{f.Open(os.O_RDWR, 0660), f.WriteString(d)}
		if err := io.Execute(fmt.Sprintf("could not set direction linux GPIO pin %v", p.identifier), actions, []io.Action{f.Close()}); err != nil {
			return err
		}
		return nil
	}
}

// Read returns the current value of the pin.
func (p *pin) Read(destination *bool) func() error {
	return func() error {
		f := p.valueFile
		rawValue := ""

		actions := []io.Action{f.Open(os.O_RDWR, 0660), f.ReadString(&rawValue, 1)}
		if err := io.Execute(fmt.Sprintf("could not read linux GPIO pin %v", p.identifier), actions, []io.Action{f.Close()}); err != nil {
			return err
		} else if v, err := strconv.ParseBool(rawValue); err != nil {
			return fmt.Errorf("could not parse linux GPIO pin %v value: %w", p.identifier, err)
		} else {
			*destination = v
			return nil
		}
	}
}

// Write sets the pin value.
func (p *pin) Write(value bool) func() error {
	return func() error {
		f := p.valueFile
		rawValue := "0"
		if value {
			rawValue = "1"
		}

		actions := []io.Action{f.Open(os.O_RDWR, 0660), f.WriteString(rawValue)}

		if err := io.Execute(fmt.Sprintf("could not write linux GPIO pin %v", p.identifier), actions, []io.Action{f.Close()}); err != nil {
			return err
		}
		return nil
	}
}
