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
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

// Program handles all the program commons.
type Program struct {
	// Wg will be waited for before shutdown.
	Wg *sync.WaitGroup
	// Ctx will be canceled when shutdown is requested.
	Ctx context.Context
	// Cancel cancels Ctx.
	Cancel func()
	// ServiceName is the FQDN of this service.
	ServiceName *string
	// REST manager for everything
	Rest rest.REST
	// RootCommand for this service.
	RootCommand *cobra.Command
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

func listeners(hostname string, binds *[]string) []net.Listener {
	if pid, pidSet := os.LookupEnv("LISTEN_PID"); pidSet && strconv.Itoa(os.Getpid()) == pid {
		os.Unsetenv("LISTEN_PID")
		listenerCount, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
		os.Unsetenv("LISTEN_FDS")
		switch {
		case err != nil:
			panic(fmt.Errorf("LISTEN_PID is set but LISTEN_FDS is invalid: %v", err))
		case listenerCount == 0:
			panic(fmt.Errorf("LISTEN_PID is set but LISTEN_FDS is 0"))
		}
		restListeners := make([]net.Listener, listenerCount)
		for i := range restListeners {
			fd := i + 3
			unix.CloseOnExec(fd)
			listener, err := net.FileListener(os.NewFile(uintptr(fd), fmt.Sprintf("LISTEN_FD_%v", fd)))
			if err != nil {
				panic(fmt.Errorf("could not create file from fd: %v", err))
			}
			restListeners[i] = listener
		}
		return restListeners
	}
	addresses := []string{fmt.Sprintf("%s:443", hostname)}
	if binds != nil {
		addresses = *binds
	}

	restListeners := make([]net.Listener, len(addresses))
	for i, address := range addresses {
		l, err := net.Listen("tcp", address)
		if err != nil {
			panic(fmt.Errorf("could not listen: %v", err))
		}
		restListeners[i] = l
	}
	return restListeners
}

func New() *Program {
	ctx, cancel := context.WithCancel(context.Background())
	program := &Program{
		&sync.WaitGroup{},
		ctx,
		cancel,
		nil,
		nil,
		nil,
	}

	flags := pflag.FlagSet{}
	flags.StringToStringP("tags", "", nil, "Tags to include in measurements")
	binds := flags.StringSliceP("binds", "", nil, "Addresses to listen on")
	program.ServiceName = flags.StringP("servicename", "", "", "Name of this service that is used to bind and pick TLS certificates")

	rootCommand := cobra.Command{
		Use:          "mauzr <subcommand>",
		Short:        "Expose devices to the network",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				env := "MAUZR_" + strings.ToUpper(f.Name)
				if value, set := os.LookupEnv(env); set {
					if err := f.Value.Set(value); err != nil {
						panic(fmt.Errorf("could not apply environment variable %v with value %v to flag %v: %v", env, value, f.Name, err))
					}
				}
			})
			program.Rest = rest.New(*program.ServiceName, listeners(*program.ServiceName, binds))
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			defer cancel()
			if cmd.Name() == "help" {
				return nil
			}
			err := program.Rest.Serve(ctx)
			return err
		},
	}
	rootCommand.PersistentFlags().AddFlagSet(&flags)
	program.RootCommand = &rootCommand
	go program.handleTerminationSignals()
	return program
}
