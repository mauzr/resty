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

// Package errors extends the generic errors package with error channel functionality.
package errors

import (
	"errors"
	"fmt"
	"sync"
)

// MultiError contains multiple errors (that may occurred independen from each other).
type MultiError struct {
	Errs []error
}

// Error returns the error as string.
func (m MultiError) Error() string {
	return fmt.Sprintf("encountered multiple errors: %v", m.Errs)
}

// Is is imported from the stdlib errors package.
var Is = errors.Is

// As is imported from the stdlib errors package.
var As = errors.As

// New is imported from the stdlib errors package.
var New = errors.New

// Merge multiple error channels into one output channel. The given onError function is called on every received error (may be nil).
func Merge(onError func(), errors ...<-chan error) <-chan error {
	merged := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(errors))
	for _, c := range errors {
		go func(errs <-chan error) {
			defer wg.Done()
			for {
				err, ok := <-errs
				if onError != nil {
					onError()
				}
				if !ok {
					return
				}
				merged <- err
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(merged)
	}()
	return merged
}

// Collect all errors from an error channel and from the parameter list and merge them into one error.
func Collect(source <-chan error, additional ...error) error {
	errors := []error{}
	if additional != nil {
		errors = append(errors, additional...)
	}
	if source != nil {
		for {
			err, ok := <-source
			if ok {
				break
			}
			errors = append(errors, err)
		}
	}
	var err error
	switch len(errors) {
	case 0:
	case 1:
		err = errors[0]
	default:
		err = &MultiError{errors}
	}
	return err
}
