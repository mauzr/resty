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

package program

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"go.eqrx.net/mauzr/pkg/rest"
)

type Program struct {
	RootCommand *cobra.Command
	Wg          *sync.WaitGroup
	Ctx         context.Context
	Cancel      func()
	Hostname    *string
	Mux         *http.ServeMux
}

// ApplyEnvsToFlags maps given environment variable names and maps them to flags.
func (p *Program) ApplyEnvsToFlags(flags *pflag.FlagSet, envsToFlags [][2]string) error {
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
func (p *Program) HandleTerminationSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	select {
	case <-c:
		p.Cancel()
	case <-p.Ctx.Done():
	}
}

func NewProgram() *Program {
	ctx, cancel := context.WithCancel(context.Background())

	flags := pflag.FlagSet{}
	flags.StringToStringP("tags", "t", nil, "Tags to include in measurements")
	hostname := flags.StringP("hostname", "n", "", "Name of this service that is used to bind and pick TLS certificates")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	var program *Program

	rootCommand := cobra.Command{
		Use:          "mauzr <subcommand>",
		Short:        "Expose devices to the network",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return program.ApplyEnvsToFlags(&flags, [][2]string{{"tags", "MAUZR_TAGS"}, {"hostname", "MAUZR_HOSTNAME"}})
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

	program = &Program{
		&rootCommand,
		&sync.WaitGroup{},
		ctx,
		cancel,
		hostname,
		mux,
	}
	go program.HandleTerminationSignals()
	return program
}
