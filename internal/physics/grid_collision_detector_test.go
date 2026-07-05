package physics

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func newGridDetectorForTest(width, height float64) *GridCollisionDetector {
	return NewGridCollisionDetector(NewSpatialGrid(width, height, DefaultCellSize))
}

func TestGridCollisionDetector_DetectsTouchingPair(t *testing.T) {
	detector := newGridDetectorForTest(800, 600)
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 110, 100, 10) // 距離10 < 20

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 1 || pairs[0].A != core.Particle(a) || pairs[0].B != core.Particle(b) {
		t.Fatalf("got %v, want [{a b}]", pairs)
	}
}

func TestGridCollisionDetector_ExactSumOfRadiiIsNotAHit(t *testing.T) {
	detector := newGridDetectorForTest(800, 600)
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 120, 100, 10) // 距離20 == 半径の合計

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 0 {
		t.Fatalf("got %v, want no pairs", pairs)
	}
}

func TestGridCollisionDetector_DoesNotDetectFarApartPair(t *testing.T) {
	detector := newGridDetectorForTest(800, 600)
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 300, 300, 10)

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 0 {
		t.Fatalf("got %v, want no pairs", pairs)
	}
}

func TestGridCollisionDetector_NoSelfOrDuplicatePairs(t *testing.T) {
	detector := newGridDetectorForTest(800, 600)
	a := newFakeParticle("H", 100, 100, 10)
	b := newFakeParticle("H", 105, 100, 10)
	c := newFakeParticle("H", 110, 100, 10)

	all := []core.Particle{a, b, c}
	pairs := detector.FindHitPairs(all)

	if len(pairs) != 3 {
		t.Fatalf("got %d pairs, want 3: %v", len(pairs), pairs)
	}
	seen := map[string]bool{}
	for _, pr := range pairs {
		if pr.A == pr.B {
			t.Fatalf("pair contains a particle paired with itself: %v", pr)
		}
		seen[normalizePairKey(all, pr)] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected 3 distinct pairs, got %d", len(seen))
	}
}

func TestGridCollisionDetector_EnumeratesPairsRegardlessOfKind(t *testing.T) {
	detector := newGridDetectorForTest(800, 600)
	o := newFakeParticle("O", 100, 100, 10)
	h2o := newFakeParticle("H2o", 105, 100, 10)

	pairs := detector.FindHitPairs([]core.Particle{o, h2o})
	if len(pairs) != 1 || pairs[0].A != core.Particle(o) || pairs[0].B != core.Particle(h2o) {
		t.Fatalf("got %v, want [{o h2o}]", pairs)
	}
}

func TestGridCollisionDetector_DetectsCollisionAcrossCellBoundary(t *testing.T) {
	cellSize := float64(DefaultCellSize)
	detector := newGridDetectorForTest(cellSize*4, cellSize*4)
	boundaryX := cellSize * 2
	a := newFakeParticle("H", boundaryX-1, boundaryX, 5) // 左側のセル
	b := newFakeParticle("H", boundaryX+1, boundaryX, 5) // 右側のセル(距離2 < 合計10)

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 1 || pairs[0].A != core.Particle(a) || pairs[0].B != core.Particle(b) {
		t.Fatalf("got %v, want [{a b}]", pairs)
	}
}

func TestGridCollisionDetector_DetectsDiagonalCollisionAcrossCellCorner(t *testing.T) {
	cellSize := float64(DefaultCellSize)
	detector := newGridDetectorForTest(cellSize*4, cellSize*4)
	boundary := cellSize * 2
	a := newFakeParticle("H", boundary-1, boundary-1, 5) // 左上のセル
	b := newFakeParticle("H", boundary+1, boundary+1, 5) // 右下のセル(対角の隣接セル)

	pairs := detector.FindHitPairs([]core.Particle{a, b})
	if len(pairs) != 1 || pairs[0].A != core.Particle(a) || pairs[0].B != core.Particle(b) {
		t.Fatalf("got %v, want [{a b}]", pairs)
	}
}
