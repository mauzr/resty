package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"mauzr.eqrx.net/go/pkg/i2c"
	"mauzr.eqrx.net/go/pkg/tsl2561"
)

func tls2561Command(mux *http.ServeMux) *cobra.Command {
	flags := pflag.FlagSet{}
	bus := flags.StringP("bus", "b", "/dev/i2c-1", "Path of the linux bus to use")
	device := flags.Uint16P("device", "d", 0x39, "I2C address of the device")

	command := cobra.Command{
		Use:   "tsl2561 location=livingroom",
		Short: "Expose a TSL2561 driver",
		Long:  "Expose a TSL2561 driver via REST. Positional arguments are interpreted as tags that will be included in measurements",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return applyEnvsToFlags(&flags, [][2]string{{"bus", "RIWERS_BUS"}, {"device", "RIWERS_DEVICE"}})
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if tags, err := cmd.Root().PersistentFlags().GetStringToString("tags"); err == nil {
				mux.Handle("/measurement", tsl2561.RESTHandler(*bus, (i2c.DeviceAddress)(*device), tags))
				mux.Handle("/metrics", promhttp.Handler())
			}
			return
		},
	}
	command.Flags().AddFlagSet(&flags)
	if err := command.MarkFlagFilename("device"); err != nil {
		panic(err)
	}
	return &command
}
