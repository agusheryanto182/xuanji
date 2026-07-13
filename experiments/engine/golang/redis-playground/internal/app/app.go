// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/agusheryanto182/redis-playground/config"
	"github.com/agusheryanto182/redis-playground/internal/controller/restapi"
	"github.com/agusheryanto182/redis-playground/internal/repo/persistent"
	"github.com/agusheryanto182/redis-playground/internal/usecase/product"
	"github.com/agusheryanto182/redis-playground/internal/usecase/user"
	"github.com/agusheryanto182/redis-playground/pkg/httpserver"
	"github.com/agusheryanto182/redis-playground/pkg/jwt"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/agusheryanto182/redis-playground/pkg/postgres"
)

type useCases struct {
	user    *user.UseCase
	product *product.UseCase
}

type servers struct {
	http *httpserver.Server
}

func initUseCases(pg *postgres.Postgres, jwtManager *jwt.Manager, l logger.Interface) useCases {
	userRepo := persistent.NewUserRepo(pg)
	productRepo := persistent.NewProductRepo(pg)

	return useCases{
		user:    user.New(userRepo, jwtManager),
		product: product.New(productRepo, l),
	}
}

func initServers(cfg *config.Config, uc useCases, jwtManager *jwt.Manager, l logger.Interface) servers {
	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, uc.user, uc.product, jwtManager, l)

	return servers{
		http: httpServer,
	}
}

func (s *servers) startServers() {
	s.http.Start()
}

func (s *servers) waitForShutdown(l logger.Interface) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case sig := <-interrupt:
		l.Info("app - Run - signal: %s", sig.String())
	case err = <-s.http.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	s.shutdownServers(l)
}

func (s *servers) shutdownServers(l logger.Interface) {
	if err := s.http.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// JWT
	jwtManager := jwt.New(cfg.JWT.Secret, cfg.JWT.TokenExpiry)

	uc := initUseCases(pg, jwtManager, l)
	s := initServers(cfg, uc, jwtManager, l)
	s.startServers()
	s.waitForShutdown(l)
}
