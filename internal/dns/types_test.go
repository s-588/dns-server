// Package dns contains simple implementation of RFC 1035.
// File types.go contains TYPE values from https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.3
package dns

import "testing"

func TestTypeString(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{TypeA, "A"},
		{TypeNS, "NS"},
		{TypeCNAME, "CNAME"},
		{TypeSOA, "SOA"},
		{TypePTR, "PTR"},
		{TypeMX, "MX"},
		{TypeTXT, "TXT"},
		{TypeAAAA, "AAAA"},
		{Type(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("Type(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		s    string
		want Type
		ok   bool
	}{
		{"A", TypeA, true},
		{"NS", TypeNS, true},
		{"CNAME", TypeCNAME, true},
		{"SOA", TypeSOA, true},
		{"PTR", TypePTR, true},
		{"MX", TypeMX, true},
		{"TXT", TypeTXT, true},
		{"AAAA", TypeAAAA, true},
		{"a", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseType(tt.s)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ParseType(%q) = (%v, %v), want (%v, %v)", tt.s, got, ok, tt.want, tt.ok)
		}
	}
}
