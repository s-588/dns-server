// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.9
package dns

import (
	"testing"
)

func TestMXRecordRoundTrip(t *testing.T) {
	pref := uint16(10)
	domain := "mail.example.com"
	mx := &MX{preference: pref, name: domain}
	data, err := mx.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	var mx2 MX
	err = mx2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if mx2.preference != pref {
		t.Errorf("preference = %d, want %d", mx2.preference, pref)
	}
	if mx2.name != domain {
		t.Errorf("name = %q, want %q", mx2.name, domain)
	}

	// Missing preference
	err = mx2.UnmarshalBinary([]byte{0x00})
	if err == nil {
		t.Error("UnmarshalBinary with too short data should error")
	}
}
