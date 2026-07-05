package core

import "testing"

func TestSeededRandom_Deterministic(t *testing.T) {
	a := NewSeededRandom(42)
	b := NewSeededRandom(42)
	for i := 0; i < 50; i++ {
		va, vb := a.Next(), b.Next()
		if va != vb {
			t.Fatalf("iteration %d: got %v and %v, want equal", i, va, vb)
		}
	}
}

func TestSeededRandom_DifferentSeedsDifferentSequences(t *testing.T) {
	a := NewSeededRandom(1)
	b := NewSeededRandom(2)

	same := true
	for i := 0; i < 10; i++ {
		if a.Next() != b.Next() {
			same = false
			break
		}
	}
	if same {
		t.Fatal("expected different sequences for different seeds, got identical sequences")
	}
}

func TestSeededRandom_RangeIsZeroToOneExclusive(t *testing.T) {
	r := NewSeededRandom(1)
	for i := 0; i < 100; i++ {
		v := r.Next()
		if v < 0 || v >= 1 {
			t.Fatalf("iteration %d: got %v, want value in [0, 1)", i, v)
		}
	}
}

func TestRandom_RangeIsZeroToOneExclusive(t *testing.T) {
	r := NewRandom()
	for i := 0; i < 100; i++ {
		v := r.Next()
		if v < 0 || v >= 1 {
			t.Fatalf("iteration %d: got %v, want value in [0, 1)", i, v)
		}
	}
}
