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

package pixels

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"go.eqrx.net/mauzr/pkg/pixels/strip"

	"unsafe"

	"go.eqrx.net/mauzr/pkg/io/file"
)

type operation struct {
	txBuf       uint64
	rxBuf       uint64 //nolint
	len         uint32 //nolint
	speedHz     uint32
	delayUsecs  uint16 //nolint
	bitsPerWord uint8  //nolint
	csChange    uint8  //nolint
	txNbits     uint8  //nolint
	rxNbits     uint8  //nolint
	pad         uint16 //nolint
}

func channelToByte(channel float64) uint8 {
	if channel < 0.0 || channel > 1.0 {
		panic(fmt.Sprintf("illegal channel value: %f", channel))
	}
	return uint8(255.0 * channel)
}

func createLut() ([][]byte, int) {
	translationFactor := 8
	lut := make([][]byte, 256)
	for dataByte := range lut {
		var translation uint64
		for dataBitPosition := 0; dataBitPosition < 8; dataBitPosition++ {
			dataBit := dataByte&(1<<dataBitPosition) != 0
			if dataBit {
				translation |= 0b11110000 << ((7 - dataBitPosition) * 8)
			} else {
				translation |= 0b11000000 << ((7 - dataBitPosition) * 8)
			}
		}
		lut[dataByte] = make([]byte, translationFactor)
		binary.LittleEndian.PutUint64(lut[dataByte], translation)
	}
	return lut, translationFactor
}

func New(input strip.Input, path string, framerate int) <-chan error {
	errors := make(chan error)
	go func() {
		defer close(errors)
		lut, translationFactor := createLut()

		ticker := time.NewTicker(time.Second / time.Duration(framerate))
		defer ticker.Stop()

		f := file.New(path)
		if err := f.Open(os.O_RDWR|os.O_SYNC, os.ModeDevice)(); err != nil {
			errors <- err
			return
		}
		defer func() {
			if err := f.Close()(); err != nil {
				errors <- err
			}
		}()

		for {
			<-ticker.C

			colors, ok := input.Get()
			if !ok {
				return
			}
			translated := make([]byte, 0, len(colors)*translationFactor)
			for i := range colors {
				translated = append(translated, lut[channelToByte(colors[i].Green)]...)
				translated = append(translated, lut[channelToByte(colors[i].Red)]...)
				translated = append(translated, lut[channelToByte(colors[i].Blue)]...)
				translated = append(translated, lut[channelToByte(colors[i].White)]...)
			}
			arg := operation{
				txBuf:   uint64(uintptr(unsafe.Pointer(&translated[0]))),
				len:     uint32(len(translated)),
				speedHz: 19000000,
			}
			err := f.Ioctl(0x40206b00, uintptr(unsafe.Pointer(&arg)))()
			if err != nil {
				errors <- err
				return
			}
		}
	}()
	return errors
}
