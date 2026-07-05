package core

import "testing"

// fakeParticle は、Particle インターフェースが後続フェーズ
// (internal/particle) で行うのと同じように、普通の struct で実装
// できることを確認するための最小限の代役。
type fakeParticle struct {
	kind Kind
	x, y float64
	dead bool
}

func (p *fakeParticle) Kind() Kind      { return p.kind }
func (p *fakeParticle) X() float64      { return p.x }
func (p *fakeParticle) Y() float64      { return p.y }
func (p *fakeParticle) Radius() float64 { return 1 }
func (p *fakeParticle) Update()         { p.x++ }
func (p *fakeParticle) IsDead() bool    { return p.dead }
func (p *fakeParticle) MarkDead()       { p.dead = true }

var _ Particle = (*fakeParticle)(nil)

func TestParticle_InterfaceIsImplementable(t *testing.T) {
	p := &fakeParticle{kind: Kind("H")}

	if p.Kind() != "H" {
		t.Fatalf("Kind() = %v, want H", p.Kind())
	}
	if p.IsDead() {
		t.Fatal("expected not dead initially")
	}

	p.Update()
	if p.X() != 1 {
		t.Fatalf("X() after Update = %v, want 1", p.X())
	}

	p.MarkDead()
	if !p.IsDead() {
		t.Fatal("expected dead after MarkDead")
	}
}
