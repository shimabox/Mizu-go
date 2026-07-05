package reaction

import "github.com/shimabox/Mizu-go/internal/core"

// fakeParticle は座標と半径が固定された core.Particle の代替であり、この
// パッケージのテストで使用される。Mizu-ts の tests/helpers/FakeParticle に
// 対応する。
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
