// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc3596#section-2
package dns

import (
	"bytes"
	"net/netip"
	"testing"
)

func TestARecordRoundTrip(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.1")
	a := &A{IP: ip}
	data, err := a.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	expected := []byte{192, 0, 2, 1}
	if !bytes.Equal(data, expected) {
		t.Errorf("MarshalBinary = %v, want %v", data, expected)
	}

	var a2 A
	err = a2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if a2.IP != ip {
		t.Errorf("UnmarshalBinary IP = %v, want %v", a2.IP, ip)
	}

	// Invalid IPv6 should error
	bad := &A{IP: netip.MustParseAddr("::1")}
	_, err = bad.MarshalBinary()
	if err == nil {
		t.Error("MarshalBinary with IPv6 should error")
	}

	// Invalid data length (too short)
	err = a2.UnmarshalBinary([]byte{192, 0, 2}) // missing last byte
	if err == nil {
		t.Error("UnmarshalBinary with truncated data should error")
	}
}
