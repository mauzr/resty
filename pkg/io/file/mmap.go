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

package file

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/io"
	"golang.org/x/sys/unix"
)

// MemoryMap represents a file that is mapped to a memory range.
type MemoryMap interface {
	Open(offset int64, length int) io.Action
	Close() io.Action
	Uint32Register(destination *[]uint32) io.Action
}

// NewMemoryMap creates a new MMap handle for the given file path.
func NewMemoryMap(path string) MemoryMap {
	return &normalMemoryMap{file: NewFile(path)}
}

type normalMemoryMap struct {
	file File
	mmap []byte
}

// Close releases the memory mapping
func (m *normalMemoryMap) Close() io.Action {
	return func() error {
		if err := m.file.Unmap(&m.mmap)(); err != nil {
			return fmt.Errorf("could not unmap memory: %v", err)
		}
		return nil
	}
}

// Open maps the file to memory
func (m *normalMemoryMap) Open(offset int64, length int) io.Action {
	return func() error {
		actions := []io.Action{
			m.file.Open(os.O_RDWR|os.O_SYNC, 0600),
			m.file.Map(offset, length, unix.PROT_WRITE|unix.PROT_READ, unix.MAP_SHARED, &m.mmap)}
		if err := io.Execute(actions, []io.Action{m.file.Close()}); err != nil {
			return fmt.Errorf("could not map file %v to memory: %v", m.file, err)
		}
		return nil
	}
}

// Uint32Register returns a slice the represents the map as uint32 array.
func (m *normalMemoryMap) Uint32Register(destination *[]uint32) io.Action {
	return func() error {
		header := *(*reflect.SliceHeader)(unsafe.Pointer(&m.mmap))
		header.Len /= 4
		header.Cap /= 4
		*destination = *(*[]uint32)(unsafe.Pointer(&header))
		return nil
	}
}
