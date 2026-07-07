package v1

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func parseUUIDParam(c *fiber.Ctx, key string) (uuid.UUID, error) {
	id := c.Params(key)
	if id == "" {
		return uuid.Nil, errors.New("id is required")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("invalid id")
	}

	return uid, nil
}
