package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"golang.org/x/net/publicsuffix"
	"google.golang.org/protobuf/proto"
)

type Transport struct {
	addr       string
	httpClient http.Client
}

func New(addr string) (*Transport, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("can't create cookie jar for http client: %w", err)
	}
	transport := http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
	return &Transport{
		addr:       "http://" + addr,
		httpClient: transport,
	}, nil
}

func (t Transport) GetAllRRs() ([]database.ResourceRecord, error) {
	// resp, err := t.httpClient.Get(t.addr + "/api/")

	return nil, nil
}

func (t Transport) UpdateRR(rr database.ResourceRecord) error {
	return nil
}

func (t Transport) GetRR(name, rrType string) database.ResourceRecord {
	return database.ResourceRecord{}
}

func (t Transport) AddRR(name, rrType, class, data string, ttl int64) error {
	return nil
}

func (t Transport) DeleteRR(id int64) error {
	return nil
}

func (t Transport) RegisterNewUser(database.User) error {
	return nil
}

func (t Transport) Login(login, password string) (*crudpb.User, error) {
	credentials := &crudpb.Login{
		Username: login,
		Password: password,
	}

	b, err := proto.Marshal(credentials)
	if err != nil {
		return nil, fmt.Errorf("can't marshal message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, t.addr+"/auth/login", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	r, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't make request to the server: %w", err)
	}
	defer r.Body.Close()

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response body: %w", err)
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", r.Status, string(msg))
	}

	user := &crudpb.User{}
	err = proto.Unmarshal(msg, user)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal response: %w", err)
	}
	return user, nil
}

func (t Transport) GetUser() database.User {
	return database.User{}
}

func (t Transport) DeleteUser(id string) error {
	return nil
}
