package product

import (
	"time"

	"github.com/google/uuid"
)

type UpdatePartialProductInput struct {
	ID uuid.UUID

	Name        *string
	Description *string
	Price       *float64
	Stock       *int

	UpdatedAt time.Time
}
