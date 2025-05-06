package database

import "github.com/prionis/dns-server/protocol"

type DB interface {
	GetRRs(domain string) ([]*protocol.RR, error)
}

type DBRR struct {
	ID int64
	RR protocol.RR
}
