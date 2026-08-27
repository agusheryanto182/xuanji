package restapi

import (
	"net/http"

	"github.com/agusheryanto182/redis-playground/config"
	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/middleware"
	v1 "github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1"
	"github.com/agusheryanto182/redis-playground/internal/usecase"
	"github.com/agusheryanto182/redis-playground/pkg/jwt"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/agusheryanto182/redis-playground/pkg/ratelimiter"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// NewRouter -.
// Swagger spec:
//
//	@title       Go Clean Template API
//	@description Multi-domain clean architecture template with translation, user, and task management
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func NewRouter(app *fiber.App, cfg *config.Config, u usecase.User, p usecase.Product, jwtManager *jwt.Manager, l logger.Interface, limiter *ratelimiter.Limiter) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewRoutes(apiV1Group, u, p, jwtManager, l, limiter)
	}
}
