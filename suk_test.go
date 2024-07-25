package suk2

import (
	"sync"
	"testing"
	"time"
)

func TestSessionStorage(t *testing.T) {
	t.Run("Session storage setting 1 key", func(t *testing.T) {
		ss := NewSessionStorage()
		got := ss.Set(10, -1*time.Second)

		if got.ID == "" {
			t.Error("got empty string for ID")
		}

		if time.Until(got.Expiration) > 0 {
			t.Error("got positive expiration time for Expiration")
		}
	})

	t.Run("Session storage setting 5 keys", func(t *testing.T) {
		ss := NewSessionStorage()

		for i := range 5 {
			ss.Set(i, -1*time.Second)
		}

		got := 0
		ss.sessions.Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 5

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})

	t.Run("Session storage setting and getting 1 key", func(t *testing.T) {
		ss := NewSessionStorage()
		gotSet := ss.Set(10, -1*time.Second)

		if gotSet.ID == "" {
			t.Error("got empty string for ID")
		}

		gotGet, err := ss.Get(gotSet.ID)

		if err != nil {
			t.Errorf("got error %s", err.Error())
		}

		if gotGet != 10 {
			t.Errorf("got %d expected %d", gotGet, 10)
		}

		got := 0
		ss.sessions.Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 1

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})

	t.Run("Session storage with 10 concurrent sets", func(t *testing.T) {
		ss := NewSessionStorage()

		wg := new(sync.WaitGroup)

		for i := range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ss.Set(i, -1*time.Second)
			}()
		}

		wg.Wait()

		got := 0
		ss.sessions.Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 10

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})

	t.Run("Session storage with 1.000 concurrent sets", func(t *testing.T) {
		ss := NewSessionStorage()

		wg := new(sync.WaitGroup)

		for i := range 1_000 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ss.Set(i, -1*time.Second)
			}()
		}

		wg.Wait()

		got := 0
		ss.sessions.Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 1_000

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})
}
