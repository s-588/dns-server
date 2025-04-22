package protocol

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

type encodeRRtest struct {
	name string
	rr   *RR
	get  *bytes.Buffer
	want *bytes.Buffer
	err  error
}

var encodeRRtests = []encodeRRtest{
	{
		"normal resource record",
		&RR{
			"example.com",
			Types["A"],
			Classes["IN"],
			3600,
			4,
			[]byte{1, 1, 1, 1},
		},
		bytes.NewBuffer([]byte{}),
		bytes.NewBuffer([]byte{
			7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
			3, 'c', 'o', 'm',
			0,
			0, 1,
			0, 1,
			0, 0, 14, 16,
			0, 4,
			1, 1, 1, 1,
		}),
		nil,
	},
	{
		"wrong resource record",
		&RR{
			"example..com",
			Types["A"],
			Classes["IN"],
			3600,
			4,
			[]byte{1, 1, 1, 1},
		},
		bytes.NewBuffer([]byte{}),
		bytes.NewBuffer([]byte{
			7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
			3, 'c', 'o', 'm',
			0,
			0, 1,
			0, 1,
			0, 0, 14, 16,
			0, 4,
			1, 1, 1, 1,
		}),
		errors.New("wrong label"),
	},
}

func TestEncodeRR(t *testing.T) {
	for _, test := range encodeRRtests {
		err := EncodeRR(test.get, test.rr)
		if err != nil && test.err == nil {
			t.Errorf("%s test: unexpected error: %v", test.name, test.err)
		}

		if !slices.Equal(test.get.Bytes(), test.want.Bytes()) && test.err == nil {
			t.Errorf("%s test: \nget:  %x\nwant: %x", test.name, test.get, test.want)
		}
	}
}

type encodeResponseTest struct {
	name    string
	dnsmsg  DNSMessage
	wantBuf []byte
	wantErr error
}

var encodeResponseTests = []encodeResponseTest{
	{
		"normal message",
		DNSMessage{
			Head: &Header{
				ID:             0x1234,
				Flags:          0x8180, // standard response flags
				QuestionsCount: 1,
			},
			Questions: []*RR{
				{
					Domain: "example.com",
					Type:   1, // A record
					Class:  1, // IN
				},
			},
			Answers: []*RR{
				{
					Domain:     "example.com",
					Type:       1,
					Class:      1,
					TimeToLive: 300,
					Data:       []byte{127, 0, 0, 1},
				},
			},
		},
		[]byte{
			// Header
			0x00, 0x01, // id
			0x81, 0x80, // flags
			0x00, 0x01, // count of questions
			0x00, 0x01, // count of answers
			0x00, 0x00, // count of authority
			0x00, 0x00, // count of addittions
			0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01, // Question
			0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2c, 0x00, 0x04, 0x7f, 0x00, 0x00, 0x01, // Answer
		},
		nil,
	},
}

func TestEncodeResponse(t *testing.T) {
	for _, test := range encodeResponseTests {
		buf, err := EncodeResponse(test.dnsmsg)
		if err != nil && test.wantErr == nil {
			t.Errorf("%s test, unexpected error: %v", test.name, err)
		}
		if !slices.Equal(buf, test.wantBuf) {
			t.Errorf("%s test\nget:  %v\nwant: %v", test.name, buf, test.wantBuf)
		}
	}
}

func TestDecodeRequest(t *testing.T) {
}

func TestDecodeHeader(t *testing.T) {
}

func TestDecodeQuestion(t *testing.T) {
}
