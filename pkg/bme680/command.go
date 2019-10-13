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

package bme680

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.eqrx.net/mauzr/pkg/program"
)

// SubCommand creates a cobra command for this driver.
func SubCommand(p *program.Program) *cobra.Command {
	flags := pflag.FlagSet{}
	bus := flags.StringP("bus", "b", "/dev/i2c-1", "Path of the linux bus to use")
	address := flags.Uint16P("address", "a", 0x77, "I2C address of the device")

	command := cobra.Command{
		Use:   "bme680 location=livingroom",
		Short: "Expose a BME680 driver",
		Long:  "Expose a BME680 driver via REST.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return p.ApplyEnvsToFlags(&flags, [][2]string{{"bus", "MAUZR_BUS"}, {"address", "MAUZR_ADDRESS"}})
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if tags, err := cmd.Root().PersistentFlags().GetStringToString("tags"); err == nil {
				chip := NewChip(*bus, *address)
				setupHandler(p.Mux, chip, tags)
				go chip.Manage(p.Ctx, p.Wg)
			}
			return
		},
	}

	if err := cobra.MarkFlagFilename(&flags, "bus"); err != nil {
		panic(err)
	}
	command.Flags().AddFlagSet(&flags)

	return &command
}
