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
	"fmt"
)

// GoOnClose calls a callback in a goroutine when the error channel closed.
func GoOnClose(cb func(), in <-chan error) <-chan error {
	return OnClose(func() { go cb() }, in)
}

// OnClose calls a callback  when the error channel closed.
func OnClose(cb func(), in <-chan error) <-chan error {
	out := make(chan error)
	go func() {
		for {
			err, ok := <-in
			if ok {
				out <- err
			} else {
				cb()
				close(out)
				return
			}
		}
	}()
	return out
}

// GoOnFirstError calls a callback in a goroutine when the error channel receives the first error.
func GoOnFirstError(cb func(error), in <-chan error) <-chan error {
	return OnFirstError(func(err error) { go cb(err) }, in)
}

// OnFirstError calls a callback  when the error channel receives the first error.
func OnFirstError(cb func(error), in <-chan error) <-chan error {
	out := make(chan error)
	go func() {
		needsCalling := true
		for {
			err, ok := <-in
			if needsCalling {
				cb(err)
				needsCalling = false
			}
			if ok {
				out <- err
			} else {
				close(out)
				return
			}
		}
	}()
	return out
}

// WrapErrorChan wraps all errors in a given channel with the given name.
func WrapErrorChan(name string, in <-chan error) <-chan error {
	out := make(chan error)
	go func() {
		for {
			err, ok := <-in
			if !ok {
				close(out)
				return
			}
			if err == nil {
				panic(fmt.Sprintf("%s received nil as error", name))
			}
			out <- fmt.Errorf("%s: %w", name, err)
		}
	}()
	return out
}
