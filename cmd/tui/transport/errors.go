package transport

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// APIError describes a non-200 response from the DNS server API.
// It is used for describing issue to user.
type APIError struct {
	Method    string
	Path      string
	Status    int
	ServerMsg string
}

// Error returns a message in format:
//
//	POST /auth/login: 401 Unauthorized: invalid password
func (e *APIError) Error() string {
	status := http.StatusText(e.Status)
	if status == "" {
		status = fmt.Sprintf("status %d", e.Status)
	}
	if e.ServerMsg == "" {
		return fmt.Sprintf("%s %s: %d %s", e.Method, e.Path, e.Status, status)
	}
	return fmt.Sprintf("%s %s: %d %s: %s",
		e.Method, e.Path, e.Status, status, e.ServerMsg)
}

// IsAuthError reports whether the server rejected the credentials.
func (e *APIError) IsAuthError() bool {
	return e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden
}

// IsValidationError reports whether the request payload was rejected.
func (e *APIError) IsValidationError() bool {
	switch e.Status {
	case http.StatusBadRequest,
		http.StatusConflict,
		http.StatusUnprocessableEntity:
		return true
	}
	return false
}

// IsServerError reports whether the server failed unexpectedly.
func (e *APIError) IsServerError() bool {
	return e.Status >= 500
}

// NetworkError describes a failure to reach the DNS server.
type NetworkError struct {
	Method string
	Addr   string // full url attempted, e.g. "http://127.0.0.1:8083/auth/login"
	Err    error
}

// Error returns the message in format:
//
//	POST http://127.0.0.1:8083/auth/login: dial tcp 127.0.0.1:8083: connect: connection refused
func (e *NetworkError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Method, e.Addr, unwrapURLError(e.Err))
}

// Unwrap allows errors.Is/As to reach the underlying cause.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

func unwrapURLError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
}

// UserMessage translates a Transport error into a message safe to show in the UI.
func UserMessage(err error) string {
	if err == nil {
		return ""
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.IsServerError() {
			return fmt.Sprintf("server error (%d), please try again later", apiErr.Status)
		}
		return apiErr.Error()
	}

	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return "cannot reach server, check that the server is running"
	}

	return err.Error()
}
