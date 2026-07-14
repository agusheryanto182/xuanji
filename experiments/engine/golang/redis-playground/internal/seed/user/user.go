package seed

import (
	"context"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/repo"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/google/uuid"
)

type UserSeeder struct {
	repo repo.UserRepo
	l    logger.Interface
}

func NewUserSeeder(repo repo.UserRepo, l logger.Interface) *UserSeeder {
	return &UserSeeder{repo, l}
}

func (s *UserSeeder) Seed(ctx context.Context) error {
	user := entity.User{
		ID:           uuid.New().String(),
		Username:     "suga",
		Email:        "suga@example.com",
		PasswordHash: "$2a$12$aD1/hXaiH2BRJu8DreHr/ehUbGi597SxMG2KiZeDFwfUgmZ9LssFq", // password: password123
	}

	if err := s.repo.Store(ctx, &user); err != nil {
		s.l.Error("Seeding user is failed ", err)
		return err
	}

	s.l.Info("Seeding user is completed: %v", user.Email)

	return nil
}
