package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func ValidateDomain(domain string) error {
	if len(domain) > 255 {
		return errors.New("domain name is too long")
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if err := ValidateLabel(label); err != nil {
			return fmt.Errorf("validate domain: %w", err)
		}
	}

	return nil
}

func ValidateLabel(label string) error {
	if len(label) > 63 {
		return errors.New("label is too long")
	}
	ok, err := regexp.MatchString("^[a-z1-9]+(-[a-z1-9]+)*$", label)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("label \"%s\" does not match RFC 1035 standard", label)
	}

	return nil
}

func DecodeDomain(buffer *bytes.Buffer) (string, error) {
	domainParts := make([]string, 0)

	// RFC 1035: "a domain name represented as a sequence of labels, where
	// each label consists of a length octet followed by that
	// number of octets."
	b, err := buffer.ReadByte()
	if err != nil {
		return "", err
	}

	for ; b != 0 && err == nil; b, err = buffer.ReadByte() {
		domainBytes := buffer.Next(int(b))
		domainParts = append(domainParts, string(domainBytes))
	}
	domain := strings.Join(domainParts, ".")
	return domain, ValidateDomain(domain)
}

func EncodeDomain(domain string, buffer *bytes.Buffer) error {
	if err := ValidateDomain(domain); err != nil {
		return err
	}

	err := buffer.WriteByte(byte(len(domain)))
	if err != nil {
		return err
	}

	_, err = buffer.WriteString(domain)
	if err != nil {
		return err
	}

	err = buffer.WriteByte(0)
	if err != nil {
		return err
	}

	return nil
}
