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
	"crypto/tls"
	"fmt"
	"net"

	"net/http"
)

// Serve blocks and runs the configured http servers.
func (r *rest) Serve() []<-chan error {
	if len(r.servers) != len(r.listeners) {
		panic(fmt.Errorf("length of servers and listeners ist different"))
	}
	if len(r.servers) == 0 {
		panic(fmt.Errorf("no servers specified"))
	}
	// TODO: Error seems to be dropped if IP is not available with systemd socket? server just returns
	errors := make([]<-chan error, len(r.servers))
	for i := range r.servers {
		err := make(chan error)
		go func(server *http.Server, listener *net.Listener, errs chan<- error) {
			tlsListener := tls.NewListener(*listener, server.TLSConfig)
			if err := server.Serve(tlsListener); err != http.ErrServerClosed {
				errs <- err
			}
			close(errs)
		}(&r.servers[i], &r.listeners[i], err)
	}
	shutdownErrors := make(chan error)
	go func() {
		<-r.webserverContext.Done()
		for i := range r.servers {
			err := r.servers[i].Shutdown(context.Background())
			if err != nil {
				shutdownErrors <- err
			}
		}
		close(shutdownErrors)
	}()

	return append(errors, shutdownErrors)
}

func (r *rest) WebserverContext() context.Context {
	return r.webserverContext
}
