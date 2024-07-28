package suk

import (
	"testing"
)

func TestRandomID(t *testing.T) {
	t.Run("Generate random string with length 0", func(t *testing.T) {
		expected := ""
		got, err := randomID(0)
		if err != nil {
			t.Error(err)
		}

		if got != expected {
			t.Errorf("got %q expected %q", got, expected)
		}
	})

	t.Run("Generate random string with length 10", func(t *testing.T) {
		expectedLen := 10
		got, err := randomID(uint64(expectedLen))
		if err != nil {
			t.Error(err)
		}

		if len(got) != expectedLen {
			t.Errorf("got %d expected %d", len(got), expectedLen)
		}
	})
}

func BenchmarkRandomIDWithIncreasingLength(b *testing.B) {
	for i := range b.N {
		randomID(uint64(i))
	}
}

func BenchmarkRandomIDLength20(b *testing.B) {
	for range b.N {
		randomID(20) // Same size as SHA-1 output
	}
}

func BenchmarkRandomIDLength32(b *testing.B) {
	for range b.N {
		randomID(32) // Same size as SHA-256 output
	}
}

func BenchmarkRandomIDLength64(b *testing.B) {
	for range b.N {
		randomID(64) // Same size as SHA-512 output
	}
}
