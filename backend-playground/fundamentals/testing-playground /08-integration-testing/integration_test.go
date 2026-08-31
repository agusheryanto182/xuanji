package integrationtesting

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

func setupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open(
		"postgres",
		"postgres://postgres:postgres@localhost:5434/integration_testing?sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	// clear the users table before each test
	_, err = db.Exec(`DELETE FROM users`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	service := NewUserService(repo)
	handler := NewHandler(service)

	body := `{
		"name": "Agus",
		"email": "agus@example.com"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"got status %d, want %d",
			rec.Code,
			http.StatusCreated,
		)
	}
}
