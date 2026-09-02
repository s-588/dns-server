// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.11
package dns

import (
	"testing"
)

func TestNSRecordRoundTrip(t *testing.T) {
	nsname := "ns1.example.com"
	ns := &NS{name: nsname}
	data, err := ns.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	var ns2 NS
	err = ns2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if ns2.name != nsname {
		t.Errorf("name = %q, want %q", ns2.name, nsname)
	}
}
