package database_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/dns"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPG database.Postgres

// newPostgres uses testcontainer to connect to the Postgres instance for testing
func TestMain(m *testing.M) {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		panic(err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	testPG, err = database.NewPostgres(connStr)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	err = pgContainer.Terminate(ctx)
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

//nolint:funlen
func TestAddUser(t *testing.T) {
	tests := []struct {
		name      string
		user      database.User
		password  string
		expectErr bool
	}{
		{
			name: "normal user",
			user: database.User{
				Login:     "normaluser",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: false,
		},
		{
			name: "empty login",
			user: database.User{
				Login:     "",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too short login",
			user: database.User{
				Login:     "a", // length < 2
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too long login",
			user: database.User{
				Login:     strings.Repeat("A", 15), // length > 14
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too short first name",
			user: database.User{
				Login:     "shortfname",
				FirstName: "A", // length < 2
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too long first name",
			user: database.User{
				Login:     "longfname",
				FirstName: strings.Repeat("A", 51), // length > 50
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too short last name",
			user: database.User{
				Login:     "shortlname",
				FirstName: "John",
				LastName:  "D", // length < 2
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too long last name",
			user: database.User{
				Login:     "longlname",
				FirstName: "John",
				LastName:  strings.Repeat("D", 51), // length > 50
				Role:      "user",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "user role",
			user: database.User{
				Login:     "userrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "password",
			expectErr: false,
		},
		{
			name: "admin role",
			user: database.User{
				Login:     "adminrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "admin",
			},
			password:  "password",
			expectErr: false,
		},
		{
			name: "empty role",
			user: database.User{
				Login:     "emptyrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "incorrect role",
			user: database.User{
				Login:     "incorrectrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "incorrect-role",
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too long role",
			user: database.User{
				Login:     "emptyrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      strings.Repeat("u", 15), // length > 14
			},
			password:  "password",
			expectErr: true,
		},
		{
			name: "too short password",
			user: database.User{
				Login:     "emptyrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  "p", // length < 4
			expectErr: true,
		},
		{
			name: "too long password",
			user: database.User{
				Login:     "emptyrole",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
			password:  strings.Repeat("p", 71), // length > 70
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, err := testPG.AddUser(t.Context(), test.user, test.password)
			if test.expectErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			test.user.ID = id

			u, err := testPG.GetUser(t.Context(), test.user.Login)
			assert.NoError(t, err)

			assert.Equal(t, test.user, u)
		})
	}
}

// TestAddRecord do not expect any validation from database layer.
// API layer should provide validation.
func TestAddRecord(t *testing.T) {
	tests := []struct {
		name      string
		rr        database.ResourceRecord
		expectErr bool
	}{
		{
			name: "normal RR",
			rr: database.ResourceRecord{
				Domain: "domain",
				Data:   "data",
				Type:   dns.TypeA.String(),
				Class:  dns.ClassIN.String(),
				TTL:    10,
			},
			expectErr: false,
		},
		{ // uniquness evaluated by domain, data, type and class
			name: "same RR",
			rr: database.ResourceRecord{
				Domain: "domain",
				Data:   "data",
				Type:   dns.TypeA.String(),
				Class:  dns.ClassIN.String(),
				TTL:    10,
			},
			expectErr: true,
		},
		{
			name: "incorrect type",
			rr: database.ResourceRecord{
				Domain: "domain",
				Data:   "data",
				Type:   "incorrect",
				Class:  dns.ClassIN.String(),
				TTL:    10,
			},
			expectErr: true,
		},
		{
			name: "incorrect class",
			rr: database.ResourceRecord{
				Domain: "domain",
				Data:   "data",
				Type:   dns.TypeA.String(),
				Class:  "incorrect",
				TTL:    10,
			},
			expectErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, err := testPG.AddRecord(t.Context(), test.rr)
			if test.expectErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			rr, err := testPG.GetRecord(t.Context(), id)
			assert.NoError(t, err)

			test.rr.ID = id
			assert.Equal(t, test.rr, rr)
		})
	}
}

func TestUpdateRecord(t *testing.T) {
	testRR := database.ResourceRecord{
		Domain: "test-update-record",
		Data:   "data",
		Type:   dns.TypeA.String(),
		Class:  dns.ClassIN.String(),
		TTL:    10,
	}
	id, err := testPG.AddRecord(t.Context(), testRR)
	assert.NoError(t, err)
	testRR.ID = id

	err = testPG.UpdateRecord(t.Context(), database.ResourceRecord{
		ID:     id,
		Domain: "test-update-update-record",
		Data:   "data-date",
		Type:   dns.TypeAAAA.String(),
		Class:  dns.ClassCH.String(),
		TTL:    1000,
	})
	assert.NoError(t, err)

	updatedRR, err := testPG.GetRecord(t.Context(), id)
	assert.NoError(t, err)

	assert.NotEqual(t, testRR, updatedRR)
}

func TestDeleteRecord(t *testing.T) {
	id, err := testPG.AddRecord(t.Context(),
		database.ResourceRecord{
			Domain: "del-domain",
			Data:   "del-data",
			Type:   dns.TypeA.String(),
			Class:  dns.ClassIN.String(),
			TTL:    10,
		})
	assert.NoError(t, err)

	err = testPG.DeleteRecord(t.Context(), id)
	assert.NoError(t, err)

	_, err = testPG.GetRecord(t.Context(), id)
	assert.Error(t, err)
}

func TestCheckUserPassword(t *testing.T) {
	u := database.User{
		Login:     "chkpass-login",
		FirstName: "chkpass-first",
		LastName:  "chkpass-last",
		Role:      "user",
	}
	id, err := testPG.AddUser(t.Context(), u, "password")
	assert.NoError(t, err)
	u.ID = id

	// correct
	newU, err := testPG.CheckUserPassword(t.Context(), u.Login, "password")
	assert.NoError(t, err)
	assert.Equal(t, u, newU)

	// incorrect
	_, err = testPG.CheckUserPassword(t.Context(), u.Login, "incorrect")
	assert.Error(t, err)
}

func TestDeleteUser(t *testing.T) {
	u := database.User{
		Login:     "del-login",
		FirstName: "del-first",
		LastName:  "del-last",
		Role:      "user",
	}
	id, err := testPG.AddUser(t.Context(), u, "password")
	assert.NoError(t, err)

	err = testPG.DeleteUser(t.Context(), id)
	assert.NoError(t, err)

	_, err = testPG.GetUser(t.Context(), u.Login)
	assert.Error(t, err)
}

func TestUpdateUser(t *testing.T) {
	u := database.User{
		Login:     "upd-login",
		FirstName: "upd-first",
		LastName:  "upd-last",
		Role:      "user",
	}
	id, err := testPG.AddUser(t.Context(), u, "password")
	assert.NoError(t, err)

	newUser := database.User{
		ID:        id,
		Login:     "new-login",
		FirstName: "new-first",
		LastName:  "new-last",
		Role:      "user",
	}
	err = testPG.UpdateUser(t.Context(), newUser, "new-password")
	assert.NoError(t, err)

	_, err = testPG.CheckUserPassword(t.Context(), newUser.Login, "password")
	assert.Error(t, err)
}

func TestFindRecords(t *testing.T) {}
