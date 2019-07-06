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
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"mauzr.eqrx.net/go/pkg/sk6812"
)

func sk6812Command(mux *http.ServeMux) *cobra.Command {
	flags := pflag.FlagSet{}
	tty := flags.StringP("tty", "y", "/dev/ttyUSB0", "TTY to use for connection")

	command := cobra.Command{
		Use:   "sk6812",
		Short: "Expose a SK6812 strip",
		Long:  "Expose a SK6812 driver via REST",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return applyEnvsToFlags(&flags, [][2]string{{"tty", "RIWERS_TTY"}})
		},
		Run: func(cmd *cobra.Command, args []string) {
			mux.Handle("/color", sk6812.RESTHandler(*tty))
			mux.Handle("/metrics", promhttp.Handler())
		},
	}
	if err := cobra.MarkFlagFilename(&flags, "tty"); err != nil {
		panic(err)
	}
	command.Flags().AddFlagSet(&flags)
	return &command
}
