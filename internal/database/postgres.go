package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/prionis/dns-server/internal/database/sqlc"
)

// Postgres struct represent connection to the PostgreSQL database.
type Postgres struct {
	db *sqlc.Queries
}

// NewPostgres create new connection to the PostgreSQL database.
func NewPostgres(connString string) (Postgres, error) {
	if connString == "" {
		connString = GetConnectionString()
	}
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return Postgres{}, fmt.Errorf("can't connect to  %w", err)
	}
	if err = conn.Ping(context.Background()); err != nil {
		return Postgres{}, fmt.Errorf("can't ping the database: %w", err)
	}

	db := sqlc.New(conn)

	return Postgres{db}, nil
}

// GetConnectionString return the formated connection string for connecting to the PostgreSQL.
func GetConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_ADDR"),
		os.Getenv("POSTGRES_DB"))
}

// GetRecord return the resource record with provided id.
func (repo Postgres) GetRecord(ctx context.Context, id int32) (ResourceRecord, error) {
	rr, err := repo.db.GetResourceRecordByID(ctx, id)
	if err != nil {
		return ResourceRecord{}, err
	}
	return ResourceRecord{
		ID:     rr.ID,
		Domain: rr.Domain,
		Data:   rr.Data,
		Type:   rr.Type,
		Class:  rr.Class,
		TTL:    rr.TimeToLive.Int32,
	}, nil
}

// AddRecord insert record in the database and return this record with ID settled ID.
func (repo Postgres) AddRecord(ctx context.Context, rr ResourceRecord) (int32, error) {
	id, err := repo.db.CreateResourceRecord(ctx, sqlc.CreateResourceRecordParams{
		Domain:     rr.Domain,
		Type:       rr.Type,
		Class:      rr.Class,
		Data:       rr.Data,
		TimeToLive: pgtype.Int4{int32(rr.TTL), true},
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetAllRecords return all the resource records from the database
func (repo Postgres) GetAllRecords(ctx context.Context) ([]ResourceRecord, error) {
	rrs, err := repo.db.GetAllResourceRecord(ctx)
	if err != nil {
		return nil, err
	}

	resourceRecords := make([]ResourceRecord, 0, len(rrs))
	for _, record := range rrs {
		resourceRecords = append(resourceRecords, ResourceRecord{
			ID:     record.ID,
			Domain: record.Domain,
			Type:   record.Type,
			Class:  record.Class,
			TTL:    record.TimeToLive.Int32,
			Data:   record.Data,
		})
	}
	return resourceRecords, nil
}

// UpdateRecord update record with provided ID and values.
func (repo Postgres) UpdateRecord(ctx context.Context, rr ResourceRecord) error {
	_, err := repo.db.UpdateResourceRecord(context.Background(), sqlc.UpdateResourceRecordParams{
		ID:         rr.ID,
		Domain:     rr.Domain,
		Data:       rr.Data,
		Type:       rr.Type,
		Class:      rr.Class,
		TimeToLive: pgtype.Int4{rr.TTL, true},
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecord delete record with provided ID.
func (repo Postgres) DeleteRecord(ctx context.Context, id int32) error {
	err := repo.db.DeleteResourceRecord(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// FindRecords return resource records with provided domain name and type.
func (repo Postgres) FindRecords(ctx context.Context, name, rrType string) ([]ResourceRecord, error) {
	rrs, err := repo.db.GetResourceRecords(ctx, sqlc.GetResourceRecordsParams{name, rrType})
	if err != nil {
		return nil, err
	}

	resourceRecords := make([]ResourceRecord, 0, len(rrs))
	for _, record := range rrs {
		resourceRecords = append(resourceRecords, ResourceRecord{
			Domain: record.Domain,
			Type:   record.Type,
			Class:  record.Class,
			TTL:    record.TimeToLive.Int32,
			Data:   record.Data,
		})
	}
	return resourceRecords, nil
}

// GetUser return user with provided login.
func (repo Postgres) GetUser(ctx context.Context, login string) (User, error) {
	user, err := repo.db.GetUser(ctx, login)
	if err != nil {
		return User{}, fmt.Errorf("can't get user from database: %w", err)
	}

	return User{
		ID:        user.ID,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

// GetAllUsers return all users from database.
func (repo Postgres) GetAllUsers(ctx context.Context) ([]User, error) {
	usersRows, err := repo.db.GetAllUsers(context.Background())
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(usersRows))
	for _, user := range usersRows {
		users = append(users, User{
			ID:        user.ID,
			Login:     user.Login,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		})
	}
	return users, nil
}

// UpdateUser update user with provided ID and values.
func (repo Postgres) UpdateUser(ctx context.Context, user User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("can't hash password: %w", err)
	}

	return repo.db.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        int32(user.ID),
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Password:  string(hash),
	})
}

// AddUser add user in the database and return this user with settled ID.
func (repo Postgres) AddUser(ctx context.Context, user User, password string) (int32, error) {
	if len(user.FirstName) < 2 {
		return 0, fmt.Errorf("can't use name %s, the length less than 2")
	}

	if len(user.LastName) < 2 {
		return 0, fmt.Errorf("can't use last name %s, the length less than 2")
	}

	if len(user.Role) < 4 {
		return 0, fmt.Errorf("length of role can't be less than 4")
	}

	if len(password) < 4 {
		return 0, fmt.Errorf("password is too weak, it must contain at least " +
			"1 special symbol and 1 number and 8 symbols in total")
	}
	if len(password) > 71 {
		return 0, fmt.Errorf("password is too long, 70 symbols max")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, fmt.Errorf("can't hash password: %w", err)
	}

	id, err := repo.db.CreateUser(ctx, sqlc.CreateUserParams{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Password:  string(hash),
	})
	if err != nil {
		return 0, fmt.Errorf("can't register new user: %w", err)
	}
	return id, nil
}

// CheckUserPassword check if the password is correct for provided user.
func (repo Postgres) CheckUserPassword(ctx context.Context, login, pass string) (User, error) {
	user, err := repo.db.GetUser(ctx, login)
	if err != nil {
		return User{}, fmt.Errorf("can't get user from database: %w", err)
	}

	return User{
		ID:        user.ID,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
}

// DeleteUser delete user with provided id.
func (repo Postgres) DeleteUser(ctx context.Context, id int32) error {
	return repo.db.DeleteUser(ctx, id)
}
