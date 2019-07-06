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

package sk6812_test

import (
	"testing"

	"mauzr.eqrx.net/go/pkg/sk6812"
)

// TestTranslate tests if a list of channel values is correctly serialized
func TestTranslate(test *testing.T) {
	colors := sk6812.StripSetting{{0, 75, 128, 255}, {255, 128, 75, 0}}
	targetData := []byte{204, 204, 204, 204, 200, 204, 140, 136, 140, 204, 204, 204, 136, 136, 136, 136,
		136, 136, 136, 136, 140, 204, 204, 204, 200, 204, 140, 136, 204, 204, 204, 204}
	actualData := sk6812.Translate(colors)

	if len(targetData) != len(actualData) {
		test.Errorf("Translate(%v) returned translation %v, expected %v", colors, actualData, targetData)
	}

	for i := range targetData {
		if targetData[i] != actualData[i] {
			test.Errorf("Translate(%v) returned translation %v, expected %v", colors, actualData, targetData)
			return
		}
	}
}
