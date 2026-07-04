package product

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductInput struct {
	Name        string
	Description string
	Price       int
	Stock       int
}

type UpdateProductInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       int
	Stock       int
}

type UpdatePartialProductInput struct {
	ID uuid.UUID

	Name        *string
	Description *string
	Price       *int
	Stock       *int

	UpdatedAt time.Time
}
