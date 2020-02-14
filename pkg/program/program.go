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
	"sync"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"go.eqrx.net/mauzr/pkg/io/rest"
)

type Program struct {
	Wg          *sync.WaitGroup
	Ctx         context.Context
	Cancel      func()
	Hostname    *string
	Rest        rest.REST
	RootCommand *cobra.Command
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

	flags := pflag.FlagSet{}
	flags.StringToStringP("tags", "", nil, "Tags to include in measurements")
	binds := flags.StringSliceP("binds", "", nil, "Addresses to listen on")
	hostname := flags.StringP("hostname", "", "", "Name of this service that is used to bind and pick TLS certificates")

	program := &Program{
		&sync.WaitGroup{},
		ctx,
		cancel,
		hostname,
		nil,
		nil,
	}

	rootCommand := cobra.Command{
		Use:          "mauzr <subcommand>",
		Short:        "Expose devices to the network",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := program.ApplyEnvsToFlags(&flags, [][2]string{
				{"tags", "MAUZR_TAGS"},
				{"hostname", "MAUZR_HOSTNAME"},
				{"binds", "MAUZR_BINDS"},
			})
			if err != nil {
				return err
			}
			program.Rest = rest.New(*hostname, listeners(*hostname, binds))
			return nil
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
