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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"mauzr.eqrx.net/go/pkg/rest"
)

func applyEnvsToFlags(flags *pflag.FlagSet, envsToFlags [][2]string) error {
	for _, envToFlag := range envsToFlags {
		env, flag := envToFlag[0], envToFlag[1]
		if value, set := os.LookupEnv(env); set {
			if err := flags.Set(flag, value); err != nil {
				return fmt.Errorf("Could not apply environment variable %v with value %v to flag %v: %v", env, value, flag, err)
			}
		}
	}
	return nil
}

func main() {
	flags := pflag.FlagSet{}

	flags.StringToStringP("tags", "t", nil, "Tags to include in measurements")
	listen := flags.StringP("listen", "l", ":443", "Listen address of the REST server")
	caPath := flags.StringP("ca", "", "/etc/ssl/certs/mauzr-ca.crt", "Path to CA file to validate clients")
	crtPath := flags.StringP("crt", "", "/etc/ssl/certs/mauzr.crt", "Path to cert file to identify to the clients")
	keyPath := flags.StringP("key", "", "/etc/ssl/private/mauzr.key", "Path to key file to identify to the clients")

	mux := http.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())

	command := cobra.Command{
		Use:   "mauzr <subcommand>",
		Short: "Expose devices to the network",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return applyEnvsToFlags(&flags, [][2]string{{"tags", "RIWERS_TAGS"}, {"listen", "RIWERS_LISTEN"},
				{"ca", "RIWERS_CA"}, {"crt", "RIWERS_CRT"}, {"key", "RIWERS_KEY"}})
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() != "help" {
				defer cancel()
				return rest.Serve(*listen, *caPath, *crtPath, *keyPath, mux)
			}
			return nil
		},
	}
	if err := cobra.MarkFlagFilename(&flags, "ca"); err != nil {
		panic(err)
	}
	if err := cobra.MarkFlagFilename(&flags, "crt"); err != nil {
		panic(err)
	}
	if err := cobra.MarkFlagFilename(&flags, "key"); err != nil {
		panic(err)
	}
	command.PersistentFlags().AddFlagSet(&flags)

	command.AddCommand(documentCmd(&command), completeCmd(&command))
	command.AddCommand(bme280Command(ctx, mux), bme680Command(ctx, mux), gpioCommand(mux))
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
