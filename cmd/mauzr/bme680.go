package main

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"mauzr.eqrx.net/go/pkg/bme680"
)

func bme680Command(ctx context.Context, mux *http.ServeMux) *cobra.Command {
	flags := pflag.FlagSet{}
	bus := flags.StringP("bus", "b", "/dev/i2c-1", "Path of the linux bus to use")
	device := flags.Uint16P("device", "d", 0x77, "I2C address of the device")

	command := cobra.Command{
		Use:   "bme680 location=livingroom",
		Short: "Expose a BME680 driver",
		Long:  "Expose a BME680 driver via REST.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return applyEnvsToFlags(&flags, [][2]string{{"bus", "RIWERS_BUS"}, {"device", "RIWERS_DEVICE"}})
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if tags, err := cmd.Root().PersistentFlags().GetStringToString("tags"); err == nil {
				mux.Handle("/metrics", promhttp.Handler())
				mux.Handle("/measurement", bme680.RESTHandler(ctx, *bus, *device, tags))
			}
			return
		},
	}
	if err := cobra.MarkFlagFilename(&flags, "device"); err != nil {
		panic(err)
	}
	command.Flags().AddFlagSet(&flags)
	return &command
}
