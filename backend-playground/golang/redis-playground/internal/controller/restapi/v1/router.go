package v1

import (
	"time"

	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/middleware"
	"github.com/agusheryanto182/redis-playground/internal/usecase"
	"github.com/agusheryanto182/redis-playground/pkg/jwt"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/agusheryanto182/redis-playground/pkg/ratelimiter"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(apiV1Group fiber.Router, u usecase.User, p usecase.Product, jwtManager *jwt.Manager, l logger.Interface, limiter *ratelimiter.Limiter) {
	r := &V1{u: u, p: p, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	// Public routes
	authGroup := apiV1Group.Group("/auth")
	{
		authGroup.Post("/register", r.register)
		authGroup.Post("/login", r.login)
	}
	// Product routes
	{
		apiV1Group.Get("/product", r.GetAll)
		apiV1Group.Get("/product/:id", r.GetByID)
	}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(jwtManager))

	userGroup := protected.Group("/user")
	{
		userGroup.Get(
			"/profile",
			limiter.Middleware(
				ratelimiter.Config{
					Namespace:   "profile",
					MaxRequests: 1,
					Window:      time.Second * 1,
				},
				func(c *fiber.Ctx) string {
					return middleware.GetUserID(c)
				},
			),
			r.profile,
		)
	}

	// Product routes
	productGroup := protected.Group("/product")
	{
		productGroup.Post("", r.Store)
		productGroup.Put("/:id", r.Update)
		productGroup.Patch("/:id", r.Patch)
		productGroup.Delete("/:id", r.Delete)
	}
}
