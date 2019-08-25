package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.eqrx.net/mauzr/pkg/gpio"
)

func gpioCommand(ctx context.Context, wg *sync.WaitGroup, mux *http.ServeMux) *cobra.Command {
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
