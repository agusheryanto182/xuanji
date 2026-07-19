package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/repo"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

const (
	productsCachePattern = "products:*"
	productsCacheTTL     = 5 * time.Minute
)

// UseCase -.
type UseCase struct {
	repo  repo.ProductRepo
	redis *redis.Client
	l     logger.Interface
}

// New -.
func New(r repo.ProductRepo, rdb *redis.Client, l logger.Interface) *UseCase {
	return &UseCase{
		repo:  r,
		redis: rdb,
		l:     l,
	}
}

func (uc *UseCase) invalidateProductsCache(ctx context.Context) {
	var cursor uint64
	var keys []string

	for {
		batch, nextCursor, err := uc.redis.Scan(ctx, cursor, productsCachePattern, 100).Result()
		if err != nil {
			uc.l.Warn("Failed to scan products cache keys: %v", err)
			return
		}

		keys = append(keys, batch...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return
	}

	if err := uc.redis.Del(ctx, keys...).Err(); err != nil {
		uc.l.Warn("Failed to invalidate products cache: %v", err)
	}
}

// Store -.
func (uc *UseCase) Store(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	now := time.Now().UTC()

	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now

	if err := uc.repo.Store(ctx, product); err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Store - uc.repo.Store: %w", err))
		return nil, entity.ErrInvalidProductCreate
	}

	uc.invalidateProductsCache(ctx)

	return product, nil
}

// Get -.
func (uc *UseCase) GetAll(
	ctx context.Context,
	limit,
	offset int,
) ([]*entity.Product, int, error) {

	cacheKey := fmt.Sprintf("products:%d:%d", limit, offset)

	value, err := uc.redis.Get(ctx, cacheKey).Result()
	switch {
	case err == nil:
		var cache ProductCache

		if err := json.Unmarshal([]byte(value), &cache); err == nil {
			uc.l.Debug("Cache hit: %s", cacheKey)

			return cache.Products, cache.Total, nil
		}

		uc.l.Warn("Failed to unmarshal cache: %v", err)

	case errors.Is(err, redis.Nil):
		uc.l.Debug("Cache miss: %s", cacheKey)

	default:
		uc.l.Warn("Redis GET failed: %v", err)
	}

	total, err := uc.repo.CountProducts(ctx)
	if err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - CountProducts: %w", err))
		return nil, 0, entity.ErrInternalServerError
	}

	products, err := uc.repo.GetAll(ctx, limit, offset)
	if err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - GetAll: %w", err))
		return nil, 0, entity.ErrInternalServerError
	}

	cache := ProductCache{
		Products: products,
		Total:    total,
	}

	bytes, err := json.Marshal(cache)
	if err != nil {
		uc.l.Warn("Failed to marshal cache: %v", err)
		return products, total, nil
	}

	if err := uc.redis.Set(ctx, cacheKey, bytes, productsCacheTTL).Err(); err != nil {
		uc.l.Warn("Redis SET failed: %v", err)
	}

	return products, total, nil
}

// GetProductByID -.
func (uc *UseCase) GetByID(ctx context.Context, ID uuid.UUID) (*entity.Product, error) {
	product, err := uc.repo.GetByID(ctx, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrProductNotFound
		}

		uc.l.Error(fmt.Errorf(
			"ProductUseCase.GetByID(id=%s): %w",
			ID,
			err,
		))

		return nil, entity.ErrInternalServerError
	}

	return product, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	product.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, product); err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return nil, err
		}

		uc.l.Error(fmt.Errorf("ProductUseCase - Update - uc.repo.Update: %w", err))
		return nil, entity.ErrInternalServerError
	}

	uc.invalidateProductsCache(ctx)

	return product, nil
}

// Patch -.
func (uc *UseCase) Patch(ctx context.Context, input PatchInput) (*entity.Product, error) {
	if input.ID == uuid.Nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Patch - input.ID is nil"))
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
		uc.l.Error(fmt.Errorf("ProductUseCase - Patch - no fields to update"))
		return nil, entity.ErrInvalidProductPatch
	}

	updatedFields["updated_at"] = time.Now().UTC()

	if err := uc.repo.Patch(ctx, input.ID, updatedFields); err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Patch - uc.repo.Patch: %w", err))
		return nil, entity.ErrInvalidProductPatch
	}

	product, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		uc.l.Error(fmt.Errorf("ProductUseCase - Patch - uc.GetProduct: %w", err))
		return nil, entity.ErrProductNotFound
	}

	uc.invalidateProductsCache(ctx)

	return product, nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return err
		}

		uc.l.Error(fmt.Errorf("ProductUseCase - Delete - uc.repo.Delete: %w", err))
		return entity.ErrInvalidProductDelete
	}

	uc.invalidateProductsCache(ctx)

	return nil
}
