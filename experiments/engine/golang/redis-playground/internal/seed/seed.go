package seed

import (
	"context"
	"fmt"

	"github.com/agusheryanto182/redis-playground/config"
	"github.com/agusheryanto182/redis-playground/internal/repo/persistent"
	seed "github.com/agusheryanto182/redis-playground/internal/seed/user"
	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/agusheryanto182/redis-playground/pkg/postgres"
)

type Seeder interface {
	Seed(ctx context.Context) error
}

func Seeding(ctx context.Context, seeders ...Seeder) error {

	for _, s := range seeders {
		if err := s.Seed(ctx); err != nil {
			return err
		}
	}

	return nil
}

func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("seeder - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	userRepo := persistent.NewUserRepo(pg)

	err = Seeding(
		context.Background(),
		seed.NewUserSeeder(userRepo),
	)
	if err != nil {
		l.Fatal(fmt.Errorf("seeder - Run - seeder.Seeding: %w", err))
	}

	l.Info("Seeding Completed")
}
