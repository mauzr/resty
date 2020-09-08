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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"net/url"
	"sort"
	"strings"
	"sync"

	"go.eqrx.net/mauzr/pkg/log"
)

const (
	landingHeader = `<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>core.eqrx.net front page</title>
</head>
<body>
`
	landingFooter = `</body>
</html>
`
)

//go:generate esc -o static.go --pkg rest --prefix=../../web/ ../../web/

// Mux is an http ServeMux with extra methods.
type Mux interface {
	// Forward a path to some other host.
	Forward(client Client, path, host, redirect string)
	// AddDefaultResponseHeader to the given header.
	AddDefaultResponseHeader(header http.Header)
	// Endpoint provides a server end point for a rest application. The given handler is called on each invoction.
	Endpoint(path string, queryHandler func(query *Request))
	// ServeHTTP just calls net/http.ServeMux.ServeHTTP.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	// Handle just calls net/http.ServeMux.Handle.
	Handle(pattern string, handler http.Handler)
	// HandleFunc just calls net/http.ServeMux.HandleFunc.
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

type mux struct {
	mux            *http.ServeMux
	pages          []string
	pageMutex      sync.Mutex
	landing        []byte
	rootRegistered bool
}

// Forward a path to some other host.
func (m *mux) Forward(client Client, path, host, redirect string) {
	director := func(req *http.Request) {
		rawURL := fmt.Sprintf("%s%s?%s", host, strings.TrimPrefix(req.URL.Path, path), req.URL.RawQuery)
		u, err := url.Parse(rawURL)
		if err != nil {
			panic(err)
		}
		req.URL = u
	}

	modifier := func(r *http.Response) error {
		h := r.Header
		if h.Get("Location") != "" {
			h.Set("Location", redirect)
		}

		return nil
	}

	m.Handle(path, &httputil.ReverseProxy{
		Director:       director,
		Transport:      client.RoundTripper(),
		ModifyResponse: modifier,
	})
}

// AddDefaultResponseHeader to the given header.
func (m *mux) AddDefaultResponseHeader(header http.Header) {
	header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	header.Add("X-XSS-Protection", "1; mode=block")
	header.Add("X-Frame-Options", "DENY")
	header.Add("X-Content-Type-Options", "nosniff")
}

// Endpoint provides a server end point for a rest application. The given handler is called on each invoction.
func (m *mux) Endpoint(path string, queryHandler func(query *Request)) {
	m.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		m.AddDefaultResponseHeader(w.Header())
		requestBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		response := Request{
			req.Context(),
			*req.URL,
			requestBody,
			nil,
			http.StatusOK,
			nil,
			nil,
			nil,
			len(req.URL.Query()) != 0,
		}
		queryHandler(&response)
		switch {
		case response.RequestErr != nil:
			http.Error(w, response.RequestErr.Error(), http.StatusBadRequest)
		case response.GatewayErr != nil:
			http.Error(w, response.GatewayErr.Error(), http.StatusBadGateway)
		case response.InternalErr != nil:
			http.Error(w, response.InternalErr.Error(), http.StatusInternalServerError)
		case req.Method != http.MethodGet && response.ResponseBody != nil:
			panic("response body only allowed for get method")
		case response.ResponseBody != nil:
			w.WriteHeader(response.Status)
			_, err := w.Write(response.ResponseBody)
			if err != nil {
				panic(err)
			}
		default:
			http.Redirect(w, req, "", http.StatusSeeOther)
		}
	})
}

// ServeHTTP just calls net/http.ServeMux.ServeHTTP.
func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !m.rootRegistered {
		m.Endpoint("/", m.land)
		m.rootRegistered = true
	}
	m.mux.ServeHTTP(w, r)
}

func (m *mux) registerPage(pattern string) {
	m.pageMutex.Lock()
	m.pages = append(m.pages, pattern)
	regenerate := m.landing != nil
	m.pageMutex.Unlock()
	if regenerate {
		m.generateLanding()
	}
	if pattern == "/" {
		m.rootRegistered = true
	}
}

// Handle just calls net/http.ServeMux.Handle.
func (m *mux) Handle(pattern string, handler http.Handler) {
	m.registerPage(pattern)
	m.mux.Handle(pattern, handler)
}

// HandleFunc just calls net/http.ServeMux.HandleFunc.
func (m *mux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.registerPage(pattern)
	m.mux.HandleFunc(pattern, handler)
}

func (m *mux) land(r *Request) {
	if r.URL.Path != "/" {
		r.Status = http.StatusNotFound
		r.ResponseBody = []byte("not found")

		return
	}
	if m.landing == nil {
		m.generateLanding()
	}
	r.ResponseBody = m.landing
}

func (m *mux) generateLanding() {
	m.pageMutex.Lock()
	defer m.pageMutex.Unlock()
	sort.Strings(m.pages)
	buf := strings.Builder{}
	buf.WriteString(landingHeader)
	for _, s := range m.pages {
		if s != "/" {
			fmt.Fprintf(&buf, "<a href=\"%s\">%s</a><br>\n", s, s)
		}
	}
	buf.WriteString(landingFooter)
	m.landing = []byte(buf.String())
}

// NewMux creates a new improves mux.
func NewMux() Mux {
	hm := http.NewServeMux()
	hm.Handle("/favicon.ico", http.FileServer(FS(false)))
	hm.HandleFunc("/debug/pprof/", pprof.Index)
	hm.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	hm.HandleFunc("/debug/pprof/profile", pprof.Profile)
	hm.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	hm.HandleFunc("/debug/pprof/trace", pprof.Trace)
	m := &mux{hm, []string{}, sync.Mutex{}, nil, false}
	m.Endpoint("/health", func(r *Request) {
		messages := log.Root.RetainedMessages()
		r.ResponseBody, r.InternalErr = json.Marshal(&messages)
	})

	return m
}
