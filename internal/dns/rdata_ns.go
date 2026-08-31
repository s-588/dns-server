// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.11
package dns

import (
	"fmt"

	"github.com/prionis/dns-server/internal/dns/codec"
)

type NS struct {
	name string
}

func (ns NS) Type() Type {
	return TypeNS
}

func (ns NS) MarshalBinary() ([]byte, error) {
	w := codec.NewWriter()
	err := w.WriteName(ns.name)
	return w.Buffer(), err
}

func (ns *NS) UnmarshalBinary(data []byte) error {
	r := codec.NewReader(data)
	name, err := r.ReadName()
	if err != nil {
		return fmt.Errorf("domain name unmarshal: %w", err)
	}
	ns.name = name
	return nil
}
