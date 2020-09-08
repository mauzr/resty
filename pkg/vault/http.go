package vault

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.eqrx.net/mauzr/pkg/errors"
	"go.eqrx.net/mauzr/pkg/log"
	"go.eqrx.net/mauzr/pkg/rest"
)

const (
	readHeaderTimeout = 3 * time.Second
	idleTimeout       = 24 * time.Hour
)

func (c *Client) serverTLSConfig(name string, pkis ...string) (*tls.Config, error) {
	cas := x509.NewCertPool()
	tlsConfig := &tls.Config{
		Certificates:             make([]tls.Certificate, len(pkis)),
		RootCAs:                  cas,
		ServerName:               name,
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                cas,
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		NextProtos:               []string{"h2"},
	}

	actions := []func() error{}
	for i, pki := range pkis {
		actions = append(actions, c.certificate(pki, "server", name, &tlsConfig.Certificates[i], tlsConfig.RootCAs).refresh)
	}
	err := errors.NewBatch(actions...).Execute("add certificates")
	if err == nil && tlsConfig.Certificates[0].PrivateKey == nil {
		panic("nope")
	}

	return tlsConfig, err
}

// RESTServer starts a new REST server in the mauzr context.
func (c *Client) RESTServer(ctx context.Context, handler http.Handler, name string, pkis ...string) <-chan error {
	errOut := make(chan error)
	go func() {
		log.Root.Debug("requesting server tls certificates for %v", name)
		tlsConfig, err := c.serverTLSConfig(name, pkis...)
		if err != nil {
			errOut <- err

			close(errOut)

			return
		}

		addrs, err := net.LookupHost(name)
		if err != nil {
			errOut <- err
			close(errOut)

			return
		}
		log.Root.Debug("found %v as addrs for %v", addrs, name)
		errs := []<-chan error{}
		shutdowns := []func(context.Context) error{}
		for _, addr := range addrs {
			if strings.Contains(addr, ".") {
				// Skip IPv4 addresses.
				continue
			}
			server := &http.Server{
				Addr:              net.JoinHostPort(addr, "443"),
				TLSConfig:         tlsConfig,
				ReadHeaderTimeout: readHeaderTimeout,
				IdleTimeout:       idleTimeout,
				Handler:           handler,
			}
			err := make(chan error)
			shutdowns = append(shutdowns, server.Shutdown)
			errs = append(errs, err)
			go func(server *http.Server, err chan<- error) {
				if e := server.ListenAndServeTLS("", ""); !errors.Is(e, http.ErrServerClosed) {
					err <- fmt.Errorf("http server for %s: %w", server.Addr, e)
				}
				close(err)
			}(server, err)
		}
		shutdownErrors := make(chan error)
		go func() {
			<-ctx.Done()
			for _, s := range shutdowns {
				if err := s(context.Background()); err != nil {
					shutdownErrors <- fmt.Errorf("http server shutdown: %w", err)
				}
			}
			close(shutdownErrors)
		}()
		errs = append(errs, shutdownErrors)
		errors.FanInto(errOut, errs...)
	}()

	return errOut
}

// RESTClient creates a new REST client for mauzr.
func (c *Client) RESTClient(name, pki string) (rest.Client, error) {
	t := &tls.Config{
		RootCAs:      x509.NewCertPool(),
		Certificates: []tls.Certificate{{}},
		MinVersion:   tls.VersionTLS13,
		NextProtos:   []string{"h2"},
	}
	cc := rest.NewClient(t)
	err := c.certificate(pki, "client", name, &t.Certificates[0], t.RootCAs).refresh()
	if err == nil && t.Certificates[0].PrivateKey == nil {
		panic("nope")
	}

	return cc, err
}
