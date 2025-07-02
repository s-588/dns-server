package server

import (
	"context"

	"github.com/prionis/dns-server/internal/database"
)

type MockDB struct{}

func (db MockDB) UpdateRR(ctx context.Context, rr database.ResourceRecord) error {
	return nil
}

func (db MockDB) AddRR(ctx context.Context, rr database.ResourceRecord) error {
	return nil
}

func (db MockDB) DelRR(ctx context.Context, id int64) error {
	return nil
}

func (db MockDB) GetAllRRs() ([]database.ResourceRecord, error) {
	return nil, nil
}

func (db MockDB) GetRRs(name, rrType string) ([]database.ResourceRecord, error) {
	return nil, nil
}

func (db MockDB) GetAllUsers(ctx context.Context) ([]database.User, error) {
	return nil, nil
}

func (db MockDB) GetUser(ctx context.Context, login string) (database.User, error) {
	return database.User{}, nil
}

func (db MockDB) RegisterNewUser(ctx context.Context, login, firstName, lastName, password, role string) (database.User, error) {
	return database.User{}, nil
}

func (db MockDB) CheckUserPassword(ctx context.Context, login, pass string) (database.User, error) {
	return database.User{}, nil
}

func (db MockDB) DeleteUser(ctx context.Context, id int32) error {
	return nil
}

func (db MockDB) UpdateUser(ctx context.Context, user database.User) error {
	return nil
}
