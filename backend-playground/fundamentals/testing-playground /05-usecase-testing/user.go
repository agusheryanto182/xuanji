package usecasetesting

import "errors"

type User struct {
	ID    int
	Name  string
	Email string
}

type UserRepository interface {
	FindByEmail(email string) (User, error)
	Create(user User) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(user User) error {
	_, err := s.repo.FindByEmail(user.Email)

	if err == nil {
		return errors.New("email already registered")
	}

	return s.repo.Create(user)
}
