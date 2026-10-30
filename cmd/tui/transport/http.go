package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prionis/dns-server/cmd/tui/structs"
	"github.com/prionis/dns-server/proto/genproto/crudpb"
	"google.golang.org/protobuf/proto"
)

// readResponse reads and closes resp.Body, returning the raw body on a 200
// response or an *APIError otherwise.
func readResponse(resp *http.Response, method, path string) ([]byte, error) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close response body: %v\n", err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return body, &APIError{
			Method:    method,
			Path:      path,
			Status:    resp.StatusCode,
			ServerMsg: strings.TrimSpace(string(body)),
		}
	}
	return body, nil
}

func (t Transport) GetAllRRs() ([]structs.RR, error) {
	const path = "/api/rrs/all"
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create %s %s request: %w", http.MethodGet, path, err)
	}
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, &NetworkError{Method: http.MethodGet, Addr: t.httpAddr + path, Err: err}
	}
	msg, err := readResponse(resp, http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	rrs := &crudpb.ResourceRecordCollection{}
	if err := proto.Unmarshal(msg, rrs); err != nil {
		return nil, fmt.Errorf("unmarshal %s %s response: %w", http.MethodGet, path, err)
	}
	records := make([]structs.RR, 0, len(rrs.Records))
	for _, record := range rrs.Records {
		records = append(records, structs.RR{
			ID:     record.Id,
			Domain: record.Domain,
			Data:   record.Data,
			Type:   record.Type,
			Class:  record.Class,
			TTL:    record.Ttl,
		})
	}
	return records, nil
}

func (t Transport) UpdateUser(u structs.User) error {
	const path = "/api/users/"
	body, err := proto.Marshal(&crudpb.User{
		Id:        u.ID,
		Login:     u.Login,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		Password:  u.Password,
	})
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}
	req, err := http.NewRequest(http.MethodPatch, t.httpAddr+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s %s request: %w", http.MethodPatch, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return &NetworkError{Method: http.MethodPatch, Addr: t.httpAddr + path, Err: err}
	}
	_, err = readResponse(resp, http.MethodPatch, path)
	return err
}

func (t Transport) UpdateRR(rr structs.RR) error {
	path := "/api/rrs/" + strconv.FormatInt(rr.ID, 10)
	body, err := proto.Marshal(&crudpb.ResourceRecord{
		Id:     rr.ID,
		Domain: rr.Domain,
		Data:   rr.Data,
		Type:   rr.Type,
		Class:  rr.Class,
		Ttl:    rr.TTL,
	})
	if err != nil {
		return fmt.Errorf("marshal resource record: %w", err)
	}
	req, err := http.NewRequest(http.MethodPatch, t.httpAddr+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s %s request: %w", http.MethodPatch, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return &NetworkError{Method: http.MethodPatch, Addr: t.httpAddr + path, Err: err}
	}
	_, err = readResponse(resp, http.MethodPatch, path)
	return err
}

func (t Transport) AddRR(rr structs.RR) (structs.RR, error) {
	const path = "/api/rrs/"
	body, err := proto.Marshal(&crudpb.ResourceRecord{
		Domain: rr.Domain,
		Data:   rr.Data,
		Type:   rr.Type,
		Class:  rr.Class,
		Ttl:    rr.TTL,
	})
	if err != nil {
		return structs.RR{}, fmt.Errorf("marshal resource record: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, t.httpAddr+path, bytes.NewReader(body))
	if err != nil {
		return structs.RR{}, fmt.Errorf("create %s %s request: %w", http.MethodPost, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return structs.RR{}, &NetworkError{Method: http.MethodPost, Addr: t.httpAddr + path, Err: err}
	}
	msg, err := readResponse(resp, http.MethodPost, path)
	if err != nil {
		return structs.RR{}, err
	}

	record := &crudpb.ResourceRecord{}
	if err := proto.Unmarshal(msg, record); err != nil {
		return structs.RR{}, fmt.Errorf("unmarshal %s %s response: %w", http.MethodPost, path, err)
	}
	return structs.RR{
		ID:     record.Id,
		Domain: record.Domain,
		Data:   record.Data,
		Type:   record.Type,
		Class:  record.Class,
		TTL:    record.Ttl,
	}, nil
}

func (t Transport) DeleteRR(id int64) error {
	path := "/api/rrs/" + strconv.FormatInt(id, 10)
	req, err := http.NewRequest(http.MethodDelete, t.httpAddr+path, http.NoBody)
	if err != nil {
		return fmt.Errorf("create %s %s request: %w", http.MethodDelete, path, err)
	}
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return &NetworkError{Method: http.MethodDelete, Addr: t.httpAddr + path, Err: err}
	}
	_, err = readResponse(resp, http.MethodDelete, path)
	return err
}

func (t Transport) RegisterNewUser(user structs.User) (structs.User, error) {
	const path = "/auth/register"
	credentials := &crudpb.Register{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Password:  user.Password,
	}
	b, err := proto.Marshal(credentials)
	if err != nil {
		return structs.User{}, fmt.Errorf("marshal register request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, t.httpAddr+path, bytes.NewReader(b))
	if err != nil {
		return structs.User{}, fmt.Errorf("create %s %s request: %w", http.MethodPost, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return structs.User{}, &NetworkError{Method: http.MethodPost, Addr: t.httpAddr + path, Err: err}
	}
	msg, err := readResponse(resp, http.MethodPost, path)
	if err != nil {
		return structs.User{}, err
	}

	u := &crudpb.User{}
	if err := proto.Unmarshal(msg, u); err != nil {
		return structs.User{}, fmt.Errorf("unmarshal %s %s response: %w", http.MethodPost, path, err)
	}
	return structs.User{
		ID:        u.Id,
		Login:     u.Login,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
	}, nil
}

func (t Transport) Login(login, password string) (*crudpb.User, error) {
	const path = "/auth/login"
	credentials := &crudpb.Login{
		Username: login,
		Password: password,
	}
	b, err := proto.Marshal(credentials)
	if err != nil {
		return nil, fmt.Errorf("marshal login request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, t.httpAddr+path, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("create %s %s request: %w", http.MethodPost, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, &NetworkError{Method: http.MethodPost, Addr: t.httpAddr + path, Err: err}
	}
	msg, err := readResponse(resp, http.MethodPost, path)
	if err != nil {
		return nil, err
	}

	user := &crudpb.User{}
	if err := proto.Unmarshal(msg, user); err != nil {
		return nil, fmt.Errorf("unmarshal %s %s response: %w", http.MethodPost, path, err)
	}
	return user, nil
}

func (t Transport) GetAllUsers() ([]structs.User, error) {
	const path = "/api/users/all"
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create %s %s request: %w", http.MethodGet, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, &NetworkError{Method: http.MethodGet, Addr: t.httpAddr + path, Err: err}
	}
	msg, err := readResponse(resp, http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	users := &crudpb.UserCollection{}
	if err := proto.Unmarshal(msg, users); err != nil {
		return nil, fmt.Errorf("unmarshal %s %s response: %w", http.MethodGet, path, err)
	}
	result := make([]structs.User, 0, len(users.Users))
	for _, user := range users.Users {
		result = append(result, structs.User{
			ID:        user.Id,
			Login:     user.Login,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		})
	}
	return result, nil
}

func (t Transport) DeleteUser(id int64) error {
	path := "/api/users/" + strconv.FormatInt(id, 10)
	req, err := http.NewRequest(http.MethodDelete, t.httpAddr+path, http.NoBody)
	if err != nil {
		return fmt.Errorf("create %s %s request: %w", http.MethodDelete, path, err)
	}
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return &NetworkError{Method: http.MethodDelete, Addr: t.httpAddr + path, Err: err}
	}
	_, err = readResponse(resp, http.MethodDelete, path)
	return err
}

func (t Transport) GetAllLogs() ([]structs.Log, error) {
	const path = "/api/logs/all"
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create %s %s request: %w", http.MethodGet, path, err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, &NetworkError{Method: http.MethodGet, Addr: t.httpAddr + path, Err: err}
	}
	msg, err := readResponse(resp, http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	logs := &crudpb.LogCollection{}
	if err := proto.Unmarshal(msg, logs); err != nil {
		return nil, fmt.Errorf("unmarshal %s %s response: %w", http.MethodGet, path, err)
	}
	result := make([]structs.Log, 0, len(logs.Logs))
	for _, log := range logs.Logs {
		result = append(result, structs.Log{
			Time:  log.Time.AsTime(),
			Level: log.Level,
			Msg:   log.Msg,
		})
	}
	return result, nil
}
