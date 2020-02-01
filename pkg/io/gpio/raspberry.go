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
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

// fetchGpioBase returns the base memory address of the GPIO chip.
func fetchGpioBase(base *uint32) io.Action {
	return func() error {
		f := file.New("/proc/device-tree/soc/ranges")
		buffer := make([]byte, 4)
		if err := io.Execute([]io.Action{f.Open(os.O_RDONLY, 0600), f.SeekTo(4), f.Read(buffer)}, []io.Action{f.Close()}); err != nil {
			return fmt.Errorf("could not read GPIO memory range: %v", err)
		}
		if err := binary.Read(bytes.NewReader(buffer), binary.BigEndian, base); err != nil {
			panic(err)
		}
		*base += 0x200000
		return nil
	}
}

// Pull sets the pull up/down setting of the pin.
func (p *pin) Pull(direction uint) func() error {
	return func() error {
		identifier, err := strconv.ParseUint(p.identifier, 10, 32)
		if err != nil {
			return fmt.Errorf("illegal GPIO identifier: %v", identifier)
		}
		if direction > 2 {
			return fmt.Errorf("illegal pull for identifier: %v", identifier)
		}

		clockRegister := identifier/32 + 38
		pullRegister := 37
		var gpioBase uint32
		var gpioMem []uint32

		set := func() error {
			gpioMem[pullRegister] = uint32(direction)
			time.Sleep(10 * time.Microsecond)
			gpioMem[clockRegister] = 1 << (identifier % 32)
			time.Sleep(10 * time.Microsecond)
			gpioMem[pullRegister] &^= 3
			gpioMem[clockRegister] = 0
			return nil
		}

		f := p.gpioMemFile
		actions := []io.Action{fetchGpioBase(&gpioBase), f.Open(int64(gpioBase), 4096), f.Uint32Register(&gpioMem), set}
		if err := io.Execute(actions, []io.Action{f.Close()}); err != nil {
			return fmt.Errorf("could set GPIO pull: %v", err)
		}

		return nil
	}
}
