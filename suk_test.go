package suk

import (
	"sync"
	"testing"
)

func TestSyncMapStorage(t *testing.T) {
	t.Run("Sync map storage setting 1 key", func(t *testing.T) {
		ss, _ := NewSessionStorage()
		got, _ := ss.Set(10)

		if got == "" {
			t.Error("got empty string for ID")
		}
	})

	t.Run("Sync map storage setting 5 keys", func(t *testing.T) {
		ss, _ := NewSessionStorage()

		for i := range 5 {
			ss.Set(i)
		}

		got := 0
		ss.storage.(*syncMap).Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 5

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})

	t.Run("Sync map storage setting and getting 1 key", func(t *testing.T) {
		ss, _ := NewSessionStorage()
		gotSet, _ := ss.Set(10)

		if gotSet == "" {
			t.Error("got empty string for ID")
		}

		gotGet, err := ss.Get(gotSet)

		if err != nil {
			t.Errorf("got error %s", err.Error())
		}

		if gotGet != 10 {
			t.Errorf("got %d expected %d", gotGet, 10)
		}

		got := 0
		ss.storage.(*syncMap).Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 1

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})

	t.Run("Sync map storage with 10 concurrent sets", func(t *testing.T) {
		ss, _ := NewSessionStorage()

		wg := new(sync.WaitGroup)

		for i := range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ss.Set(i)
			}()
		}

		wg.Wait()

		got := 0
		ss.storage.(*syncMap).Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 10

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})

	t.Run("Sync map storage with 1.000 concurrent sets", func(t *testing.T) {
		ss, _ := NewSessionStorage()

		wg := new(sync.WaitGroup)

		for i := range 1_000 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ss.Set(i)
			}()
		}

		wg.Wait()

		got := 0
		ss.storage.(*syncMap).Range(func(key, value any) bool {
			got++
			return true
		})

		expected := 1_000

		if got != expected {
			t.Errorf("got %d expected %d", got, expected)
		}
	})
}
