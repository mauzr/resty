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

	"mauzr.eqrx.net/go/pkg/io"
)

// File represents a file.
type File interface {
	Open(int, os.FileMode) io.Action
	Close() io.Action
	Write([]byte, *int) io.Action
	WriteString(string, *int) io.Action
	WriteBinary(order binary.ByteOrder, data interface{}) io.Action
	Seek(int64, int, *int64) io.Action
	Read([]byte, *int) io.Action
	ReadString(*string, int) io.Action
	Ioctl(uintptr, uintptr) io.Action
	Unmap(*[]byte) io.Action
	Map(int64, int, int, int, *[]byte) io.Action
}

type normalFile struct {
	path   string
	handle *os.File
}

// NewFile creates a new Device. this function can be overridden to mock the device.
var NewFile = newNormalFile

func newNormalFile(path string) File {
	return &normalFile{path: path}
}

func (f *normalFile) Fd() uintptr {
	return f.handle.Fd()
}

// Open the file.
func (f *normalFile) Open(flags int, mask os.FileMode) io.Action {
	return func() error {
		h, err := os.OpenFile(f.path, flags, mask)
		if err != nil {
			return fmt.Errorf("Could not not open file %v with flags %v and mask %v: %v", f.path, flags, mask, err)
		}
		f.handle = h
		return nil
	}
}

// Unmap file from memory
func (f *normalFile) Unmap(memoryMap *[]byte) io.Action {
	return func() error {
		defer func() { *memoryMap = nil }()
		if err := unix.Munmap(*memoryMap); err != nil {
			return fmt.Errorf("Could not unmap memory: %v", err)
		}
		return nil
	}
}

// Map file to memory
func (f *normalFile) Map(offset int64, length, prot, flags int, memoryMap *[]byte) io.Action {
	return func() error {
		if mmem, err := unix.Mmap(int(f.Fd()), offset, length, prot, flags); err == nil {
			*memoryMap = mmem
		} else {
			return fmt.Errorf("Could not map memory: %v", err)
		}
		return nil
	}
}

// Close the file.
func (f *normalFile) Close() io.Action {
	return func() error {
		err := f.handle.Close()
		f.handle = nil
		if err != nil {
			return fmt.Errorf("Could not not close file %v: %v", f.path, err)
		}
		return nil
	}
}

// Write bytes to the file.
func (f *normalFile) Write(data []byte, count *int) io.Action {
	return func() error {
		c, err := f.handle.Write(data)
		if err != nil {
			return fmt.Errorf("Could not write %v to file %v: %v", data, f.path, err)
		}
		if count != nil {
			*count = c
		}
		return nil
	}
}

// Write bytes to the file.
func (f *normalFile) WriteBinary(order binary.ByteOrder, data interface{}) io.Action {
	return func() error {
		err := binary.Write(f.handle, order, data)
		if err != nil {
			return fmt.Errorf("Could not binary write %v to file %v: %v", data, f.path, err)
		}
		return nil
	}
}

// WriteString to the file.
func (f *normalFile) WriteString(data string, count *int) io.Action {
	return func() error {
		c, err := f.handle.WriteString(data)
		if err != nil {
			return fmt.Errorf("Could not write %v to file %v: %v", data, f.path, err)
		}
		if count != nil {
			*count = c
		}
		return nil
	}
}

// Seek in the file.
func (f *normalFile) Seek(offset int64, whence int, new *int64) io.Action {
	return func() error {
		n, err := f.handle.Seek(offset, whence)
		if err != nil {
			return fmt.Errorf("Could not seek to %v as %v in file %v: %v", offset, whence, f.path, err)
		}
		if new != nil {
			*new = n
		}
		return nil
	}
}

// Read from the file.
func (f *normalFile) Read(destination []byte, count *int) io.Action {
	return func() error {
		c, err := f.handle.Read(destination)
		if err != nil {
			return fmt.Errorf("Could not read #%v from file %v: %v", len(destination), f.path, err)
		}
		if count != nil {
			*count = c
		}
		return nil
	}
}

// Read from the file.
func (f *normalFile) ReadString(destination *string, length int) io.Action {
	buf := make([]byte, length)
	return func() error {
		err := f.Read(buf, nil)()
		if err != nil {
			return fmt.Errorf("Could not read #%v from file %v: %v", length, f.path, err)
		}
		*destination = string(buf)
		return nil
	}
}

// Ioctl execute an IOCTL command.
func (f *normalFile) Ioctl(operation, argument uintptr) io.Action {
	return func() error {
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.handle.Fd(), operation, argument); errno != 0 {
			return fmt.Errorf("Ioctl %v failed with handle %v and argument %v: %v", operation, f.handle.Fd(), argument, errno)
		}
		return nil
	}
}
