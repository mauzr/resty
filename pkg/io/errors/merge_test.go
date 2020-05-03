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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.eqrx.net/mauzr/pkg/io/errors"
)

func expectNothing(t *testing.T, timer *time.Timer, c <-chan error) {
	select {
	case <-timer.C:
	case <-c:
		assert.FailNow(t, "expected no err")
	}
}

func TestMerge(t *testing.T) {
	wasCalled := false
	once := sync.Once{}
	onError := func() { once.Do(func() { wasCalled = true }) }
	a, b, c := make(chan error), make(chan error), make(chan error)
	merged := errors.Merge(onError, a, b, c)

	assert.Equal(t, 0, len(merged))
	select {
	case <-merged:
		assert.FailNow(t, "expected no err")
	default:
	}
	assert.False(t, wasCalled)
	err := fmt.Errorf("somerror") // nolint
	b <- err

	holdTime := 1 * time.Millisecond
	select {
	case <-time.NewTimer(holdTime).C:
		assert.FailNow(t, "expected err")
	case e, ok := <-merged:
		assert.True(t, ok)
		assert.Equal(t, err, e)
	}
	assert.True(t, wasCalled)
	expectNothing(t, time.NewTimer(holdTime), merged)

	close(c)

	expectNothing(t, time.NewTimer(holdTime), merged)

	a <- err

	select {
	case <-time.NewTimer(holdTime).C:
		assert.FailNow(t, "expected err")
	case e, ok := <-merged:
		assert.True(t, ok)
		assert.Equal(t, err, e)
	}

	close(a)

	expectNothing(t, time.NewTimer(holdTime), merged)

	close(b)

	select {
	case <-time.NewTimer(holdTime).C:
		assert.FailNow(t, "expected err")
	case _, ok := <-merged:
		assert.False(t, ok)
	}
}
