package v1

import (
	"github.com/agusheryanto182/ecommerce-monolith/internal/entity"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) bindAndValidate(ctx *fiber.Ctx, dst any) error {
	if err := ctx.BodyParser(dst); err != nil {
		return entity.ErrorInvalidRequestBody
	}

	if err := r.v.Struct(dst); err != nil {
		return entity.ErrorInvalidRequestBody
	}

	return nil
}
