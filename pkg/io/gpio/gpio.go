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

package gpio

const (
	// Input means the pin is an input
	Input = true
	// Output means the pin is an output
	Output = false
	// PullUp means that the pin is pulled to VCC
	PullUp = 1
	// PullDown means that the pin is pulled to ground
	PullDown = 1
	// PullNone means that the pin is not pulled
	PullNone = 2
)

// Pin represents a GPIO pin
type Pin interface {
	Pull(direction uint) func() error
	Direction(input bool) func() error
	Export() func() error
	Read(destination *bool) func() error
	Write(source bool) func() error
}

// NewPin creates a new Pin. this function can be overridden to
// mock the device
var NewPin = newNormalPin
