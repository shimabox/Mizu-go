package physics

import "github.com/shimabox/Mizu-go/internal/core"

// Pair は接触している粒子のペア。
type Pair struct {
	A, B core.Particle
}

// CollisionDetector は接触している粒子のペアを列挙する。実装は、ある
// 粒子を自分自身とペアにして報告してはならず、同じペアについて
// (a, b) と (b, a) の両方を報告してもならない。
type CollisionDetector interface {
	FindHitPairs(particles []core.Particle) []Pair
}
