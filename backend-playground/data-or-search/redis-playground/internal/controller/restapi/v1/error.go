package v1

import (
	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

func errorResponse(ctx *fiber.Ctx, status int, code, msg string) error {
	return ctx.Status(status).JSON(response.Response[any]{
		Status: response.StatusError,
		Error: &response.Error{
			Code:    code,
			Message: msg,
		},
	})
}
