package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])

if current == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end

return current
`)

type Limiter struct {
	client *redis.Client
}

func New(client *redis.Client) *Limiter {
	return &Limiter{
		client: client,
	}
}

func (l *Limiter) Allow(
	ctx context.Context,
	cfg Config,
	key string,
) (bool, int64, error) {

	cfg.setDefaults()

	redisKey := fmt.Sprintf(
		"%s:%s:%s",
		cfg.Prefix,
		cfg.Namespace,
		key,
	)

	count, err := rateLimitScript.Run(
		ctx,
		l.client,
		[]string{redisKey},
		int(cfg.Window/time.Second),
	).Int64()
	if err != nil {
		return false, 0, err
	}

	remaining := cfg.MaxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return count <= cfg.MaxRequests, remaining, nil
}
