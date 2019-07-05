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

import "os"

// DirectDevice connect to a device behind an SPI bus
type DirectDevice struct {
	bus   *os.File
	speed uint32
}

// AttachDirectDevice connects to a device behind an SPI bus
func attachDirectDevice(path string) (Device, error) {
	bus, err := openLinuxBus(path)
	device := DirectDevice{bus: bus}
	return device, err
}

// Close closes device
func (d DirectDevice) Close() error {
	return closeLinuxDevice(d.bus)
}

// Exchange sends data to an SPI device while receiving the same amount of data
func (d DirectDevice) Exchange(mosi []byte, miso []byte) error {
	return exchangeLinux(d.bus, d.speed, mosi, miso)
}
