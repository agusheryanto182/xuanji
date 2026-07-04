package entity

import (
	"time"

	"github.com/google/uuid"
)

// Product -.
type Product struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name"        example:"Product Name"`
	Description string    `json:"description" example:"Product Description"`
	Price       float64   `json:"price"       example:"19.99"`
	Stock       int       `json:"stock"       example:"100"`
	CreatedAt   time.Time `json:"created_at"  example:"2026-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"  example:"2026-01-01T00:00:00Z"`
} // @name entity.Product
