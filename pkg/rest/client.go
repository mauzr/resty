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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.eqrx.net/mauzr/pkg/errors"
	"golang.org/x/net/http2"
)

type client struct {
	*http.Client
}

// Client is an improved http client.
type Client interface {
	// Request begins a new HTTP request.
	Request(ctx context.Context, url string, method string) ClientRequest

	// RoundTripper returns the used transport for the Client.
	RoundTripper() http.RoundTripper
}

type clientResponse struct {
	Response   *http.Response
	RequestErr error
	DataErr    error
}

// ClientResponse represents an received HTTP response.
type ClientResponse interface {
	// JSONBody extracts a JSON string from a HTTP body request.
	JSONBody(data interface{}) ClientResponse

	// ByteSliceBody extracts a byte slice from a HTTP body request.
	ByteSliceBody(data *[]byte) ClientResponse

	// StringBody extracts a string from a HTTP body request.
	StringBody(dst *string) ClientResponse

	// Check finalizes the request and returns an error on failure or unexpected status code.
	Check() error
}

type clientRequest struct {
	Client *client
	ctx    context.Context
	url    string
	method string
	body   *bytes.Buffer
	header http.Header
}

// ClientRequest is an improved HTTP client request.
type ClientRequest interface {
	// JSONBody includes a JSON string into a HTTP body request.
	JSONBody(data interface{}) ClientRequest

	// StringBody includes a string into a HTTP body request.
	StringBody(data string) ClientRequest

	// Header adds a header to the request.
	Header(key string, value ...string) ClientRequest

	// Send a request on its way.
	Send(okCode int) ClientResponse
}

// NewClient creates a new improved http client.
func NewClient(tls *tls.Config) Client {
	return &client{
		&http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http2.Transport{
				TLSClientConfig: tls,
			},
		},
	}
}

// Request begins a new HTTP request.
func (c *client) Request(ctx context.Context, url string, method string) ClientRequest {
	return &clientRequest{c, ctx, url, method, &bytes.Buffer{}, http.Header{}}
}

// RoundTripper returns the used transport for the Client.
func (c *client) RoundTripper() http.RoundTripper {
	return c.Transport
}

// JSONBody includes a JSON string into a HTTP body request.
func (c *clientRequest) JSONBody(data interface{}) ClientRequest {
	if c.body.Len() != 0 {
		panic("body already set")
	}

	if err := json.NewEncoder(c.body).Encode(data); err != nil {
		panic(err)
	}
	return c
}

// StringBody includes a string into a HTTP body request.
func (c *clientRequest) StringBody(data string) ClientRequest {
	if c.body.Len() != 0 {
		panic("body already set")
	}
	c.body.Write([]byte(data))
	return c
}

// Header adds a header to the request.
func (c *clientRequest) Header(key string, value ...string) ClientRequest {
	c.header[key] = value
	return c
}

// Send a request on its way.
func (c *clientRequest) Send(okCode int) ClientResponse {
	request, err := http.NewRequestWithContext(c.ctx, c.method, c.url, c.body)
	request.Header = c.header
	if err != nil {
		panic(err)
	}
	cc := &clientResponse{}
	cc.Response, cc.RequestErr = c.Client.Do(request) //nolint:bodyclose
	if cc.RequestErr == nil {
		if okCode != cc.Response.StatusCode {
			if cc.Response.StatusCode == http.StatusMovedPermanently || cc.Response.StatusCode == http.StatusTemporaryRedirect || cc.Response.StatusCode == http.StatusSeeOther {
				cc.RequestErr = HTTPError{URL: request.URL.String(), StatusCode: cc.Response.StatusCode, Text: fmt.Sprintf("unexpected redirect to %v", cc.Response.Header["Location"])}
			} else {
				data, err := ioutil.ReadAll(cc.Response.Body)
				if err != nil {
					panic(err)
				}
				cc.RequestErr = HTTPError{URL: request.URL.String(), StatusCode: cc.Response.StatusCode, Text: string(data)}
			}
		}
	} else {
		cc.RequestErr = HTTPError{URL: request.URL.String(), Text: cc.RequestErr.Error()}
	}
	if cc.Response == nil && cc.RequestErr == nil {
		panic("invalid state")
	}
	return cc
}

// JSONBody extracts a JSON string from a HTTP body request.
func (c *clientResponse) JSONBody(data interface{}) ClientResponse {
	if c.RequestErr == nil {
		c.DataErr = json.NewDecoder(c.Response.Body).Decode(data)
	}
	return c
}

// ByteSliceBody extracts a byte slice from a HTTP body request.
func (c *clientResponse) ByteSliceBody(data *[]byte) ClientResponse {
	if c.RequestErr != nil {
		d, err := ioutil.ReadAll(c.Response.Body)
		*data = d
		c.DataErr = err
	}
	return c
}

// StringBody extracts a string from a HTTP body request.
func (c *clientResponse) StringBody(dst *string) ClientResponse {
	if c.RequestErr == nil {
		b, err := ioutil.ReadAll(c.Response.Body)
		if err != nil {
			*dst = string(b)
		}
		c.DataErr = err
	}
	return c
}

// Check finalizes the request and returns an error on failure or unexpected status code.
func (c *clientResponse) Check() error {
	if c.Response != nil {
		_ = c.Response.Body.Close()
	}
	if c.RequestErr != nil {
		return c.RequestErr
	}
	if c.DataErr != nil {
		return c.DataErr
	}
	return nil
}

// SendAll sends all ClientRequests and returns the error.
func SendAll(okCode int, clients ...ClientRequest) error {
	errs := make([]<-chan error, 0, len(clients))
	for _, c := range clients {
		err := make(chan error)
		errs = append(errs, err)
		go func(c ClientRequest, err chan error) {
			err <- c.Send(okCode).Check()
			close(err)
		}(c, err)
	}
	return errors.Aggregate(errors.FanIn(errs...), nil)
}

// GoSendAll sends all ClientRequests.
// The log parameter is expected to have fmt.Printf like functionality and
// log errors somewhere is not nil.
func GoSendAll(okCode int, log func(string, ...interface{}), clients ...ClientRequest) {
	go func() {
		err := SendAll(okCode, clients...)
		if err != nil && log != nil {
			log("send all REST requests failed: %v", err)
		}
	}()
}
