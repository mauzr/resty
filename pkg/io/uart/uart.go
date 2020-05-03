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

package uart

import (
	"encoding/binary"
	"os"
	"unsafe"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"

	"golang.org/x/sys/unix"
)

// Port is an UART port handle.
type Port interface {
	// Open the connection to the port.
	Open() io.Action
	// Close the port connection.
	Close() io.Action
	// Write data over the port.
	Write([]byte) io.Action
	// WriteBinary data over the port.
	WriteBinary(binary.ByteOrder, interface{}) io.Action
	// RTS state setting.
	RTS(bool) io.Action
	// DTR state setting.
	DTR(bool) io.Action
	// ResetOutput purges UART output that hasn't been sent yet.
	ResetOutput() io.Action
	// ResetInput purges UART input that hasn't been handled yet.
	ResetInput() io.Action
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
func (p *normalPort) Open() io.Action {
	return func() error {
		settings := unix.Termios{}
		settings.Cflag |= unix.CS8 | p.baud
		settings.Cc[unix.VMIN] = 1
		settings.Cc[unix.VTIME] = 0
		actions := []io.Action{
			p.file.Open(unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|os.O_RDWR, 0666),
			p.file.Ioctl(unix.TCSETS, uintptr(unsafe.Pointer(&settings))),
		}
		return io.Execute("open uart", actions, []io.Action{})
	}
}

// Close the port connection.
func (p *normalPort) Close() io.Action {
	return p.file.Close()
}

// Write data over the port.
func (p *normalPort) Write(data []byte) io.Action {
	return p.file.Write(data)
}

// WriteBinary data over the port.
func (p *normalPort) WriteBinary(order binary.ByteOrder, data interface{}) io.Action {
	return p.file.WriteBinary(order, data)
}

// RTS state setting.
func (p *normalPort) RTS(value bool) io.Action {
	var mask int = unix.TIOCM_RTS

	if value {
		return p.file.Ioctl(unix.TIOCMBIS, uintptr(unsafe.Pointer(&mask)))
	}
	return p.file.Ioctl(unix.TIOCMBIC, uintptr(unsafe.Pointer(&mask)))
}

// DTR state setting.
func (p *normalPort) DTR(value bool) io.Action {
	var mask int = unix.TIOCM_DTR
	if value {
		return p.file.Ioctl(unix.TIOCMBIS, uintptr(unsafe.Pointer(&mask)))
	}
	return p.file.Ioctl(unix.TIOCMBIC, uintptr(unsafe.Pointer(&mask)))
}

// ResetOutput purges UART output that hasn't been sent yet.
func (p *normalPort) ResetOutput() io.Action {
	return p.file.Ioctl(unix.TCFLSH, unix.TCOFLUSH)
}

// ResetInput purges UART input that hasn't been handled yet.
func (p *normalPort) ResetInput() io.Action {
	return p.file.Ioctl(unix.TCFLSH, unix.TCIFLUSH)
}
