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

// Package uart interfaces with devices behind UART.
package uart

import (
	"encoding/binary"
	"os"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/errors"
	"go.eqrx.net/mauzr/pkg/file"

	"golang.org/x/sys/unix"
)

// Port is an UART port handle.
type Port interface {
	// Open the connection to the port.
	Open() error
	// Close the port connection.
	Close() error
	// Write data over the port.
	Write([]byte) func() error
	// WriteBinary data over the port.
	WriteBinary(binary.ByteOrder, interface{}) func() error
	// RTS state setting.
	RTS(bool) func() error
	// DTR state setting.
	DTR(bool) func() error
	// ResetOutput purges UART output that hasn't been sent yet.
	ResetOutput() func() error
	// ResetInput purges UART input that hasn't been handled yet.
	ResetInput() func() error
}

type normalPort struct {
	file file.File
	baud uint32
	path string
}

// NewPort creates a new UART port handler.
func NewPort(path string, baud uint32) Port {
	return &normalPort{file.New(path), baud, path}
}

// Open a connection to the port.
func (p *normalPort) Open() error {
	settings := unix.Termios{}
	settings.Cflag |= unix.CS8 | p.baud
	settings.Cc[unix.VMIN] = 1
	settings.Cc[unix.VTIME] = 0
	return errors.NewBatch(
		p.file.Open(unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|os.O_RDWR, 0666),
		p.file.IoctlPointerArgument(unix.TCSETS, unsafe.Pointer(&settings)),
	).Execute("open uart")
}

// Close the port connection.
func (p *normalPort) Close() error {
	return p.file.Close()
}

// Write data over the port.
func (p *normalPort) Write(data []byte) func() error {
	return p.file.Write(data)
}

// WriteBinary data over the port.
func (p *normalPort) WriteBinary(order binary.ByteOrder, data interface{}) func() error {
	return p.file.WriteBinary(order, data)
}

// RTS state setting.
func (p *normalPort) RTS(value bool) func() error {
	var mask int = unix.TIOCM_RTS

	if value {
		return p.file.IoctlPointerArgument(unix.TIOCMBIS, unsafe.Pointer(&mask))
	}
	return p.file.IoctlPointerArgument(unix.TIOCMBIC, unsafe.Pointer(&mask))
}

// DTR state setting.
func (p *normalPort) DTR(value bool) func() error {
	var mask int = unix.TIOCM_DTR
	if value {
		return p.file.IoctlPointerArgument(unix.TIOCMBIS, unsafe.Pointer(&mask))
	}
	return p.file.IoctlPointerArgument(unix.TIOCMBIC, unsafe.Pointer(&mask))
}

// ResetOutput purges UART output that hasn't been sent yet.
func (p *normalPort) ResetOutput() func() error {
	return p.file.IoctlGenericArgument(unix.TCFLSH, unix.TCOFLUSH)
}

// ResetInput purges UART input that hasn't been handled yet.
func (p *normalPort) ResetInput() func() error {
	return p.file.IoctlGenericArgument(unix.TCFLSH, unix.TCIFLUSH)
}
