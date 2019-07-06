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

package i2c

import (
	"fmt"
	"os"
	"unsafe"

	"mauzr.eqrx.net/go/pkg/io"
	"mauzr.eqrx.net/go/pkg/io/file"
)

const (
	ioctlReadOperation      = 0x0001
	ioctlReadWriteOperation = 0x0707
)

type ioctlArg struct {
	msgs  uintptr
	nmsgs uint32
}

type ioctlPart struct {
	addr  uint16  //nolint
	flags uint16  //nolint
	len   uint16  //nolint
	buf   uintptr //nolint
}

type normalDevice struct {
	file    file.File
	address uint16
}

// Open the connection to the device.
func (d *normalDevice) Open() io.Action {
	return d.file.Open(os.O_RDWR, 0660)
}

// Close the connection to the device
func (d *normalDevice) Close() io.Action {
	return d.file.Close()
}

// WriteRead execute an I2C write followed by a read in the same transaction.
func (d *normalDevice) WriteRead(source []byte, destination []byte) io.Action {
	return func() error {
		amount := int16(len(destination))
		parts := []ioctlPart{
			{addr: d.address, flags: 0, len: uint16(len(source)), buf: uintptr(unsafe.Pointer(&source[0]))},
			{addr: d.address, flags: ioctlReadOperation, len: uint16(amount), buf: uintptr(unsafe.Pointer(&destination[0]))},
		}
		msg := ioctlArg{
			msgs:  uintptr(unsafe.Pointer(&parts[0])),
			nmsgs: uint32(len(parts)),
		}

		if err := d.file.Ioctl(ioctlReadWriteOperation, uintptr(unsafe.Pointer(&msg)))(); err != nil {
			return fmt.Errorf("Failed to write %v and read #%v to I2C address %v because \"%v\"", source, len(destination), d.address, err)
		}
		return nil
	}
}

// Write to an I2C device.
func (d *normalDevice) Write(source []byte) io.Action {
	return func() error {
		parts := []ioctlPart{
			{
				addr:  uint16(0x77),
				flags: 0,
				len:   uint16(len(source)),
				buf:   uintptr(unsafe.Pointer(&source[0])),
			},
		}

		msg := ioctlArg{
			msgs:  uintptr(unsafe.Pointer(&parts[0])),
			nmsgs: uint32(len(parts)),
		}

		if err := d.file.Ioctl(ioctlReadWriteOperation, uintptr(unsafe.Pointer(&msg)))(); err != nil {
			return fmt.Errorf("Failed to write %v to I2C address %v because \"%v\"", source, d.address, err)
		}
		return nil
	}
}

func newNormalDevice(path string, address uint16) Device {
	return &normalDevice{file: file.NewFile(path), address: address}
}
