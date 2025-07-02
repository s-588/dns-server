package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/prionis/dns-server/cmd/tui/structs"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"google.golang.org/protobuf/proto"
)

func (t Transport) GetAllRRs() ([]structs.RR, error) {
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+"/api/rrs/all", http.NoBody)
	if err != nil {
		return []structs.RR{}, fmt.Errorf("can't create new request: %w", err)
	}

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return []structs.RR{}, fmt.Errorf("can't make request to the server: %w", err)
	}
	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return []structs.RR{}, fmt.Errorf("can't read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return []structs.RR{}, fmt.Errorf("%s: %s", resp.Status, string(msg))
	}

	rrs := &crudpb.ResourceRecordCollection{}
	err = proto.Unmarshal(msg, rrs)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal response: %w", err)
	}
	records := make([]structs.RR, 0, len(rrs.Records))
	for _, record := range rrs.Records {
		records = append(records, structs.RR{
			ID:     record.Id,
			Domain: record.Domain,
			Data:   record.Data,
			Type:   record.Type,
			Class:  record.Class,
			TTL:    record.TimeToLive,
		})
	}
	return records, nil
}

func (t Transport) UpdateUser(u structs.User) error {
	body, err := proto.Marshal(&crudpb.User{
		Id:        u.ID,
		Login:     u.Login,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		Password:  u.Password,
	})

	req, err := http.NewRequest(http.MethodPatch, t.httpAddr+"/api/users/", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("can't make request to the server: %w", err)
	}
	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, string(msg))
	}

	return nil
}

func (t Transport) UpdateRR(rr structs.RR) error {
	body, err := proto.Marshal(&crudpb.ResourceRecord{
		Id:         rr.ID,
		Domain:     rr.Domain,
		Data:       rr.Data,
		Type:       rr.Type,
		Class:      rr.Class,
		TimeToLive: rr.TTL,
	})

	req, err := http.NewRequest(http.MethodPatch, t.httpAddr+"/api/rrs/"+strconv.FormatInt(int64(rr.ID), 10), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("can't make request to the server: %w", err)
	}
	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, string(msg))
	}

	return nil
}

func (t Transport) AddRR(rr structs.RR) (structs.RR, error) {
	body, err := proto.Marshal(&crudpb.ResourceRecord{
		Domain:     rr.Domain,
		Data:       rr.Data,
		Type:       rr.Type,
		Class:      rr.Class,
		TimeToLive: rr.TTL,
	})

	req, err := http.NewRequest(http.MethodPost, t.httpAddr+"/api/rrs/", bytes.NewReader(body))
	if err != nil {
		return structs.RR{}, fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return structs.RR{}, fmt.Errorf("can't make request to the server: %w", err)
	}
	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return structs.RR{}, fmt.Errorf("can't read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return structs.RR{}, fmt.Errorf("%s: %s", resp.Status, string(msg))
	}

	record := &crudpb.ResourceRecord{}
	err = proto.Unmarshal(msg, record)
	if err != nil {
		return structs.RR{}, fmt.Errorf("unmarshal server response: %w", err)
	}

	return structs.RR{
		ID:     record.Id,
		Domain: record.Domain,
		Data:   record.Data,
		Type:   record.Type,
		Class:  record.Class,
		TTL:    record.TimeToLive,
	}, nil
}

func (t Transport) DeleteRR(id int32) error {
	req, err := http.NewRequest(http.MethodDelete, t.httpAddr+"/api/rrs/"+strconv.FormatInt(int64(id), 10), http.NoBody)
	if err != nil {
		return fmt.Errorf("can't create new request: %w", err)
	}

	r, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("can't make request to the server: %w", err)
	}
	defer r.Body.Close()

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("can't read response body: %w", err)
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", r.Status, string(msg))
	}

	return nil
}

func (t Transport) RegisterNewUser(user structs.User) (structs.User, error) {
	credentials := &crudpb.Register{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Password:  user.Password,
	}

	b, err := proto.Marshal(credentials)
	if err != nil {
		return structs.User{}, fmt.Errorf("can't marshal message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, t.httpAddr+"/auth/register", bytes.NewReader(b))
	if err != nil {
		return structs.User{}, fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	r, err := t.HTTPClient.Do(req)
	if err != nil {
		return structs.User{}, fmt.Errorf("can't make request to the server: %w", err)
	}
	defer r.Body.Close()

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return structs.User{}, fmt.Errorf("can't read response body: %w", err)
	}

	if r.StatusCode != http.StatusOK {
		return structs.User{}, fmt.Errorf("%s: %s", r.Status, string(msg))
	}

	u := &crudpb.User{}
	err = proto.Unmarshal(msg, u)
	if err != nil {
		return structs.User{}, fmt.Errorf("can't unmarshal response: %w", err)
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
	credentials := &crudpb.Login{
		Username: login,
		Password: password,
	}

	b, err := proto.Marshal(credentials)
	if err != nil {
		return nil, fmt.Errorf("can't marshal message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, t.httpAddr+"/auth/login", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	r, err := t.HTTPClient.Do(req)
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

func (t Transport) GetAllUsers() ([]structs.User, error) {
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+"/api/users/all", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	r, err := t.HTTPClient.Do(req)
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

	users := &crudpb.UserCollection{}
	err = proto.Unmarshal(msg, users)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal response: %w", err)
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

func (t Transport) DeleteUser(id int32) error {
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+"/api/users/"+strconv.FormatInt(int64(id), 10), http.NoBody)
	if err != nil {
		return fmt.Errorf("can't create new request: %w", err)
	}

	r, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("can't make request to the server: %w", err)
	}
	defer r.Body.Close()

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("can't read response body: %w", err)
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", r.Status, string(msg))
	}

	return nil
}

func (t Transport) GetAllLogs() ([]structs.Log, error) {
	req, err := http.NewRequest(http.MethodGet, t.httpAddr+"/api/logs/all", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("can't create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("Accept", "application/protobuf")

	r, err := t.HTTPClient.Do(req)
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

	logs := &crudpb.LogCollection{}
	err = proto.Unmarshal(msg, logs)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal response: %w", err)
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
