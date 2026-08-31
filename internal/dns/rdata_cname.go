// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.1
package dns

import (
	"fmt"

	"github.com/prionis/dns-server/internal/dns/codec"
)

type CNAME struct {
	name string
}

func (cname CNAME) Type() Type {
	return TypeCNAME
}

func (cname CNAME) MarshalBinary() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteName(cname.name)
	return w.Buffer(), nil
}

func (cname *CNAME) UnmarshalBinary(data []byte) error {
	r := codec.NewReader(data)
	name, err := r.ReadName()
	if err != nil {
		return fmt.Errorf("domain name unmarshal: %w", err)
	}
	cname.name = name
	return nil
}
