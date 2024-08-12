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
	_maxCookieSize = 4096

	// defaultKeyLength gives an entropy of 192 for each key, which should be fine
	// for most applications.
	defaultKeyLength = 32
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
	mu *sync.Mutex

	// stopChannel is only used when WithAutoClearExpiredKeys is set, to finish
	// the underlying go routine that keeps ticking the autoclear.
	stopChannel chan struct{}
}

// New creates a new session storage.
func New(opts ...Option) (*SessionStorage, error) {
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

	ss := SessionStorage{config: c, mu: &sync.Mutex{}}

	var keyLength uint64 = defaultKeyLength
	if c.customKeyLength != nil {
		keyLength = *c.customKeyLength
	}

	var durationToExpire time.Duration
	if c.customKeyDuration != nil {
		durationToExpire = *c.customKeyDuration
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

	if c.autoClearExpiredKeys {
		ss.stopChannel = make(chan struct{})

		go func() {
			ticker := time.NewTicker(durationToExpire)
			defer ticker.Stop()

			for {
				select {
				case <-ss.stopChannel:
					return
				case <-ticker.C:
					ss.ClearExpired()
				}
			}
		}()
	}

	return &ss, nil
}

// Destroy cleans up and removes a session storage.
func Destroy(ss *SessionStorage) {
	if ss.config.autoClearExpiredKeys {
		close(ss.stopChannel)
	}

	ss = nil
}

// Set assigns the session and returns a key for it.
func (ss *SessionStorage) Set(session any) (string, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	key, err := ss.storage.set(session)
	if err != nil {
		return "", err
	}

	return key, nil
}

// Get retrieves the session and generates a new key for it.
func (ss *SessionStorage) Get(key string) (any, string, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	session, newKey, err := ss.storage.get(key)
	if err != nil {
		return struct{}{}, "", err
	}

	return session, newKey, nil
}

// Remove deletes the specified key and its associated value.
func (ss *SessionStorage) Remove(key string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	err := ss.storage.remove(key)
	if err != nil {
		return err
	}

	return nil
}

// ClearExpired removes all expired keys. For Redis, this function is a no-op
// as Redis handles expiration automatically. To enable similar behavior for
// the default syncMap, start the SessionStorage with the
// WithAutoClearExpiredKeys option.
func (ss *SessionStorage) ClearExpired() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	err := ss.storage.clearExpired()
	if err != nil {
		return err
	}
	
	return nil
}
