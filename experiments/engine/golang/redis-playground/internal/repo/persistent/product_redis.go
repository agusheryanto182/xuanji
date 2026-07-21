package persistent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/repo"
	"github.com/redis/go-redis/v9"
)

const (
	productsCachePattern = "products:*"
	productsCacheTTL     = 5 * time.Minute
)

type ProductRedis struct {
	client *redis.Client
}

var _ repo.ProductCache = (*ProductRedis)(nil)

func NewProductRedis(client *redis.Client) *ProductRedis {
	return &ProductRedis{
		client: client,
	}
}

type productCache struct {
	Products []*entity.Product `json:"products"`
	Total    int               `json:"total"`
}

func (r *ProductRedis) GetAll(
	ctx context.Context,
	limit,
	offset int,
) ([]*entity.Product, int, error) {

	cacheKey := fmt.Sprintf("products:%d:%d", limit, offset)

	value, err := r.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, 0, redis.Nil
		}

		return nil, 0, err
	}

	var cache productCache
	if err := json.Unmarshal([]byte(value), &cache); err != nil {
		return nil, 0, err
	}

	return cache.Products, cache.Total, nil
}

func (r *ProductRedis) SetAll(
	ctx context.Context,
	limit,
	offset int,
	products []*entity.Product,
	total int,
) error {

	cacheKey := fmt.Sprintf("products:%d:%d", limit, offset)

	cache := productCache{
		Products: products,
		Total:    total,
	}

	bytes, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return r.client.Set(
		ctx,
		cacheKey,
		bytes,
		productsCacheTTL,
	).Err()
}

func (r *ProductRedis) Invalidate(ctx context.Context) error {

	var (
		cursor uint64
		keys   []string
	)

	for {

		batch, nextCursor, err := r.client.Scan(
			ctx,
			cursor,
			productsCachePattern,
			100,
		).Result()

		if err != nil {
			return err
		}

		keys = append(keys, batch...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}
