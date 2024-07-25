package suk

import (
	"database/sql"
	"errors"
)

var (
	ErrNilCacheDB = errors.New("The given cache DB is nil.")
	ErrCacheDBAlreadySet = errors.New("A cache DB was already registered for this session storage.")
)

type config struct {
	cacheDB *sql.DB
}

type Option interface {
	apply(*config) error
}

type option func(*config) error

func (o option) apply(c *config) error {
	return o(c)
}

// WithCacheDB replaces
func WithCacheDB(db *sql.DB) Option {
	return option(func(c *config) error {
		if c.cacheDB != nil {
			return ErrCacheDBAlreadySet
		}

		if db == nil {
			return ErrNilCacheDB
		}

		c.cacheDB = db
		return nil
	})
}
