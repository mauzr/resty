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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.eqrx.net/mauzr/pkg/etc/pixels/strip"
	"go.eqrx.net/mauzr/pkg/io/rest"
)

// PostStrip posts the values of a strip to a REST endpoint.
func PostStrip(wg *sync.WaitGroup, input strip.Input, url string, c rest.REST) {
	refresh := time.Second / 30
	wg.Add(1)
	defer wg.Done()

	channels := make(chan []uint8)
	go StripToGRBWBytes(input, channels)

	timer := time.NewTicker(refresh)
	defer timer.Stop()

	for {
		<-timer.C
		channels, ok := <-channels
		if !ok {
			return
		}
		data, err := json.Marshal(channels)
		if err != nil {
			panic(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		response, err := c.PostRaw(ctx, url, bytes.NewReader(data))
		if err != nil {
			fmt.Printf("error setting colors: %v\n", err)
		} else {
			response.Body.Close()
		}
	}
}
