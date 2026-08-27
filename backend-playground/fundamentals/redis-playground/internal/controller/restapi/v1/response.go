package v1

import (
	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

func successResponse[T any](
	ctx *fiber.Ctx,
	status int,
	data T,
	meta *response.Meta,
) error {
	return ctx.Status(status).JSON(response.Response[T]{
		Status: response.StatusSuccess,
		Data:   data,
		Meta:   meta,
	})
}
