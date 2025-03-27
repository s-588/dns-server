package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

var (
	Types = map[string]uint16{
		"A":     uint16(1),
		"NS":    uint16(2),
		"MD":    uint16(3),
		"MF":    uint16(4),
		"CNAME": uint16(5),
		"SOA":   uint16(6),
		"MB":    uint16(7),
		"MG":    uint16(8),
		"MR":    uint16(9),
		"NULL":  uint16(10),
		"WKS":   uint16(11),
		"PTR":   uint16(12),
		"HINFO": uint16(13),
		"MINFO": uint16(14),
		"MX":    uint16(15),
		"TXT":   uint16(16),
	}

	Classes = map[string]uint16{
		"IN": uint16(1),
		"CS": uint16(2),
		"CH": uint16(3),
		"HS": uint16(4),
	}
)

type DNSMessage struct {
	Head        *Header
	Questions   []*RR
	Answers     []*RR
	Authorities []*RR
	Additionals []*RR
}

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

// Represent resource record
type RR struct {
	Domain string
	Type   uint16
	Class  uint16

	// Specifies the interval (in seconds) that the resource record may be
	// cached before it should be discarded.
	TimeToLive uint32

	DataLen uint16
	Data    []byte
}

// Encode header and answers, return encoded bytes
func EncodeResponse(message DNSMessage) ([]byte, error) {
	response := bytes.NewBuffer(make([]byte, 0))

	log.Println(message.Head)
	header := Header{
		ID:              uint16(message.Head.ID),
		Flags:           uint16(1 << 15),
		QuestionsCount:  message.Head.QuestionsCount,
		AnswersCount:    uint16(len(message.Answers)),
		AuthorityCount:  uint16(len(message.Authorities)),
		AdditionalCount: uint16(len(message.Additionals)),
	}
	log.Println(header)

	err := binary.Write(response, binary.BigEndian, &header)
	if err != nil {
		return response.Bytes(), err
	}
	log.Println(response.Bytes())

	for _, question := range message.Questions {
		err := EncodeDomain(response, question.Domain)
		if err != nil {
			return response.Bytes(), err
		}

		err = binary.Write(response, binary.BigEndian, question.Type)
		if err != nil {
			return response.Bytes(), err
		}

		err = binary.Write(response, binary.BigEndian, question.Class)
		if err != nil {
			return response.Bytes(), err
		}
	}

	for _, answer := range message.Answers {
		err := EncodeRR(response, answer)
		if err != nil {
			return response.Bytes(), err
		}
	}

	for _, authority := range message.Authorities {
		err := EncodeRR(response, authority)
		if err != nil {
			return response.Bytes(), err
		}
	}

	for _, addition := range message.Additionals {
		err := EncodeRR(response, addition)
		if err != nil {
			return response.Bytes(), err
		}
	}

	return response.Bytes(), nil
}

func EncodeRR(buffer *bytes.Buffer, rr *RR) error {
	err := EncodeDomain(buffer, rr.Domain)
	if err != nil {
		return fmt.Errorf("encode RR: %w", err)
	}

	err = binary.Write(buffer, binary.BigEndian, rr.Type)
	if err != nil {
		return fmt.Errorf("encode RR type: %w", err)
	}

	err = binary.Write(buffer, binary.BigEndian, rr.Class)
	if err != nil {
		return fmt.Errorf("encode RR class: %w", err)
	}

	err = binary.Write(buffer, binary.BigEndian, rr.TimeToLive)
	if err != nil {
		return fmt.Errorf("encode RR TTL: %w", err)
	}

	err = binary.Write(buffer, binary.BigEndian, rr.DataLen)
	if err != nil {
		return fmt.Errorf("encode RR data len: %w", err)
	}

	err = binary.Write(buffer, binary.BigEndian, rr.Data)
	if err != nil {
		return fmt.Errorf("encode RR data: %w", err)
	}

	return nil
}

func EncodeDomain(buffer *bytes.Buffer, domain string) error {
	err := buffer.WriteByte(byte(len(domain)))
	if err != nil {
		return fmt.Errorf("encode domain len: %w", err)
	}

	_, err = buffer.WriteString(domain)
	if err != nil {
		return fmt.Errorf("encode domain: %w", err)
	}

	return nil
}

// Read raw request, return request header and slice of questions
func DecodeRequest(request []byte) (DNSMessage, error) {
	message := DNSMessage{}
	reqBuffer := bytes.NewBuffer(request)

	header, err := DecodeHeader(reqBuffer)
	if err != nil {
		return message, fmt.Errorf("decode request: %w", err)
	}
	message.Head = header

	questions := make([]*RR, header.QuestionsCount)
	for i := range questions {
		questions[i], err = DecodeQuestion(reqBuffer)
		if err != nil {
			return message, fmt.Errorf("decode request: %w", err)
		}
	}
	message.Questions = questions

	answers := make([]*RR, header.AnswersCount)
	for i := range answers {
		answers[i], err = DecodeQuestion(reqBuffer)
		if err != nil {
			return message, fmt.Errorf("decode request: %w", err)
		}
	}
	message.Answers = answers

	authorities := make([]*RR, header.AnswersCount)
	for i := range authorities {
		authorities[i], err = DecodeQuestion(reqBuffer)
		if err != nil {
			return message, fmt.Errorf("decode request: %w", err)
		}
	}
	message.Authorities = authorities

	additionals := make([]*RR, header.AdditionalCount)
	for i := range additionals {
		additionals[i], err = DecodeQuestion(reqBuffer)
		if err != nil {
			return message, fmt.Errorf("decode request: %w", err)
		}
	}
	message.Additionals = additionals

	return message, nil
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

// Read only question fields: domain, type and class from request buffer.
// Must be used after the header was readed.
func DecodeQuestion(buffer *bytes.Buffer) (*RR, error) {
	body := &RR{}

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

// Read all resource record fields from request buffer.
func DecodeRR(buffer *bytes.Buffer) (*RR, error) {
	body := &RR{}

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
	body.TimeToLive = binary.BigEndian.Uint32(buffer.Next(4))
	body.DataLen = binary.BigEndian.Uint16(buffer.Next(2))

	data := make([]byte, body.DataLen)
	_, err = binary.Decode(buffer.Bytes(), binary.BigEndian, data)
	if err != err {
		return body, err
	}
	body.Data = data

	return body, nil
}
