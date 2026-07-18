package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
)

// Option -.
type Option func(*Postgres)

// MaxPoolSize -.
func MaxPoolSize(size int) Option {
	return func(c *Postgres) {
		c.maxPoolSize = size
	}
}

// ConnAttempts -.
func ConnAttempts(attempts int) Option {
	return func(c *Postgres) {
		c.connAttempts = attempts
	}
}

// ConnTimeout -.
func ConnTimeout(timeout time.Duration) Option {
	return func(c *Postgres) {
		c.connTimeout = timeout
	}
}

// WithTracer -.
func WithTracer(tracer pgx.QueryTracer) Option {
	return func(p *Postgres) {
		p.tracer = tracer
	}
}
