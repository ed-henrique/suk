package suk

import (
	"testing"
)

func TestRandomID(t *testing.T) {
	t.Run("Generate random string with length 1_000_000", func(t *testing.T) {
		expectedLen := 1_000_000
		got, err := randomID(uint64(expectedLen))
		if err != nil {
			t.Error(err)
		}

		if len(got) != expectedLen {
			t.Errorf("got %d expected %d", len(got), expectedLen)
		}
	})

	t.Run("1.000 concurrent random strings with length 16", func(t *testing.T) {
		var idLen uint64 = 16
		concurrentStrings := 1_000

		ids := make([]string, 0, concurrentStrings)
		idChannel := make(chan string)

		for range concurrentStrings {
			go func() {
				id, err := randomID(idLen)
				if err != nil {
					t.Error(err)
				}

				idChannel <- id
			}()
		}

		for range concurrentStrings {
			ids = append(ids, <-idChannel)
		}

		if len(ids) != int(concurrentStrings) {
			t.Errorf("got %d expected %d", len(ids), concurrentStrings)
		}
	})

	t.Run("10.000 concurrent random strings with length 16", func(t *testing.T) {
		var idLen uint64 = 16
		concurrentStrings := 10_000

		ids := make([]string, 0, concurrentStrings)
		idChannel := make(chan string)

		for range concurrentStrings {
			go func() {
				id, err := randomID(idLen)
				if err != nil {
					t.Error(err)
				}

				idChannel <- id
			}()
		}

		for range concurrentStrings {
			ids = append(ids, <-idChannel)
		}

		if len(ids) != int(concurrentStrings) {
			t.Errorf("got %d expected %d", len(ids), concurrentStrings)
		}
	})

	t.Run("100.000 concurrent random strings with length 16", func(t *testing.T) {
		var idLen uint64 = 16
		concurrentStrings := 100_000

		ids := make([]string, 0, concurrentStrings)
		idChannel := make(chan string)

		for range concurrentStrings {
			go func() {
				id, err := randomID(idLen)
				if err != nil {
					t.Error(err)
				}

				idChannel <- id
			}()
		}

		for range concurrentStrings {
			ids = append(ids, <-idChannel)
		}

		if len(ids) != int(concurrentStrings) {
			t.Errorf("got %d expected %d", len(ids), concurrentStrings)
		}
	})
}

func BenchmarkRandomID(b *testing.B) {
	for i := range b.N {
		randomID(uint64(i * 10))
	}
}
