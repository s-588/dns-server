package protocol

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

type readTest struct {
	name   string
	buf    *bytes.Buffer
	domain string
	err    error
}

var readTests = []readTest{
	{
		"normal domain test",
		bytes.NewBuffer([]byte{
			6, 'd', 'o', 'm', 'a', 'i', 'n', 4, 'n', 'a', 'm', 'e', 0,
		},
		),
		"domain.name",
		nil,
	},
	{
		"normal domain with subdomain",
		bytes.NewBuffer([]byte{
			3, 'w', 'w', 'w', 6, 'd', 'o', 'm', 'a', 'i', 'n', 4, 'n', 'a', 'm', 'e', 0,
		},
		),
		"www.domain.name",
		nil,
	},
	{
		"domain name with empty label",
		bytes.NewBuffer([]byte{
			6, 'd', 'o', 'm', 'a', 'i', 'n', 1, 4, 'n', 'a', 'm', 'e', 0,
		},
		),
		"",
		errors.New("not nil"),
	},
	{
		"big domain name",
		bytes.NewBuffer([]byte{
			65, 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 'n', 'd', 'o', 'm', 'a', 'i', 4, 'n', 'a', 'm', 'e', 0,
		},
		),
		"",
		errors.New("not nil"),
	},
}

func TestReadDomainName(t *testing.T) {
	for _, test := range readTests {
		resultDomain, resultErr := ReadDomainName(test.buf)

		// If we not expecting error, but get it
		if resultErr != nil && test.err == nil {
			t.Errorf("%s test: error is unexpected, want nil, get: %v", test.name, resultErr)
		}
		// If we expect that function encode domain right
		if resultDomain != test.domain && test.err == nil {
			t.Errorf("%s test: wrong domain, want: %s, get: %s", test.name, test.domain, resultDomain)
		}
	}
}

type writeTest struct {
	name      string
	domain    string
	resultBuf *bytes.Buffer
	wantBuf   *bytes.Buffer
	err       error
}

var writeTests = []writeTest{
	{
		"normal domain name",
		"domain.name",
		bytes.NewBuffer([]byte{}),
		bytes.NewBuffer([]byte{6, 'd', 'o', 'm', 'a', 'i', 'n', 4, 'n', 'a', 'm', 'e', 0}),
		nil,
	},
	{
		"normal domain name with subdomain",
		"www.domain.name",
		bytes.NewBuffer([]byte{}),
		bytes.NewBuffer([]byte{
			3, 'w', 'w', 'w', 6, 'd', 'o', 'm', 'a', 'i', 'n', 4, 'n', 'a', 'm', 'e', 0,
		}),
		nil,
	},
}

func TestWriteDomainName(t *testing.T) {
	for _, test := range writeTests {
		resultErr := WriteDomainName(test.domain, test.resultBuf)
		if resultErr != nil && test.err == nil {
			t.Errorf("%s test: unexpected error, want nil, get %v", test.name, test.err)
		}
		if !slices.Equal(test.resultBuf.Bytes(), test.wantBuf.Bytes()) {
			t.Errorf("%s test: wrong result,\nwant: %v \nget:  %v", test.name, test.wantBuf, test.resultBuf)
		}
	}
}

type labelTest struct {
	name  string
	label string
	want  error
}

var labelTests = []labelTest{
	{
		"normal label",
		"com",
		nil,
	},
	{
		"big label",
		"comcomcomcomcomcomcomcomcomcomcomcomcomcomcomcomcomcomcomcomomcomcomcomcomcomcomcomcom",
		errors.New("too long"),
	},
	{
		"label with dot",
		"c.om",
		errors.New("contain dot"),
	},
	{
		"label with in the start",
		".com",
		errors.New("contain dot"),
	},
	{
		"start with hyphen",
		"-com",
		errors.New("start with hyphen"),
	},
}

func TestValidateLabel(t *testing.T) {
	for _, test := range labelTests {
		resultErr := ValidateLabel(test.label)
		if resultErr != nil && test.want == nil {
			t.Errorf("%s test: unexpected error, want nil, get: %v", test.name, resultErr)
		}
		if test.want != nil && resultErr == nil {
			t.Errorf("%s test: exepcted error, but don't get it", test.name)
		}
	}
}

type domainTest struct {
	name    string
	domain  string
	wantErr error
}

var domainTests = []domainTest{
	{
		"normal domain",
		"domain.normal",
		nil,
	},
	{
		"domain with subdomain",
		"www.domain.normal",
		nil,
	},
	{
		"domain with many labels",
		"domain.www.com.ru.en.gov.normal",
		nil,
	},
	{
		"domain with empty label",
		"domain..normal",
		errors.New("wrong label"),
	},
	{
		"domain with wrong label",
		"domain.-normal",
		errors.New("wrong label"),
	},
}

func TestValidateDomainName(t *testing.T) {
	for _, test := range domainTests {
		resultErr := ValidateDomainName(test.domain)
		if resultErr != nil && test.wantErr == nil {
			t.Errorf("%s test: unexpected error: %v", test.name, resultErr)
		}
		if resultErr == nil && test.wantErr != nil {
			t.Errorf("%s test: want error, but not get", test.name)
		}
	}
}
