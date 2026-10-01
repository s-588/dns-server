package server

import "github.com/prionis/dns-server/internal/database"

type options struct {
	dnsPort  string
	httpPort string
	db       database.Repository
	https    bool
}

// Option represents a configuration option for the Server.
type Option interface {
	apply(*options)
}

// dnsPort option represents an option to set the DNS port for the TCP and UDP connections.
type dnsPort string

func (p dnsPort) apply(opts *options) {
	opts.dnsPort = string(p)
}

// WithDNSPort returns an Option that sets the DNS port for the server.
func WithDNSPort(p string) Option {
	return dnsPort(p)
}

// httpPort option
type httpPort string

// apply method applies the HTTP port option to the provided options.
func (p httpPort) apply(opts *options) {
	opts.httpPort = string(p)
}

// WithHTTPPort returns an Option that sets the HTTP port for the server.
func WithHTTPPort(p string) Option {
	return httpPort(p)
}

// dbOption represents an option to set the database repository for the server.
type dbOption struct {
	db database.Repository
}

// apply method applies the database option to the provided options.
func (d dbOption) apply(opts *options) {
	opts.db = d.db
}

// WithDB returns an Option that sets the database repository for the server.
func WithDB(db database.Repository) Option {
	return dbOption{db}
}

type https bool

func (p https) apply(opts *options) {
	opts.https = bool(p)
}

func WithHTTPS(enable bool) Option {
	return https(enable)
}
