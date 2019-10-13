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

package sk6812

import (
	"encoding/binary"
	"fmt"
	"os"

	"go.eqrx.net/mauzr/pkg/io"
	"go.eqrx.net/mauzr/pkg/io/file"
)

func apply(tty string, setting []uint8) error {
	f := file.NewFile(tty)

	actions := []io.Action{
		f.Open(os.O_RDWR|os.O_CREATE, 0644),
		f.WriteBinary(binary.LittleEndian, uint16(len(setting))),
		f.WriteBinary(binary.LittleEndian, setting),
	}
	if err := io.Execute(actions, []io.Action{f.Close()}); err != nil {
		return fmt.Errorf("could set write setting: %v", err)
	}
	return nil
}
