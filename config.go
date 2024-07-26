package suk

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// WithRedis Errors

	ErrNilRedisClient        = errors.New("The given Redis client is nil.")
	ErrRedisClientAlreadySet = errors.New("A Redis client was already registered for this session storage.")

	// WithKeyLength Errors

	ErrZeroKeyLength = errors.New("The given key length must be at least 1.")

	// WithKeyDuration Errors

	ErrNonPositiveKeyDuration = errors.New("The given key duration must be positive.")

	// Option Already Set Errors

	ErrCustomKeyLengthAlreadySet   = errors.New("A custom key length was already registered for this session storage.")
	ErrCustomKeyDurationAlreadySet = errors.New("A custom key duration was already registered for this session storage.")
	ErrAutoExpiredClearAlreadySet  = errors.New("Auto clear for expired keys was already set for this session storage.")
)

type config struct {
	autoExpiredClear  bool
	customKeyLength   *uint64
	customKeyDuration *time.Duration
	redisCtx          context.Context
	redisClient       *redis.Client
}

type Option interface {
	apply(*config) error
}

type option func(*config) error

func (o option) apply(c *config) error {
	return o(c)
}

// WithRedis uses the given Redis client to store the sessions, instead of using
// an in-memory storage. It also may receive a custom context to work on, but by
// default it uses context.Background().
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

// WithKeyLength sets a custom key length for generated keys. The default
// is 32, which gives an entropy of 192 for each key, which should be fine for
// most applications.
//
// Be aware that low length keys are susceptible to guessing, as the entropy
// level goes down. To calculate entropy, do keyLength * log2(65).
func WithKeyLength(keyLength uint64) Option {
	return option(func(c *config) error {
		if c.customKeyLength != nil {
			return ErrCustomKeyLengthAlreadySet
		}

		if keyLength == 0 {
			return ErrZeroKeyLength
		}

		c.customKeyLength = &keyLength
		return nil
	})
}

// WithKeyDuration sets a custom duration for generated keys. The default is
// 10 minutes.
func WithKeyDuration(duration time.Duration) Option {
	return option(func(c *config) error {
		if c.customKeyDuration != nil {
			return ErrCustomKeyDurationAlreadySet
		}

		if duration <= 0 {
			return ErrNonPositiveKeyDuration
		}

		c.customKeyDuration = &duration
		return nil
	})
}

// WithAutoClearExpiredKeys automatically clears expired keys at intervals
// based on the set key expiration time. By default, the clearing process occurs
// every 10 minutes, but this can be adjusted by setting a different key
// expiration time using WithTokenDuration.
func WithAutoClearExpiredKeys() Option {
	return option(func(c *config) error {
		c.autoExpiredClear = true
		return nil
	})
}
