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

package sk6812

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/uart"
	"golang.org/x/sys/unix"
)

// Strip represents a SK6812 strip.
type Strip struct {
	port     uart.Port
	requests chan setRequest
}

type setRequest struct {
	channels chan []uint8
	result   chan error
}

// NewStrip creates a new SK6812 strip manager.
func NewStrip(path string) *Strip {
	return &Strip{uart.NewPort(path, unix.B115200), make(chan setRequest)}
}

func (m *Strip) sendChannels(channels []uint8) io.Action {
	return func() error {
		actions := []io.Action{
			m.port.WriteBinary(binary.LittleEndian, uint16(len(channels))),
			m.port.Write(channels),
			m.port.ResetInput(),
		}
		return io.Execute(actions, []io.Action{})
	}
}

// Manage performs management operations until canceled.
func (m *Strip) Manage(ctx context.Context) {
	defer close(m.requests)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		lastLength := 0
		actions := []io.Action{
			m.port.Open(),
			m.port.ResetOutput(),
			m.port.DTR(false),
			m.port.RTS(true),
			io.Sleep(1 * time.Second),
			m.port.RTS(false),
			m.port.ResetInput(),
			func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					case request := <-m.requests:
						channels, valid := <-request.channels
						if !valid {
							continue
						}
						err := m.sendChannels(channels)()
						request.result <- err
						if err != nil {
							return err
						}
					}
				}
			},
		}
		err := io.Execute(actions, []io.Action{
			func() error {
				m.sendChannels(make([]uint8, lastLength))
				return nil
			},
			m.port.Close(),
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}

// Set lets the manager set the strip to the given channel setting.
func (m *Strip) Set(ctx context.Context, channels []uint8) error {
	request := setRequest{make(chan []uint8), make(chan error)}
	defer close(request.channels)
	defer close(request.result)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.requests <- request:
		request.channels <- channels
		return <-request.result
	}
}
