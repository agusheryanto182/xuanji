// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/agusheryanto182/ecommerce-monolith/internal/entity"
	"github.com/agusheryanto182/ecommerce-monolith/internal/usecase/product"
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
		Store(ctx context.Context, input product.CreateProductInput) (*entity.Product, error)
		GetProduct(ctx context.Context, column, value string) (*entity.Product, error)
		Update(ctx context.Context, input product.UpdateProductInput) (*entity.Product, error)
		UpdatePartial(ctx context.Context, input product.UpdatePartialProductInput) (*entity.Product, error)
		Delete(ctx context.Context, id string) error
	}
)
