// Package suk offers easy server-side session management using single-use
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
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// maxCookieSize is used for reference, as stipulated by RFC 2109, RFC 2965
	// and RFC 6265.
	maxCookieSize = 4096
)

var (
	defaultDurationToExpire = 10 * time.Minute

	ErrKeyWasExpired = errors.New("The given key has expired.")
	ErrNoKeyFound    = errors.New("No value was found with the given key.")
	ErrNilSession    = errors.New("The session passed can't be nil.")
)

type storage interface {
	set(any) (string, error)
	get(string) (any, string, error)
	remove(string) error
	clearExpired() error
}

type value struct {
	data       any
	expiration time.Time
}

type syncMap struct {
	*sync.Map

	keyLength        uint64
	durationToExpire time.Duration
}

func (s *syncMap) set(session any) (string, error) {
	if session == nil {
		return "", ErrNilSession
	}

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

	v := value{data: session, expiration: time.Now().Add(s.durationToExpire)}
	s.Store(id, v)
	return id, nil
}

func (s *syncMap) get(key string) (any, string, error) {
	session, loaded := s.LoadAndDelete(key)
	if !loaded {
		return nil, "", ErrNoKeyFound
	}

	v := session.(value)
	if time.Until(v.expiration) <= 0 {
		return nil, "", ErrKeyWasExpired
	}

	newKey, err := s.set(session)
	if err != nil {
		return nil, "", err
	}

	return v.data, newKey, nil
}

func (s *syncMap) remove(key string) error {
	s.Delete(key)
	return nil
}

func (s *syncMap) clearExpired() error {
	s.Range(func(k, v any) bool {
		vl := v.(value)
		if time.Until(vl.expiration) <= 0 {
			s.Delete(k)
		}
		return true
	})
	return nil
}

type redisDB struct {
	*redis.Client

	ctx              context.Context
	keyLength        uint64
	durationToExpire time.Duration
}

func (r *redisDB) set(session any) (string, error) {
	if session == nil {
		return "", ErrNilSession
	}

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

	err = r.Set(r.ctx, id, session, r.durationToExpire).Err()
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

func (r *redisDB) clearExpired() error {
	return nil
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

	var durationToExpire time.Duration
	if c.customTokenDuration != nil {
		durationToExpire = *c.customTokenDuration
	} else {
		durationToExpire = defaultDurationToExpire
	}

	if c.redisClient != nil {
		cd := redisDB{new(redis.Client), c.redisCtx, keyLength, durationToExpire}
		ss.storage = &cd

		return &ss, nil
	}

	sm := syncMap{new(sync.Map), keyLength, durationToExpire}
	ss.storage = &sm

	var autoClearExpiredKeys bool
	if c.autoExpiredClear != nil {
		autoClearExpiredKeys = *c.autoExpiredClear
	}

	if autoClearExpiredKeys {
		ticker := time.NewTicker(durationToExpire)

		go func() {
			for {
				select {
				case <-ticker.C:
					ss.ClearExpired()
				}
			}
		}()
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

func (ss *SessionStorage) ClearExpired() error {
	return ss.storage.clearExpired()
}
