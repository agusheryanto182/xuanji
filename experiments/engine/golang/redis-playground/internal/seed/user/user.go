package seed

import (
	"context"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/repo"
	"github.com/google/uuid"
)

type UserSeeder struct {
	repo repo.UserRepo
}

func NewUserSeeder(repo repo.UserRepo) *UserSeeder {
	return &UserSeeder{repo}
}

func (s *UserSeeder) Seed(ctx context.Context) error {
	user := entity.User{
		ID:           uuid.New().String(),
		Username:     "suga",
		Email:        "suga@example.com",
		PasswordHash: "$2a$12$aD1/hXaiH2BRJu8DreHr/ehUbGi597SxMG2KiZeDFwfUgmZ9LssFq", // password: password123
	}

	return s.repo.Store(ctx, &user)
}
