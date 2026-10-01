// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.1
package dns

import (
	"fmt"

	"github.com/prionis/dns-server/internal/dns/codec"
)

// CNAME represents a DNS CNAME record.
type CNAME struct {
	name string
}

// Type returns the DNS record type for CNAME records.
func (cname CNAME) Type() Type {
	return TypeCNAME
}

// MarshalBinary marshals the CNAME record into binary format.
func (cname CNAME) MarshalBinary() ([]byte, error) {
	w := codec.NewWriter()
	err := w.WriteName(cname.name)
	if err != nil {
		return nil, fmt.Errorf("domain name marshal: %w", err)
	}
	return w.Buffer(), nil
}

// UnmarshalBinary unmarshals the binary data into the CNAME record.
func (cname *CNAME) UnmarshalBinary(data []byte) error {
	r := codec.NewReader(data)
	name, err := r.ReadName()
	if err != nil {
		return fmt.Errorf("domain name unmarshal: %w", err)
	}
	cname.name = name
	return nil
}
