// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc3596#section-2
package dns

import (
	"errors"
	"fmt"
	"net/netip"
)

type A struct {
	IP netip.Addr
}

func (a A) Type() Type {
	return TypeA
}

func (a *A) MarshalBinary() ([]byte, error) {
	if !a.IP.Is4() {
		return nil, errors.New("A record must be IPv4")
	}
	return a.IP.AsSlice(), nil
}

func (a *A) UnmarshalBinary(data []byte) error {
	ip, ok := netip.AddrFromSlice(data)
	if !ok {
		return fmt.Errorf("can't parse IP")
	}
	a.IP = ip
	return nil
}
