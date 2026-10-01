// Package dns contains simple implementation of RFC 1035.
// File classes.go contains CLASS values according to RFC 1035 https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.4
package dns

// Class represents the CLASS field in a DNS Resource Record.
type Class uint16

// DNS record classes.
const (
	ClassIN Class = 1 // Internet
	ClassCS Class = 2 // CSNET (obsolete)
	ClassCH Class = 3 // CHAOS
	ClassHS Class = 4 // Hesiod
)

// String returns the string representation of the Class.
func (c Class) String() string {
	switch c {
	case ClassIN:
		return "IN"
	case ClassCS:
		return "CS"
	case ClassCH:
		return "CH"
	case ClassHS:
		return "HS"
	default:
		return "UNKNOWN"
	}
}

// ParseClass parses string representation of CLASS to Class type.
func ParseClass(s string) (Class, bool) {
	switch s {
	case "IN":
		return ClassIN, true
	case "CS":
		return ClassCS, true
	case "CH":
		return ClassCH, true
	case "HS":
		return ClassHS, true
	default:
		return 0, false
	}
}
