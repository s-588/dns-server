// Package dns contains simple implementation of RFC 1035.
// File classes.go contains CLASS values according to RFC 1035 https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.4
package dns

import "testing"

func TestClassString(t *testing.T) {
	tests := []struct {
		class Class
		want  string
	}{
		{ClassIN, "IN"},
		{ClassCS, "CS"},
		{ClassCH, "CH"},
		{ClassHS, "HS"},
		{Class(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.class.String(); got != tt.want {
			t.Errorf("Class(%d).String() = %q, want %q", tt.class, got, tt.want)
		}
	}
}

func TestParseClass(t *testing.T) {
	tests := []struct {
		s    string
		want Class
		ok   bool
	}{
		{"IN", ClassIN, true},
		{"CS", ClassCS, true},
		{"CH", ClassCH, true},
		{"HS", ClassHS, true},
		{"in", 0, false},
		{"", 0, false},
		{"UNKNOWN", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseClass(tt.s)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ParseClass(%q) = (%v, %v), want (%v, %v)", tt.s, got, ok, tt.want, tt.ok)
		}
	}
}
