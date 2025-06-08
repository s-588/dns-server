package database

import "context"

type Repository interface {
	UpdateRR(ctx context.Context, rr ResourceRecord) error
	AddRR(ctx context.Context, rr ResourceRecord) error
	DelRR(ctx context.Context, id int64) error
	GetAllRRs() ([]ResourceRecord, error)
	GetRRs(name, rrType string) ([]ResourceRecord, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, login string) (User, error)
	CheckUserPassword(ctx context.Context, login, pass string) (User, error)
	RegisterNewUser(ctx context.Context, login, firstName, lastName, password, role string) (User, error)
	DeleteUser(ctx context.Context, id int32) error
	UpdateUser(ctx context.Context, user User) error
}

// ResourceRecord structure represent resource record in the dabase.
type ResourceRecord struct {
	ID      int64
	Domain  string
	Data    string
	Type    string
	Class   string
	ClassID int
	TypeID  int
	TTL     int32
}

type User struct {
	ID        int64
	Login     string
	FirstName string
	LastName  string
	Role      string
}
