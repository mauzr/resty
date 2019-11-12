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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"net"
)

type Server struct {
	http.Server
}

// Serve creates a REST server secured by TLS with client cert authentication.
func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer httpCancel()
		if err := s.Shutdown(httpCtx); err != nil {
			panic(err)
		}
	}()

  listener, err := net.Listen("tcp6", s.Addr)
	if err == nil {
		err = s.ServeTLS(listener, "", "")
		if err == http.ErrServerClosed {
			err = nil
		}
	}
	return err
}

func NewServer(handler http.Handler, caPath, crtPath, keyPath, listen string) *Server {
	config := TLSConfig(crtPath, keyPath)

	server := Server{
		http.Server{
			Addr:              listen,
			TLSConfig:         config,
			ReadHeaderTimeout: 3 * time.Second,
			IdleTimeout:       120 * time.Second,
			Handler:           handler,
		},
	}

	if pem, err := ioutil.ReadFile(caPath); err != nil {
		panic(fmt.Errorf("failed to load CA file from %v: %v", caPath, err))
	} else if !config.ClientCAs.AppendCertsFromPEM(pem) {
		panic(fmt.Errorf("failed to parse CA file from %v", caPath))
	}
	return &server
}

func DefaultResponseHeader(header http.Header) {
	header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}
