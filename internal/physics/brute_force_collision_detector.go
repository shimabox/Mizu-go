package physics

import (
	"math"

	"github.com/shimabox/Mizu-go/internal/core"
)

// BruteForceCollisionDetector は O(n^2) の衝突検出器。実行時には
// GridCollisionDetector に置き換わった後も、property テストの基準
// (reference)実装として残してある。
type BruteForceCollisionDetector struct{}

// NewBruteForceCollisionDetector は BruteForceCollisionDetector を返す。
func NewBruteForceCollisionDetector() *BruteForceCollisionDetector {
	return &BruteForceCollisionDetector{}
}

// FindHitPairs は CollisionDetector を実装する。
func (d *BruteForceCollisionDetector) FindHitPairs(particles []core.Particle) []Pair {
	var pairs []Pair
	for i := 0; i < len(particles); i++ {
		for j := i + 1; j < len(particles); j++ {
			if bruteForceIsHit(particles[i], particles[j]) {
				pairs = append(pairs, Pair{A: particles[i], B: particles[j]})
			}
		}
	}
	return pairs
}

func bruteForceIsHit(a, b core.Particle) bool {
	dx := b.X() - a.X()
	dy := b.Y() - a.Y()
	distance := math.Sqrt(dx*dx + dy*dy) // 2点間の距離(三平方の定理)
	hitDistance := a.Radius() + b.Radius()

	return distance < hitDistance
}
