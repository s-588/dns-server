// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.14
package dns

import (
	"errors"

	"github.com/prionis/dns-server/internal/dns/codec"
)

// TXT represents a DNS TXT record.
type TXT struct {
	data []byte
}

// Type returns the DNS record type for TXT records.
func (txt TXT) Type() Type {
	return TypeTXT
}

// MarshalBinary marshals the TXT record into binary format.
func (txt TXT) MarshalBinary() ([]byte, error) {
	if len(txt.data) == 0 {
		return nil, errors.New("empty txt data")
	}
	w := codec.NewWriter()
	w.Bytes(txt.data)
	return w.Buffer(), nil
}

// UnmarshalBinary unmarshals the binary data into the TXT record.
func (txt *TXT) UnmarshalBinary(data []byte) error {
	txt.data = data
	return nil
}
