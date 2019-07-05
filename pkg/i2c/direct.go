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
	"os"
)

type directDevice struct {
	bus     *os.File
	address DeviceAddress
}

// Write represents a write action to the device
type directWrite struct {
	directDevice
	source []byte
}

// Read represents the reading of a device register to a sice
type directWriteRead struct {
	directDevice
	source      []byte
	destination []byte
}

// AttachDirectDevices create an Device handle for the given bus and address
func attachDirectDevice(path string, address DeviceAddress) (Device, error) {
	bus, err := openLinuxBus(path, address)
	if err != nil {
		return nil, err
	}
	return &directDevice{bus: bus, address: address}, nil
}

// WriteRead creates an action that executes a write followed by a read in the same I2C transaction
func (device directDevice) WriteRead(source []byte, destination []byte) Action {
	return &directWriteRead{directDevice: device, source: source, destination: destination}
}

// Write creates an action that writes the given bytes to the I2c device
func (device directDevice) Write(source []byte) Action {
	return &directWrite{directDevice: device, source: source}
}

// Close closes the given Device
func (device directDevice) Close() error {
	return closeLinuxBus(device.bus)
}

// Execute executes a write followed by a read in the same I2C transaction
func (action directWriteRead) Execute() error {
	return readRegisterLinux(action.bus, action.address, action.source, action.destination)
}

// Execute writes the given bytes to the I2c device
func (action directWrite) Execute() error {
	return writeLinux(action.bus, action.address, action.source)
}
