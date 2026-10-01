// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc3596#section-2
package dns

import (
	"errors"
	"fmt"
	"net/netip"
)

// A represents a DNS A record.
type A struct {
	IP netip.Addr
}

// Type returns the DNS record type for A records.
func (a A) Type() Type {
	return TypeA
}

// MarshalBinary marshals the A record into binary format.
func (a *A) MarshalBinary() ([]byte, error) {
	if !a.IP.Is4() {
		return nil, errors.New("A record must be IPv4")
	}
	return a.IP.AsSlice(), nil
}

// UnmarshalBinary unmarshals the binary data into the A record.
func (a *A) UnmarshalBinary(data []byte) error {
	ip, ok := netip.AddrFromSlice(data)
	if !ok {
		return fmt.Errorf("can't parse IP")
	}
	a.IP = ip
	return nil
}
