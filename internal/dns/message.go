// Package dns contains simple implementation of RFC 1035.
// File message.go contains of Message according to https://datatracker.ietf.org/doc/html/rfc1035#section-4
package dns

import (
	"fmt"
	"math"

	"github.com/prionis/dns-server/internal/dns/codec"
)

// Message represents a DNS message.
type Message struct {
	Header      Header
	Questions   []Question
	Answers     []RR
	Authorities []RR
	Additionals []RR
}

// Question represents a DNS question section.
type Question struct {
	Name  string
	Type  Type
	Class Class
}

// encodeQuestion encodes a DNS question into the provided codec.Writer.
func encodeQuestion(w *codec.Writer, q Question) error {
	if err := w.WriteName(q.Name); err != nil {
		return fmt.Errorf("write domain name: %w", err)
	}

	w.Uint16(uint16(q.Type))
	w.Uint16(uint16(q.Class))

	return nil
}

// decodeQuestion decodes a DNS question from the provided codec.Reader.
func decodeQuestion(r *codec.Reader) (Question, error) {
	var q Question
	var err error
	q.Name, err = r.ReadName()
	if err != nil {
		return Question{}, fmt.Errorf("read domain name: %w", err)
	}

	t, err := r.Uint16()
	if err != nil {
		return Question{}, fmt.Errorf("read type: %w", err)
	}
	q.Type = Type(t)

	c, err := r.Uint16()
	if err != nil {
		return Question{}, fmt.Errorf("read class: %w", err)
	}
	q.Class = Class(c)

	return q, nil
}

// encodeRR encodes a DNS resource record into the provided codec.Writer.
func encodeRR(w *codec.Writer, rr RR) error {
	if err := w.WriteName(rr.Name); err != nil {
		return fmt.Errorf("write domain name: %w", err)
	}

	w.Uint16(uint16(rr.Type))
	w.Uint16(uint16(rr.Class))
	w.Uint32(rr.TTL)

	rdata, err := rr.RData.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal RR data: %w", err)
	}
	l := len(rdata)
	if l < 0 || l > math.MaxUint16 {
		return fmt.Errorf("length of rdata is too big: %d", l)
	}
	w.Uint16(uint16(l))
	w.Bytes(rdata)

	return nil
}

// decodeRR decodes a DNS resource record from the provided codec.Reader.
func decodeRR(r *codec.Reader) (RR, error) {
	var rr RR
	var err error
	rr.Name, err = r.ReadName()
	if err != nil {
		return RR{}, fmt.Errorf("read domain name: %w", err)
	}

	t, err := r.Uint16()
	if err != nil {
		return RR{}, fmt.Errorf("read type: %w", err)
	}
	rr.Type = Type(t)

	c, err := r.Uint16()
	if err != nil {
		return RR{}, fmt.Errorf("read class: %w", err)
	}
	rr.Class = Class(c)

	rr.TTL, err = r.Uint32()
	if err != nil {
		return RR{}, fmt.Errorf("read TTL: %w", err)
	}

	rr.RDLength, err = r.Uint16()
	if err != nil {
		return RR{}, fmt.Errorf("read RDLength: %w", err)
	}

	raw, err := r.Bytes(int(rr.RDLength))
	if err != nil {
		return RR{}, fmt.Errorf("marshal RR data: %w", err)
	}

	rdata, err := decodeRData(rr.Type, raw)
	if err != nil {
		return RR{}, fmt.Errorf("decode RData: %w", err)
	}
	rr.RData = rdata

	return rr, nil
}

// MarshalBinary marshals the Message into binary format.
func (m *Message) MarshalBinary() ([]byte, error) {
	w := codec.NewWriter()

	w.Uint16(m.Header.ID)
	w.Uint16(m.Header.Flags)

	err := putUint16(len(m.Questions), w)
	if err != nil {
		return nil, err
	}

	err = putUint16(len(m.Answers), w)
	if err != nil {
		return nil, err
	}

	err = putUint16(len(m.Authorities), w)
	if err != nil {
		return nil, err
	}

	err = putUint16(len(m.Additionals), w)
	if err != nil {
		return nil, err
	}

	for _, q := range m.Questions {
		if err := encodeQuestion(w, q); err != nil {
			return nil, fmt.Errorf("encode question: %w", err)
		}
	}

	for _, a := range m.Answers {
		if err := encodeRR(w, a); err != nil {
			return nil, fmt.Errorf("encode answer: %w", err)
		}
	}

	for _, a := range m.Authorities {
		if err := encodeRR(w, a); err != nil {
			return nil, fmt.Errorf("encode authority: %w", err)
		}
	}

	for _, a := range m.Additionals {
		if err := encodeRR(w, a); err != nil {
			return nil, fmt.Errorf("encode additional: %w", err)
		}
	}

	return w.Buffer(), nil
}

func putUint16(v int, w *codec.Writer) error {
	if v < 0 || v > math.MaxUint16 {
		return fmt.Errorf("too many questions: %d", v)
	}
	w.Uint16(uint16(v))
	return nil
}

// UnmarshalBinary unmarshals the binary data into the Message.
func (m *Message) UnmarshalBinary(data []byte) error {
	r := codec.NewReader(data)
	var err error

	m.Header.ID, err = r.Uint16()
	if err != nil {
		return fmt.Errorf("header ID: %w", err)
	}
	m.Header.Flags, err = r.Uint16()
	if err != nil {
		return fmt.Errorf("header flags: %w", err)
	}
	m.Header.QDCount, err = r.Uint16()
	if err != nil {
		return fmt.Errorf("header QDCount: %w", err)
	}
	m.Header.ANCount, err = r.Uint16()
	if err != nil {
		return fmt.Errorf("header ANCount: %w", err)
	}
	m.Header.NSCount, err = r.Uint16()
	if err != nil {
		return fmt.Errorf("header NSCount: %w", err)
	}
	m.Header.ARCount, err = r.Uint16()
	if err != nil {
		return fmt.Errorf("header ARCount: %w", err)
	}

	for range m.Header.QDCount {
		var q Question
		q, err := decodeQuestion(r)
		if err != nil {
			return fmt.Errorf("decode question: %w", err)
		}
		m.Questions = append(m.Questions, q)
	}

	for range m.Header.ANCount {
		var rr RR
		rr, err := decodeRR(r)
		if err != nil {
			return fmt.Errorf("decode answers: %w", err)
		}
		m.Answers = append(m.Answers, rr)
	}

	for range m.Header.NSCount {
		var rr RR
		rr, err := decodeRR(r)
		if err != nil {
			return fmt.Errorf("decode authorities: %w", err)
		}
		m.Authorities = append(m.Authorities, rr)
	}

	for range m.Header.ARCount {
		var rr RR
		rr, err := decodeRR(r)
		if err != nil {
			return fmt.Errorf("decode additionals: %w", err)
		}
		m.Additionals = append(m.Additionals, rr)
	}
	return nil
}
