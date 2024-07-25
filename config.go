package suk

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNilRedisClient        = errors.New("The given Redis client is nil.")
	ErrRedisClientAlreadySet = errors.New("A Redis client was already registered for this session storage.")

	ErrCustomKeyLengthAlreadySet = errors.New("A custom key length was already registered for this session storage.")
)

type config struct {
	customKeyLength *uint64
	redisCtx        context.Context
	redisClient     *redis.Client
}

type Option interface {
	apply(*config) error
}

type option func(*config) error

func (o option) apply(c *config) error {
	return o(c)
}

// WithRedis uses the given Redis client to store the sessions, instead of using
// an in-memory storage.
func WithRedis(client *redis.Client, ctx context.Context) Option {
	return option(func(c *config) error {
		if c.redisClient != nil || c.redisCtx != nil {
			return ErrRedisClientAlreadySet
		}

		if client == nil {
			return ErrNilRedisClient
		}

		if ctx == nil {
			c.redisCtx = context.Background()
		} else {
			c.redisCtx = ctx
		}

		c.redisClient = client
		return nil
	})
}

// WithCustomKeyLength sets a custom key length for generated keys.
func WithCustomKeyLength(keyLength uint64) Option {
	return option(func(c *config) error {
		if c.customKeyLength != nil {
			return ErrCustomKeyLengthAlreadySet
		}

		c.customKeyLength = &keyLength
		return nil
	})
}
