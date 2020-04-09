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
	"encoding/binary"
	"fmt"
	"os"

	"golang.org/x/sys/unix"

	"go.eqrx.net/mauzr/pkg/io"
)

// File represents a file.
type File interface {
	// Open the file.
	Open(int, os.FileMode) io.Action
	// Close the file.
	Close() io.Action
	// Write writes bytes to the file.
	Write([]byte) io.Action
	// WriteString writes a string to the file.
	WriteString(string) io.Action
	// WriteBinary uses binary.Write to write and interface to the file.
	WriteBinary(order binary.ByteOrder, data interface{}) io.Action
	// SeekTo a location in the file.
	SeekTo(int64) io.Action
	// Read bytes from the file.
	Read([]byte) io.Action
	// ReadString reads a string from the file.
	ReadString(*string, int) io.Action
	// Ioctl execute an IOCTL command.
	Ioctl(uintptr, uintptr) io.Action
	// Unmap file from memory.
	Unmap(*[]byte) io.Action
	// Map file to memory.
	Map(int64, int, int, int, *[]byte) io.Action
}

// file is just a plain old os.File.
type file struct {
	path   string
	handle *os.File
}

// New create a new file representation.
func New(path string) File {
	return &file{path: path}
}

// Open the file.
func (f *file) Open(flags int, mask os.FileMode) io.Action {
	return func() error {
		h, err := os.OpenFile(f.path, flags, mask)
		if err != nil {
			return fmt.Errorf("could not not open file %v with flags %v and mask %v: %v", f.path, flags, mask, err)
		}
		f.handle = h
		return nil
	}
}

// Unmap file from memory.
func (f *file) Unmap(memoryMap *[]byte) io.Action {
	return func() error {
		defer func() { *memoryMap = nil }()
		if err := unix.Munmap(*memoryMap); err != nil {
			return fmt.Errorf("could not unmap memory: %v", err)
		}
		return nil
	}
}

// Map file to memory.
func (f *file) Map(offset int64, length, prot, flags int, memoryMap *[]byte) io.Action {
	return func() error {
		if mmem, err := unix.Mmap(int(f.handle.Fd()), offset, length, prot, flags); err == nil {
			*memoryMap = mmem
		} else {
			return fmt.Errorf("could not map memory: %v", err)
		}
		return nil
	}
}

// Close the file.
func (f *file) Close() io.Action {
	return func() error {
		err := f.handle.Close()
		f.handle = nil
		if err != nil {
			return fmt.Errorf("could not not close file %v: %v", f.path, err)
		}
		return nil
	}
}

// Write writes bytes to the file.
func (f *file) Write(data []byte) io.Action {
	return func() error {
		_, err := f.handle.Write(data)
		if err != nil {
			return fmt.Errorf("could not write %v to file %v: %v", data, f.path, err)
		}
		return nil
	}
}

// WriteBinary uses binary.Write to write and interface to the file.
func (f *file) WriteBinary(order binary.ByteOrder, data interface{}) io.Action {
	return func() error {
		err := binary.Write(f.handle, order, data)
		if err != nil {
			return fmt.Errorf("could not binary write %v to file %v: %v", data, f.path, err)
		}
		return nil
	}
}

// WriteString writes a string to the file.
func (f *file) WriteString(data string) io.Action {
	return func() error {
		_, err := f.handle.WriteString(data)
		if err != nil {
			return fmt.Errorf("could not write %v to file %v: %v", data, f.path, err)
		}
		return nil
	}
}

// SeekTo a location in the file.
func (f *file) SeekTo(offset int64) io.Action {
	return func() error {
		_, err := f.handle.Seek(offset, 0)
		if err != nil {
			return fmt.Errorf("could not seek to %v in file %v: %v", offset, f.path, err)
		}
		return nil
	}
}

// Read bytes from the file.
func (f *file) Read(destination []byte) io.Action {
	return func() error {
		_, err := f.handle.Read(destination)
		if err != nil {
			return fmt.Errorf("could not read #%v from file %v: %v", len(destination), f.path, err)
		}
		return nil
	}
}

// ReadString reads a string from the file.
func (f *file) ReadString(destination *string, length int) io.Action {
	buf := make([]byte, length)

	return func() error {
		err := f.Read(buf)()
		if err != nil {
			return fmt.Errorf("could not read #%v from file %v: %v", length, f.path, err)
		}
		*destination = string(buf)
		return nil
	}
}

// Ioctl execute an IOCTL command.
func (f *file) Ioctl(operation, argument uintptr) io.Action {
	return func() error {
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.handle.Fd(), operation, argument); errno != 0 {
			return fmt.Errorf("ioctl %v failed with handle %v: %v", operation, f.handle.Name(), errno)
		}
		return nil
	}
}
