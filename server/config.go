package server

import "github.com/prionis/dns-server/database"

type options struct {
	port   string
	logger Logger
	db     database.DB
}

type Option interface {
	apply(*options)
}

// Port option

type port string

func (p port) apply(opts *options) {
	opts.port = ":" + string(p)
}

func SetPort(p string) Option {
	return port(p)
}

// Logger option

type loggerOption struct {
	logger Logger
}

func (l loggerOption) apply(opts *options) {
	opts.logger = l.logger
}

func WithLogger(l Logger) Option {
	return loggerOption{l}
}

// Database option

type dbOption struct {
	db database.DB
}

func (d dbOption) apply(opts *options) {
	opts.db = d.db
}

func WithDB(db database.DB) Option {
	return dbOption{db}
}
