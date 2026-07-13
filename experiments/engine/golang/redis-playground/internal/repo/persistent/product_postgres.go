package persistent

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/pkg/postgres"
	"github.com/google/uuid"
)

type ProductRepo struct {
	*postgres.Postgres
}

// NewProductRepo -.
func NewProductRepo(pg *postgres.Postgres) *ProductRepo {
	return &ProductRepo{pg}
}

// Store -.
func (r *ProductRepo) Store(ctx context.Context, product *entity.Product) error {
	sql, args, err := r.Builder.
		Insert("products").
		Columns("id, name, description, price, stock, created_at, updated_at").
		Values(product.ID, product.Name, product.Description, product.Price, product.Stock, product.CreatedAt, product.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// Get By ID
func (r *ProductRepo) GetByID(ctx context.Context, ID uuid.UUID) (*entity.Product, error) {
	sql, args, err := r.Builder.
		Select("id, name, description, price, stock, created_at, updated_at").
		From("products").
		Where(sq.Eq{"id": ID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var product entity.Product
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// Full Update -.
func (r *ProductRepo) Update(ctx context.Context, product *entity.Product) error {
	sql, args, err := r.Builder.
		Update("products").
		Set("name", product.Name).
		Set("description", product.Description).
		Set("price", product.Price).
		Set("stock", product.Stock).
		Set("updated_at", product.UpdatedAt).
		Where(sq.Eq{"id": product.ID}).
		ToSql()
	if err != nil {
		return err
	}

	cmdTag, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return entity.ErrProductNotFound
	}

	return nil
}

// Patch -.
func (r *ProductRepo) Patch(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	sql := r.Builder.Update("products")

	for column, value := range updates {
		sql = sql.Set(column, value)
	}

	sql = sql.Where(sq.Eq{"id": id})

	sqlStr, args, err := sql.ToSql()
	if err != nil {
		return err
	}

	_, err = r.Pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return err
	}
	return nil
}

// Delete -.
func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	sql, args, err := r.Builder.
		Delete("products").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	cmdTag, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return entity.ErrProductNotFound
	}

	return nil
}
