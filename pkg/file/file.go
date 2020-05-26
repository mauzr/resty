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

// Package file contains tools to run operations in batch.
package file

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

// File represents a file.
type File interface {
	// Open the file.
	Open(int, os.FileMode) func() error
	// Close the file.
	Close() error
	// Write writes bytes to the file.
	Write([]byte) func() error
	// WriteString writes a string to the file.
	WriteString(string) func() error
	// WriteBinary uses binary.Write to write an interface to the file.
	WriteBinary(order binary.ByteOrder, data interface{}) func() error
	// SeekTo a location in the file.
	SeekTo(int64) func() error
	// Read bytes from the file.
	Read([]byte) func() error
	// ReadString reads a string from the file.
	ReadString(*string, int) func() error
	// ReadBinary uses binary.Write to read an interface from the file.
	ReadBinary(order binary.ByteOrder, data interface{}) func() error
	// IoctlGeneric execute an IOCTL command with uintptr as argument.
	IoctlGenericArgument(request, argument uintptr) func() error
	// IoctlGeneric execute an IOCTL command with uintptr as argument.
	IoctlPointerArgument(request uintptr, argument unsafe.Pointer) func() error
	// Unmap file from memory.
	Unmap(*[]byte) func() error
	// Map file to memory.
	Map(int64, int, int, int, *[]byte) func() error
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

// NewFromFd creates a new file from the given file descriptor.
func NewFromFd(fd uintptr, name string) File {
	f := file{name, os.NewFile(fd, name)}
	if f.handle == nil {
		panic("passed fd is invalid")
	}
	return &f
}

// Open the file.
func (f *file) Open(flags int, mask os.FileMode) func() error {
	return func() error {
		h, err := os.OpenFile(f.path, flags, mask)
		if err != nil {
			return fmt.Errorf("could not not open file %v with flags %v and mask %v: %w", f.path, flags, mask, err)
		}
		f.handle = h
		return nil
	}
}

// Close the file.
func (f *file) Close() error {
	err := f.handle.Close()
	f.handle = nil
	if err != nil {
		return fmt.Errorf("could not not close file %v: %w", f.path, err)
	}
	return nil
}

// Write writes bytes to the file.
func (f *file) Write(data []byte) func() error {
	return func() error {
		_, err := f.handle.Write(data)
		if err != nil {
			return fmt.Errorf("could not write %v to file %v: %w", data, f.path, err)
		}
		return nil
	}
}

// WriteBinary uses binary.Write to write an interface to the file.
func (f *file) WriteBinary(order binary.ByteOrder, data interface{}) func() error {
	return func() error {
		err := binary.Write(f.handle, order, data)
		if err != nil {
			return fmt.Errorf("could not binary write %v to file %v: %w", data, f.path, err)
		}
		return nil
	}
}

// WriteString writes a string to the file.
func (f *file) WriteString(data string) func() error {
	return func() error {
		_, err := f.handle.WriteString(data)
		if err != nil {
			return fmt.Errorf("could not write %v to file %v: %w", data, f.path, err)
		}
		return nil
	}
}

// SeekTo a location in the file.
func (f *file) SeekTo(offset int64) func() error {
	return func() error {
		_, err := f.handle.Seek(offset, 0)
		if err != nil {
			return fmt.Errorf("could not seek to %v in file %v: %w", offset, f.path, err)
		}
		return nil
	}
}

// Read bytes from the file.
func (f *file) Read(destination []byte) func() error {
	return func() error {
		_, err := f.handle.Read(destination)
		if err != nil {
			return fmt.Errorf("could not read #%v from file %v: %w", len(destination), f.path, err)
		}
		return nil
	}
}

// ReadString reads a string from the file.
func (f *file) ReadString(destination *string, length int) func() error {
	buf := make([]byte, length)

	return func() error {
		err := f.Read(buf)()
		if err != nil {
			return fmt.Errorf("could not read #%v from file %v: %w", length, f.path, err)
		}
		*destination = string(buf)
		return nil
	}
}

// ReadBinary uses binary.Write to read an interface from the file.
func (f *file) ReadBinary(order binary.ByteOrder, data interface{}) func() error {
	return func() error {
		err := binary.Read(f.handle, order, data)
		if err != nil {
			return fmt.Errorf("could not binary read %v from file %v: %w", data, f.path, err)
		}
		return nil
	}
}
