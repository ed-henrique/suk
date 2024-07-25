// Package sucks offers easy server-side session management using single-use
// keys.
package suk2

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

// maxCookieSize is used for reference, as stipulated by RFC 2109, RFC 2965 and
// RFC 6265.
const maxCookieSize = 4096

var (
	// byteBuffer contains all characters used to randomly generate keys.
	byteBuffer = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	// Errors
	ErrNoKeyFound = errors.New("No cookie was found with the given value")
)

type SingleUseKey struct {
	ID         string
	Expiration time.Time
}

type SessionStorage struct {
	sessions *sync.Map
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

func NewSessionStorage() *SessionStorage {
	ss := &SessionStorage{
		sessions: new(sync.Map),
	}
	return ss
}

func (ss *SessionStorage) Set(session any, duration time.Duration) SingleUseKey {
	id := randomID()

	for {
		_, ok := ss.sessions.Load(id)
		if !ok {
			break
		}

		id = randomID()
	}

	suc := SingleUseKey{
		ID:         id,
		Expiration: time.Now().Add(duration),
	}

	ss.sessions.Store(id, session)
	return suc
}

func (ss *SessionStorage) Get(key string) (any, error) {
	session, loaded := ss.sessions.LoadAndDelete(key)
	if !loaded {
		return nil, ErrNoKeyFound
	}

	ss.Set(session, 0)
	return session, nil
}

func (ss *SessionStorage) Remove(key string) {
	ss.sessions.Delete(key)
}
