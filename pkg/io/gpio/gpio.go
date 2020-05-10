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

// Package gpio interface with GPIO pins.
package gpio

import (
	"os"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

const (
	flagInput     uint32 = 0b00000001
	flagOutput    uint32 = 0b00000010
	flagActiveLow uint32 = 0b00000100
)

// Chip represents an gpio chip. You can request I/O lines from it.
type Chip interface {
	Open() io.Action
	NewInput(number uint32, active bool) Input
	NewOutput(number uint32, active bool, value bool) Output
}

type chip struct {
	file file.File
}

func (c *chip) Open() io.Action {
	return c.file.Open(os.O_RDWR, 0660)
}

// NewChip creates a new representation for the GPIO chip behind the given block device.
func NewChip(path string) Chip {
	return &chip{file.New(path)}
}
