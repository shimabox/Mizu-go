package particle

import (
	"image/color"
	"testing"

	"github.com/shimabox/Mizu-go/internal/behavior"
	"github.com/shimabox/Mizu-go/internal/core"
)

func TestH2o_ImplementsParticle(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	var _ core.Particle = newH2o(100, 100, 10, movement)
}

func TestH2o_InitialState(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	p := newH2o(100, 200, 10, movement)

	if p.Kind() != KindH2o {
		t.Errorf("Kind() = %v, want %v", p.Kind(), KindH2o)
	}
	if p.X() != 100 || p.Y() != 200 {
		t.Errorf("X,Y = %v,%v want 100,200", p.X(), p.Y())
	}
	if p.Radius() != 10 {
		t.Errorf("Radius() = %v, want 10", p.Radius())
	}
	if p.IsDead() {
		t.Error("expected not dead initially")
	}
}

func TestH2o_Color(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	p := newH2o(0, 0, 10, movement)

	want := color.RGBA{R: 0x00, G: 0x7f, B: 0xff, A: 0xff}
	if p.Color() != want {
		t.Errorf("Color() = %v, want %v", p.Color(), want)
	}
}

func TestH2o_UpdateFalls(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	p := newH2o(100, 100, 10, movement)

	y := p.Y()
	p.Update()
	if p.Y() <= y {
		t.Errorf("expected y to increase after Update, before=%v after=%v", y, p.Y())
	}
}

func TestH2o_LandsAndDies(t *testing.T) {
	bounds := core.NewBounds(800, 50) // small height so landing triggers immediately
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	p := newH2o(100, bounds.H(), 10, movement)

	if p.IsDead() {
		t.Fatal("expected not dead before update")
	}
	p.Update()
	if !p.IsDead() {
		t.Error("expected dead after reaching bottom of screen")
	}
}

func TestH2o_NotLandedBeforeReachingBottom(t *testing.T) {
	bounds := core.NewBounds(800, 10000) // tall screen: far from landing
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	p := newH2o(100, 0, 10, movement)

	p.Update()
	if p.IsDead() {
		t.Error("expected not dead when far from the bottom of the screen")
	}
}

func TestH2o_MarkDead(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	movement := behavior.NewFallAndSway(bounds, 20, random)
	p := newH2o(0, 0, 10, movement)
	p.MarkDead()
	if !p.IsDead() {
		t.Error("expected dead after MarkDead")
	}
}
