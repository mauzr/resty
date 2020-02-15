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
	"time"

	"go.eqrx.net/mauzr/pkg/etc/pixels/strip"
)

// Feeder feeds a strip into an SK6812 strip.
func Feeder(ctx context.Context, wg *sync.WaitGroup, input strip.Input, bus string, address uint16) {
	data := make(chan []uint8)

	outputter := NewManager(bus, address)
	subCtx, subCancel := context.WithCancel(ctx)
	defer subCancel()
	go StripToGRBWBytes(input, data)
	go outputter.Manage(subCtx, wg)
	go func() {
		timer := time.NewTicker(time.Second / 10)
		for {
			<-timer.C
			data, ok := <-data
			if !ok {
				return
			}
			if ctx.Err() == nil {
				err := outputter.Set(ctx, data)
				if err != nil && err != ctx.Err() {
					fmt.Printf("set error: %v", err)
				}
			}
		}
	}()
}
