// Package dns contains simple implementation of RFC 1035.
// https://datatracker.ietf.org/doc/html/rfc1035#section-3.3.14
package dns

import (
	"bytes"
	"testing"
)

func TestTXTRecordRoundTrip(t *testing.T) {
	txtData := []byte("hello world")
	txt := &TXT{data: txtData}
	data, err := txt.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	if !bytes.Equal(data, txtData) {
		t.Errorf("MarshalBinary = %v, want %v", data, txtData)
	}
	var txt2 TXT
	err = txt2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if !bytes.Equal(txt2.data, txtData) {
		t.Errorf("UnmarshalBinary data = %v, want %v", txt2.data, txtData)
	}

	// Empty data should error on Marshal
	empty := &TXT{data: []byte{}}
	_, err = empty.MarshalBinary()
	if err == nil {
		t.Error("MarshalBinary with empty data should error")
	}
}
