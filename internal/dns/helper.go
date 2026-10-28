package dns

import (
	"strings"
	"unicode"
)

func IsValidDomain(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	name = strings.TrimSuffix(name, ".")
	if name == "" { // root zone
		return true
	}

	labels := strings.SplitSeq(name, ".")
	for label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(unicode.Is(unicode.Latin, c) ||
				unicode.IsDigit(c) ||
				c == '-') {
				return false
			}
		}
	}
	return true
}
