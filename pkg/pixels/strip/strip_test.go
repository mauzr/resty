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

package strip_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.eqrx.net/mauzr/pkg/pixels/color"
	"go.eqrx.net/mauzr/pkg/pixels/strip"
)

// TestName: Test if name is correctly set.
func TestName(t *testing.T) {
	assembly := strip.New()
	name := "name"
	strip := assembly.New(name, 3)
	assert.Equal(t, name, strip.Name(), "Name is not correct")
}

// TestInvalidSet: Test if strip blocks invalid value sets.
func TestInvalidSet(t *testing.T) {
	assembly := strip.New()
	name := "name"
	strip := assembly.New(name, 3)
	assert.Panics(t, func() { strip.Set(make([]color.RGBW, 8)) }, "Strip accepted value for set that is too large")
	assert.Panics(t, func() { strip.Set(make([]color.RGBW, 1)) }, "Strip accepted value for set that is too small")
}
