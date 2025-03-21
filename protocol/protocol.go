// In this file defines DNS messages from RFC 1035
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Represent header of any message
type Header struct {
	ID uint16

	// See RFC 1035 25-27 pages for details
	// 0... .... .... .... = QR: Message is a query,
	// .000 0... .... .... = OPCODE: Standard query (0),
	// .... .0.. .... .... = AA:  Authoritative Answer
	// .... ..0. .... .... = TC: Message is not truncated,
	// .... ...1 .... .... = RD: Do query recursively,
	// .... .... 1... .... = RA: Recursion AvaRecursion Availableilable
	// .... .... .0.. .... = Z: reserved (0),
	// .... .... ..00 00.. = RCODE: No error condition

	Flags uint16

	// Specify the number of entries in the question section.
	QuestionsCount uint16

	// Specify the number of resource records in the answer section.
	AnswersCount uint16

	// Specify the number of name server resource records
	// in the authority records section.
	AuthorityCount uint16

	// Specify the number of
	// resource records in the additional records section.
	AdditionalCount uint16
}

type ResourceRecord struct {
	Domain string
	Type   uint16
	Class  uint16

	// specifies the interval (in seconds) that the resource record may be
	// cached before it should be discarded.
	TTL uint32

	DataLength uint16
	Data       []byte
}

// Read raw request, return request header and slice questions
func DecodeRequest(request []byte) (*Header, []*ResourceRecord, error) {
	reqBuffer := bytes.NewBuffer(request)

	header, err := DecodeHeader(reqBuffer)
	if err != nil {
		return header, make([]*ResourceRecord, 0), fmt.Errorf("decode request: %w", err)
	}

	questions := make([]*ResourceRecord, header.QuestionsCount)
	for i := range questions {
		questions[i], err = DecodeRecordBody(reqBuffer)
		if err != nil {
			return header, questions, fmt.Errorf("decode request: %w", err)
		}
	}

	return header, questions, nil
}

// Read from reader a undecoded binary header
func DecodeHeader(buffer *bytes.Buffer) (*Header, error) {
	header := &Header{}
	err := binary.Read(buffer, binary.BigEndian, header)
	if err != nil {
		return header, fmt.Errorf("decode header: %w", err)
	}
	return header, nil
}

// Read domain, type and class from request buffer.
// Must be used after the header was readed.
func DecodeRecordBody(buffer *bytes.Buffer) (*ResourceRecord, error) {
	body := &ResourceRecord{}
	// RFC 1035: "a domain name represented as a sequence of labels, where
	// each label consists of a length octet followed by that
	// number of octets."
	domainLen, err := buffer.ReadByte()
	if err != nil {
		return body, fmt.Errorf("decode body: %w", err)
	}

	domainBytes := make([]byte, domainLen)
	if _, err := buffer.Read(domainBytes); err != nil {
		return body, fmt.Errorf("decode body: %w", err)
	}
	body.Domain = string(domainBytes)
	body.Type = binary.BigEndian.Uint16(buffer.Next(2))
	body.Class = binary.BigEndian.Uint16(buffer.Next(2))

	return body, nil
}
