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

	"mauzr.eqrx.net/go/pkg/io"
	"mauzr.eqrx.net/go/pkg/io/file"
)

const (
	ioctlSpiOperation = 0x40206b00
)

type spiIoctlArg struct {
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

type normalDevice struct {
	file  file.File
	speed uint32
}

// Open connection to the device.
func (d *normalDevice) Open() io.Action {
	return d.file.Open(os.O_RDWR|os.O_SYNC, 0660)
}

// Close the connection to the device.
func (d *normalDevice) Close() io.Action {
	return d.file.Close()
}

// Exchange sends data to an SPI device while receiving the same amount of data.
func (d *normalDevice) Exchange(mosi []byte, miso []byte) io.Action {
	if len(mosi) != len(miso) {
		panic(fmt.Sprintf("SPI MOSI and MISO arrays have different lengths (%v %v)/n", len(mosi), len(miso)))
	}
	arg := spiIoctlArg{
		txBuf:   uint64(uintptr(unsafe.Pointer(&miso[0]))),
		rxBuf:   uint64(uintptr(unsafe.Pointer(&mosi[0]))),
		len:     uint32(len(mosi)),
		speedHz: d.speed,
	}
	return d.file.Ioctl(ioctlSpiOperation, uintptr(unsafe.Pointer(&arg)))
}

func newNormalDevice(path string) Device {
	return &normalDevice{file: file.NewFile(path)}
}
