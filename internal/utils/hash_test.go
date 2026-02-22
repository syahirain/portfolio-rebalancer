package utils

import (
	"testing"
)

func TestCanonicalHash(t *testing.T) {
	t.Run("Deterministic output", func(t *testing.T) {
		m1 := map[string]float64{
			"stocks": 60,
			"bonds":  30,
			"gold":   10,
		}

		m2 := map[string]float64{
			"gold":   10,
			"stocks": 60,
			"bonds":  30,
		}

		h1 := CanonicalHash(m1)
		h2 := CanonicalHash(m2)

		if h1 != h2 {
			t.Errorf("expected hashes to be equal, got %s and %s", h1, h2)
		}
	})

	t.Run("Different inputs produce different hashes", func(t *testing.T) {
		m1 := map[string]float64{
			"stocks": 60,
			"bonds":  30,
			"gold":   10,
		}

		m2 := map[string]float64{
			"stocks": 70,
			"bonds":  20,
			"gold":   10,
		}

		h1 := CanonicalHash(m1)
		h2 := CanonicalHash(m2)

		if h1 == h2 {
			t.Errorf("expected hashes to be different, got %s", h1)
		}
	})

	t.Run("Empty map", func(t *testing.T) {
		m1 := map[string]float64{}
		h1 := CanonicalHash(m1)

		if len(h1) == 0 {
			t.Error("expected non-empty hash for empty map")
		}
	})
}
