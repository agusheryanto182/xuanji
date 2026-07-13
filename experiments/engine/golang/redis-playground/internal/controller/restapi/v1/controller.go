package v1

import (
	"github.com/agusheryanto182/redis-playground/internal/usecase"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	u usecase.User
	p usecase.Product
	l logger.Interface
	v *validator.Validate
}
