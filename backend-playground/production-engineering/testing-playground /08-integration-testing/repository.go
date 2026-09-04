package integrationtesting

import "database/sql"

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user User) (User, error) {
	var created User

	err := r.db.QueryRow(
		`INSERT INTO users (name, email)
		 VALUES ($1, $2)
		 RETURNING id, name, email`,
		user.Name,
		user.Email,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Email,
	)

	return created, err
}
