package usecasetesting

import (
	"errors"
	"testing"
)

type fakeUserRepository struct {
	foundUser User
	findErr   error

	createdUser User
	createErr   error

	createCalled bool
}

func (f *fakeUserRepository) FindByEmail(email string) (User, error) {
	return f.foundUser, f.findErr
}

func (f *fakeUserRepository) Create(user User) error {
	f.createCalled = true
	f.createdUser = user

	return f.createErr
}

func TestUserService_CreateUser(t *testing.T) {
	t.Run("email is not registered", func(t *testing.T) {
		repo := &fakeUserRepository{
			findErr: errors.New("user not found"),
		}

		service := NewUserService(repo)

		user := User{
			Name:  "Agus",
			Email: "agus@example.com",
		}

		err := service.CreateUser(user)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !repo.createCalled {
			t.Error("expected Create to be called")
		}

		if repo.createdUser.Email != user.Email {
			t.Errorf(
				"got email %s, want %s",
				repo.createdUser.Email,
				user.Email,
			)
		}
	})

	t.Run("email is already registered", func(t *testing.T) {
		repo := &fakeUserRepository{
			foundUser: User{
				ID:    1,
				Name:  "Agus",
				Email: "agus@example.com",
			},
			findErr: nil,
		}

		service := NewUserService(repo)

		user := User{
			Name:  "Agus 2",
			Email: "agus@example.com",
		}

		err := service.CreateUser(user)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if repo.createCalled {
			t.Error("Create should not be called")
		}
	})
}
