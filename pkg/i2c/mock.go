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

type mockDevice struct {
	bus     string
	address DeviceAddress
	handler MockDeviceHandler
}

type mockWrite struct {
	mockDevice
	source []byte
}

type mockWriteRead struct {
	mockDevice
	source      []byte
	destination []byte
}

// MockDeviceHandler handles queries from a mocked I2C device
type MockDeviceHandler interface {
	WriteRead(bus string, address DeviceAddress, source []byte, destination []byte) error
	OnWrite(bus string, address DeviceAddress, data []byte) error
}

// DefaultMockDeviceHandler will be assigned to new device handles when set
var DefaultMockDeviceHandler MockDeviceHandler

// AttachMockDevice creates a mock device for the given setting with the DefaultMockDeviceHandler
func attachMockDevice(bus string, address DeviceAddress) Device {
	if DefaultMockDeviceHandler != nil {
		return mockDevice{bus, address, DefaultMockDeviceHandler}
	}
	return nil
}

// Write returns a write action for the mock
func (device mockDevice) Write(source []byte) Action {
	return mockWrite{device, source}
}

// Execute executes the mock write action
func (action mockWrite) Execute() error {
	return action.handler.OnWrite(action.bus, action.address, action.source)
}

// WriteRead creates an action that consists of a write followes by a read
func (device mockDevice) WriteRead(source []byte, destination []byte) Action {
	return mockWriteRead{device, source, destination}
}

// Execute executes the write and read action
func (action mockWriteRead) Execute() error {
	return action.handler.WriteRead(action.bus, action.address, action.source, action.destination)
}

// Close does nothing
func (device mockDevice) Close() error {
	return nil
}
