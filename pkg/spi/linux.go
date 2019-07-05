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

	"golang.org/x/sys/unix"
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

func openLinuxBus(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		err = fmt.Errorf("Could not open SPI bus %v: %v", path, err)
	}
	return file, err
}

func closeLinuxDevice(bus *os.File) error {
	err := bus.Close()
	if err != nil {
		err = fmt.Errorf("Could not close SPI bus: %v", err)
	}
	return err
}

func exchangeLinux(bus *os.File, speed uint32, mosi []byte, miso []byte) error {
	if len(mosi) != len(miso) {
		panic(fmt.Sprintf("SPI MOSI and MISO arrays have different lengths (%v %v)/n", len(mosi), len(miso)))
	}
	arg := spiIoctlArg{
		txBuf:   uint64(uintptr(unsafe.Pointer(&miso[0]))),
		rxBuf:   uint64(uintptr(unsafe.Pointer(&mosi[0]))),
		len:     uint32(len(mosi)),
		speedHz: speed,
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, bus.Fd(), ioctlSpiOperation, uintptr(unsafe.Pointer(&arg))); errno != 0 {
		return fmt.Errorf("SPI exchange failed with errno %v", errno)
	}
	return nil
}
