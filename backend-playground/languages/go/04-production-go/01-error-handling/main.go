package main

import (
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Reason
}

type Repository struct{}

func (Repository) FindUser(id int) error {
	if id == 42 {
		return ErrUserNotFound
	}
	return nil
}

type Service struct {
	repo Repository
}

func (s Service) GetUser(id int) error {
	if id <= 0 {
		return &ValidationError{
			Field:  "id",
			Reason: "must be greater than zero",
		}
	}

	if err := s.repo.FindUser(id); err != nil {
		return fmt.Errorf("get user %d: %w", id, err)
	}

	return nil
}

func main() {
	service := Service{repo: Repository{}}

	err := service.GetUser(42)
	if err != nil {
		fmt.Println("error:", err)

		if errors.Is(err, ErrUserNotFound) {
			fmt.Println("classification: user not found")
		}
	}

	err = service.GetUser(0)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf(
				"classification: validation error (field=%s, reason=%s)\n",
				validationErr.Field,
				validationErr.Reason,
			)
		}
	}
}
