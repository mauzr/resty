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
	"net/http"
	"time"
)

// Serve creates a REST server secured by TLS with client cert authentication.
func Serve(ctx context.Context, listen string, tlsConfig *tls.Config, mux *http.ServeMux) error {
	server := http.Server{
		Addr:              listen,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           mux,
		// Disable HTTP v2.0 since that would require 128 bit ciphers.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	go func() {
		<-ctx.Done()
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer httpCancel()
		if err := server.Shutdown(httpCtx); err != nil {
			panic(err)
		}
	}()

	err := server.ListenAndServeTLS("", "")
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}
