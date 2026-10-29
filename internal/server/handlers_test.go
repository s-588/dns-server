package server

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/dns"
	"github.com/prionis/dns-server/internal/testutil"
	"github.com/prionis/dns-server/proto/genproto/crudpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func newTestServer(t *testing.T) (Server, *testutil.MockRepository) {
	t.Helper()
	db := testutil.NewMockRepository(t)
	return Server{
		db:      db,
		metrics: NewMetrics(),
		https:   false,
	}, db
}

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

func TestServer_dnsHandler(t *testing.T) {
	s, db := newTestServer(t)

	rr := database.ResourceRecord{
		Domain: "test.com",
		Data:   "127.0.0.1",
		TTL:    10,
		Type:   dns.TypeA.String(),
		Class:  dns.ClassIN.String(),
	}

	db.EXPECT().
		FindRecords(context.Background(), "test.com", dns.TypeA.String()).
		Return([]database.ResourceRecord{rr}, nil).
		Once()

	msg := dns.Message{
		Header: dns.Header{
			ID:      0x1234,
			QDCount: 1,
		},
		Questions: []dns.Question{
			{
				Name:  "test.com",
				Type:  dns.TypeA,
				Class: dns.ClassIN,
			},
		},
	}
	b, err := msg.MarshalBinary()
	require.NoError(t, err)

	gotBytes := s.dnsHandler(b)
	require.NotNil(t, gotBytes)

	var got dns.Message
	err = got.UnmarshalBinary(gotBytes)
	require.NoError(t, err)

	// we need to check every field because dnsHandler set QR flag,
	// so messages will never be the same.
	assert.Equal(t, msg.Header.ID, got.Header.ID)
	assert.NotZero(t, got.Header.Flags&dns.FlagQR, "QR flag must be set")
	assert.Equal(t, uint16(1), got.Header.QDCount)
	assert.Len(t, got.Questions, 1)
	assert.Equal(t, "test.com", got.Questions[0].Name)
	assert.Equal(t, dns.TypeA, got.Questions[0].Type)
	assert.Equal(t, dns.ClassIN, got.Questions[0].Class)
	assert.Equal(t, uint16(1), got.Header.ANCount)
	assert.Len(t, got.Answers, 1)
}

func TestServer_loginHandler(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	s, db := newTestServer(t)

	goodUser := database.User{
		ID:        1,
		Login:     "login",
		FirstName: "first",
		LastName:  "last",
		Role:      "user",
	}
	db.EXPECT().
		CheckUserPassword(context.Background(), "login", "password").
		Return(goodUser, nil).Once()

	loginPb := &crudpb.Login{
		Username: goodUser.Login,
		Password: "password",
	}
	body, err := proto.Marshal(loginPb)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	r.Header.Set("content-type", "application/protobuf")
	w := httptest.NewRecorder()

	s.loginHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Result().Cookies()) // JWT was set

	got := new(crudpb.User)
	err = proto.Unmarshal(w.Body.Bytes(), got)
	assert.NoError(t, err)
	assert.Equal(t, goodUser.Login, got.Login)
	assert.Equal(t, goodUser.FirstName, got.FirstName)
	assert.Equal(t, goodUser.LastName, got.LastName)
	assert.Equal(t, goodUser.Role, got.Role)
}

func TestServer_registerHandler(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	s, db := newTestServer(t)

	goodUser := database.User{
		Login:     "login",
		FirstName: "first",
		LastName:  "last",
		Role:      "user",
	}
	db.EXPECT().
		AddUser(context.Background(), goodUser, "password").
		Return(1, nil).Once()

	regPb := &crudpb.Register{
		Login:     "login",
		FirstName: "first",
		LastName:  "last",
		Role:      "user",
		Password:  "password",
	}
	body, err := proto.Marshal(regPb)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	r.Header.Set("content-type", "application/protobuf")
	w := httptest.NewRecorder()

	s.registerHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	got := new(crudpb.User)
	err = proto.Unmarshal(w.Body.Bytes(), got)
	assert.NoError(t, err)
	assert.Equal(t, goodUser.Login, got.Login)
	assert.Equal(t, goodUser.FirstName, got.FirstName)
	assert.Equal(t, goodUser.LastName, got.LastName)
	assert.Equal(t, goodUser.Role, got.Role)
}

func TestServer_getRecordHandler(t *testing.T) {
	s, db := newTestServer(t)

	goodRR := database.ResourceRecord{
		ID:     1,
		Domain: "good.test.com",
		Data:   "127.0.0.1",
		Type:   dns.TypeA.String(),
		Class:  dns.ClassIN.String(),
		TTL:    10,
	}
	tests := []struct {
		name       string
		idStr      string
		setupMock  func()
		wantStatus int
	}{
		{
			name:  "normal request",
			idStr: "1",
			setupMock: func() {
				db.EXPECT().
					GetRecord(context.Background(), int64(1)).
					Return(goodRR, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "RR does not exist",
			idStr: "2",
			setupMock: func() {
				db.EXPECT().
					GetRecord(context.Background(), int64(2)).
					Return(database.ResourceRecord{}, database.ErrNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no ID",
			idStr:      "",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not a number",
			idStr:      "id",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			r := httptest.NewRequest(http.MethodGet,
				"/api/rrs/"+tt.idStr, http.NoBody)
			if tt.idStr != "" {
				r.SetPathValue("id", tt.idStr)
			}
			w := httptest.NewRecorder()

			s.getRecordHandler(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				got := &crudpb.ResourceRecord{}
				err := proto.Unmarshal(w.Body.Bytes(), got)
				require.NoError(t, err)
				assert.Equal(t, goodRR.ID, got.Id)
				assert.Equal(t, goodRR.Domain, got.Domain)
				assert.Equal(t, goodRR.Data, got.Data)
				assert.Equal(t, goodRR.Type, got.Type)
				assert.Equal(t, goodRR.Class, got.Class)
				assert.Equal(t, goodRR.TTL, got.Ttl)
			}
		})
	}
}

func TestServer_getUserHandler(t *testing.T) {
	s, db := newTestServer(t)

	goodUser := database.User{
		ID:        1,
		Login:     "login",
		FirstName: "first",
		LastName:  "last",
		Role:      "user",
	}
	tests := []struct {
		name       string
		loginStr   string
		setupMock  func()
		wantStatus int
	}{
		{
			name:     "normal request",
			loginStr: goodUser.Login,
			setupMock: func() {
				db.EXPECT().
					GetUser(context.Background(), goodUser.Login).
					Return(goodUser, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "User does not exist",
			loginStr: "not-exist",
			setupMock: func() {
				db.EXPECT().
					GetUser(context.Background(), "not-exist").
					Return(database.User{}, database.ErrNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no ID",
			loginStr:   "",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			r := httptest.NewRequest(http.MethodGet,
				"/api/users/"+tt.loginStr, http.NoBody)
			if tt.loginStr != "" {
				r.SetPathValue("login", tt.loginStr)
			}
			w := httptest.NewRecorder()

			s.getUserHandler(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				got := &crudpb.User{}
				err := proto.Unmarshal(w.Body.Bytes(), got)
				require.NoError(t, err)
				assert.Equal(t, goodUser.ID, got.Id)
				assert.Equal(t, goodUser.Login, got.Login)
				assert.Equal(t, goodUser.FirstName, got.FirstName)
				assert.Equal(t, goodUser.LastName, got.LastName)
				assert.Equal(t, goodUser.Role, got.Role)
			}
		})
	}
}

func TestServer_deleteUserHandler(t *testing.T) {
	s, db := newTestServer(t)

	tests := []struct {
		name       string
		idStr      string
		setupMock  func()
		wantStatus int
	}{
		{
			name:  "normal request",
			idStr: "1",
			setupMock: func() {
				db.EXPECT().
					DeleteUser(context.Background(), int64(1)).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "User does not exist",
			idStr: "-1",
			setupMock: func() {
				db.EXPECT().
					DeleteUser(context.Background(), int64(-1)).
					Return(database.ErrNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no ID",
			idStr:      "",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not a number",
			idStr:      "id",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			r := httptest.NewRequest(http.MethodDelete,
				"/api/users/"+tt.idStr, http.NoBody)
			if tt.idStr != "" {
				r.SetPathValue("id", tt.idStr)
			}
			w := httptest.NewRecorder()

			s.deleteUserHandler(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				got := html.UnescapeString(w.Body.String())
				assert.Equal(t, got, fmt.Sprintf("User with id %s successful deleted", tt.idStr))
			}
		})
	}
}

func TestServer_deleteRRHandler(t *testing.T) {
	s, db := newTestServer(t)

	tests := []struct {
		name       string
		idStr      string
		setupMock  func()
		wantStatus int
	}{
		{
			name:  "normal request",
			idStr: "1",
			setupMock: func() {
				db.EXPECT().
					DeleteRecord(context.Background(), int64(1)).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "User does not exist",
			idStr: "-1",
			setupMock: func() {
				db.EXPECT().
					DeleteRecord(context.Background(), int64(-1)).
					Return(database.ErrNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no ID",
			idStr:      "",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not a number",
			idStr:      "id",
			setupMock:  func() {}, // handler returns before touching DB
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			r := httptest.NewRequest(http.MethodDelete,
				"/api/rrs/"+tt.idStr, http.NoBody)
			if tt.idStr != "" {
				r.SetPathValue("id", tt.idStr)
			}
			w := httptest.NewRecorder()

			s.deleteRRHandler(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				got := html.UnescapeString(w.Body.String())
				assert.Equal(t, got, fmt.Sprintf("Resource record with id %s successful deleted", tt.idStr))
			}
		})
	}
}

func TestServer_postRRHandler(t *testing.T) {
	s, db := newTestServer(t)

	goodRR := database.ResourceRecord{
		Domain: "domain.com",
		Data:   "127.0.0.1",
		Type:   dns.TypeA.String(),
		Class:  dns.ClassIN.String(),
		TTL:    10,
	}
	db.EXPECT().
		AddRecord(context.Background(), goodRR).
		Return(1, nil).Once()

	regPb := &crudpb.ResourceRecord{
		Domain: "domain.com",
		Data:   "127.0.0.1",
		Type:   dns.TypeA.String(),
		Class:  dns.ClassIN.String(),
		Ttl:    10,
	}
	body, err := proto.Marshal(regPb)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/api/rrs", bytes.NewReader(body))
	r.Header.Set("content-type", "application/protobuf")
	w := httptest.NewRecorder()

	s.postRRHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	got := new(crudpb.ResourceRecord)
	err = proto.Unmarshal(w.Body.Bytes(), got)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), got.Id)
	assert.Equal(t, goodRR.Domain, got.Domain)
	assert.Equal(t, goodRR.Data, got.Data)
	assert.Equal(t, goodRR.Type, got.Type)
	assert.Equal(t, goodRR.Class, got.Class)
	assert.Equal(t, goodRR.TTL, got.Ttl)
}
