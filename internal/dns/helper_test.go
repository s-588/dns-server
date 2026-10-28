package dns

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isValidDomain(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		ok     bool
	}{
		{
			name:   "normal domain",
			domain: "www.example.com",
			ok:     true,
		},
		{
			name:   "root zone",
			domain: ".",
			ok:     true,
		},
		{
			name:   "too long",
			domain: strings.Repeat("a", 270) + ".com",
			ok:     false,
		},
		{
			name:   "double hyphen at the start",
			domain: "--a.com",
			ok:     false,
		},
		{
			name:   "double dots",
			domain: "www..example..com",
			ok:     false,
		},
		{
			name:   "prefix and postfix dots",
			domain: ".www.example.com.",
			ok:     false,
		},
		{
			name:   "incorrect symbols",
			domain: "><).&!@#$|%.^*(",
			ok:     false,
		},
		{
			name:   "non latin symbols",
			domain: "ввв.экзампл.ком",
			ok:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := IsValidDomain(tt.domain)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
