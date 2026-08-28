// Package dns contains simple implementation of RFC 1035.
// File message.go contains of Message according to https://datatracker.ietf.org/doc/html/rfc1035#section-4
package dns

type Message struct {
	Header     Header
	Question   Question
	Answer     RR
	Authority  RR
	Additional RR
}

type Question struct {
	Name  string
	Type  Type
	Class Class
}
