package repositorytesting

import (
	"database/sql"
)

type User struct {
	ID    int
	Name  string
	Email string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(user User) (int, error) {
	var id int

	err := r.db.QueryRow(
		`INSERT INTO users (name, email)
		 VALUES ($1, $2)
		 RETURNING id`,
		user.Name,
		user.Email,
	).Scan(&id)

	return id, err
}

func (r *UserRepository) GetByID(id int) (User, error) {
	var user User

	err := r.db.QueryRow(
		`SELECT id, name, email FROM users WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
	)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *UserRepository) Update(user User) error {
	_, err := r.db.Exec(
		`UPDATE users
		 SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $3`,
		user.Name,
		user.Email,
		user.ID,
	)

	return err
}