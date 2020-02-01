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

package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error with http specific information.
type Error interface {
	// Error returns the error as string.
	Error() string
	// StatusCode returns the http that caused the error.
	StatusCode() int
}

// httpError implements error.
type httpError struct {
	statusCode int
	cause      error
}

// StatusCode returns the http that caused the error.
func (h httpError) StatusCode() int {
	return h.statusCode
}

// Error returns the error as string.
func (h httpError) Error() string {
	switch {
	case h.cause != nil:
		return h.cause.Error()
	case h.statusCode != 0:
		return http.StatusText(h.statusCode)
	default:
		panic("empty error")
	}
}

// GetRaw response from a remote site.
func (r *rest) GetRaw(ctx context.Context, url string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	return r.client.Do(request)
}

// GetJSON from a remote site. It gets serialized into the given interface.
func (r *rest) GetJSON(ctx context.Context, url string, target interface{}) Error {
	response, err := r.GetRaw(ctx, url)
	if err != nil {
		return &httpError{0, err}
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return &httpError{response.StatusCode, nil}
	}
	if err := json.NewDecoder(response.Body).Decode(&target); err != nil {
		return &httpError{0, fmt.Errorf("could not deserialize JSON - %v", err)}
	}
	return nil
}

// PostRaw from the given reader to a remote site.
func (r *rest) PostRaw(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		panic(err)
	}

	return r.client.Do(request)
}
