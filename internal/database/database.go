package database

import "context"

// Repository interface represent database.
type Repository interface {
	// AddRecord add resource record to the database and return this resource record with inserted ID.
	AddRecord(ctx context.Context, rr ResourceRecord) (int32, error)
	// GetAllRecords return all resource records that database contain.
	GetAllRecords(ctx context.Context) ([]ResourceRecord, error)
	// GetRecord return one resource record with provided ID.
	GetRecord(ctx context.Context, id int32) (ResourceRecord, error)
	// FindRecords find the resource record based on the provided domain name and type.
	FindRecords(ctx context.Context, name, rrType string) ([]ResourceRecord, error)
	// UpdateRecord update the resource record.
	// Provided resource record contain ID of the record that need to be updated
	// and other fields contains the new data.
	UpdateRecord(ctx context.Context, rr ResourceRecord) error
	// DeleteRecord delete resource record with provided ID.
	DeleteRecord(ctx context.Context, id int32) error
	// GetAllUsers return all users from database.
	GetAllUsers(ctx context.Context) ([]User, error)
	// GetUser return user with provided login.
	GetUser(ctx context.Context, login string) (User, error)
	// CheckUserPassword check if the password is correct for user with provided login.
	CheckUserPassword(ctx context.Context, login, pass string) (User, error)
	// AddUser add user to the database. Return this user with seted ID.
	AddUser(ctx context.Context, user User, password string) (int32, error)
	// DeleteUser delete user with provided ID.
	DeleteUser(ctx context.Context, id int32) error
	// UpdateUser update user with provided ID and values from the struct.
	UpdateUser(ctx context.Context, user User, password string) error
}

// ResourceRecord structure represent resource record in the dabase.
type ResourceRecord struct {
	// ID of the resource record in the database.
	ID int32
	// Domain that this resource record contain.
	Domain string
	// Data that resource record contain.
	Data string
	// Type of the resource record.
	Type string
	// Class of the resource record.
	Class string
	// Time To Live of the resource record.
	TTL int32
}

// User represent any people in database.
type User struct {
	// ID of user in the database.
	ID int32
	// Login of the user.
	Login string
	// First name of the user.
	FirstName string
	// Last name of the user.
	LastName string
	// Role of the user(admin, user, etc.).
	Role string
}
