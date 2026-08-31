// Package dns contains simple implementation of RFC 1035.
// File types.go contains TYPE values from https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.3
package dns

type Type uint16

const (
	TypeA     Type = 1
	TypeNS    Type = 2
	TypeCNAME Type = 5
	TypeSOA   Type = 6
	TypePTR   Type = 12
	TypeMX    Type = 15
	TypeTXT   Type = 16
	TypeAAAA  Type = 28
)

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

func (t *Type) IsValid() bool {
	switch *t {
	case TypeA, TypeNS, TypeCNAME, TypeSOA, TypePTR, TypeMX, TypeTXT, TypeAAAA:
		return true
	default:
		return false
	}
}
