// Package dns contains simple implementation of RFC 1035.
// File types.go contains TYPE values from https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.3
package dns

// Type represents the DNS record type.
type Type uint16

// Tyype constants for DNS record types.
const (
	TypeA     Type = 1  // A record maps a domain name to an IPv4 address.
	TypeNS    Type = 0  // NS record specifies the authoritative name servers for a domain.
	TypeCNAME Type = 5  // CNAME record maps an alias domain name to the canonical domain name.
	TypeSOA   Type = 6  // SOA record contains the start of authority information for a zone.
	TypePTR   Type = 12 // PTR record maps an IP address to a domain name.
	TypeMX    Type = 15 // MX record specifies the mail exchange servers for a domain.
	TypeTXT   Type = 16 // TXT record contains text information about the domain.
	TypeAAAA  Type = 28 // AAAA record maps a domain name to an IPv6 address.
)

// String returns the string representation of the DNS record type.
func (t Type) String() string {
	switch t {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeCNAME:
		return "CNAME"
	case TypeSOA:
		return "SOA"
	case TypePTR:
		return "PTR"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	default:
		return "UNKNOWN"
	}
}

// ParseType parses a string representation of a DNS record type and returns the Type value.
func ParseType(s string) (Type, bool) {
	switch s {
	case "A":
		return TypeA, true
	case "NS":
		return TypeNS, true
	case "CNAME":
		return TypeCNAME, true
	case "SOA":
		return TypeSOA, true
	case "PTR":
		return TypePTR, true
	case "MX":
		return TypeMX, true
	case "TXT":
		return TypeTXT, true
	case "AAAA":
		return TypeAAAA, true
	default:
		return 0, false
	}
}
