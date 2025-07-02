package util

import (
	"errors"
	"fmt"
	"time"
)

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

func ValidateTimeFunc(s string) error {
	_, err := ParseTime(s)
	return err
}
