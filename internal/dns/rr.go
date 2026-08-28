// Package dns contains simple implementation of RFC 1035.
// File rr.go contains implementation of Rersource Record according to https://datatracker.ietf.org/doc/html/rfc1035#section-3.2
package dns

import "io"

type RR struct {
	Name  string
	Type  Type
	Class Class
	TTL   uint32
	Data  RData
}

type RData interface {
	Type() Type
	Marshal(w *io.Writer) []byte
}
