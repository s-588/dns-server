package server

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const (
	testSecret = "test_secret"
)

func Test_parseJWT(t *testing.T) {
	makeToken := func(claims jwt.Claims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, err := token.SignedString([]byte(testSecret))
		assert.NoError(t, err)
		return s
	}
	tests := []struct {
		name     string
		tokenStr string
		login    string
		wantErr  bool
		secret   string
	}{
		{
			name:  "normal token",
			login: "normal-login",
			tokenStr: makeToken(jwt.MapClaims{
				"login": "normal-login",
			}),
			wantErr: false,
			secret:  testSecret,
		},
		{
			name:     "empty token",
			tokenStr: "",
			wantErr:  true,
			secret:   testSecret,
		},
		{
			name:     "missing login claim",
			tokenStr: makeToken(jwt.MapClaims{}),
			wantErr:  true,
			secret:   testSecret,
		},
		{
			name: "login is not a string",
			tokenStr: makeToken(jwt.MapClaims{
				"login": 12345,
			}),
			wantErr: true,
			secret:  testSecret,
		},
		{
			name:    "empty JWT secret",
			wantErr: true,
			secret:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.secret != "" {
				t.Setenv("JWT_SECRET", tt.secret)
			}
			got, err := parseJWT(tt.tokenStr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.login, got)
		})
	}
}
