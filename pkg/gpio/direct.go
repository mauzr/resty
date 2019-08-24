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

import "strconv"

import (
	"fmt"
	"os"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

type normalPin struct {
	identifier    string
	exportFile    file.File
	directionFile file.File
	valueFile     file.File
	gpioMemFile   file.MemoryMap
}

func newNormalPin(identifier string) Pin {
	return &normalPin{
		identifier:    identifier,
		exportFile:    file.NewFile("/sys/class/gpio/export"),
		directionFile: file.NewFile(fmt.Sprintf("/sys/class/gpio/gpio%v/direction", identifier)),
		valueFile:     file.NewFile(fmt.Sprintf("/sys/class/gpio/gpio%v/value", identifier)),
		gpioMemFile:   file.NewMemoryMap("/dev/gpiomem"),
	}
}

// Export a pin
func (p *normalPin) Export() func() error {
	return func() error {
		if _, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%v/direction", p.identifier)); os.IsNotExist(err) {
			f := p.exportFile

			actions := []io.Action{f.Open(os.O_WRONLY, 0660), f.WriteString(p.identifier, nil)}
			if err := io.Execute(actions, []io.Action{f.Close()}); err != nil {
				return fmt.Errorf("Could not export linux GPIO pin %v: %v", p.identifier, err)
			}
		}
		return nil
	}
}

// Direction of the pin is set
func (p *normalPin) Direction(direction bool) func() error {
	return func() error {
		d := "out"
		if direction {
			d = "in"
		}
		f := p.directionFile

		actions := []io.Action{f.Open(os.O_RDWR, 0660), f.WriteString(d, nil)}
		if err := io.Execute(actions, []io.Action{f.Close()}); err != nil {
			return fmt.Errorf("Could not set direction linux GPIO pin %v: %v", p.identifier, err)
		}
		return nil
	}

}

// Read the pin value
func (p *normalPin) Read(destination *bool) func() error {
	return func() error {
		f := p.valueFile
		rawValue := ""

		actions := []io.Action{f.Open(os.O_RDWR, 0660), f.ReadString(&rawValue, 1)}
		if err := io.Execute(actions, []io.Action{f.Close()}); err != nil {
			return fmt.Errorf("Could not read linux GPIO pin %v: %v", p.identifier, err)
		} else if v, err := strconv.ParseBool(rawValue); err != nil {
			return fmt.Errorf("Could not parse linux GPIO pin %v value: %v", p.identifier, err)
		} else {
			*destination = v
			return nil
		}
	}
}

// Write the pin value
func (p *normalPin) Write(value bool) func() error {
	return func() error {
		f := p.valueFile
		rawValue := "0"
		if value {
			rawValue = "1"
		}

		actions := []io.Action{f.Open(os.O_RDWR, 0660), f.WriteString(rawValue, nil)}

		if err := io.Execute(actions, []io.Action{f.Close()}); err != nil {
			return fmt.Errorf("Could not write linux GPIO pin %v: %v", p.identifier, err)
		}
		return nil
	}
}
