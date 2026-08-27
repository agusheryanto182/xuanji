package v1

import (
	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func parseUUIDParam(c *fiber.Ctx, key string) (uuid.UUID, error) {
	id := c.Params(key)
	if id == "" {
		return uuid.Nil, entity.ErrIdIsRequired
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func parseLimitAndOffset(c *fiber.Ctx, limitDefault, offsetDefault int) (limit, offset int) {
	limit = limitDefault
	offset = offsetDefault

	if c.Query("limit") != "" {
		limit = c.QueryInt("limit")
	}

	if c.Query("offset") != "" {
		offset = c.QueryInt("offset")
	}

	return limit, offset
}
