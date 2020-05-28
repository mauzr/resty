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

// Package assert provides basic helpers for testing.
package assert

import (
	"reflect"
	"testing"
)

// Assert provides assertion helpers.
type Assert interface {
	Equal(expected, actual interface{}, err string)
	True(actual bool, err string)
	False(actual bool, err string)
	Panics(actual func(), err string)
	Errorf(format string, args ...interface{})
	FailNow()
}

type assert struct {
	t *testing.T
}

func (a assert) Equal(expected, actual interface{}, err string) {
	if !reflect.DeepEqual(expected, actual) {
		a.Errorf("%s: expected %v, actual %v", err, expected, actual)
	}
}

func (a assert) True(actual bool, err string) {
	if !actual {
		a.Errorf("%s: not true", err)
	}
}

func (a assert) False(actual bool, err string) {
	if actual {
		a.Errorf("%s: not false", err)
	}
}

func (a assert) Panics(actual func(), err string) {
	defer func() {
		if r := recover(); r == nil {
			a.Errorf("%s: did not panic", err)
			a.FailNow()
		}
	}()
	actual()
}

func (a assert) Errorf(format string, args ...interface{}) {
	a.t.Errorf(format, args...)
}

func (a assert) FailNow() {
	a.t.FailNow()
}

// New creates a new Assert instance.
func New(t *testing.T) Assert {
	return &assert{t}
}
