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

package errors

import (
	"fmt"
	"time"
)

// StepError represents an error inside a batch.
type StepError struct {
	Previous *StepError
	Cause    error
}

func (s StepError) Error() string {
	return s.Cause.Error()
}

// Unwrap the error before it.
func (s StepError) Unwrap() error {
	if s.Previous != nil {
		return s.Previous
	}
	return s.Cause
}

// Step of a batch job.
type Step struct {
	start     *Step
	actions   []func() error
	onError   *Step
	onSuccess *Step
}

// OnError executes the given actions if the current step execution fails.
func (s *Step) OnError(actions ...func() error) *Step {
	s.onError = &Step{s.start, actions, nil, nil}
	return s.onError
}

// OnSuccess executes the given actions if the current step execution succeeds.
func (s *Step) OnSuccess(actions ...func() error) *Step {
	s.onSuccess = &Step{s.start, actions, nil, nil}
	return s.onSuccess
}

// Always always executes the given actions afterwards.
func (s *Step) Always(actions ...func() error) *Step {
	s.onSuccess = &Step{s.start, actions, nil, nil}
	s.onError = s.onSuccess
	return s.onSuccess
}

func (s *Step) executeStep(stepErr *StepError) *StepError {
	for _, a := range s.actions {
		if err := a(); err != nil {
			stepErr = &StepError{stepErr, err}
		}
	}
	switch {
	case stepErr == nil && s.onSuccess != nil:
		return s.onSuccess.executeStep(stepErr)
	case s.onError != nil:
		return s.onError.executeStep(stepErr)
	default:
		return stepErr
	}
}

// Execute the batch and return the first error.
func (s *Step) Execute(message string) error {
	err := s.start.executeStep(nil)
	if err != nil {
		return fmt.Errorf("failed to %s: %w", message, err)
	}
	return nil
}

// NewBatch creates a new batch execution.
func NewBatch(actions ...func() error) *Step {
	s := &Step{nil, actions, nil, nil}
	s.start = s
	return s
}

// BatchSleepAction is a batch action for time.Sleep.
func BatchSleepAction(duration time.Duration) func() error {
	return func() error {
		time.Sleep(duration)
		return nil
	}
}

// BatchNoError wraps a method that returns nothing for easy integration.
func BatchNoError(real func()) func() error {
	return func() error {
		real()
		return nil
	}
}
