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

// Package program manages mauzr subcommand instances.
package program

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"go.eqrx.net/mauzr/pkg/io/errors"
	"go.eqrx.net/mauzr/pkg/io/rest"
)

// Program handles all the program commons.
type Program struct {
	// Ctx will be canceled when shutdown is requested.
	Ctx context.Context

	Errors []<-chan error

	// ServiceName is the FQDN of this service.
	ServiceName *string
	// REST manager for everything.
	Rest rest.REST
	// RootCommand for this service.
	RootCommand *cobra.Command
}

func listeners(binds []string) []net.Listener {
	if pid, pidSet := os.LookupEnv("LISTEN_PID"); pidSet && strconv.Itoa(os.Getpid()) == pid {
		_ = os.Unsetenv("LISTEN_PID")
		listenerCount, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
		_ = os.Unsetenv("LISTEN_FDS")
		switch {
		case err != nil:
			panic(fmt.Errorf("LISTEN_PID is set but LISTEN_FDS is invalid: %w", err))
		case listenerCount == 0:
			panic("LISTEN_PID is set but LISTEN_FDS is 0")
		}
		restListeners := make([]net.Listener, listenerCount)
		for i := range restListeners {
			fd := i + 3
			unix.CloseOnExec(fd)
			listener, err := net.FileListener(os.NewFile(uintptr(fd), fmt.Sprintf("LISTEN_FD_%v", fd)))
			if err != nil {
				panic(fmt.Errorf("could not create file from fd: %w", err))
			}
			restListeners[i] = listener
		}
		return restListeners
	}

	restListeners := make([]net.Listener, len(binds))
	for i, address := range binds {
		l, err := net.Listen("tcp", address)
		if err != nil {
			panic(fmt.Errorf("could not listen: %w", err))
		}
		restListeners[i] = l
	}
	return restListeners
}

// New creates a new program manager.
func New() *Program {
	runtimeCtx, programCancel := context.WithCancel(context.Background())
	webserverCtx, webserverCancel := context.WithCancel(runtimeCtx)
	program := &Program{runtimeCtx, []<-chan error{}, nil, nil, nil}

	flags := pflag.FlagSet{}
	flags.StringToStringP("tags", "", nil, "Tags to include in measurements")
	binds := flags.StringSliceP("binds", "", nil, "Addresses to listen on")
	program.ServiceName = flags.StringP("servicename", "", "", "Name of this service that is used to bind and pick TLS certificates")
	program.RootCommand = createCommand(preCommand(webserverCtx, program, *binds), postCommand(program, programCancel, webserverCancel))
	program.RootCommand.PersistentFlags().AddFlagSet(&flags)

	return program
}

func createCommand(pre func(cmd *cobra.Command, args []string) error, post func(cmd *cobra.Command, args []string) error) *cobra.Command {
	return &cobra.Command{
		Use:                "mauzr <subcommand>",
		Short:              "Expose devices to the network",
		SilenceUsage:       true,
		PersistentPreRunE:  pre,
		PersistentPostRunE: post,
	}
}

func preCommand(webserverCtx context.Context, p *Program, binds []string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		errors := []error{}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			env := "MAUZR_" + strings.ToUpper(f.Name)
			if value, set := os.LookupEnv(env); set {
				if err := f.Value.Set(value); err != nil {
					errors = append(errors, fmt.Errorf("could not apply environment variable %v with value %v to flag %v: %w", env, value, f.Name, err))
				}
			}
		})
		p.Rest = rest.New(webserverCtx, *p.ServiceName, listeners(binds))
		if len(errors) > 1 {
			return errors[0]
		}
		return nil
	}
}

func postCommand(p *Program, programCancel, webserverCancel func()) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			webserverCancel()
			programCancel()
			return nil
		}

		runtimeErrorSource := errors.Merge(webserverCancel, p.Errors...)
		webserverErrorSource := errors.Merge(webserverCancel, p.Rest.Serve()...)
		webserverErrorSink := make(chan error)
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			<-c
			webserverCancel()
		}()
		go func() {
			defer close(webserverErrorSink)
			defer programCancel()
			for {
				err, ok := <-webserverErrorSource
				if !ok {
					return
				}
				webserverErrorSink <- err
			}
		}()

		return errors.Collect(errors.Merge(nil, runtimeErrorSource, webserverErrorSource))
	}
}
