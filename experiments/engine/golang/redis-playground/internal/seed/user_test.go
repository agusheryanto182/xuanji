package seed_test

import (
	"context"
	"errors"
	"testing"

	entity "github.com/agusheryanto182/redis-playground/internal/entity"
	seed "github.com/agusheryanto182/redis-playground/internal/seed/user"
	gomock "github.com/golang/mock/gomock"
)

func newUserSeeder(t *testing.T) (*seed.UserSeeder, *MockUserRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repo := NewMockUserRepo(ctrl)
	seeder := seed.NewUserSeeder(repo)

	return seeder, repo
}

func TestUserSeeder_Seed(t *testing.T) {
	t.Run("seed success", func(t *testing.T) {
		seeder, repo := newUserSeeder(t)
		repo.EXPECT().
			Store(gomock.Any(), gomock.AssignableToTypeOf(&entity.User{})).
			DoAndReturn(func(ctx context.Context, user *entity.User) error {
				if user.Username != "suga" {
					t.Errorf("unexpected username")
				}

				if user.Email != "suga@example.com" {
					t.Errorf("unexpected email")
				}

				if user.ID == "" {
					t.Error("id should not be empty")
				}

				return nil
			})

		err := seeder.Seed(context.Background())

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		t.Log("seeder.Seed() completed")
	})
}

func TestUserSeeder_StoreError(t *testing.T) {
	seeder, repo := newUserSeeder(t)

	expected := errors.New("db error")

	repo.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return(expected)

	err := seeder.Seed(context.Background())

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
