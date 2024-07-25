// Package sucks offers easy server-side session management using single-use
// keys.
//
// You may use an in-memory map (default) or a external cache DB (such as Redis, Valkey,
// Dragonfly, etc.) to hold your sessions. Do note that, when using an in-memory map,
// the session data is lost as soon as the program stops.
package suk

import (
	"database/sql"
	"errors"
	"math/rand/v2"
	"sync"
)

// maxCookieSize is used for reference, as stipulated by RFC 2109, RFC 2965 and
// RFC 6265.
const maxCookieSize = 4096

var (
	// byteBuffer contains all characters used to randomly generate keys.
	byteBuffer = []byte(" !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	// Errors
	ErrNoKeyFound = errors.New("No cookie was found with the given value.")
)

type storage interface {
	set(any) (string, error)
	get(string) (any, error)
	remove(string) error
	clear() error
}

type syncMap struct {
	*sync.Map
}

func (s *syncMap) set(session any) (string, error) {
	id := randomID()

	for {
		_, ok := s.Load(id)
		if !ok {
			break
		}

		id = randomID()
	}

	s.Store(id, session)
	return id, nil
}

func (s *syncMap) get(key string) (any, error) {
	session, loaded := s.LoadAndDelete(key)
	if !loaded {
		return nil, ErrNoKeyFound
	}

	s.set(session)
	return session, nil
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

type cacheDB *sql.DB

type SessionStorage struct {
	config  config
	storage storage
}

// randomID Generates random ID with a length of maxCookieSize.
func randomID() string {
	// TODO: Improve speed and entropy
	b := make([]byte, maxCookieSize)
	for i := range b {
		b[i] = byteBuffer[rand.IntN(len(byteBuffer))]
	}
	return string(b)
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
	if c.cacheDB != nil {
		ss.storage = cacheDB(c.cacheDB)
	} else {
		sm := syncMap{new(sync.Map)}
		ss.storage = &sm
	}

	return &ss, nil
}

func (ss *SessionStorage) Set(session any) (string, error) {
	return ss.storage.set(session)
}

func (ss *SessionStorage) Get(key string) (any, error) {
	return ss.storage.get(key)
}

func (ss *SessionStorage) Remove(key string) error {
	return ss.storage.remove(key)
}

func (ss *SessionStorage) Clear() error {
	return ss.storage.clear()
}
