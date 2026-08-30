package mockStubFake

import (
	"errors"
	"testing"
)

// Stub
// Digunakan ketika kita hanya membutuhkan
// return value tertentu dari dependency.
type stubUserRepository struct {
	user User
	err  error
}

func (s *stubUserRepository) GetByID(id int) (User, error) {
	return s.user, s.err
}

func TestUserService_GetUser_Stub(t *testing.T) {
	repo := &stubUserRepository{
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

// Mock
// Digunakan ketika kita ingin mengetahui
// bagaimana dependency digunakan.
type mockUserRepository struct {
	calledID int
}

func (m *mockUserRepository) GetByID(id int) (User, error) {
	m.calledID = id

	return User{
		ID:   id,
		Name: "Agus",
	}, nil
}

func TestUserService_GetUser_Mock(t *testing.T) {
	repo := &mockUserRepository{}
	service := NewUserService(repo)

	_, err := service.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}

	if repo.calledID != 1 {
		t.Errorf("got ID %d, want 1", repo.calledID)
	}
}

// Fake
// Implementasi sederhana yang benar-benar
// memiliki behavior.
type fakeUserRepository struct {
	users map[int]User
}

func (f *fakeUserRepository) GetByID(id int) (User, error) {
	user, ok := f.users[id]
	if !ok {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

func TestUserService_GetUser_Fake(t *testing.T) {
	repo := &fakeUserRepository{
		users: map[int]User{
			1: {
				ID:   1,
				Name: "Agus",
			},
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
