package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDecodeHeader(t *testing.T) {
	headers := []*Header{
		{
			ID:              1,
			Flags:           uint16(0b0000_0000_0000_0000),
			QuestionsCount:  0,
			AnswersCount:    0,
			AuthorityCount:  0,
			AdditionalCount: 0,
		},
		{
			ID:              2,
			Flags:           uint16(0b1101_0100_1000_0000),
			QuestionsCount:  1,
			AnswersCount:    2,
			AuthorityCount:  3,
			AdditionalCount: 0,
		},
		{
			ID:              3,
			Flags:           uint16(0b0000_0000_0000_0000),
			QuestionsCount:  3,
			AnswersCount:    1,
			AuthorityCount:  0,
			AdditionalCount: 9,
		},
	}
	for _, h := range headers {
		data := make([]byte, 13)
		n, err := binary.Encode(data, binary.BigEndian, h)
		if err != nil {
			t.Errorf("can't encode header to binary format: %v", err)
		}
		if n != 13 {
			t.Errorf("header was encoded wrong, count of writen bytes must be 13, n = %d", n)
		}

		header, err := DecodeHeader(bytes.NewBuffer(data))
		if err != nil {
			t.Errorf("function return error, but must not: %v", err)
		}
		if *header != *h {
			t.Errorf("Processed header not equals original:\n%v\n%v", header, h)
		}
	}
}

func TestDecodeRecordBody(t *testing.T) {
	records := []*RR{
		{
			Domain: "10google.com",
			Type:   binary.BigEndian.Uint16([]byte("A0")),
			Class:  binary.BigEndian.Uint16([]byte("IN")),
		},
	}
	for _, r := range records {
		data := make([]byte, len(r.Domain)+4)
		n, err := binary.Encode(data, binary.BigEndian, r)
		if err != nil {
			t.Errorf("can't encode record body to binary format: %v", err)
		}
		if n != len(r.Domain)+4 {
			t.Errorf("record body was encoded wrong,"+
				" count of writen bytes must be %d, n = %d", len(r.Domain)+4, n)
		}
		buffer := bytes.NewBuffer(data)

		result, err := DecodeRR(buffer)
		if err != nil {
			t.Errorf("function return error, but must not: %v", err)
		}
		if result.Type != r.Type || result.Class != result.Type {
			t.Errorf("body was decoded wrong:\nexpected%v\nget%v", r, result)
		}
	}
}
