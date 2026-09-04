package e2etesting

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

func setupE2EDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open(
		"postgres",
		"postgres://postgres:postgres@localhost:5435/e2e_testing?sslmode=disable",
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

func TestCreateAndGetUser(t *testing.T) {
	db := setupE2EDB(t)
	defer db.Close()

	server := httptest.NewServer(NewServer(db))
	defer server.Close()

	// Create user.
	body := `{
		"name": "Agus",
		"email": "agus@example.com"
	}`

	resp, err := http.Post(
		server.URL+"/users",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"got status %d, want %d",
			resp.StatusCode,
			http.StatusCreated,
		)
	}

	var created User

	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	// Get the created user.
	resp, err = http.Get(
		server.URL + "/users/" + strconv.Itoa(created.ID),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"got status %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	var got User

	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Name != "Agus" {
		t.Errorf("got name %s, want Agus", got.Name)
	}
}
