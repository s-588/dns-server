// Package structs provides data structures for TUI.
package structs

import "time"

// RR represents a DNS resource record with an ID, domain, data, type, class, and TTL.
type RR struct {
	ID     int64  // Unique identifier for the resource record
	Domain string // Domain name
	Data   string // Resource data
	Type   string // Record type
	Class  string // Record class
	TTL    uint32 // Time to live in seconds
}

// User represents a user with an ID, login, first name, last name, role, and password.
type User struct {
	ID        int64  // Unique identifier for the user
	Login     string // User's login name
	FirstName string // User's first name
	LastName  string // User's last name
	Role      string // User's role
	Password  string // User's password
}

// Log represents a log entry with a timestamp, log level, and message.
type Log struct {
	Time  time.Time // Timestamp of the log entry
	Level string    // Log level (e.g., INFO, ERROR, WARN)
	Msg   string    // Message of the log entry
}
