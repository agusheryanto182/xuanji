package seed

import (
	"context"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/repo"
	"github.com/google/uuid"
)

type ProductSeeder struct {
	repo repo.ProductRepo
}

func NewProductSeeder(repo repo.ProductRepo) *ProductSeeder {
	return &ProductSeeder{repo}
}

func (s *ProductSeeder) Seed(ctx context.Context) error {
	products := []*entity.Product{
		{
			ID:          uuid.New(),
			Name:        "Test Product 1",
			Description: "This is a test product 1",
			Price:       9.99,
			Stock:       10,
		},
		{
			ID:          uuid.New(),
			Name:        "Test Product 2",
			Description: "This is a test product 2",
			Price:       19.99,
			Stock:       20,
		},
		{
			ID:          uuid.New(),
			Name:        "Test Product 3",
			Description: "This is a test product 3",
			Price:       29.99,
			Stock:       30,
		},
	}

	for _, product := range products {
		err := s.repo.Store(ctx, product)
		if err != nil {
			return err
		}
	}

	return nil
}
