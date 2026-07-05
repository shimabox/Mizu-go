package physics

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func TestBruteForceCollisionDetector_DetectsTouchingPair(t *testing.T) {
	detector := NewBruteForceCollisionDetector()
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 110, 100, 10) // 距離10 < 20

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 1 || pairs[0].A != core.Particle(a) || pairs[0].B != core.Particle(b) {
		t.Fatalf("got %v, want [{a b}]", pairs)
	}
}

func TestBruteForceCollisionDetector_ExactSumOfRadiiIsNotAHit(t *testing.T) {
	detector := NewBruteForceCollisionDetector()
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 120, 100, 10) // 距離20 == 半径の合計

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 0 {
		t.Fatalf("got %v, want no pairs", pairs)
	}
}

func TestBruteForceCollisionDetector_DoesNotDetectFarApartPair(t *testing.T) {
	detector := NewBruteForceCollisionDetector()
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 300, 300, 10)

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 0 {
		t.Fatalf("got %v, want no pairs", pairs)
	}
}

func TestBruteForceCollisionDetector_NoSelfOrDuplicatePairs(t *testing.T) {
	detector := NewBruteForceCollisionDetector()
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 105, 100, 10)
	c := newFakeParticle("H", 110, 100, 10)

	all := []core.Particle{a, b, c}
	pairs := detector.FindHitPairs(all)

	// 3つとも互いに重なり合っている: ペアはちょうど3組。
	if len(pairs) != 3 {
		t.Fatalf("got %d pairs, want 3: %v", len(pairs), pairs)
	}
	seen := map[string]bool{}
	for _, pr := range pairs {
		if pr.A == pr.B {
			t.Fatalf("pair contains a particle paired with itself: %v", pr)
		}
		key := normalizePairKey(all, pr)
		seen[key] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected 3 distinct pairs, got %d", len(seen))
	}
}

func TestBruteForceCollisionDetector_EnumeratesPairsRegardlessOfKind(t *testing.T) {
	detector := NewBruteForceCollisionDetector()
	o := newFakeParticle("O", 100, 100, 10)
	h2o := newFakeParticle("H2o", 105, 100, 10)

	pairs := detector.FindHitPairs([]core.Particle{o, h2o})
	if len(pairs) != 1 || pairs[0].A != core.Particle(o) || pairs[0].B != core.Particle(h2o) {
		t.Fatalf("got %v, want [{o h2o}]", pairs)
	}
}
