// Package sucks offers easy server-side session management using single-use
// keys.
//
// You may use an in-memory map (default) or a Redis client to hold your
// sessions. Do note that, when using an in-memory map, the session data is lost
// as soon as the program stops.
package suk

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	// maxCookieSize is used for reference, as stipulated by RFC 2109, RFC 2965
	// and RFC 6265.
	maxCookieSize = 4096
)

var (
	ErrNoKeyFound = errors.New("No value was found with the given key.")
)

type storage interface {
	set(any) (string, error)
	get(string) (any, string, error)
	remove(string) error
	clear() error
}

type syncMap struct {
	*sync.Map

	keyLength uint64
}

func (s *syncMap) set(session any) (string, error) {
	id, err := randomID(s.keyLength)
	if err != nil {
		return "", err
	}

	var ok bool
	for {
		_, ok = s.Load(id)
		if !ok {
			break
		}

		id, err = randomID(s.keyLength)
		if err != nil {
			return "", err
		}
	}

	s.Store(id, session)
	return id, nil
}

func (s *syncMap) get(key string) (any, string, error) {
	session, loaded := s.LoadAndDelete(key)
	if !loaded {
		return nil, "", ErrNoKeyFound
	}

	newKey, err := s.set(session)
	if err != nil {
		return nil, "", err
	}

	return session, newKey, nil
}

func (s *syncMap) remove(key string) error {
	s.Delete(key)
	return nil
}

func (s *syncMap) clear() error {
	s.Range(func(key, value any) bool {
		s.Delete(key)
		return true
	})
	return nil
}

type redisDB struct {
	*redis.Client

	ctx       context.Context
	keyLength uint64
}

func (r *redisDB) set(session any) (string, error) {
	id, err := randomID(r.keyLength)
	if err != nil {
		return "", err
	}

	for {
		_, err = r.Get(r.ctx, id).Result()
		if err == redis.Nil {
			break
		} else if err != nil {
			return "", err
		}

		id, err = randomID(r.keyLength)
		if err != nil {
			return "", err
		}
	}

	err = r.Set(r.ctx, id, session, 0).Err()
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *redisDB) get(key string) (any, string, error) {
	session, err := r.GetDel(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, "", ErrNoKeyFound
	} else if err != nil {
		return nil, "", err
	}

	newKey, err := r.set(session)
	if err != nil {
		return nil, "", err
	}

	return session, newKey, nil
}

func (r *redisDB) remove(key string) error {
	return r.Del(r.ctx, key).Err()
}

func (r *redisDB) clear() error {
	return r.FlushDB(r.ctx).Err()
}

type SessionStorage struct {
	config  config
	storage storage
}

func NewSessionStorage(opts ...Option) (*SessionStorage, error) {
	var c config
	errs := make([]error, 0, len(opts))
	for _, opt := range opts {
		if err := opt.apply(&c); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	ss := SessionStorage{config: c}

	var keyLength uint64 = maxCookieSize
	if c.customKeyLength != nil {
		keyLength = *c.customKeyLength
	}

	if c.redisClient != nil {
		cd := redisDB{new(redis.Client), c.redisCtx, keyLength}
		ss.storage = &cd
	} else {
		sm := syncMap{new(sync.Map), keyLength}
		ss.storage = &sm
	}

	return &ss, nil
}

func (ss *SessionStorage) Set(session any) (string, error) {
	return ss.storage.set(session)
}

func (ss *SessionStorage) Get(key string) (any, string, error) {
	return ss.storage.get(key)
}

func (ss *SessionStorage) Remove(key string) error {
	return ss.storage.remove(key)
}

func (ss *SessionStorage) Clear() error {
	return ss.storage.clear()
}
