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
	"fmt"
	"sync"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/i2c"
)

const (
	I2CCommand = 0x01
)

type manager struct {
	device   i2c.Device
	requests chan setRequest
}

type setRequest struct {
	channels chan []uint8
	result   chan error
}

type Manager interface {
	Manage(context.Context, *sync.WaitGroup)
	Set(context.Context, []uint8) error
}

func NewManager(path string, address uint16) Manager {
	return &manager{i2c.NewDevice(path, address), make(chan setRequest)}
}

func (m *manager) Manage(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		actions := []io.Action{
			m.device.Open(),
			func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					case request := <-m.requests:
						channels, ok := <-request.channels
						if ok {
							data := append([]uint8{I2CCommand}, channels...)
							err := io.Execute([]io.Action{m.device.Write(data)}, []io.Action{})
							request.result <- err
							if err != nil {
								return err
							}
						}
					}
				}
			},
		}
		err := io.Execute(actions, []io.Action{
			m.device.Close(),
		})
		if err != nil {
			fmt.Println(fmt.Errorf("ignoring error while writing to sk6812 manager: %v", err))
		}
	}
}

// Set lets the manager set the strip to the given channel setting.
func (m *manager) Set(ctx context.Context, channels []uint8) error {
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
