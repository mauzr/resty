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

import (
	"fmt"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/file"
)

// Output represents a general purpose output. Ask a chip instance to create one.
type Output interface {
	Open() error
	Set(value bool) func() error
	Close() error
}

type output struct {
	setup func() error
	file  file.File
}

func (o *output) Open() error {
	return o.setup()
}

func (o *output) Close() error {
	if o.file == nil {
		return nil
	}
	err := o.file.Close()
	if err == nil {
		o.file = nil
	}
	return err
}

func (o *output) Set(value bool) func() error {
	ioctl := file.IoctlRequestNumber(true, true, unsafe.Sizeof([64]uint8{}), 0xb4, 9)
	return func() error {
		payload := [64]uint8{}
		if value {
			payload[0] = 1
		}
		return o.file.IoctlPointerArgument(ioctl, unsafe.Pointer(&payload[0]))()
	}
}

func (c *chip) NewOutput(number uint32, active bool, value bool) Output {
	flags := flagOutput
	if !active {
		flags |= flagActiveLow
	}
	var intValue uint8 = 0
	if value {
		intValue = 1
	}
	r := struct {
		offsets  [64]uint32
		flags    uint32
		defaults [64]uint8
		consumer [32]byte
		amount   uint32
		fd       uintptr
	}{
		[64]uint32{number},
		flags,
		[64]uint8{intValue},
		[32]byte{},
		1,
		0,
	}
	copy(r.consumer[:], []byte("mauzr"))

	ioctl := file.IoctlRequestNumber(true, true, unsafe.Sizeof(r), 0xb4, 3)

	o := output{}
	o.setup = func() error {
		if err := c.file.IoctlPointerArgument(ioctl, unsafe.Pointer(&r))(); err != nil {
			return err
		}
		o.file = file.NewFromFd(r.fd, fmt.Sprintf("gpio-%v", number))
		return nil
	}
	return &o
}
