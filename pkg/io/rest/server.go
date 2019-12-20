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
	"time"

	"net/http"
)

func (r *rest) Serve(ctx context.Context) error {
	for i := range r.servers {
		go func(server *http.Server) {
			err := server.ListenAndServeTLS("", "")
			if err != http.ErrServerClosed {
				r.serverErrors <- err
			}
			r.serverErrors <- nil
		}(&r.servers[i])
	}
	remaining := len(r.servers)
	terminated := false
	errors := []error{}
	for remaining != 0 {
		select {
		case <-ctx.Done():
			if terminated {
				continue
			}
		case err := <-r.serverErrors:
			remaining -= 1
			switch {
			case err != nil:
				errors = append(errors, err)
			case err == nil && !terminated:
				panic(fmt.Errorf("server terminated error free without being asked so"))
			case terminated:
				continue
			}
		}
		for i := range r.servers {
			httpCtx, httpCancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := r.servers[i].Shutdown(httpCtx); err != nil {
				errors = append(errors, err)
			}
			httpCancel()
		}
		terminated = true
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
