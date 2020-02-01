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

package tradfri

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.eqrx.net/mauzr/pkg/program"
)

// SubCommand creates a cobra command for this driver.
func SubCommand(p *program.Program) *cobra.Command {
	flags := pflag.FlagSet{}
	client := flags.StringP("client", "c", "", "Client ID to use")
	psk := flags.StringP("psk", "p", "", "PSK to use")
	gateway := flags.StringP("gateway", "g", "", "Gateway to use")
	mapping := flags.StringToStringP("mapping", "m", nil, "Name to device ID mapping")

	command := cobra.Command{
		Use:   "tradfri",
		Short: "Expose tradfri LEDs behind a gateway",
		Long:  "Expose tradfri LEDs behind a gateway via REST",
		Run: func(cmd *cobra.Command, args []string) {
			if *client == "" {
				panic("Client ID must be set")
			}
			if *psk == "" {
				panic("PSK must be set")
			}
			setupKeys(*client, *psk)
			params := setupParams(*client, *gateway)
			for name, group := range *mapping {
				setupMapping(p.Rest, name, group, params)
			}
		},
	}
	command.Flags().AddFlagSet(&flags)
	return &command
}
