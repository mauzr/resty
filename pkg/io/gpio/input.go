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

	mauzrio "go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

// RawInputEvent represents a GPIO event returned by the kernel.
type RawInputEvent struct {
	Timestamp uint64
	ID        uint32
	_         uint32
}

// Input represents a general purpose input. Ask a chip instance to create one.
type Input interface {
	OpenForPoll() mauzrio.Action
	OpenForEvents(context.Context, *<-chan InputEvent) mauzrio.Action
	Close() mauzrio.Action
	Current(target *bool) mauzrio.Action
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
	file   file.File
}

func (i *input) OpenForEvents(ctx context.Context, eventsDestination *<-chan InputEvent) mauzrio.Action {
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
		if err := i.chip.file.Ioctl(ioctlRequest, uintptr(unsafe.Pointer(&r)))(); err != nil {
			return err
		}
		file := file.NewFromFd(r.fd, fmt.Sprintf("gpio-%v", i.number))
		i.file = file

		go func() {
			defer close(events)
			defer file.Close()
			for {
				var rawEvent RawInputEvent
				if err := i.file.ReadBinary(binary.LittleEndian, &rawEvent)(); err != nil {
					panic(err)
				}

				event := InputEvent{When: time.Unix(0, int64(rawEvent.Timestamp))}
				if rawEvent.ID == 1 {
					event.NewValue = true
				}
				select {
				case events <- event:
				case <-ctx.Done():
					return
				}
			}
		}()

		return nil
	}
}

func (i *input) OpenForPoll() mauzrio.Action {
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

	return func() error {
		if err := i.chip.file.Ioctl(ioctl, uintptr(unsafe.Pointer(&r)))(); err != nil {
			return err
		}
		i.file = file.NewFromFd(r.fd, fmt.Sprintf("gpio-%v", i.number))
		return nil
	}
}

func (i *input) Current(target *bool) mauzrio.Action {
	ioctlRead := file.IoctlRequestNumber(true, true, unsafe.Sizeof([64]uint8{}), 0xb4, 8)
	return func() error {
		values := [64]uint8{}
		if err := i.file.Ioctl(ioctlRead, uintptr(unsafe.Pointer(&values[0])))(); err != nil {
			return err
		}
		*target = values[0] != 0
		return nil
	}
}

func (i *input) Close() mauzrio.Action {
	return func() (err error) {
		if i.file != nil {
			err = i.file.Close()()
			if err == nil {
				i.file = nil
			}
		}
		return
	}
}

func (c *chip) NewInput(number uint32, active bool) Input {
	return &input{c, number, active, nil}
}
