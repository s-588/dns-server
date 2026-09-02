// Package dns contains simple implementation of RFC 1035.
// File rr.go contains implementation of Rersource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-3.2
package dns

import (
	"bytes"
	"net/netip"
	"testing"
)

func TestRDataTypes(t *testing.T) {
	// Ensure each RData type returns the correct Type() value.
	tests := []struct {
		r    RData
		want Type
	}{
		{&A{}, TypeA},
		{&AAAA{}, TypeAAAA},
		{&CNAME{}, TypeCNAME},
		{&MX{}, TypeMX},
		{&NS{}, TypeNS},
		{&TXT{}, TypeTXT},
	}
	for _, tt := range tests {
		if got := tt.r.Type(); got != tt.want {
			t.Errorf("%T.Type() = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestParseRData(t *testing.T) {
	tests := []struct {
		typ   Type
		input string
		want  RData
		err   bool
	}{
		{TypeA, "192.0.2.1", &A{IP: netip.MustParseAddr("192.0.2.1")}, false},
		{TypeA, "invalid", nil, true},
		{TypeA, "::1", nil, true}, // IPv6 not allowed
		{TypeAAAA, "2001:db8::1", &AAAA{IP: netip.MustParseAddr("2001:db8::1")}, false},
		{TypeAAAA, "192.0.2.1", nil, true},
		{TypeNS, "ns1.example.com.", &NS{name: "ns1.example.com."}, false},
		{TypeCNAME, "alias.example.com.", &CNAME{name: "alias.example.com."}, false},
		{TypeMX, "10 mail.example.com", &MX{preference: 10, name: "mail.example.com"}, false},
		{TypeMX, "invalid", nil, true},
		{TypeMX, "abc mail", nil, true}, // invalid preference
		{TypeTXT, "some text", &TXT{data: []byte("some text")}, false},
		{Type(99), "anything", nil, true},
	}
	for _, tt := range tests {
		got, err := ParseRData(tt.typ, tt.input)
		if tt.err && err == nil {
			t.Errorf("ParseRData(%v, %q) expected error, got nil", tt.typ, tt.input)
			continue
		}
		if !tt.err {
			if err != nil {
				t.Errorf("ParseRData(%v, %q) unexpected error: %v", tt.typ, tt.input, err)
				continue
			}
			// Compare by marshal equality (simpler)
			wantData, _ := tt.want.MarshalBinary()
			gotData, _ := got.MarshalBinary()
			if !bytes.Equal(wantData, gotData) {
				t.Errorf("ParseRData(%v, %q) = %v, want %v", tt.typ, tt.input, got, tt.want)
			}
		}
	}
}

func TestDecodeRData(t *testing.T) {
	// We'll test by encoding some records then decoding.
	tests := []struct {
		typ  Type
		data []byte
		want RData
		err  bool
	}{
		{TypeA, []byte{192, 0, 2, 1}, &A{IP: netip.MustParseAddr("192.0.2.1")}, false},
		{TypeA, []byte{1, 2, 3}, nil, true}, // wrong length
		{TypeAAAA, []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, &AAAA{IP: netip.MustParseAddr("2001:db8::1")}, false},
		{TypeNS, []byte{3, 'n', 's', '1', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}, &NS{name: "ns1.example.com"}, false},
		{TypeCNAME, []byte{5, 'a', 'l', 'i', 'a', 's', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}, &CNAME{name: "alias.example.com"}, false},
		{TypeMX, []byte{0x00, 0x0a, 4, 'm', 'a', 'i', 'l', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}, &MX{preference: 10, name: "mail.example.com"}, false},
		{TypeTXT, []byte("hello"), &TXT{data: []byte("hello")}, false},
		{Type(99), []byte{1, 2, 3}, nil, true},
	}
	for _, tt := range tests {
		got, err := decodeRData(tt.typ, tt.data)
		if tt.err && err == nil {
			t.Errorf("decodeRData(%v, %v) expected error, got nil", tt.typ, tt.data)
			continue
		}
		if !tt.err {
			if err != nil {
				t.Errorf("decodeRData(%v, %v) unexpected error: %v", tt.typ, tt.data, err)
				continue
			}
			// Compare by marshal
			wantData, _ := tt.want.MarshalBinary()
			gotData, _ := got.MarshalBinary()
			if !bytes.Equal(wantData, gotData) {
				t.Errorf("decodeRData(%v, %v) = %v, want %v", tt.typ, tt.data, got, tt.want)
			}
		}
	}
}
