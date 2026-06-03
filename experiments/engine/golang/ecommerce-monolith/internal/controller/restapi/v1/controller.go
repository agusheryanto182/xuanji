package v1

import (
	"github.com/agusheryanto182/ecommerce-monolith/internal/usecase"
	"github.com/agusheryanto182/ecommerce-monolith/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	u usecase.User
	l logger.Interface
	v *validator.Validate
}
