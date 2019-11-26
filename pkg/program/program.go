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

func (p *Program) ApplyEnvsToFlags(flags *pflag.FlagSet, envsToFlags [][2]string) error {
	for _, envToFlag := range envsToFlags {
		flag, env := envToFlag[0], envToFlag[1]
		if value, set := os.LookupEnv(env); set {
			if err := flags.Set(flag, value); err != nil {
				return fmt.Errorf("could not apply environment variable %v with value %v to flag %v: %v", env, value, flag, err)
			}
		}
	}
	return nil
}

func (p *Program) handleTerminationSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	select {
	case <-c:
		p.Cancel()
	case <-p.Ctx.Done():
	}
}

func New() *Program {
	ctx, cancel := context.WithCancel(context.Background())

	flags := pflag.FlagSet{}
	flags.StringToStringP("tags", "", nil, "Tags to include in measurements")
	binds := flags.StringSliceP("binds", "", nil, "Addresses to listen on")
	hostname := flags.StringP("hostname", "", "", "Name of this service that is used to bind and pick TLS certificates")

	mux := http.NewServeMux()

	var program *Program

	rootCommand := cobra.Command{
		Use:          "mauzr <subcommand>",
		Short:        "Expose devices to the network",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return program.ApplyEnvsToFlags(&flags, [][2]string{
				{"tags", "MAUZR_TAGS"},
				{"hostname", "MAUZR_HOSTNAME"},
				{"binds", "MAUZR_BINDS"},
			})
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			defer cancel()
			if cmd.Name() == "help" {
				return nil
			}
			if *binds == nil {
				*binds = []string{fmt.Sprintf("%s:443", *hostname)}
			}
			err := rest.ServeAll(
				ctx,
				mux,
				fmt.Sprintf("/etc/ssl/certs/%s-ca.crt", *hostname),
				fmt.Sprintf("/etc/ssl/certs/%s.crt", *hostname),
				fmt.Sprintf("/etc/ssl/private/%s.key", *hostname),
				*binds,
			)
			return err
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
	go program.handleTerminationSignals()
	return program
}
