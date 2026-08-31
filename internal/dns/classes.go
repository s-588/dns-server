// Package dns contains simple implementation of RFC 1035.
// File classes.go contains CLASS values according to RFC 1035 https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.4
package dns

type Class uint16

const (
	ClassIN Class = 1
	ClassCS Class = 2
	ClassCH Class = 3
	ClassHS Class = 4
)

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

func (c Class) ParseClass(s string) (Class, bool) {
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

func (c Class) IsValid() bool {
	switch c {
	case ClassIN, ClassCS, ClassCH, ClassHS:
		return true
	default:
		return false
	}
}
