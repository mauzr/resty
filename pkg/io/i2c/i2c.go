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

// Package i2c interface with I2C devices.
package i2c

import (
	"fmt"
	"os"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

// Device represents a device behind an I2C bus.
type Device interface {
	// Open the connection to the device.
	Open() io.Action
	// Close the connection to the device.
	Close() io.Action
	// Write to an I2C device.
	Write(source ...byte) io.Action
	// WriteRead execute an I2C write followed by a read in the same transaction.
	WriteRead(source []byte, destination []byte) io.Action
}

// operation represents an i2c operation.
type operation struct {
	// addr is the I2C device address.
	addr uint16 //nolint
	// flags for the operation (like, is this a read or write).
	flags uint16 //nolint
	// len is the data length.
	len uint16 //nolint
	// buf points to the data.
	buf uintptr //nolint
}

// operations is a list of operation that C understands.
type operations struct {
	msgs  uintptr
	nmsgs uint32
}

// device presents a device behind an I2C bus.
type device struct {
	file    file.File
	address uint16
}

// Open the connection to the device.
func (d *device) Open() io.Action {
	return d.file.Open(os.O_RDWR, 0660)
}

// Close the connection to the device.
func (d *device) Close() io.Action {
	return d.file.Close()
}

// WriteRead execute an I2C write followed by a read in the same transaction.
func (d *device) WriteRead(source []byte, destination []byte) io.Action {
	ioctl := uintptr(0x0707) // I2C IOCTL does not follow usual naming for some reason.
	return func() error {
		parts := []operation{
			{addr: d.address, flags: 0, len: uint16(len(source)), buf: uintptr(unsafe.Pointer(&source[0]))},           // write
			{addr: d.address, flags: 1, len: uint16(len(destination)), buf: uintptr(unsafe.Pointer(&destination[0]))}, // read
		}
		msg := operations{msgs: uintptr(unsafe.Pointer(&parts[0])), nmsgs: 2}

		if err := d.file.Ioctl(ioctl, uintptr(unsafe.Pointer(&msg)))(); err != nil {
			return fmt.Errorf("failed to write %v and read #%v to I2C address %v because: %w", source, len(destination), d.address, err)
		}
		return nil
	}
}

// Write to an I2C device.
func (d *device) Write(source ...byte) io.Action {
	ioctl := uintptr(0x0707) // I2C IOCTL does not follow usual naming for some reason.
	return func() error {
		parts := []operation{
			{addr: d.address, flags: 0, len: uint16(len(source)), buf: uintptr(unsafe.Pointer(&source[0]))},
		}
		msg := operations{msgs: uintptr(unsafe.Pointer(&parts[0])), nmsgs: 1}

		if err := d.file.Ioctl(ioctl, uintptr(unsafe.Pointer(&msg)))(); err != nil {
			return fmt.Errorf("failed to write %v to I2C address %v: %w", source, d.address, err)
		}
		return nil
	}
}

// New creates a new I2C device.
func new(path string, address uint16) Device {
	return &device{file: file.New(path), address: address}
}

// New can be overridden for test mockups.
var New = new
