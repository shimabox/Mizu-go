package physics

import (
	"fmt"
	"sort"

	"github.com/shimabox/Mizu-go/internal/core"
)

// fakeParticle は、このパッケージのテスト全体で使われる、位置・半径が
// 固定された最小限の代役。TS 版の tests/helpers/FakeParticle.ts に対応
// する。
type fakeParticle struct {
	kind core.Kind
	x, y float64
	r    float64
	dead bool
}

func newFakeParticle(kind core.Kind, x, y, r float64) *fakeParticle {
	return &fakeParticle{kind: kind, x: x, y: y, r: r}
}

func (p *fakeParticle) Kind() core.Kind { return p.kind }
func (p *fakeParticle) X() float64      { return p.x }
func (p *fakeParticle) Y() float64      { return p.y }
func (p *fakeParticle) Radius() float64 { return p.r }
func (p *fakeParticle) Update()         {}
func (p *fakeParticle) IsDead() bool    { return p.dead }
func (p *fakeParticle) MarkDead()       { p.dead = true }

var _ core.Particle = (*fakeParticle)(nil)

// createRandomParticles は tests/helpers/randomParticles.ts を移植した
// もの: core.Random を使って、半径が MaxParticleRadius 以下に収まるように
// 粒子を決定的に散らばらせるので、SpatialGrid の cellSize に関する不変
// 条件が常に成り立つ。
func createRandomParticles(count int, random core.Random, width, height float64) []core.Particle {
	particles := make([]core.Particle, 0, count)
	for i := 0; i < count; i++ {
		x := random.Next() * width
		y := random.Next() * height
		r := 4 + random.Next()*(MaxParticleRadius-4)
		particles = append(particles, newFakeParticle(core.Kind("H"), x, y, r))
	}
	return particles
}

// normalizePairKey は Pair を particles 内でのインデックスとして
// (小さい方, 大きい方)の順で表現する。これにより、異なる検出器で
// 見つかったペアを、A/B の順序に関係なく等価比較できるようになる。
func normalizePairKey(particles []core.Particle, pr Pair) string {
	ia, ib := indexOf(particles, pr.A), indexOf(particles, pr.B)
	if ia > ib {
		ia, ib = ib, ia
	}
	return fmt.Sprintf("%d-%d", ia, ib)
}

// normalizePairs は全てのペアを normalizePairKey で表現してから結果を
// ソートすることで、同じ particles スライスに対して実行した2つの
// CollisionDetector 実装同士を直接比較できる値を作る。
func normalizePairs(particles []core.Particle, pairs []Pair) []string {
	keys := make([]string, 0, len(pairs))
	for _, pr := range pairs {
		keys = append(keys, normalizePairKey(particles, pr))
	}
	sort.Strings(keys)
	return keys
}

func indexOf(particles []core.Particle, target core.Particle) int {
	for i, p := range particles {
		if p == target {
			return i
		}
	}
	return -1
}
