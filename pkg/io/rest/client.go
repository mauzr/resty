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
	"io"
	"net/http"
)

// HTTPError represents an HTTP error in combination with an HTTP status code.
type HTTPError struct {
	StatusCode int
	Cause      error
}

// Error returns the error as string.
func (h HTTPError) Error() string {
	switch {
	case h.Cause != nil:
		return h.Cause.Error()
	case h.StatusCode != 0:
		return http.StatusText(h.StatusCode)
	default:
		panic("empty error")
	}
}

// Unwrap the causing error.
func (h HTTPError) Unwrap() error {
	return h.Cause
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
func (r *rest) GetJSON(ctx context.Context, url string, target interface{}) error {
	response, err := r.GetRaw(ctx, url)
	if err != nil {
		return &HTTPError{0, err}
	}

	if response.StatusCode != http.StatusOK {
		err = &HTTPError{response.StatusCode, nil}
	} else if err = json.NewDecoder(response.Body).Decode(&target); err != nil {
		err = &HTTPError{0, err}
	}
	_ = response.Body.Close()
	return err
}

// PostRaw from the given reader to a remote site.
func (r *rest) PostRaw(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		panic(err)
	}

	return r.client.Do(request)
}
