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

	"go.eqrx.net/mauzr/pkg/io"
	"golang.org/x/sys/unix"
)

// IoctlRequestNumber calculates an IOCTL request number based of some magic attributes.
func IoctlRequestNumber(read, write bool, argumentSize, group, number uintptr) uintptr {
	var direction uintptr
	if read {
		direction |= 0b10
	}
	if write {
		direction |= 0b01
	}
	return direction<<30 | argumentSize<<16 | group<<8 | number
}

// Ioctl execute an IOCTL command.
func (f *file) Ioctl(request, argument uintptr) io.Action {
	return func() error {
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.handle.Fd(), request, argument); errno != 0 {
			return fmt.Errorf("ioctl %v failed with handle %v: %w", request, f.handle.Name(), errno)
		}
		return nil
	}
}
