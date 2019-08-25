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

package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.eqrx.net/mauzr/pkg"
	"go.eqrx.net/mauzr/pkg/rest"
)

func main() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	go pkg.HandleTerminationSignals(ctx, cancel)
	defer wg.Wait()
	defer cancel()

	flags := pflag.FlagSet{}
	flags.StringToStringP("tags", "t", nil, "Tags to include in measurements")
	hostname := flags.StringP("hostname", "n", "", "Name of this service that is used to bind and pick TLS certificates")
	mux := http.NewServeMux()

	rootCommand := cobra.Command{
		Use:          "mauzr <subcommand>",
		Short:        "Expose devices to the network",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return pkg.ApplyEnvsToFlags(&flags, [][2]string{{"tags", "MAUZR_TAGS"}, {"hostname", "MAUZR_HOSTNAME"}})
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" {
				return nil
			}
			tlsConfig, err := rest.ServerConfig(
				"/etc/ssl/certs/mauzr-ca.crt",
				fmt.Sprintf("/etc/ssl/certs/%s.crt", *hostname),
				fmt.Sprintf("/etc/ssl/private/%s.key", *hostname),
			)
			if err != nil {
				return err
			}
			return rest.Serve(ctx, fmt.Sprintf("%s:443", *hostname), tlsConfig, mux)
		},
	}
	rootCommand.PersistentFlags().AddFlagSet(&flags)

	rootCommand.AddCommand(documentCmd(&rootCommand), completeCmd(&rootCommand))

	subCommands := map[string]*cobra.Command{
		"bme280": bme280Command(ctx, &wg, mux),
		"bme680": bme680Command(ctx, &wg, mux),
		"gpio":   gpioCommand(ctx, &wg, mux),
		"sk6812": sk6812Command(ctx, &wg, mux),
	}
	for _, subCommand := range subCommands {
		rootCommand.AddCommand(subCommand)
	}
	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}
