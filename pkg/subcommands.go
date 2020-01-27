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
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"go.eqrx.net/mauzr/pkg/bme"
	"go.eqrx.net/mauzr/pkg/program"
	"go.eqrx.net/mauzr/pkg/sk6812"
	"go.eqrx.net/mauzr/pkg/tradfri"
)

func completeCmd(p *program.Program) *cobra.Command {
	var cmd = &cobra.Command{
		Use:       "completion <bash|zsh>",
		Short:     "Generates completion scripts for bash and zsh",
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			switch args[0] {
			case "bash":
				err = p.RootCommand.GenBashCompletion(os.Stdout)
			case "zsh":
				err = p.RootCommand.GenZshCompletion(os.Stdout)
			}
			return
		},
	}
	return cmd
}

func documentCmd(p *program.Program) *cobra.Command {
	var path string
	var cmd = &cobra.Command{
		Use:       "document <man|md|rest|yaml>",
		Short:     "Generates documentation",
		ValidArgs: []string{"man", "md", "rest", "yaml"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			switch args[0] {
			case "man":
				header := &doc.GenManHeader{
					Title:   "Mauzr",
					Section: "1",
				}
				err = doc.GenManTree(p.RootCommand, header, path)
			case "md":
				err = doc.GenMarkdownTree(p.RootCommand, path)
			case "rest":
				err = doc.GenReSTTree(p.RootCommand, path)
			case "yaml":
				err = doc.GenYamlTree(p.RootCommand, path)
			}
			return
		},
	}

	cmd.Flags().StringVarP(&path, "output-dir", "o", "/tmp/", "Directory to populate with documentation")

	return cmd
}

func healthcheckCmd(p *program.Program) *cobra.Command {
	command := cobra.Command{
		Use:   "healthcheck",
		Short: "Check the health of the configured agent",
		Long:  "Check the health of the configured agent via REST",
		Run: func(cmd *cobra.Command, args []string) {
			names := []string{}
			results := make(chan bool)
			for _, name := range names {
				go func(n string) {
					ctx, cancel := context.WithTimeout(p.Ctx, 4*time.Second)
					defer cancel()
					r, err := p.Rest.GetRaw(ctx, fmt.Sprintf("https://%s/health", n))
					r.Body.Close()
					results <- err == nil && r.StatusCode == http.StatusOK
				}(name)
			}
			for i := 0; i < len(names); i++ {
				if !<-results {
					os.Exit(1)
				}
			}
			os.Exit(0)
		},
	}
	return &command
}

// SetupCommands adds subcommands of this pkg.
func SetupCommands(p *program.Program) {
	subCommands := []*cobra.Command{
		documentCmd(p),
		completeCmd(p),
		healthcheckCmd(p),
		sk6812.SubCommand(p),
		tradfri.SubCommand(p),
	}
	subCommands = append(subCommands, bme.SubCommands(p)...)
	for _, subCommand := range subCommands {
		p.RootCommand.AddCommand(subCommand)
	}
}
