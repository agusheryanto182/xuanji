package seed

import (
	"context"
	"fmt"

	"github.com/agusheryanto182/ecommerce-monolith/config"
	"github.com/agusheryanto182/ecommerce-monolith/internal/repo/persistent"
	seed "github.com/agusheryanto182/ecommerce-monolith/internal/seed/user"
	"github.com/agusheryanto182/ecommerce-monolith/pkg/logger"
	"github.com/agusheryanto182/ecommerce-monolith/pkg/postgres"
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
