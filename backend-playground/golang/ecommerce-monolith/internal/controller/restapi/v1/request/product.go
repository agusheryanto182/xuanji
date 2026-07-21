package request

// Store -.
type Store struct {
	Name        string  `json:"name"        validate:"required,min=3,max=255" example:"Product Name"`
	Description string  `json:"description" validate:"required,min=10"         example:"Product Description"`
	Price       float64 `json:"price"       validate:"required,gt=0"            example:"99.99"`
	Stock       int     `json:"stock"       validate:"required,gt=0"            example:"10"`
} // @name v1.Store

// Update
type Update struct {
	Name        string  `json:"name"        validate:"omitempty,min=3,max=255" example:"Product Name"`
	Description string  `json:"description" validate:"omitempty,min=10"         example:"Product Description"`
	Price       float64 `json:"price"       validate:"omitempty,gt=0"            example:"99.99"`
	Stock       int     `json:"stock"       validate:"omitempty,gt=0"            example:"10"`
}

// Patch
type Patch struct {
	Name        *string  `json:"name"        validate:"omitempty,min=3,max=255" example:"Product Name"`
	Description *string  `json:"description" validate:"omitempty,min=10"         example:"Product Description"`
	Price       *float64 `json:"price"       validate:"omitempty,gt=0"            example:"99.99"`
	Stock       *int     `json:"stock"       validate:"omitempty,gt=0"            example:"10"`
}
