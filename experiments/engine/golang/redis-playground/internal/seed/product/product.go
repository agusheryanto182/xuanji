package seed

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/repo"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/google/uuid"
)

type ProductSeeder struct {
	repo repo.ProductRepo
	l    logger.Interface
}

func NewProductSeeder(repo repo.ProductRepo, l logger.Interface) *ProductSeeder {
	return &ProductSeeder{repo, l}
}

func (s *ProductSeeder) Seed(ctx context.Context) error {
	products := generateProducts(1000)

	if err := s.repo.BatchStore(ctx, products); err != nil {
		s.l.Error(fmt.Errorf("Seeding products is failed: %w", err))

		return err
	}

	s.l.Info("Seeding %d products is completed", len(products))

	return nil
}

func generateProducts(n int) []*entity.Product {
	products := make([]*entity.Product, n)
	now := time.Now().UTC()

	for i := range products {
		price := math.Round(rand.Float64()*10000) / 100

		products[i] = &entity.Product{
			ID:          uuid.New(),
			Name:        fmt.Sprintf("Test Product %d", i+1),
			Description: fmt.Sprintf("This is test product %d", i+1),
			Price:       price,
			Stock:       rand.IntN(100),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	return products
}
