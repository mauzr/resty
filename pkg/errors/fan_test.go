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

	"go.eqrx.net/mauzr/pkg/errors"
	"go.eqrx.net/mauzr/pkg/testing/assert"
)

func expectNothing(assert assert.Assert, timer *time.Timer, c <-chan error) {
	select {
	case <-timer.C:
	case <-c:
		assert.Errorf("expected no err")
		assert.FailNow()
	}
}

func expectErrChannelClosed(assert assert.Assert, timer *time.Timer, c <-chan error) {
	select {
	case <-timer.C:
		assert.Errorf("expected err")
		assert.FailNow()
	case e, ok := <-c:
		assert.True(ok, "no error received")
		assert.True(errors.Is(e, errors.ErrChannelClosed), "expected errors.ErrChannelClosed")
	}
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	a, b, c := make(chan error), make(chan error), make(chan error)
	merged := errors.FanIn(a, b, c)

	assert.Equal(0, len(merged), "unexpected output")
	select {
	case <-merged:
		t.Errorf("expected no err")
		t.FailNow()
	default:
	}
	err := fmt.Errorf("somerror") // nolint
	b <- err

	holdTime := 1 * time.Millisecond
	select {
	case <-time.NewTimer(holdTime).C:
		t.Errorf("expected err")
		t.FailNow()
	case e, ok := <-merged:
		assert.True(ok, "no error received")
		assert.Equal(err, e, "expected other error")
	}
	expectNothing(assert, time.NewTimer(holdTime), merged)

	close(c)

	expectErrChannelClosed(assert, time.NewTimer(holdTime), merged)

	a <- err

	select {
	case <-time.NewTimer(holdTime).C:
		t.Errorf("expected err")
		t.FailNow()
	case e, ok := <-merged:
		assert.True(ok, "expected valid error")
		assert.Equal(err, e, "expected other error")
	}

	close(a)

	expectErrChannelClosed(assert, time.NewTimer(holdTime), merged)

	close(b)

	expectErrChannelClosed(assert, time.NewTimer(holdTime), merged)

	select {
	case <-time.NewTimer(holdTime).C:
		t.Errorf("expected err")
		t.FailNow()
	case _, ok := <-merged:
		assert.False(ok, "expected no error")
	}
}
