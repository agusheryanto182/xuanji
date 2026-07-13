package product

import (
	"time"

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
