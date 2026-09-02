// Package dns contains simple implementation of RFC 1035.
// File header.go contains Header of Resource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-4.1.1
package dns

import "testing"

func TestHeaderFlags(t *testing.T) {
	t.Run("SetFlag", func(t *testing.T) {
		h := &Header{}
		h.SetFlag(FlagQR)
		if h.Flags&FlagQR == 0 {
			t.Error("SetFlag(FlagQR) did not set bit")
		}
	})

	t.Run("Opcode", func(t *testing.T) {
		h := Header{Flags: uint16(OpcodeIQUERY) << FlagOpcodeshift}
		if h.Opcode() != OpcodeIQUERY {
			t.Errorf("Opcode() = %v, want %v", h.Opcode(), OpcodeIQUERY)
		}
	})

	t.Run("SetOpcode", func(t *testing.T) {
		h := &Header{}
		h.SetOpcode(OpcodeSTATUS)
		if h.Opcode() != OpcodeSTATUS {
			t.Errorf("SetOpcode(STATUS) resulted in Opcode() = %v, want STATUS", h.Opcode())
		}
		// ensure other flags are preserved
		h.SetFlag(FlagAA)
		h.SetOpcode(OpcodeQUERY)
		if h.Flags&FlagAA == 0 {
			t.Error("SetOpcode cleared unrelated flag AA")
		}
		if h.Opcode() != OpcodeQUERY {
			t.Errorf("Opcode after SetOpcode(QUERY) = %v, want QUERY", h.Opcode())
		}
	})

	t.Run("RCode", func(t *testing.T) {
		h := Header{Flags: uint16(RCodeNameError)}
		if h.RCode() != RCodeNameError {
			t.Errorf("RCode() = %v, want %v", h.RCode(), RCodeNameError)
		}
	})

	t.Run("SetRCode", func(t *testing.T) {
		h := &Header{Flags: 0}
		h.SetRCode(RCodeRefused)
		if h.RCode() != RCodeRefused {
			t.Errorf("SetRCode(REFUSED) resulted in RCode() = %v, want REFUSED", h.RCode())
		}
	})
}

func TestOpcodeString(t *testing.T) {
	tests := []struct {
		op   OPCODE
		want string
	}{
		{OpcodeQUERY, "QUERY"},
		{OpcodeIQUERY, "IQUERY"},
		{OpcodeSTATUS, "STATUS"},
		{OPCODE(7), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("OPCODE(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestParseOpcode(t *testing.T) {
	tests := []struct {
		s    string
		want OPCODE
		ok   bool
	}{
		{"QUERY", OpcodeQUERY, true},
		{"IQUERY", OpcodeIQUERY, true},
		{"STATUS", OpcodeSTATUS, true},
		{"query", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseOpcode(tt.s)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ParseOpcode(%q) = (%v, %v), want (%v, %v)", tt.s, got, ok, tt.want, tt.ok)
		}
	}
}

func TestRCodeString(t *testing.T) {
	tests := []struct {
		rc   RCode
		want string
	}{
		{RCodeNoError, "NOERROR"},
		{RCodeFormatError, "FORMERR"},
		{RCodeServerFailure, "SERVFAIL"},
		{RCodeNameError, "NXDOMAIN"},
		{RCodeNotImplemented, "NOTIMP"},
		{RCodeRefused, "REFUSED"},
		{RCode(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.rc.String(); got != tt.want {
			t.Errorf("RCode(%d).String() = %q, want %q", tt.rc, got, tt.want)
		}
	}
}

func TestParseRCode(t *testing.T) {
	tests := []struct {
		s    string
		want RCode
		ok   bool
	}{
		{"NOERROR", RCodeNoError, true},
		{"FORMERR", RCodeFormatError, true},
		{"SERVFAIL", RCodeServerFailure, true},
		{"NXDOMAIN", RCodeNameError, true},
		{"NOTIMP", RCodeNotImplemented, true},
		{"REFUSED", RCodeRefused, true},
		{"noerror", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseRCode(tt.s)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ParseRCode(%q) = (%v, %v), want (%v, %v)", tt.s, got, ok, tt.want, tt.ok)
		}
	}
}

func TestRCodeDescription(t *testing.T) {
	tests := []struct {
		rc   RCode
		want string
	}{
		{RCodeNoError, "No error"},
		{RCodeFormatError, "Format error"},
		{RCodeServerFailure, "Server failure"},
		{RCodeNameError, "Non-existent domain"},
		{RCodeNotImplemented, "Not implemented"},
		{RCodeRefused, "Query refused"},
		{RCode(99), "Unknown response code"},
	}
	for _, tt := range tests {
		if got := tt.rc.Description(); got != tt.want {
			t.Errorf("RCode(%d).Description() = %q, want %q", tt.rc, got, tt.want)
		}
	}
}
