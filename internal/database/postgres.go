package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/prionis/dns-server/internal/database/sqlc"
)

type Postgres struct {
	db *sqlc.Queries
}

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

func GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@postgres:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
}

func (repo Postgres) GetRecord(id int32) (ResourceRecord, error) {
	return ResourceRecord{}, nil
}

func (repo Postgres) AddRecord(ctx context.Context, rr ResourceRecord) error {
	_, err := repo.db.CreateResourceRecord(ctx, sqlc.CreateResourceRecordParams{
		Domain:     rr.Domain,
		Type:       rr.Type,
		Class:      rr.Class,
		Data:       rr.Data,
		TimeToLive: pgtype.Int4{int32(rr.TTL), true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo Postgres) GetAllRecords() ([]ResourceRecord, error) {
	rrs, err := repo.db.GetAllResourceRecord(context.Background())
	if err != nil {
		return nil, err
	}

	resourceRecords := make([]ResourceRecord, 0, len(rrs))
	for _, record := range rrs {
		resourceRecords = append(resourceRecords, ResourceRecord{
			ID:      record.ID,
			Domain:  record.Domain,
			Type:    record.Type,
			Class:   record.Class,
			ClassID: int(record.ClassID),
			TypeID:  int(record.TypeID),
			TTL:     record.TimeToLive.Int32,
			Data:    record.Data,
		})
	}
	if len(resourceRecords) <= 0 {
		return resourceRecords, errors.New("not found")
	}
	return resourceRecords, nil
}

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

func (repo Postgres) DeleteRecord(ctx context.Context, id int32) error {
	err := repo.db.DeleteResourceRecord(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (repo Postgres) FindRecords(name, rrType string) ([]ResourceRecord, error) {
	rrs, err := repo.db.GetResourceRecords(context.Background(), sqlc.GetResourceRecordsParams{name, rrType})
	if err != nil {
		return nil, err
	}

	resourceRecords := make([]ResourceRecord, 0, len(rrs))
	for _, record := range rrs {
		resourceRecords = append(resourceRecords, ResourceRecord{
			Domain:  record.Domain,
			Type:    record.Type,
			Class:   record.Class,
			TypeID:  int(record.TypeID),
			ClassID: int(record.ClassID),
			TTL:     record.TimeToLive.Int32,
			Data:    record.Data,
		})
	}
	return resourceRecords, nil
}

func (repo Postgres) GetUser(ctx context.Context, login string) (User, error) {
	user, err := repo.db.GetUser(ctx, login)
	if err != nil {
		return User{}, fmt.Errorf("can't get user from database: %w", err)
	}

	return User{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

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

func (repo Postgres) UpdateUser(ctx context.Context, user User) error {
	return repo.db.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        int32(user.ID),
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	})
}

func (repo Postgres) AddUser(ctx context.Context, login, firstName, lastName, password, role string) (User, error) {
	if len(firstName) < 2 {
		return User{}, fmt.Errorf("can't use name %s, the length less than 2")
	}

	if len(lastName) < 2 {
		return User{}, fmt.Errorf("can't use last name %s, the length less than 2")
	}

	if len(role) < 4 {
		return User{}, fmt.Errorf("length of role can't be less than 4")
	}

	if !strings.ContainsAny(password, "+[{(&=)}]*!$|1234567890") && len(password) < 8 {
		return User{}, fmt.Errorf("password is too weak, it must contain at least " +
			"1 special symbol and 1 number and 8 symbols in total")
	}
	if len(password) > 71 {
		return User{}, fmt.Errorf("password is too long, 70 symbols max")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return User{}, fmt.Errorf("can't hash password: %w", err)
	}

	user, err := repo.db.CreateUser(ctx, sqlc.CreateUserParams{
		Login:     login,
		FirstName: firstName,
		LastName:  lastName,
		Password:  string(hash),
		Role:      role,
	})
	if err != nil {
		return User{}, fmt.Errorf("can't register new user: %w", err)
	}
	u, err := repo.db.GetUser(ctx, user.Login)
	fmt.Println(u.Password)
	return User{
		Login:     u.Login,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
	}, nil
}

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

func (repo Postgres) DeleteUser(ctx context.Context, id int32) error {
	return repo.db.DeleteUser(ctx, id)
}
