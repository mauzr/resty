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

package spi

import (
	"fmt"
	"os"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

// Device represents a device behind an SPI bus.
type Device interface {
	// Open a connection to the device connection.
	Open() io.Action
	// Close the device connection.
	Close() io.Action
	// Exchange sends data to an SPI device while receiving the same amount of data.
	Exchange(mosi []byte, miso []byte) io.Action
}

const (
	ioctl = 0x40206b00
)

type operation struct {
	txBuf       uint64 //nolint
	rxBuf       uint64 //nolint
	len         uint32 //nolint
	speedHz     uint32 //nolint
	delayUsecs  uint16 //nolint
	bitsPerWord uint8  //nolint
	csChange    uint8  //nolint
	txNbits     uint8  //nolint
	rxNbits     uint8  //nolint
	pad         uint16 //nolint
}

type device struct {
	file  file.File
	speed uint32
}

// Open a connection to the device connection.
func (d *device) Open() io.Action {
	return d.file.Open(os.O_RDWR|os.O_SYNC, 0660)
}

// Close the device connection.
func (d *device) Close() io.Action {
	return d.file.Close()
}

// Exchange sends data to an SPI device while receiving the same amount of data.
func (d *device) Exchange(mosi []byte, miso []byte) io.Action {
	if len(mosi) != len(miso) {
		panic(fmt.Sprintf("SPI MOSI and MISO arrays have different lengths (%v %v)/n", len(mosi), len(miso)))
	}

	arg := operation{
		txBuf:   uint64(uintptr(unsafe.Pointer(&miso[0]))),
		rxBuf:   uint64(uintptr(unsafe.Pointer(&mosi[0]))),
		len:     uint32(len(mosi)),
		speedHz: d.speed,
	}
	return d.file.Ioctl(ioctl, uintptr(unsafe.Pointer(&arg)))
}

// New creates an new SPI device.
func New(path string) Device {
	return &device{file: file.New(path)}
}
