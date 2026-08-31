package mockstubfake

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	GetByID(id int) (User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetUser(id int) (User, error) {
	return s.repo.GetByID(id)
}
