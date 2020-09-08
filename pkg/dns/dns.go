// Package dns allows interfacing mit DNS servers.
package dns

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

const (
	ttl   = 300
	fudge = 300
)

// ErrResult indicates that the DNS server replied with an error code.
var ErrResult error = errors.New("DNS returned error as result")

// Update the given DNS AAAA record with the given ip addresses using TSIG.
func Update(serverAddress, signatureName, signatureHash, zone, fqdn string, targets ...net.IP) error {
	c := &dns.Client{SingleInflight: true, TsigSecret: map[string]string{signatureName: signatureHash}}
	query := dns.Msg{}
	query.SetQuestion(fqdn, dns.TypeAAAA)
	existing, _, err := c.Exchange(&query, serverAddress)
	if err != nil {
		panic(err)
	}

	add := []dns.RR{}
	remove := []dns.RR{}
	for _, answer := range existing.Answer {
		for _, target := range targets {
			if target.Equal(answer.(*dns.AAAA).AAAA) {
				remove = append(remove, answer.(*dns.AAAA))
			}
		}
	}

	for _, target := range targets {
		targetExists := false
		for _, answer := range existing.Answer {
			if target.Equal(answer.(*dns.AAAA).AAAA) {
				targetExists = true
			} else {
				remove = append(remove, answer.(*dns.AAAA))
			}
		}
		if !targetExists {
			add = append(add, &dns.AAAA{Hdr: dns.RR_Header{Name: fqdn, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}, AAAA: target})
		}
	}

	update := dns.Msg{}
	update.SetUpdate(zone)
	switch {
	case len(add) != 0:
		update.Insert(add)
	case len(remove) != 0:
		update.RemoveRRset(remove)
	default:
		return nil
	}

	update.SetTsig(signatureName, dns.HmacSHA512, fudge, time.Now().Unix())
	in, _, err := c.Exchange(&update, serverAddress)
	switch {
	case err != nil:
		return err
	case in.Rcode != 0:
		return fmt.Errorf("%w: %v", ErrResult, in)
	default:
		return nil
	}
}
