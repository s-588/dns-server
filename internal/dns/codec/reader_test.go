package dns

import (
	"reflect"
	"testing"
)

func TestNewReader(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want *Reader
	}{
		{
			name: "nil slice",
			args: args{
				data: nil,
			},
			want: &Reader{data: nil},
		},
		{
			name: "empty slice",
			args: args{
				data: []byte{},
			},
			want: &Reader{data: []byte{}},
		},
		{
			name: "slice with values",
			args: args{
				data: []byte{1, 2, 3},
			},
			want: &Reader{data: []byte{1, 2, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewReader(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_Uint8(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		pos     int
		want    uint8
		wantErr bool
	}{
		{"ok", []byte{1}, 0, 1, false},
		{"ok; pos in the middle", []byte{1, 2, 3, 4, 5}, 2, 3, false},
		{"eof", []byte{}, 0, 0, true},
		{"pos over the len", []byte{1}, 1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reader{
				data: tt.data,
				pos:  tt.pos,
			}
			got, err := r.Uint8()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Reader.Uint8() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Reader.Uint8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_Uint16(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		pos     int
		want    uint16
		wantErr bool
	}{
		{"ok", []byte{1, 2}, 0, 0x0102, false},
		{"ok; pos in the middle", []byte{1, 2, 3, 4, 5}, 2, 0x0304, false},
		{"eof at start", []byte{}, 0, 0, true},
		{"pos past end", []byte{1}, 1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reader{
				data: tt.data,
				pos:  tt.pos,
			}
			got, err := r.Uint16()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Reader.Uint16() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Reader.Uint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_Bytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		pos     int
		n       int
		want    []byte
		wantErr bool
	}{
		{"ok", []byte{1, 2, 3}, 0, 3, []byte{1, 2, 3}, false},
		{"pos in the middle", []byte{1, 2, 3, 4, 5}, 2, 3, []byte{3, 4, 5}, false},
		{"eof at start", []byte{}, 0, 3, nil, true},
		{"pos over the len", []byte{1, 2}, 2, 1, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reader{
				data: tt.data,
				pos:  tt.pos,
			}
			got, err := r.Bytes(tt.n)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Reader.Bytes() error = %v, wantErr %v; got = %v, want = %v", err, tt.wantErr, got, tt.want)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reader.Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadName(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{"root", []byte{0}, "", false},
		{"simple", []byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}, "www.example.com", false},
		{"truncated; missing length", []byte{3, 'w', 'w'}, "", true},
		{"end of domain name at the start", []byte{0, 3, 'a', 'b', 'c', 0}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reader{
				data: tt.data,
			}
			got, err := r.ReadName()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Reader.ReadName() error = %v, wantErr %v; got = %v, want %v", err, tt.wantErr, got, tt.want)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Reader.ReadName() = %v, want %v", got, tt.want)
			}
		})
	}
}
