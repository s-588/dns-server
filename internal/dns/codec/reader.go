package codec

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

func (r *Reader) Uint32() (uint32, error) {
	if len(r.data)-r.pos < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	v := binary.BigEndian.Uint32(r.data[r.pos : r.pos+4])

	r.pos += 4

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

	jumped := false
	start := r.pos

	for {
		length, err := r.Uint8()
		if err != nil {
			return "", fmt.Errorf("read label length: %w", err)
		}

		// length&0xC0 need to ensure that we check only first two bits
		if length&0xC0 == 0xC0 {
			secondHalf, err := r.Uint8()
			if err != nil {
				return "", fmt.Errorf("read pointer: %w", err)
			}
			// length&0b0011_1111 clears lower 6 bits
			// << 8 shift 6 lower bits and makes them highest in 14 bit number
			// | secondHalf fills lower 8 bits
			ptr := length&0b0011_1111<<8 | secondHalf
			if !jumped {
				start = r.pos
				jumped = true
			}
			r.pos = int(ptr)
			continue
		}

		if length == 0 {
			break
		}
		if length > 63 {
			return "", errors.New("label too long")
		}

		label, err := r.Bytes(int(length))
		if err != nil {
			return "", err
		}

		labels = append(labels, string(label))
	}

	if jumped {
		r.pos = start
	}

	return strings.Join(labels, "."), nil
}
