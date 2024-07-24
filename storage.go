package sucks

import (
	"errors"
	"math/rand/v2"
	"time"
)

const maxCookieSize = 4096

var (
	byteBuffer = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	ErrNoCookieFound = errors.New("No cookie was found with the given value")
)

// randomID Generates random ID with a length of maxCookieSize.
func randomID() string {
	// TODO: Improve speed and entropy
	b := make([]byte, maxCookieSize)
	for i := range b {
		b[i] = byteBuffer[rand.IntN(len(byteBuffer))]
	}
	return string(b)
}

type SingleUseCookie struct {
	ID         string
	Expiration time.Time
}

type SessionStorage struct {
	Sessions map[string]any
}

func NewSessionStorage() *SessionStorage {
	ss := &SessionStorage{Sessions: make(map[string]any)}
	return ss
}

func (ss *SessionStorage) NewCookie(session any, duration time.Duration) SingleUseCookie {
	id := randomID()
	storageHasKey := false

	for !storageHasKey {
		_, ok := ss.Sessions[id]
		if !ok {
			break
		}

		id = randomID()
	}

	suc := SingleUseCookie{
		ID:         id,
		Expiration: time.Now().Add(duration),
	}

	ss.Sessions[id] = session
	return suc
}

func (ss *SessionStorage) GetCookie(value string) (any, error) {
	session, ok := ss.Sessions[value]
	if !ok {
		return nil, ErrNoCookieFound
	}

	delete(ss.Sessions, value)

	ss.NewCookie(session, 0)
	return session, nil
}

func (ss *SessionStorage) RemoveCookie(value string) {
	delete(ss.Sessions, value)
}
