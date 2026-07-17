package middleware

import (
	"strings"

	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/response"
	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

const _bearerParts = 2

type errorResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Auth returns a JWT authentication middleware for Fiber.
func Auth(jwtManager *jwt.Manager) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get("Authorization")
		if header == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Status: "error", Code: response.ErrMissingAuthorizationHeader, Message: entity.ErrMissingAuthorizationHeader.Error()})
		}

		parts := strings.SplitN(header, " ", _bearerParts)
		if len(parts) != _bearerParts || parts[0] != "Bearer" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Status: "error", Code: response.ErrInvalidAuthorizationHeader, Message: entity.ErrInvalidAuthorizationHeader.Error()})
		}

		userID, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Status: "error", Code: response.ErrInvalidOrExpiredToken, Message: entity.ErrInvalidOrExpiredToken.Error()})
		}

		ctx.Locals("userID", userID)

		return ctx.Next()
	}
}
