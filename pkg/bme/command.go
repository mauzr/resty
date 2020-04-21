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

package bme

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.eqrx.net/mauzr/pkg/program"
)

// SubCommands creates a cobra command for BME drivers.
func SubCommands(p *program.Program) []*cobra.Command {
	flags := pflag.FlagSet{}
	bus := flags.StringP("bus", "b", "/dev/i2c-1", "Path of the linux bus to use")
	address := flags.Uint16P("address", "a", 0x77, "I2C address of the device")
	if err := cobra.MarkFlagFilename(&flags, "bus"); err != nil {
		panic(err)
	}
	goff := flags.Float64P("goff", "", 0.0, "Gas resistance offset to apply to measurements")
	hoff := flags.Float64P("hoff", "", 0.0, "Humidity offset to apply to measurements")
	poff := flags.Float64P("poff", "", 0.0, "Pressure offset to apply to measurements")
	toff := flags.Float64P("toff", "", 0.0, "Temperature offset to apply to measurements")

	bme280 := cobra.Command{
		Use:   "bme280 location=livingroom",
		Short: "Expose a BME280 driver",
		Long:  "Expose a BME280 driver via REST.",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := cmd.Root().PersistentFlags().GetStringToString("tags")
			if err != nil {
				return err
			}

			requests := make(chan Request)
			go func() {
				<-p.Rest.WebserverContext().Done()
				close(requests)
			}()

			NewBME280(*bus, *address, Measurement{GasResistance: *goff, Humidity: *hoff, Pressure: *poff, Temperature: *toff}, requests)
			setupHandler(p.Rest, requests, tags)

			return nil
		},
	}
	bme280.Flags().AddFlagSet(&flags)

	bme680 := cobra.Command{
		Use:   "bme680 location=livingroom",
		Short: "Expose a BME680 driver",
		Long:  "Expose a BME680 driver via REST.",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := cmd.Root().PersistentFlags().GetStringToString("tags")
			if err != nil {
				return err
			}

			requests := make(chan Request)
			go func() {
				<-p.Rest.WebserverContext().Done()
				close(requests)
			}()

			NewBME680(*bus, *address, Measurement{GasResistance: *goff, Humidity: *hoff, Pressure: *poff, Temperature: *toff}, requests)
			setupHandler(p.Rest, requests, tags)
			return nil
		},
	}
	bme680.Flags().AddFlagSet(&flags)

	return []*cobra.Command{&bme280, &bme680}
}
