// Package codec package contains utilities for encoding and decoding DNS messages, including support for name compression.
package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// Writer is a utility for writing DNS messages with support for name compression.
type Writer struct {
	buf []byte // underlying buffer

	// map for DNS name compression
	// mapping domain names to their offsets in the buffer
	compression map[string]uint16
}

// NewWriter creates a new Writer.
// Initial buffer size is 22 bytes.
func NewWriter() *Writer {
	return &Writer{
		// 22 bytes because: minimal header length is 12
		// and other 8 bytes is QTYPE(2 byte), QCLASS(2 byte), and
		// smallest QNAME is '.', but 6 is more realistic in real world
		buf: make([]byte, 0, 22),

		compression: make(map[string]uint16),
	}
}

// Uint8 appends a uint8 value to the Writer's buffer.
func (w *Writer) Uint8(v uint8) {
	w.buf = append(w.buf, v)
}

// Uint16 appends a uint16 value to the Writer's buffer in big-endian order.
func (w *Writer) Uint16(v uint16) {
	var buf [2]byte

	binary.BigEndian.PutUint16(buf[:], v)

	w.buf = append(w.buf, buf[:]...)
}

// Uint32 appends a uint32 value to the Writer's buffer in big-endian order.
func (w *Writer) Uint32(v uint32) {
	var buf [4]byte

	binary.BigEndian.PutUint32(buf[:], v)

	w.buf = append(w.buf, buf[:]...)
}

// Bytes appends the given byte slice to the Writer's buffer.
func (w *Writer) Bytes(data []byte) {
	w.buf = append(w.buf, data...)
}

// WriteName writes a domain name to the buffer, using DNS name compression if possible.
func (w *Writer) WriteName(name string) error {
	name = strings.TrimSuffix(name, ".")

	if name == "" {
		w.Uint8(0)
		return nil
	}

	labels := strings.Split(name, ".")

	for i, label := range labels {
		if len(label) > 63 {
			return fmt.Errorf("DNS label too long: %q", label)
		}

		if label == "" {
			return errors.New("empty DNS label")
		}

		domain := strings.Join(labels[i:], ".")
		if off, ok := w.compression[domain]; ok {
			w.writePointer(off)
			return nil
		}

		//
		if offset := len(w.buf); offset < 0x4000 {
			w.compression[domain] = uint16(offset)
		}

		l := len(label)
		if l < 0 || l > math.MaxUint8 {
			return fmt.Errorf("length of label too big: %d", l)
		}
		w.Uint8(uint8(l))
		w.Bytes([]byte(label))
	}

	w.Uint8(0)

	return nil
}

// writePointer writes a pointer to the given offset in the message.
func (w *Writer) writePointer(offset uint16) {
	w.Uint16(0xC000 | offset)
}

// Buffer returns the underlying byte slice of the Writer.
func (w *Writer) Buffer() []byte {
	return w.buf
}
