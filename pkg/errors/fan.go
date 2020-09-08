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

package errors

import (
	"sync"
)

// FanIn multiple error channels into one output channel.
func FanIn(errors ...<-chan error) <-chan error {
	merged := make(chan error)
	FanInto(merged, errors...)

	return merged
}

// FanInto multiple error channels a given output channel.
func FanInto(merged chan<- error, errors ...<-chan error) {
	wg := sync.WaitGroup{}
	wg.Add(len(errors))
	for _, c := range errors {
		go func(errs <-chan error) {
			for {
				err, ok := <-errs
				if ok {
					merged <- err
				} else {
					merged <- ErrChannelClosed
					wg.Done()

					return
				}
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(merged)
	}()
}
