package codec

import (
	"bytes"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w := NewWriter()
	if w == nil {
		t.Fatal("writer is nil")
	}
	if w.buf == nil || w.compression == nil {
		t.Fatalf("buf = %v; compression = %v", w.buf, w.compression)
	}
}

func TestWriter_Uint8_Uint16_Uint32_Bytes(t *testing.T) {
	w := NewWriter()
	w.Uint8(1)
	w.Uint16(0x0101)
	w.Uint32(0x01010101)
	w.Bytes([]byte{4, 5})
	want := []byte{1,
		1, 1,
		1, 1, 1, 1,
		4, 5}
	if !slices.Equal(w.buf, want) {
		t.Fatalf("got %v, want %v", w.buf, want)
	}
}

func TestWriter_WriteName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"root", ".", []byte{0}, false},
		{"empty", "", []byte{0}, false},
		{"simple", "www.example.com",
			[]byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}, false},
		{"label too long", strings.Repeat("a", 64), nil, true},
		{"empty label", "a..b", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Writer{
				buf:         make([]byte, 0),
				compression: make(map[string]uint16),
			}
			if err := w.WriteName(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestWriter_WriteNameWithPointer(t *testing.T) {
	w := NewWriter()

	if err := w.WriteName("www.example.com"); err != nil {
		t.Fatalf("WriteName: %v", err)
	}

	// example.com - compressed suffix
	if err := w.WriteName("mail.example.com"); err != nil {
		t.Fatalf("WriteName: %v", err)
	}

	got := w.Buffer()

	//  0: 3 www
	//  4: 7 example
	// 12: 3 com
	// 16: 0
	// 17: 4 mail
	// 22: C0 04
	want := []byte{
		3, 'w', 'w', 'w',
		7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		3, 'c', 'o', 'm',
		0,
		4, 'm', 'a', 'i', 'l',
		0xC0, 0x04,
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("compressed wire format\ngot  %v\nwant %v", got, want)
	}
}

func TestWriter_writePointer(t *testing.T) {
	w := NewWriter()
	if err := w.WriteName("example.com"); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteName("www.example.com"); err != nil {
		t.Fatal(err)
	}

	if off, ok := w.compression["example.com"]; !ok || off != 0 {
		t.Fatalf("compression[\"example.com\"] = %v, %v; want 0, true", off, ok)
	}

	want := []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		3, 'c', 'o', 'm',
		0,
		3, 'w', 'w', 'w',
		0xC0, 0}
	if !slices.Equal(w.buf, want) {
		t.Fatalf("got: %s(%v), want %s(%v)", w.buf, w.buf, want, want)
	}
}

func TestWriter_Buffer(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		want []byte
	}{
		{"empty buffer", []byte{}, []byte{}},
		{"non empty buffer", []byte{1, 2, 3}, []byte{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Writer{
				buf:         tt.buf,
				compression: make(map[string]uint16),
			}
			if got := w.Buffer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Writer.Buffer() = %v, want %v", got, tt.want)
			}
		})
	}
}
