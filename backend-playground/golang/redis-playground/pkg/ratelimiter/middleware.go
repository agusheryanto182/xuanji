package ratelimiter

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type KeyFunc func(*fiber.Ctx) string

func (l *Limiter) Middleware(
	cfg Config,
	keyFunc KeyFunc,
) fiber.Handler {

	cfg.setDefaults()

	return func(c *fiber.Ctx) error {

		key := keyFunc(c)

		allowed, remaining, err := l.Allow(
			c.Context(),
			cfg,
			key,
		)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		c.Set(
			"X-RateLimit-Limit",
			strconv.FormatInt(cfg.MaxRequests, 10),
		)

		c.Set(
			"X-RateLimit-Remaining",
			strconv.FormatInt(remaining, 10),
		)

		if !allowed {
			return fiber.ErrTooManyRequests
		}

		return c.Next()
	}
}
