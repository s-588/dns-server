// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.1
package dns

import (
	"testing"
)

func TestCNAMERecordRoundTrip(t *testing.T) {
	name := "www.example.com"
	cname := &CNAME{name: name}
	data, err := cname.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	// We'll just check that we can unmarshal back.
	var cname2 CNAME
	err = cname2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if cname2.name != name {
		t.Errorf("UnmarshalBinary name = %q, want %q", cname2.name, name)
	}

	// Empty data should error
	err = cname2.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("UnmarshalBinary with empty data should error")
	}
}
