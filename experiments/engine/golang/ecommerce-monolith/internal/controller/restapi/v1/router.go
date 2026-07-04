package v1

import (
	"github.com/agusheryanto182/ecommerce-monolith/internal/controller/restapi/middleware"
	"github.com/agusheryanto182/ecommerce-monolith/internal/usecase"
	"github.com/agusheryanto182/ecommerce-monolith/pkg/jwt"
	"github.com/agusheryanto182/ecommerce-monolith/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(apiV1Group fiber.Router, u usecase.User, p usecase.Product, jwtManager *jwt.Manager, l logger.Interface) {
	r := &V1{u: u, p: p, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	// Public routes
	authGroup := apiV1Group.Group("/auth")
	{
		authGroup.Post("/register", r.register)
		authGroup.Post("/login", r.login)
	}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(jwtManager))

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
	}

	productGroup := protected.Group("/product")
	{
		productGroup.Post("", r.Store)
	}
}
