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

package sk6812

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.eqrx.net/mauzr/pkg/program"
)

// SubCommand creates a cobra command for this driver.
func SubCommand(p *program.Program) *cobra.Command {
	flags := pflag.FlagSet{}
	tty := flags.StringP("tty", "y", "/dev/ttyUSB0", "TTY to use for connection")

	command := cobra.Command{
		Use:   "sk6812",
		Short: "Expose a SK6812 strip",
		Long:  "Expose a SK6812 driver via REST",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return p.ApplyEnvsToFlags(&flags, [][2]string{{"tty", "MAUZR_TTY"}})
		},
		Run: func(cmd *cobra.Command, args []string) {
			strip := NewManager(*tty)
			setupHandler(p.Rest, strip)
			go strip.Manage(p.Ctx, p.Wg)
		},
	}

	if err := cobra.MarkFlagFilename(&flags, "tty"); err != nil {
		panic(err)
	}

	command.Flags().AddFlagSet(&flags)
	return &command
}
