package repositorytesting

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open(
		"postgres",
		"postgres://postgres:postgres@localhost:5433/testing?sslmode=disable",
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

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	user := User{
		Name:  "Agus",
		Email: "agus@example.com",
	}

	id, err := repo.Create(user)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != user.Name {
		t.Errorf("got name %s, want %s", got.Name, user.Name)
	}

	if got.Email != user.Email {
		t.Errorf("got email %s, want %s", got.Email, user.Email)
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	user := User{
		Name:  "Agus",
		Email: "update@example.com",
	}

	id, err := repo.Create(user)
	if err != nil {
		t.Fatal(err)
	}

	created, err := repo.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}

	created.Name = "Agus Heryanto"
	created.Email = "agus.h@example.com"

	err = repo.Update(created)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != created.Name {
		t.Errorf("got name %s, want %s", got.Name, created.Name)
	}

	if got.Email != created.Email {
		t.Errorf("got email %s, want %s", got.Email, created.Email)
	}
}
