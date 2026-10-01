// Package dns contains simple implementation of RFC 1035.
// File header.go contains Header of Resource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-4.1.1
package dns

// Header represents the DNS message header as defined in RFC 1035 section 4.1.1.
type Header struct {
	ID      uint16 // Identifier for the DNS message
	Flags   uint16 // Flags containing QR, Opcode, AA, TC, RD, RA, and RCODE
	QDCount uint16 // Question Count
	ANCount uint16 // Answer Count
	NSCount uint16 // Authority Count
	ARCount uint16 // Additional Count
}

// Flag constants for the DNS header flags field.
const (
	FlagQR          uint16 = 1 << 15               // Query/Response Flag
	FlagOpcodeMask  uint16 = 0b0111_1000_0000_0000 // Mask for the OPCODE bits (bits 14-11)
	FlagOpcodeshift        = 11                    // Shift for the OPCODE bits
	FlagAA          uint16 = 1 << 10               // Authoritative Answer Flag
	FlagTC          uint16 = 1 << 9                // Truncation Flag
	FlagRD          uint16 = 1 << 8                // Recursion Desired Flag
	FlagRA          uint16 = 1 << 7                // Recursion Available Flag
	FlagRcodeMask   uint16 = 0x000F                // Mask for the RCODE bits (bits 3-0)
)

// SetFlag sets the specified flag in the header's flags field.
func (h *Header) SetFlag(flag uint16) {
	h.Flags |= flag
}

// OPCODE represents the operation code in DNS header.
type OPCODE uint8

// String returns the string representation of the OPCODE.
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

// ParseOpcode parses string representation of OPCODE to OPCODE type.
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
	// OpcodeQUERY is the standard query operation, used for most DNS queries.
	OpcodeQUERY OPCODE = 0
	// OpcodeIQUERY is used for inverse queries.
	OpcodeIQUERY OPCODE = 1
	// OpcodeSTATUS is reserved for future use. It is not used in standard DNS queries or responses.
	OpcodeSTATUS OPCODE = 2
)

// Opcode returns the OPCODE from the header's flags.
func (h Header) Opcode() OPCODE {
	return OPCODE(h.Flags & FlagOpcodeMask >> FlagOpcodeshift)
}

// SetOpcode sets the OPCODE in the header's flags.
func (h *Header) SetOpcode(opcode OPCODE) {
	h.Flags = (h.Flags &^ FlagOpcodeMask) | (uint16(opcode)&0x0F)<<FlagOpcodeshift
}

// RCode represents the response code in DNS header.
type RCode uint8

const (
	// RCodeNoError indicates that the query completed successfully.
	RCodeNoError RCode = 0

	// RCodeFormatError indicates that the name server was unable to interpret the query.
	RCodeFormatError RCode = 1

	// RCodeServerFailure indicates that the name server was unable to process the query
	// due to a problem with the name server.
	RCodeServerFailure RCode = 2

	// RCodeNameError indicates that the domain name referenced in the query does not exist.
	RCodeNameError RCode = 3

	// RCodeNotImplemented indicates that the name server does not support the requested kind of query.
	RCodeNotImplemented RCode = 4

	// RCodeRefused indicates that the name server refuses to perform the specified operation for policy reasons.
	RCodeRefused RCode = 5
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

// ParseRCode parses string representation of RCode to RCode type.
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

// Description returns a human-readable description of the RCode.
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

// RCode returns the RCODE from the header's flags.
func (h Header) RCode() RCode {
	return RCode(h.Flags & FlagRcodeMask)
}

// SetRCode sets the RCODE in the header's flags.
func (h *Header) SetRCode(rcode RCode) {
	h.Flags = (h.Flags &^ FlagRcodeMask) | (uint16(rcode) & 0x0F)
}
