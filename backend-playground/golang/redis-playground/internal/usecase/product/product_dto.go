package product

import (
	"time"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/google/uuid"
)

type PatchInput struct {
	ID uuid.UUID

	Name        *string
	Description *string
	Price       *float64
	Stock       *int

	UpdatedAt time.Time
}

type ProductCache struct {
	Products []*entity.Product `json:"products"`
	Total    int               `json:"total"`
}
