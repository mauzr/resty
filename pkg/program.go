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

package pkg

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
)

// ApplyEnvsToFlags maps given environment variable names and maps them to flags.
func ApplyEnvsToFlags(flags *pflag.FlagSet, envsToFlags [][2]string) error {
	for _, envToFlag := range envsToFlags {
		flag, env := envToFlag[0], envToFlag[1]
		if value, set := os.LookupEnv(env); set {
			if err := flags.Set(flag, value); err != nil {
				return fmt.Errorf("Could not apply environment variable %v with value %v to flag %v: %v", env, value, flag, err)
			}
		}
	}
	return nil
}

// HandleTerminationSignals executes the given cancel function when a shutdown signal is received.
func HandleTerminationSignals(ctx context.Context, cancel func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	select {
	case <-c:
		cancel()
	case <-ctx.Done():
	}
}
