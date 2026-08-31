// Package dns contains simple implementation of RFC 1035.
// File rr.go contains implementation of Rersource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-3.2
package dns

import "errors"

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
