package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*redis.Client
}

type Option func(*redis.Options)

func New(url string, opts ...Option) (*redis.Client, error) {
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(options)
	}

	client := redis.NewClient(options)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
