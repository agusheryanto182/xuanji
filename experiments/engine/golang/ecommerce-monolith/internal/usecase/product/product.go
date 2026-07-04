package product

import (
	"context"
	"fmt"
	"time"

	"github.com/agusheryanto182/ecommerce-monolith/internal/entity"
	"github.com/agusheryanto182/ecommerce-monolith/internal/repo"
	"github.com/agusheryanto182/ecommerce-monolith/pkg/logger"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo repo.ProductRepo
	l    logger.Interface
}

// New -.
func New(r repo.ProductRepo, l logger.Interface) *UseCase {
	return &UseCase{
		repo: r,
		l:    l,
	}
}

// Store -.
func (uc *UseCase) Store(ctx context.Context, product entity.Product) (*entity.Product, error) {
	now := time.Now().UTC()

	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now

	if err := uc.repo.Store(ctx, &product); err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Store - uc.repo.Store: %w", err))
		return nil, entity.ErrInvalidProductCreate
	}

	return &product, nil
}

// GetProduct -.
func (uc *UseCase) GetProduct(ctx context.Context, column, value string) (*entity.Product, error) {
	product, err := uc.repo.GetProduct(ctx, column, value)
	if err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - GetProduct - uc.repo.GetProduct: %w", err))
		return nil, entity.ErrProductNotFound
	}

	return &product, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, product entity.Product) (*entity.Product, error) {
	if product.ID == uuid.Nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Update - product.ID is nil"))
		return nil, entity.ErrInvalidIdProduct
	}

	product.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, &product); err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Update - uc.repo.Update: %w", err))
		return nil, entity.ErrInvalidProductUpdate
	}

	return &product, nil
}

// UpdatePartial -.
func (uc *UseCase) UpdatePartial(ctx context.Context, input UpdatePartialProductInput) (*entity.Product, error) {
	if input.ID == uuid.Nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - UpdatePartial - input.ID is nil"))
		return nil, entity.ErrInvalidIdProduct
	}

	updatedFields := make(map[string]any)
	if input.Name != nil {
		updatedFields["name"] = *input.Name
	}
	if input.Description != nil {
		updatedFields["description"] = *input.Description
	}
	if input.Price != nil {
		updatedFields["price"] = *input.Price
	}
	if input.Stock != nil {
		updatedFields["stock"] = *input.Stock
	}

	if len(updatedFields) == 0 {
		uc.l.Error(fmt.Errorf("ProductUseCase - UpdatePartial - no fields to update"))
		return nil, entity.ErrInvalidProductPartialUpdate
	}

	updatedFields["updated_at"] = time.Now().UTC()

	if err := uc.repo.PartialUpdate(ctx, input.ID, updatedFields); err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - UpdatePartial - uc.repo.PartialUpdate: %w", err))
		return nil, entity.ErrInvalidProductPartialUpdate
	}

	product, err := uc.GetProduct(ctx, "id", input.ID.String())
	if err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - UpdatePartial - uc.GetProduct: %w", err))
		return nil, entity.ErrProductNotFound
	}

	return product, nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Delete - uc.repo.Delete: %w", err))
		return entity.ErrInvalidProductDelete
	}

	return nil
}
