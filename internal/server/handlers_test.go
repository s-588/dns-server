package server

import (
	"net/netip"
	"testing"

	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/dns"
	"github.com/stretchr/testify/assert"
)

//nolint:funlen
func Test_parseRR(t *testing.T) {
	tests := []struct {
		name    string
		a       database.ResourceRecord
		want    dns.RR
		wantErr bool
	}{
		{
			name: "normal A record",
			a: database.ResourceRecord{
				ID:     10,
				Domain: "www.domain.com",
				Data:   "127.0.0.1",
				Type:   dns.TypeA.String(),
				Class:  dns.ClassIN.String(),
				TTL:    10,
			},
			want: dns.RR{
				Name:     "www.domain.com",
				Type:     dns.TypeA,
				Class:    dns.ClassIN,
				TTL:      10,
				RDLength: 4,
				RData: &dns.A{
					IP: netip.MustParseAddr("127.0.0.1"),
				},
			},
		},
		{
			name: "incorrect type",
			a: database.ResourceRecord{
				Domain: "example.com",
				Data:   "127.0.0.1",
				Type:   "NOTATYPE",
				Class:  dns.ClassIN.String(),
				TTL:    300,
			},
			wantErr: true,
		},
		{
			name: "incorrect class",
			a: database.ResourceRecord{
				Domain: "example.com",
				Data:   "127.0.0.1",
				Type:   dns.TypeA.String(),
				Class:  "NOTACLASS",
				TTL:    300,
			},
			wantErr: true,
		},
		{
			name: "incorrect data (bad IP)",
			a: database.ResourceRecord{
				Domain: "example.com",
				Data:   "400.1000",
				Type:   dns.TypeA.String(),
				Class:  dns.ClassIN.String(),
				TTL:    300,
			},
			wantErr: true,
		},
		{
			name: "incorrect data (IPv6 for A record)",
			a: database.ResourceRecord{
				Domain: "example.com",
				Data:   "2001:db8::1",
				Type:   dns.TypeA.String(),
				Class:  dns.ClassIN.String(),
				TTL:    300,
			},
			wantErr: true,
		},
		{
			name: "incorrect domain name",
			a: database.ResourceRecord{
				Domain: ".incorrect.#@$@(&..domain))",
				Data:   "127.0.0.1",
				Type:   dns.TypeA.String(),
				Class:  dns.ClassIN.String(),
				TTL:    60,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRR(tt.a)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if got.Name != tt.want.Name ||
				got.Type != tt.want.Type ||
				got.Class != tt.want.Class ||
				got.TTL != tt.want.TTL ||
				got.RDLength != tt.want.RDLength {
				t.Errorf("parseRR() mismatch in header fields\ngot:  %+v\nwant: %+v", got, tt.want)
			}

			assert.Equal(t, got.RData, tt.want.RData)
		})
	}
}
