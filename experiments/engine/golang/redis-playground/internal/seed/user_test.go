package seed_test

import (
	"context"
	"errors"
	"testing"

	entity "github.com/agusheryanto182/redis-playground/internal/entity"
	seed "github.com/agusheryanto182/redis-playground/internal/seed/user"
	gomock "github.com/golang/mock/gomock"
)

func newUserSeeder(t *testing.T) (*seed.UserSeeder, *MockUserRepo, *MockInterface) {
	t.Helper()

	ctrl := gomock.NewController(t)

	l := NewMockInterface(ctrl)

	repo := NewMockUserRepo(ctrl)
	seeder := seed.NewUserSeeder(repo, l)

	return seeder, repo, l
}

func TestUserSeeder_Seed(t *testing.T) {
	t.Run("seed success", func(t *testing.T) {
		seeder, repo, l := newUserSeeder(t)

		l.EXPECT().
			Info(gomock.Any(), gomock.Any()).Times(1)

		repo.EXPECT().
			BatchStore(gomock.Any(), gomock.AssignableToTypeOf([]*entity.User{})).
			DoAndReturn(func(ctx context.Context, users []*entity.User) error {
				if len(users) != 1000 {
					t.Errorf("expected 1000 users, got %v", len(users))
				}

				if users[0].Username != "user 1" {
					t.Errorf("unexpected username")
				}

				if users[0].Email != "user1@example.com" {
					t.Errorf("unexpected email")
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
	seeder, repo, l := newUserSeeder(t)

	expected := errors.New("db error")

	l.EXPECT().
		Error(gomock.Any(), gomock.Any()).Times(1)

	repo.EXPECT().
		BatchStore(gomock.Any(), gomock.Any()).
		Return(expected)

	err := seeder.Seed(context.Background())

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
