package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.eqrx.net/mauzr/pkg/gpio"
)

func gpioCommand(mux *http.ServeMux) *cobra.Command {
	command := cobra.Command{
		Use:   "gpio",
		Short: "Expose a GPIO driver",
		Long:  "Expose a GPIO driver via REST.",
		Run: func(cmd *cobra.Command, args []string) {
			mux.Handle("/metrics", promhttp.Handler())
			mux.Handle("/input", gpio.InputHandler())
			mux.Handle("/output", gpio.OutputHandler())
		},
	}
	return &command
}
