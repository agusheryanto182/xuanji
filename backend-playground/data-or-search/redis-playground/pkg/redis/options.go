package redis

import "github.com/redis/go-redis/v9"

func Addr(addr string) Option {
	return func(o *redis.Options) {
		o.Addr = addr
	}
}

func PoolSize(size int) Option {
	return func(o *redis.Options) {
		o.PoolSize = size
	}
}

func MinIdleConns(n int) Option {
	return func(o *redis.Options) {
		o.MinIdleConns = n
	}
}
