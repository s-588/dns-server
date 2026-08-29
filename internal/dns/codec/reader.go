package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Reader struct {
	data []byte
	pos  int
}

func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
	}
}

func (r *Reader) Uint8() (uint8, error) {
	if r.pos >= len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}

	v := r.data[r.pos]
	r.pos++

	return v, nil
}

func (r *Reader) Uint16() (uint16, error) {
	if len(r.data)-r.pos < 2 {
		return 0, io.ErrUnexpectedEOF
	}

	v := binary.BigEndian.Uint16(r.data[r.pos : r.pos+2])

	r.pos += 2

	return v, nil
}

func (r *Reader) Bytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("negative length")
	}

	if len(r.data)-r.pos < n {
		return nil, io.ErrUnexpectedEOF
	}

	b := r.data[r.pos : r.pos+n]

	r.pos += n

	return b, nil
}

func (r *Reader) ReadName() (string, error) {
	var labels []string

	for {
		length, err := r.Uint8()
		if err != nil {
			return "", fmt.Errorf("read label length: %w", err)
		}

		if length == 0 {
			break
		}

		label, err := r.Bytes(int(length))
		if err != nil {
			return "", err
		}

		labels = append(labels, string(label))
	}

	return strings.Join(labels, "."), nil
}
