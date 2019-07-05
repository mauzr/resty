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

	"golang.org/x/sys/unix"
)

const (
	ioctlReadOperation           = 0x0001
	ioctlReadWriteOperation      = 0x0707
	ioctlI2CSlaveSelectOperation = 0x0703 //nolint
)

type ioctlPart struct {
	addr  uint16  //nolint
	flags uint16  //nolint
	len   uint16  //nolint
	buf   uintptr //nolint
}

type ioctlArg struct {
	msgs  uintptr
	nmsgs uint32
}

func openLinuxBus(path string, address DeviceAddress) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		return nil, fmt.Errorf("Could not open I2C bus %v: %v", path, err)
	}
	defer func() {
		if lerr := file.Close(); err == nil {
			err = lerr
		}
	}()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Could not stat I2C bus %v: %v", path, err)
	}

	if info.Mode()&os.ModeDevice == 0 {
		return nil, fmt.Errorf("I2C bus %v is not a device", path)
	}
	return file, nil
}

func closeLinuxBus(bus *os.File) error {
	if err := bus.Close(); err != nil {
		return fmt.Errorf("Could not close I2C bus: %v", err)
	}
	return nil
}

func readRegisterLinux(bus *os.File, address DeviceAddress, source []byte, destination []byte) error {
	amount := int16(len(destination))
	parts := []ioctlPart{
		{
			addr:  uint16(address),
			flags: 0,
			len:   uint16(len(source)),
			buf:   uintptr(unsafe.Pointer(&source[0])),
		},
		{
			addr:  uint16(address),
			flags: ioctlReadOperation,
			len:   uint16(amount),
			buf:   uintptr(unsafe.Pointer(&destination[0])),
		},
	}
	msg := ioctlArg{
		msgs:  uintptr(unsafe.Pointer(&parts[0])),
		nmsgs: uint32(len(parts)),
	}

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, bus.Fd(), ioctlReadWriteOperation, uintptr(unsafe.Pointer(&msg))); errno != 0 {
		return fmt.Errorf("Failed to read register %v from I2C address %v: %v", source, address, errno)
	}
	return nil
}

func writeLinux(bus *os.File, address DeviceAddress, source []byte) error {
	parts := []ioctlPart{
		{
			addr:  uint16(address),
			flags: 0,
			len:   uint16(len(source)),
			buf:   uintptr(unsafe.Pointer(&source[0])),
		},
	}

	msg := ioctlArg{
		msgs:  uintptr(unsafe.Pointer(&parts[0])),
		nmsgs: uint32(len(parts)),
	}

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, bus.Fd(), ioctlReadWriteOperation, uintptr(unsafe.Pointer(&msg))); errno != 0 {
		return fmt.Errorf("Failed to write %v to I2C address %v: %v", source, address, errno)
	}
	return nil
}

func writeLinuxSimple(bus *os.File, address DeviceAddress, source []byte) error { //nolint
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, bus.Fd(), ioctlI2CSlaveSelectOperation, uintptr(address)); errno != 0 {
		return fmt.Errorf("Failed to select I2C device %v because \"%v\"", address, errno)
	}
	if _, err := bus.Write(source); err != nil {
		return fmt.Errorf("Failed to write %v to I2C address %v: %v", source, address, err)
	}
	return nil
}
