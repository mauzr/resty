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

// Package errors extends the generic errors package with error channel and batch functionality.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

var errChannelClosed = New("error channel closed")

// MultiError contains multiple errors (that may occurred independen from each other).
type MultiError struct {
	Errs []error
}

// Error returns the error as string.
func (m MultiError) Error() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "encountered multiple errors:\n")
	for _, e := range m.Errs {
		buf.WriteString(e.Error())
	}
	return buf.String()
}

// Is is imported from the stdlib errors package.
var Is = errors.Is

// As is imported from the stdlib errors package.
var As = errors.As

// New is imported from the stdlib errors package.
var New = errors.New
