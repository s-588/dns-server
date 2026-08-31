// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.9
package dns

import (
	"fmt"

	"github.com/prionis/dns-server/internal/dns/codec"
)

type MX struct {
	preference uint16
	name       string
}

func (mx MX) Type() Type {
	return TypeMX
}

func (mx MX) MarshalBinary() ([]byte, error) {
	w := codec.NewWriter()
	w.Uint16(mx.preference)
	err := w.WriteName(mx.name)
	return w.Buffer(), err
}

func (mx *MX) UnmarshalBinary(data []byte) error {
	r := codec.NewReader(data)
	pref, err := r.Uint16()
	if err != nil {
		return fmt.Errorf("domain name unmarshal: %w", err)
	}
	mx.preference = pref
	name, err := r.ReadName()
	if err != nil {
		return fmt.Errorf("domain name unmarshal: %w", err)
	}
	mx.name = name
	return nil
}
