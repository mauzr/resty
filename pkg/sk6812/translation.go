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

const (
	translationSizeFactor = 4
	channelCount          = 4
)

var (
	// Input byte as index, get its translation as field
	translation [256][translationSizeFactor]uint8
)

// StripSetting represents an array of SK6812 channel settings
type StripSetting [][channelCount]uint8

func init() {
	// A single bit is converted into a nibble to we need to translate
	// two bits of a color channel to get a translated byte
	// Put the value of the half nibble in here to get the translation for it.
	//
	// Technical Reason: The signal needs to be clocked between 2.85 MHz and 4MHz
	// A bitwise 1 is translated to 1000, a 0 to 1100.
	//
	// Notice: If you use a Raspberry PI you need to pin the GPU frequency to a
	// value of your choosing (e.g. 250MHz) since the clock speed may change
	// while inside a translation.
	halfNibbleTranslation := [4]byte{0xcc, 0xc8, 0x8c, 0x88}

	for i := range translation {
		for j := 0; j < 4; j++ {
			offset := 6 - 2*j
			halfNibble := i >> uint(offset) & 0x3
			translation[i][j] = halfNibbleTranslation[halfNibble]
		}
	}
}

// Translate translates a given array (One entry representing one LED)
// of arrays containing the color channels (GRBW) of one LED
func Translate(colors StripSetting) []byte {
	translated := make([]byte, len(colors)*channelCount*translationSizeFactor)
	translatedIndex := 0
	for _, color := range colors {
		for _, channel := range color {
			for _, translation := range translation[channel] {
				translated[translatedIndex] = translation
				translatedIndex++
			}
		}
	}
	return translated
}
