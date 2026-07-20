package seed

import (
	"context"
	"fmt"
	"time"

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
	users := generateUsers(1000)

	if err := s.repo.BatchStore(ctx, users); err != nil {
		s.l.Error("Seeding user is failed ", err)
		return err
	}

	s.l.Info("Seeding user is completed: %v", len(users))

	return nil
}

func generateUsers(n int) []*entity.User {
	users := make([]*entity.User, n)
	now := time.Now().UTC()

	for i := range users {
		users[i] = &entity.User{
			ID:           uuid.New().String(),
			Username:     fmt.Sprintf("user %d", i+1),
			Email:        fmt.Sprintf("user%d@example.com", i+1),
			PasswordHash: "$2a$12$aD1/hXaiH2BRJu8DreHr/ehUbGi597SxMG2KiZeDFwfUgmZ9LssFq", // password: password123
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	return users
}
