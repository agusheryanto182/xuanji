// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// UserRepo -.
	UserRepo interface {
		Store(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
	}

	// ProductRepo -.
	ProductRepo interface {
		Store(ctx context.Context, product *entity.Product) error
		BatchStore(ctx context.Context, products []*entity.Product) error
		GetAll(ctx context.Context, limit, offset int) ([]*entity.Product, error)
		GetByID(ctx context.Context, ID uuid.UUID) (*entity.Product, error)
		Update(ctx context.Context, product *entity.Product) error
		Patch(ctx context.Context, id uuid.UUID, updates map[string]any) error
		Delete(ctx context.Context, id string) error
	}
)
