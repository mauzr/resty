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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"go.eqrx.net/mauzr/pkg/bme280"
	"go.eqrx.net/mauzr/pkg/bme680"
	"go.eqrx.net/mauzr/pkg/program"
	"go.eqrx.net/mauzr/pkg/sk6812"
	"go.eqrx.net/mauzr/pkg/tradfri"
)

func completeCmd(rootCmd *cobra.Command) *cobra.Command {
	var cmd = &cobra.Command{
		Use:       "completion <bash|zsh>",
		Short:     "Generates completion scripts for bash and zsh",
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			switch args[0] {
			case "bash":
				err = rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				err = rootCmd.GenZshCompletion(os.Stdout)
			}
			return
		},
	}
	return cmd
}

func documentCmd(rootCmd *cobra.Command) *cobra.Command {
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
				err = doc.GenManTree(rootCmd, header, path)
			case "md":
				err = doc.GenMarkdownTree(rootCmd, path)
			case "rest":
				err = doc.GenReSTTree(rootCmd, path)
			case "yaml":
				err = doc.GenYamlTree(rootCmd, path)
			}
			return
		},
	}

	cmd.Flags().StringVarP(&path, "output-dir", "o", "/tmp/", "Directory to populate with documentation")

	return cmd
}

// SetupCommands adds subcommands of this pkg.
func SetupCommands(p *program.Program) {
	subCommands := []*cobra.Command{
		documentCmd(p.RootCommand),
		completeCmd(p.RootCommand),
		bme280.SubCommand(p),
		bme680.SubCommand(p),
		sk6812.SubCommand(p),
		tradfri.SubCommand(p),
	}
	for _, subCommand := range subCommands {
		p.RootCommand.AddCommand(subCommand)
	}
}
