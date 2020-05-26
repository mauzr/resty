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

// Aggregate all errors from an error channel and from the parameter list and merge them into one error.
// If onFirstError is not nil, it will be called on the first error.
func Aggregate(source <-chan error, additional ...error) error {
	errors := []error{}

	for _, err := range additional {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if source != nil {
		for {
			err, ok := <-source
			if !ok {
				break
			}
			if err == nil {
				panic("received null as err")
			}
			if !Is(errChannelClosed, err) {
				errors = append(errors, err)
			}
		}
	}
	switch len(errors) {
	case 0:
		return nil
	case 1:
		return errors[0]
	default:
		return &MultiError{errors}
	}
}
