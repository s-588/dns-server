package database

import "context"

type Repository interface {
	AddRecord(ctx context.Context, rr ResourceRecord) error
	GetAllRecords() ([]ResourceRecord, error)
	GetRecord(id int32) (ResourceRecord, error)
	FindRecords(name, rrType string) ([]ResourceRecord, error)
	UpdateRecord(ctx context.Context, rr ResourceRecord) error
	DeleteRecord(ctx context.Context, id int32) error
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, login string) (User, error)
	CheckUserPassword(ctx context.Context, login, pass string) (User, error)
	AddUser(ctx context.Context, login, firstName, lastName, password, role string) (User, error)
	DeleteUser(ctx context.Context, id int32) error
	UpdateUser(ctx context.Context, user User) error
}

// ResourceRecord structure represent resource record in the dabase.
type ResourceRecord struct {
	ID      int32
	Domain  string
	Data    string
	Type    string
	Class   string
	ClassID int
	TypeID  int
	TTL     int32
}

type User struct {
	ID        int32
	Login     string
	FirstName string
	LastName  string
	Role      string
}
