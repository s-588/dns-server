// Package dns contains simple implementation of RFC 1035.
package dns

import (
	"errors"
	"fmt"
	"net/netip"
)

// AAAA represents a DNS AAAA record.
type AAAA struct {
	IP netip.Addr
}

// Type returns the DNS record type for AAAA records.
func (aaaa AAAA) Type() Type {
	return TypeAAAA
}

// MarshalBinary marshals the AAAA record into binary format.
func (aaaa AAAA) MarshalBinary() ([]byte, error) {
	if !aaaa.IP.Is6() {
		return nil, errors.New("A record must be IPv6")
	}
	return aaaa.IP.AsSlice(), nil
}

// UnmarshalBinary unmarshals the binary data into the AAAA record.
func (aaaa *AAAA) UnmarshalBinary(data []byte) error {
	ip, ok := netip.AddrFromSlice(data)
	if !ok {
		return fmt.Errorf("can't parse IP")
	}
	if !ip.Is6() {
		return errors.New("AAAA record should use IPv6")
	}
	aaaa.IP = ip
	return nil
}
