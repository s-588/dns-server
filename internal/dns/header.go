// Package dns contains simple implementation of RFC 1035.
// File header.go contains Header of Resource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-4.1.1
package dns

type Header struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

const (
	FlagQR          uint16 = 1 << 15
	FlagOpcodeMask  uint16 = 0b0000_1111_0000_0000
	FlagOpcodeshift        = 11
	FlagAA          uint16 = 1 << 10
	FlagTC          uint16 = 1 << 9
	FlagRD          uint16 = 1 << 8
	FlagRA          uint16 = 1 << 7
	FlagRcodeMask   uint16 = 0b0000_0000_0000_1111
)

type OPCODE uint8

func (o OPCODE) String() string {
	switch o {
	case OpcodeQUERY:
		return "QUERY"
	case OpcodeIQUERY:
		return "IQUERY"
	case OpcodeSTATUS:
		return "STATUS"
	default:
		return "UNKNOWN"
	}
}

func ParseOpcode(s string) (OPCODE, bool) {
	switch s {
	case "QUERY":
		return OpcodeQUERY, true
	case "IQUERY":
		return OpcodeIQUERY, true
	case "STATUS":
		return OpcodeSTATUS, true
	default:
		return 0, false
	}
}

const (
	OpcodeQUERY  OPCODE = 0
	OpcodeIQUERY OPCODE = 1
	OpcodeSTATUS OPCODE = 2
)

func (h Header) Opcode() OPCODE {
	return OPCODE(uint8(h.Flags&FlagOpcodeMask) >> FlagOpcodeshift)
}

func (h *Header) SetOpcode(opcode OPCODE) {
	h.Flags = (h.Flags &^ FlagOpcodeMask) | (uint16(opcode)&0x0F)<<FlagOpcodeshift
}

type RCode uint8

const (
	RCodeNoError        RCode = 0
	RCodeFormatError    RCode = 1
	RCodeServerFailure  RCode = 2
	RCodeNameError      RCode = 3
	RCodeNotImplemented RCode = 4
	RCodeRefused        RCode = 5
)

func (r RCode) String() string {
	switch r {
	case RCodeNoError:
		return "NOERROR"
	case RCodeFormatError:
		return "FORMERR"
	case RCodeServerFailure:
		return "SERVFAIL"
	case RCodeNameError:
		return "NXDOMAIN"
	case RCodeNotImplemented:
		return "NOTIMP"
	case RCodeRefused:
		return "REFUSED"
	default:
		return "UNKNOWN"
	}
}

func ParseRCode(s string) (RCode, bool) {
	switch s {
	case "NOERROR":
		return RCodeNoError, true
	case "FORMERR":
		return RCodeFormatError, true
	case "SERVFAIL":
		return RCodeServerFailure, true
	case "NXDOMAIN":
		return RCodeNameError, true
	case "NOTIMP":
		return RCodeNotImplemented, true
	case "REFUSED":
		return RCodeRefused, true
	default:
		return 0, false
	}
}

func (r RCode) Description() string {
	switch r {
	case RCodeNoError:
		return "No error"
	case RCodeFormatError:
		return "Format error"
	case RCodeServerFailure:
		return "Server failure"
	case RCodeNameError:
		return "Non-existent domain"
	case RCodeNotImplemented:
		return "Not implemented"
	case RCodeRefused:
		return "Query refused"
	default:
		return "Unknown response code"
	}
}

func (h Header) RCode() RCode {
	return RCode(h.Flags & FlagRcodeMask)
}

func (h *Header) SetRCode(rcode RCode) {
	h.Flags = (h.Flags &^ FlagRcodeMask) | (uint16(rcode) & 0x0F)
}
