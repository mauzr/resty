package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"mauzr.eqrx.net/go/pkg/bme680"
)

func bme680Command(ctx context.Context, mux *http.ServeMux) *cobra.Command {
	flags := pflag.FlagSet{}
	bus := flags.StringP("bus", "b", "/dev/i2c-1", "Path of the linux bus to use")
	address := flags.Uint16P("address", "a", 0x77, "I2C address of the device")

	command := cobra.Command{
		Use:   "bme680 location=livingroom",
		Short: "Expose a BME680 driver",
		Long:  "Expose a BME680 driver via REST.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return applyEnvsToFlags(&flags, [][2]string{{"bus", "MAUZR_BUS"}, {"address", "MAUZR_ADDRESS"}})
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			logger := log.New(os.Stderr, "", 0)
			if tags, err := cmd.Root().PersistentFlags().GetStringToString("tags"); err == nil {
				chip := bme680.NewChip(*bus, *address)
				mux.Handle("/metrics", promhttp.Handler())
				mux.Handle("/measurement", bme680.RESTHandler(ctx, logger, chip, tags))
				go chip.Manage(ctx, logger)
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
