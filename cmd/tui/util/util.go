// Package util provides utility functions for the TUI.
package util

import (
	"errors"
	"fmt"
	"time"
)

// ParseTime parses a time string in various formats (DateOnly, DateTime, TimeOnly) and returns a time.Time object.
func ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("time is empty")
	}
	layouts := []string{time.DateOnly, time.DateTime, time.TimeOnly}
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			break
		}
	}
	if t.IsZero() {
		return time.Time{}, fmt.Errorf("incorrect date: %s", s)
	}
	return t, err
}

// ValidateTimeFunc validates the time string using ParseTime function.
func ValidateTimeFunc(s string) error {
	_, err := ParseTime(s)
	return err
}
