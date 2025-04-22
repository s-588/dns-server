package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func ValidateDomainName(domain string) error {
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

func ReadDomainName(buffer *bytes.Buffer) (string, error) {
	var domainName string

	b, err := buffer.ReadByte()

	for ; b != 0 && err == nil; b, err = buffer.ReadByte() {
		labelLength := int(b)
		labelBytes := buffer.Next(labelLength)
		labelName := string(labelBytes)

		if len(domainName) == 0 {
			domainName = labelName
		} else {
			domainName += "." + labelName
		}
	}

	return domainName, err
}

func WriteDomainName(domainName string, responseBuffer *bytes.Buffer) error {
	labels := strings.Split(domainName, ".")

	for _, label := range labels {
		labelLength := len(label)
		labelBytes := []byte(label)

		responseBuffer.WriteByte(byte(labelLength))
		responseBuffer.Write(labelBytes)
	}

	err := responseBuffer.WriteByte(byte(0))

	return err
}
