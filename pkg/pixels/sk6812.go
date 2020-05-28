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

// Package pixels manages hardware pixels like SK6812 chains.
package pixels

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"time"

	"unsafe"

	"go.eqrx.net/mauzr/pkg/file"
	"go.eqrx.net/mauzr/pkg/log"
	"go.eqrx.net/mauzr/pkg/pixels/color"
)

type operation struct {
	txBuf       uint64
	rxBuf       uint64 //nolint:structcheck // Keep the name for future use.
	len         uint32
	speedHz     uint32
	delayUsecs  uint16 //nolint:structcheck // Keep the name for future use.
	bitsPerWord uint8  //nolint:structcheck // Keep the name for future use.
	csChange    uint8  //nolint:structcheck // Keep the name for future use.
	txNbits     uint8  //nolint:structcheck // Keep the name for future use.
	rxNbits     uint8  //nolint:structcheck // Keep the name for future use.
	pad         uint16 //nolint:structcheck // Keep the name for future use.
}

func channelToByte(channel float64) uint8 {
	if channel < 0.0 || channel > 1.0 {
		panic(fmt.Sprintf("illegal channel value: %f", channel))
	}
	return uint8(255.0 * channel)
}

func translate(colors []color.RGBW, translated []byte, lut [][]byte, translationFactor int) {
	for i := range colors {
		copy(translated[(i*4+0)*translationFactor:], lut[channelToByte(colors[i].Green())])
		copy(translated[(i*4+1)*translationFactor:], lut[channelToByte(colors[i].Red())])
		copy(translated[(i*4+2)*translationFactor:], lut[channelToByte(colors[i].Blue())])
		copy(translated[(i*4+3)*translationFactor:], lut[channelToByte(colors[i].White())])
	}
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

func determineSpeed() uint32 {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		panic(err)
	}

	var revision string

	matcher := regexp.MustCompile(`Revision\W+:\W+(\w+)`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := matcher.FindStringSubmatch(line)
		if parts != nil {
			revision = parts[1]
		}
	}
	_ = file.Close()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	speed, ok := map[string]uint32{
		"c03112": 19000000,
		"a22082": 6400000,
		"a02082": 6400000,
	}[revision]
	if !ok {
		panic("speed not found")
	}
	return speed
}

// New creates a new manager the outputs pixel data from a strip input to the actual pixels.
func New(colors []color.RGBW, sources []Source, path string, framerate int) <-chan error {
	if framerate <= 0 {
		panic("invalid framerate")
	}
	if len(colors) == 0 {
		panic("invalid colors")
	}
	if len(sources) == 0 {
		panic("invalid sources")
	}
	speed := determineSpeed()

	errors := make(chan error)
	ioctl := file.IoctlRequestNumber(false, true, unsafe.Sizeof(operation{}), 0x6b, 0)
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
			if err := f.Close(); err != nil {
				errors <- err
			}
		}()

		translated := make([]byte, len(colors)*translationFactor*4)
		arg := operation{
			txBuf:   uint64(uintptr(unsafe.Pointer(&translated[0]))),
			len:     uint32(len(translated)),
			speedHz: speed,
		}
		for allClosed := false; !allClosed; {
			<-ticker.C
			allClosed = handleSources(ticker.C, sources)
			translate(colors, translated, lut, translationFactor)
			if err := f.IoctlPointerArgument(ioctl, unsafe.Pointer(&arg))(); err != nil {
				log.Root.Warning("could not update pixels: %v", err)
			}
		}
	}()
	return errors
}

func handleSources(ticker <-chan time.Time, sources []Source) bool {
	for _, s := range sources {
		s.SendTick(ticker)
	}
	allClosed := true
	for _, s := range sources {
		if ok := s.AwaitDone(ticker); ok {
			allClosed = false
		}
	}
	return allClosed
}
