package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"

	"github.com/prionis/dns-server/internal/database/sqlc"
)

// Errors returned by the repository.
var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrForeignKey       = errors.New("foreign key violation")
	ErrNotNullViolation = errors.New("not null violation")
	ErrCheckViolation   = errors.New("check violation")
	ErrInvalidInput     = errors.New("invalid input")
)

// Constants for validating inputs
const (
	minimalNameLength = 2
	minimalLength     = 4
	maxLength         = 14
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Postgres struct represent connection to the PostgreSQL database.
type Postgres struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

// NewPostgres create new connection to the PostgreSQL database.
func NewPostgres(connString string) (Postgres, error) {
	if connString == "" {
		connString = GetConnectionString()
	}
	p := Postgres{}
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return Postgres{}, fmt.Errorf("can't connect to  %w", err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		return Postgres{}, fmt.Errorf("can't ping the database: %w", err)
	}
	p.pool = pool

	p.db = sqlc.New(pool)
	err = p.initDB()
	if err != nil {
		return Postgres{}, fmt.Errorf("can't init database: %w", err)
	}

	return p, nil
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
func (repo Postgres) GetRecord(ctx context.Context, id int64) (ResourceRecord, error) {
	rr, err := repo.db.GetResourceRecordByID(ctx, id)
	if err != nil {
		return ResourceRecord{}, MapPostgresError(err)
	}
	ttl := rr.TimeToLive.Int64
	if ttl < 0 || ttl > math.MaxUint32 {
		return ResourceRecord{}, fmt.Errorf("TTL too big")
	}
	return ResourceRecord{
		ID:     rr.ID,
		Domain: rr.Domain,
		Data:   rr.Data,
		Type:   rr.Type,
		Class:  rr.Class,
		TTL:    uint32(ttl),
	}, nil
}

// AddRecord insert record in the database and return this record with ID settled ID.
func (repo Postgres) AddRecord(ctx context.Context, rr ResourceRecord) (int64, error) {
	id, err := repo.db.CreateResourceRecord(ctx, sqlc.CreateResourceRecordParams{
		Domain:     rr.Domain,
		Type:       rr.Type,
		Class:      rr.Class,
		Data:       rr.Data,
		TimeToLive: pgtype.Int8{Int64: int64(rr.TTL), Valid: true},
	})
	if err != nil {
		return 0, MapPostgresError(err)
	}
	return id, nil
}

// GetAllRecords return all the resource records from the database
func (repo Postgres) GetAllRecords(ctx context.Context) ([]ResourceRecord, error) {
	rrs, err := repo.db.GetAllResourceRecord(ctx)
	if err != nil {
		return nil, MapPostgresError(err)
	}

	resourceRecords := make([]ResourceRecord, 0, len(rrs))
	for _, record := range rrs {
		ttl := record.TimeToLive.Int64
		if ttl < 0 || ttl > math.MaxUint32 {
			return nil, fmt.Errorf("TTL too big")
		}
		resourceRecords = append(resourceRecords, ResourceRecord{
			ID:     record.ID,
			Domain: record.Domain,
			Type:   record.Type,
			Class:  record.Class,
			TTL:    uint32(ttl),
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
		TimeToLive: pgtype.Int8{Int64: int64(rr.TTL), Valid: true},
	})
	if err != nil {
		return MapPostgresError(err)
	}

	return nil
}

// DeleteRecord delete record with provided ID.
func (repo Postgres) DeleteRecord(ctx context.Context, id int64) error {
	err := repo.db.DeleteResourceRecord(ctx, id)
	if err != nil {
		return MapPostgresError(err)
	}

	return nil
}

// FindRecords return resource records with provided domain name and type.
func (repo Postgres) FindRecords(ctx context.Context, name, rrType string) ([]ResourceRecord, error) {
	rrs, err := repo.db.GetResourceRecords(ctx, sqlc.GetResourceRecordsParams{Domain: name, Type: rrType})
	if err != nil {
		return nil, MapPostgresError(err)
	}

	resourceRecords := make([]ResourceRecord, 0, len(rrs))
	for _, record := range rrs {
		ttl := record.TimeToLive.Int64
		if ttl < 0 || ttl > math.MaxUint32 {
			return nil, fmt.Errorf("TTL too big")
		}
		resourceRecords = append(resourceRecords, ResourceRecord{
			Domain: record.Domain,
			Type:   record.Type,
			Class:  record.Class,
			TTL:    uint32(ttl),
			Data:   record.Data,
		})
	}
	return resourceRecords, nil
}

// GetUser return user with provided login.
func (repo Postgres) GetUser(ctx context.Context, login string) (User, error) {
	user, err := repo.db.GetUser(ctx, login)
	if err != nil {
		return User{}, MapPostgresError(err)
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
		return nil, MapPostgresError(err)
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
		ID:        user.ID,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Password:  string(hash),
	})
}

// AddUser add user in the database and return this user with settled ID.
func (repo Postgres) AddUser(ctx context.Context, user User, password string) (int64, error) {
	err := parseUserCredentials(user, password)
	if err != nil {
		return 0, MapPostgresError(err)
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
		return 0, MapPostgresError(err)
	}
	return id, nil
}

func parseUserCredentials(user User, password string) error {
	if len(user.Login) < minimalLength {
		return fmt.Errorf("too short user name")
	}
	if len(user.Login) > maxLength {
		return fmt.Errorf("too long user name")
	}

	if len(user.FirstName) < minimalNameLength {
		return fmt.Errorf("too short first name")
	}

	if len(user.FirstName) > maxLength {
		return fmt.Errorf("too long first name")
	}

	if len(user.LastName) < minimalNameLength {
		return fmt.Errorf("too short last name")
	}
	if len(user.LastName) > maxLength {
		return fmt.Errorf("too long last name")
	}

	if len(user.Role) < minimalLength {
		return fmt.Errorf("role is too short")
	}
	if len(user.Role) > maxLength {
		return fmt.Errorf("role is too long")
	}

	if len(password) < minimalLength {
		return fmt.Errorf("password is too short, must be %d symbols minimum", minimalLength)
	}
	if len(password) > 70 {
		return fmt.Errorf("password is too long")
	}
	return nil
}

// CheckUserPassword check if the password is correct for provided user.
func (repo Postgres) CheckUserPassword(ctx context.Context, login, pass string) (User, error) {
	user, err := repo.db.GetUser(ctx, login)
	if err != nil {
		return User{}, MapPostgresError(err)
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
func (repo Postgres) DeleteUser(ctx context.Context, id int64) error {
	return MapPostgresError(repo.db.DeleteUser(ctx, id))
}

// initDB methods connects to Postgres and execute migrations
func (p Postgres) initDB() error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("set dialect for goose: %w", err)
	}

	if err := goose.Up(stdlib.OpenDBFromPool(p.pool), "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

// MapPostgresError converts a pgx/pgconn error into a domain error.
// Unknown errors are wrapped so the caller can still inspect the original.
func MapPostgresError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		// not a Postgres error (network, context cancelled, etc.)
		return fmt.Errorf("database error: %w", err)
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		return fmt.Errorf("%w: %s", ErrAlreadyExists, pgErr.ConstraintName)

	case "23503": // foreign_key_violation
		return fmt.Errorf("%w: %s", ErrForeignKey, pgErr.ConstraintName)

	case "23502": // not_null_violation
		return fmt.Errorf("%w: %s", ErrNotNullViolation, pgErr.ColumnName)

	case "23514": // check_violation
		return fmt.Errorf("%w: %s", ErrCheckViolation, pgErr.ConstraintName)

	default:
		return fmt.Errorf("database error (%s): %w", pgErr.Code, err)
	}
}
