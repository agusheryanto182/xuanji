package dependencyinjection

import (
	"fmt"
	"testing"
)

type fakeUserRepository struct {
	user User
	err  error
}

func (f *fakeUserRepository) GetByID(id int) (User, error) {
	return f.user, f.err
}

func TestUserService_GetUserSuccess(t *testing.T) {
	repo := &fakeUserRepository{
		user: User{
			ID:   1,
			Name: "Agus",
		},
	}

	service := NewUserService(repo)

	got, err := service.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "Agus" {
		t.Errorf("got %s, want Agus", got.Name)
	}
}

func TestUserService_GetUserError(t *testing.T) {
	repo := &fakeUserRepository{
		user: User{
			ID:   1,
			Name: "Agus",
		},
		err: fmt.Errorf("user not found"),
	}

	service := NewUserService(repo)

	got, err := service.GetUser(2)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "Agus" {
		t.Errorf("got %s, want Agus", got.Name)
	}
}
