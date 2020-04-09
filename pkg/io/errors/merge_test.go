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

package errors_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.eqrx.net/mauzr/pkg/io/errors"
)

func TestMerge(t *testing.T) {
	wasCalled := false
	onError := func() {
		wasCalled = true
	}
	a := make(chan error)
	b := make(chan error)
	c := make(chan error)
	merged := errors.Merge(onError, a, b, c)

	assert.Equal(t, 0, len(merged))
	select {
	case <-merged:
		assert.FailNow(t, "expected no err")
	default:
	}
	assert.False(t, wasCalled)
	err := fmt.Errorf("somerror")
	b <- err

	holdTime := 1 * time.Millisecond
	timer := time.NewTimer(holdTime)
	defer timer.Stop()

	select {
	case <-timer.C:
		assert.FailNow(t, "expected err")
	case e, ok := <-merged:
		assert.True(t, ok)
		assert.Equal(t, err, e)
	}
	timer.Reset(holdTime)
	assert.True(t, wasCalled)
	select {
	case <-timer.C:
	case <-merged:
		assert.FailNow(t, "expected no err")
	}

	close(c)
	timer.Reset(holdTime)

	select {
	case <-timer.C:
	case <-merged:
		assert.FailNow(t, "expected no err")
	}

	a <- err
	timer.Reset(holdTime)

	select {
	case <-timer.C:
		assert.FailNow(t, "expected err")
	case e, ok := <-merged:
		assert.True(t, ok)
		assert.Equal(t, err, e)
	}

	close(a)
	timer.Reset(holdTime)

	select {
	case <-timer.C:
	case <-merged:
		assert.FailNow(t, "expected no err")
	}

	close(b)
	timer.Reset(holdTime)

	select {
	case <-timer.C:
		assert.FailNow(t, "expected err")
	case _, ok := <-merged:
		assert.False(t, ok)
	}
}
