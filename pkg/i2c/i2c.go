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

// Action represents an I2C acton
type Action interface {
	Execute() error
}

// Device is a handler for an device identified by a bus and an address
type Device interface {
	Write(source []byte) Action
	WriteRead(source []byte, destination []byte) Action
	Close() error
}

// DeviceAddress represents the I2C address of a device
type DeviceAddress uint16

// Execute the given action set and stop at errors
func Execute(actions []Action) error {
	for _, action := range actions {
		if err := action.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// AttachDevice checks available methods and connects to and I2C connected device
func AttachDevice(bus string, address DeviceAddress) (device Device, err error) {
	if device = attachMockDevice(bus, address); device != nil {
		return
	}
	device, err = attachDirectDevice(bus, address)

	return
}
