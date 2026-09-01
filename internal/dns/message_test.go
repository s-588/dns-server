// Package dns contains simple implementation of RFC 1035.
// File message.go contains of Message according to https://datatracker.ietf.org/doc/html/rfc1035#section-4
package dns

import (
	"bytes"
	"net/netip"
	"reflect"
	"testing"

	"github.com/prionis/dns-server/internal/dns/codec"
)

func Test_QuestionRoundTrip(t *testing.T) {
	w := codec.NewWriter()

	q := Question{
		Name:  "example.com",
		Type:  TypeA,
		Class: ClassIN,
	}
	err := encodeQuestion(w, q)
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	r := codec.NewReader(w.Buffer())
	newQ, err := decodeQuestion(r)
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	if !reflect.DeepEqual(newQ, q) {
		t.Errorf("got %v, want %v", newQ, q)
	}
}

func Test_RR_RoundTrip(t *testing.T) {
	w := codec.NewWriter()

	a := &A{
		IP: netip.MustParseAddr("240.99.18.23"),
	}
	b, err := a.MarshalBinary()
	rr := RR{
		Name:     "example.com",
		Type:     TypeA,
		Class:    ClassIN,
		TTL:      88,
		RDLength: uint16(len(b)),
		RData:    a,
	}
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	err = encodeRR(w, rr)
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	r := codec.NewReader(w.Buffer())
	newQ, err := decodeQuestion(r)
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	if !reflect.DeepEqual(newQ, rr) {
		t.Errorf("got %v, want %v", newQ, rr)
	}
}

func TestMessage_BinaryRoundTrip(t *testing.T) {
	a := &A{
		IP: netip.MustParseAddr("240.99.18.23"),
	}
	b, err := a.MarshalBinary()
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	rr := RR{
		Name:     "example.com",
		Type:     TypeA,
		Class:    ClassIN,
		TTL:      88,
		RDLength: uint16(len(b)),
		RData:    a,
	}
	m := &Message{
		Header: Header{
			ID:      88,
			Flags:   0,
			QDCount: 1,
			ANCount: 1,
			NSCount: 0,
			ARCount: 0,
		},
		Questions: []Question{
			{
				Name:  "example.com",
				Type:  TypeA,
				Class: ClassIN,
			},
		},
		Answers: []RR{
			rr,
		},
		Authorities: []RR{},
		Additionals: []RR{},
	}
	d, err := m.MarshalBinary()
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	var newM Message
	err = newM.UnmarshalBinary(d)
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}

	bm, err := m.MarshalBinary()
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	bNew, err := newM.MarshalBinary()
	if err != nil {
		t.Fatalf("error was not expected: %s", err)
	}
	if !bytes.Equal(bm, bNew) {
		t.Fatalf("got %v, want %v", m, newM)
	}
}
