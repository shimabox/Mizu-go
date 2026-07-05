package sim

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

// fakeParticle は位置・半径が固定された、テスト専用のスタブであり、
// Mizu-ts の tests/helpers/FakeParticle.ts に対応する。
type fakeParticle struct {
	kind   core.Kind
	x, y   float64
	radius float64
	dead   bool
}

func newFakeParticle(kind core.Kind, x, y, radius float64) *fakeParticle {
	return &fakeParticle{kind: kind, x: x, y: y, radius: radius}
}

func (p *fakeParticle) Kind() core.Kind { return p.kind }
func (p *fakeParticle) X() float64      { return p.x }
func (p *fakeParticle) Y() float64      { return p.y }
func (p *fakeParticle) Radius() float64 { return p.radius }
func (p *fakeParticle) Update()         {}
func (p *fakeParticle) IsDead() bool    { return p.dead }
func (p *fakeParticle) MarkDead()       { p.dead = true }

var _ core.Particle = (*fakeParticle)(nil)

// TestWorld_AddCountAll は World.test.ts の "add / count / all が機能する
// こと" に対応する。
func TestWorld_AddCountAll(t *testing.T) {
	w := NewWorld()
	w.Add(newFakeParticle("H", 0, 0, 1))
	w.Add(newFakeParticle("H", 10, 10, 1))
	w.Add(newFakeParticle("O", 20, 20, 1))

	if got := w.Count("H"); got != 2 {
		t.Errorf("Count(H) = %d, want 2", got)
	}
	if got := w.Count("O"); got != 1 {
		t.Errorf("Count(O) = %d, want 1", got)
	}
	if got := w.Count("H2"); got != 0 {
		t.Errorf("Count(H2) = %d, want 0", got)
	}
	if got := len(w.All()); got != 3 {
		t.Errorf("len(All()) = %d, want 3", got)
	}
}

// TestWorld_SweepRemovesDead は "sweep 後に dead な粒子が残らないこと" に
// 対応する。
func TestWorld_SweepRemovesDead(t *testing.T) {
	w := NewWorld()
	p1 := newFakeParticle("H", 0, 0, 1)
	p2 := newFakeParticle("H", 10, 10, 1)
	p3 := newFakeParticle("H", 20, 20, 1)
	w.Add(p1)
	w.Add(p2)
	w.Add(p3)

	p2.MarkDead()
	w.Sweep()

	if got := w.Count("H"); got != 2 {
		t.Errorf("Count(H) after sweep = %d, want 2", got)
	}
	for _, p := range w.All() {
		if p.IsDead() {
			t.Errorf("All() contains a dead particle after Sweep")
		}
	}
}

// TestWorld_SweepPreservesOrder は "sweep 後も生存粒子の順序が保たれること"
// に対応する。
func TestWorld_SweepPreservesOrder(t *testing.T) {
	w := NewWorld()
	p1 := newFakeParticle("H", 0, 0, 1)
	p2 := newFakeParticle("H2", 10, 10, 1)
	p3 := newFakeParticle("O", 20, 20, 1)
	w.Add(p1)
	w.Add(p2)
	w.Add(p3)

	p2.MarkDead()
	w.Sweep()

	all := w.All()
	if len(all) != 2 || all[0] != core.Particle(p1) || all[1] != core.Particle(p3) {
		t.Errorf("All() after sweep = %v, want [p1, p3]", all)
	}
}

// TestWorld_EmptyWorld は "空の World でも count / sweep がエラーにならない
// こと" に対応する。
func TestWorld_EmptyWorld(t *testing.T) {
	w := NewWorld()
	if got := w.Count("H"); got != 0 {
		t.Errorf("Count(H) on empty world = %d, want 0", got)
	}

	// 空の world に対して Sweep が panic してはならない。
	w.Sweep()

	if got := len(w.All()); got != 0 {
		t.Errorf("len(All()) on empty world = %d, want 0", got)
	}
}
