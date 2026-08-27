// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/agusheryanto182/ecommerce-monolith/internal/entity"
	"github.com/agusheryanto182/ecommerce-monolith/internal/usecase/product"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// User -.
	User interface {
		Register(ctx context.Context, username, email, password string) (entity.User, error)
		Login(ctx context.Context, email, password string) (string, error)
		GetUser(ctx context.Context, userID string) (entity.User, error)
	}

	// Product -.
	Product interface {
		Store(ctx context.Context, product *entity.Product) (*entity.Product, error)
		GetByID(ctx context.Context, ID uuid.UUID) (*entity.Product, error)
		Update(ctx context.Context, product *entity.Product) (*entity.Product, error)
		Patch(ctx context.Context, input product.PatchInput) (*entity.Product, error)
		Delete(ctx context.Context, id string) error
	}
)
