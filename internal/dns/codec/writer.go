package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

type Writer struct {
	buf         []byte
	compression map[string]uint16
}

func NewWriter() *Writer {
	return &Writer{
		// 22 bytes because: minimal header length is 12
		// and other 8 bytes is QTYPE(2 byte), QCLASS(2 byte), and
		// smallest QNAME is '.', but 6 is more realistic in real world
		buf: make([]byte, 0, 22),

		compression: make(map[string]uint16),
	}
}

func (w *Writer) Uint8(v uint8) {
	w.buf = append(w.buf, v)
}

func (w *Writer) Uint16(v uint16) {
	var buf [2]byte

	binary.BigEndian.PutUint16(buf[:], v)

	w.buf = append(w.buf, buf[:]...)
}

func (w *Writer) Bytes(data []byte) {
	w.buf = append(w.buf, data...)
}

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
		if len(w.buf) < 0x4000 {
			w.compression[domain] = uint16(len(w.buf))
		}

		w.Uint8(uint8(len(label)))
		w.Bytes([]byte(label))
	}

	w.Uint8(0)

	return nil
}

func (w *Writer) writePointer(offset uint16) {
	w.Uint16(0xC000 | offset)
}
