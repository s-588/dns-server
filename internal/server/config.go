package server

import "github.com/prionis/dns-server/internal/database"

type options struct {
	dnsPort  string
	httpPort string
	db       database.Repository
}

type Option interface {
	apply(*options)
}

// DNS port option

type dnsPort string

func (p dnsPort) apply(opts *options) {
	opts.dnsPort = string(p)
}

func SetDNSPort(p string) Option {
	return dnsPort(p)
}

// HTTP port option

type httpPort string

func (p httpPort) apply(opts *options) {
	opts.dnsPort = string(p)
}

func SetHTTPPort(p string) Option {
	return dnsPort(p)
}

// Database option

type dbOption struct {
	db database.Repository
}

func (d dbOption) apply(opts *options) {
	opts.db = d.db
}

func WithDB(db database.Repository) Option {
	return dbOption{db}
}
