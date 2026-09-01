// Package dns contains simple implementation of RFC 1035.
package dns

import (
	"errors"
	"fmt"
	"net/netip"
)

type AAAA struct {
	IP netip.Addr
}

func (aaaa AAAA) Type() Type {
	return TypeAAAA
}

func (aaaa AAAA) MarshalBinary() ([]byte, error) {
	if !aaaa.IP.Is6() {
		return nil, errors.New("A record must be IPv6")
	}
	return aaaa.IP.AsSlice(), nil
}

func (aaaa *AAAA) UnmarshalBinary(data []byte) error {
	ip, ok := netip.AddrFromSlice(data)
	if !ok {
		return fmt.Errorf("can't parse IP")
	}
	if !aaaa.IP.Is6() {
		return errors.New("AAAA record should use IPv6")
	}
	aaaa.IP = ip
	return nil
}
