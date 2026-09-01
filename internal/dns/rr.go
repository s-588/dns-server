// Package dns contains simple implementation of RFC 1035.
// File rr.go contains implementation of Rersource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-3.2
package dns

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

type RR struct {
	Name     string
	Type     Type
	Class    Class
	TTL      uint32
	RDLength uint16
	RData    RData
}

type RData interface {
	Type() Type
	MarshalBinary() ([]byte, error)
}

func decodeRData(t Type, data []byte) (RData, error) {
	switch t {
	case TypeA:
		a := &A{}
		err := a.UnmarshalBinary(data)
		return a, err
	case TypeAAAA:
		aaaa := &AAAA{}
		err := aaaa.UnmarshalBinary(data)
		return aaaa, err
	case TypeNS:
		ns := &NS{}
		err := ns.UnmarshalBinary(data)
		return ns, err
	case TypeCNAME:
		cname := &CNAME{}
		err := cname.UnmarshalBinary(data)
		return cname, err
	case TypeMX:
		mx := &MX{}
		err := mx.UnmarshalBinary(data)
		return mx, err
	case TypeTXT:
		txt := &TXT{}
		err := txt.UnmarshalBinary(data)
		return txt, err
	default:
		return nil, errors.New("Uknown RDATA type")
	}

}

func ParseRData(t Type, s string) (RData, error) {
	switch t {
	case TypeA:
		ip, err := netip.ParseAddr(s)
		if err != nil {
			return nil, fmt.Errorf("parse addr: %w", err)
		}
		if !ip.Is4() {
			return nil, fmt.Errorf("IP should be IPv4")
		}
		return &A{IP: ip}, nil

	case TypeAAAA:
		ip, err := netip.ParseAddr(s)
		if err != nil {
			return nil, fmt.Errorf("parse addr: %w", err)
		}
		if !ip.Is6() {
			return nil, fmt.Errorf("IP should be IPv6")
		}
		return &AAAA{IP: ip}, nil

	case TypeNS:
		return &NS{name: s}, nil

	case TypeCNAME:
		return &CNAME{name: s}, nil

	case TypeMX:
		parts := strings.SplitN(s, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid MX format: expected 'preference name'")
		}
		pref, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid preference: %w", err)
		}
		return &MX{preference: uint16(pref), name: parts[1]}, nil

	case TypeTXT:
		return &TXT{data: []byte(s)}, nil

	default:
		return nil, fmt.Errorf("unknown type: %v", t)
	}
}
