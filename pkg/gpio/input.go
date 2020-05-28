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
	"context"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/file"
)

// RawInputEvent represents a GPIO event returned by the kernel.
type RawInputEvent struct {
	Timestamp uint64
	ID        uint32
	_         uint32
}

// Input represents a general purpose input. Ask a chip instance to create one.
type Input interface {
	Events(context.Context, *<-chan InputEvent) func() error
	Current(target *bool) func() error
}

// InputEvent marks that the value of an input changed at the given time.
type InputEvent struct {
	When     time.Time `json:"when"`
	NewValue bool      `json:"new_value"`
}

type input struct {
	chip   *chip
	number uint32
	active bool
}

func (i *input) Events(ctx context.Context, eventsDestination *<-chan InputEvent) func() error {
	handleFlags := flagInput
	if !i.active {
		handleFlags |= flagActiveLow
	}
	r := struct {
		offset      uint32
		handleFlags uint32
		eventFlags  uint32
		consumer    [32]byte
		fd          uintptr
	}{
		i.number,
		handleFlags,
		0b11,
		[32]byte{},
		0,
	}
	copy(r.consumer[:], []byte("mauzr"))
	events := make(chan InputEvent)
	*eventsDestination = events

	ioctlRequest := file.IoctlRequestNumber(true, true, unsafe.Sizeof(r), 0xb4, 4)
	return func() error {
		if err := i.chip.file.IoctlPointerArgument(ioctlRequest, unsafe.Pointer(&r))(); err != nil {
			return err
		}
		f := file.NewFromFd(r.fd, fmt.Sprintf("gpio-%v", i.number))

		go func() {
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			defer close(events)
			for {
				var rawEvent RawInputEvent
				if err := f.ReadBinary(binary.LittleEndian, &rawEvent)(); err != nil {
					panic(err)
				}

				event := InputEvent{When: time.Unix(0, int64(rawEvent.Timestamp))}
				if rawEvent.ID == 1 {
					event.NewValue = true
				}
				select {
				case events <- event:
				case <-ctx.Done():
					if err := f.Close(); err != nil {
						panic(err)
					}
					return
				}
			}
		}()

		return nil
	}
}

func (i *input) Current(target *bool) func() error {
	return func() (err error) {
		flags := flagInput
		if !i.active {
			flags |= flagActiveLow
		}
		r := struct {
			offsets  [64]uint32
			flags    uint32
			defaults [64]uint8
			consumer [32]byte
			amount   uint32
			fd       uintptr
		}{
			[64]uint32{i.number},
			flags,
			[64]uint8{},
			[32]byte{},
			1,
			0,
		}
		copy(r.consumer[:], []byte("mauzr"))

		ioctl := file.IoctlRequestNumber(true, true, unsafe.Sizeof(r), 0xb4, 3)

		if err := i.chip.file.IoctlPointerArgument(ioctl, unsafe.Pointer(&r))(); err != nil {
			return err
		}
		f := file.NewFromFd(r.fd, fmt.Sprintf("gpio-%v", i.number))
		ioctlRead := file.IoctlRequestNumber(true, true, unsafe.Sizeof([64]uint8{}), 0xb4, 8)
		values := [64]uint8{}
		if err := f.IoctlPointerArgument(ioctlRead, unsafe.Pointer(&values[0]))(); err != nil {
			return err
		}
		*target = values[0] != 0
		return f.Close()
	}
}

func (c *chip) NewInput(number uint32, active bool) Input {
	return &input{c, number, active}
}
