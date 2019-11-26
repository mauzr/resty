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
	"reflect"
	"time"
)

type server struct {
	http.Server
	errChan chan error
}

func (s *server) Run() {
	err := s.ListenAndServeTLS("", "")
	if err != http.ErrServerClosed {
		s.errChan <- err
	}
	close(s.errChan)
}

func ServeAll(ctx context.Context, handler http.Handler, caPath, crtPath, keyPath string, listenAddresses []string) error {
	if len(listenAddresses) == 0 {
		return fmt.Errorf("no listeners specified")
	}
	servers := make([]*server, len(listenAddresses))
	cases := make([]reflect.SelectCase, len(listenAddresses)+1)
	for i, address := range listenAddresses {
		servers[i] = newServer(handler, caPath, crtPath, keyPath, address)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(servers[i].errChan)}
		go servers[i].Run()
	}
	cases[len(listenAddresses)] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}

	errors := []error{}
	index, value, _ := reflect.Select(cases)
	if index != len(listenAddresses) {
		errors = append(errors, value.Interface().(error))
	}

	for _, server := range servers {
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := server.Shutdown(httpCtx); err != nil {
			errors = append(errors, err)
		}
		httpCancel()
	}

	for _, server := range servers {
		for {
			err, ok := <-server.errChan
			if !ok {
				break
			}
			if err != nil {
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
		return fmt.Errorf("multiple webservers failed :%v", errors)
	}
}

func newServer(handler http.Handler, caPath, crtPath, keyPath string, listenAddress string) *server {
	server := server{
		http.Server{
			Addr:              listenAddress,
			TLSConfig:         TLSConfig(crtPath, keyPath),
			ReadHeaderTimeout: 3 * time.Second,
			IdleTimeout:       120 * time.Second,
			Handler:           handler,
		},
		make(chan error),
	}

	if pem, err := ioutil.ReadFile(caPath); err != nil {
		panic(fmt.Errorf("failed to load CA file from %v: %v", caPath, err))
	} else if !server.TLSConfig.ClientCAs.AppendCertsFromPEM(pem) {
		panic(fmt.Errorf("failed to parse CA file from %v", caPath))
	}
	return &server
}

func DefaultResponseHeader(header http.Header) {
	header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}
