// Package dns contains simple implementation of RFC 1035.
package dns

import (
	"bytes"
	"net/netip"
	"testing"
)

func TestAAAARecordRoundTrip(t *testing.T) {
	ip := netip.MustParseAddr("2001:db8::1")
	aaaa := &AAAA{IP: ip}
	data, err := aaaa.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	expected := ip.AsSlice()
	if !bytes.Equal(data, expected) {
		t.Errorf("MarshalBinary = %v, want %v", data, expected)
	}

	var aaaa2 AAAA
	err = aaaa2.UnmarshalBinary(data)
	if err == nil {
		if aaaa2.IP != ip {
			t.Errorf("UnmarshalBinary IP = %v, want %v", aaaa2.IP, ip)
		}
	}

	// Invalid IPv4 should error
	bad := &AAAA{IP: netip.MustParseAddr("192.0.2.1")}
	_, err = bad.MarshalBinary()
	if err == nil {
		t.Error("MarshalBinary with IPv4 should error")
	}
}
