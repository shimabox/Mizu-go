package particle

import (
	"image/color"
	"testing"

	"github.com/shimabox/Mizu-go/internal/behavior"
	"github.com/shimabox/Mizu-go/internal/core"
)

// stepMovement は atom.Update を RandomWalk のランダム性に依存せず単体で
// テストするための、テスト用の MovementBehavior。
type stepMovement struct {
	dx, dy float64
}

func (m *stepMovement) Next(x, y float64) (float64, float64) {
	return x + m.dx, y + m.dy
}

var _ behavior.MovementBehavior = (*stepMovement)(nil)

func TestAtom_ImplementsParticle(t *testing.T) {
	var _ core.Particle = newAtom(KindH, 0, 0, 1, &stepMovement{}, color.RGBA{})
}

func TestAtom_KindConstants(t *testing.T) {
	cases := []struct {
		name string
		kind core.Kind
		want string
	}{
		{"H", KindH, "H"},
		{"H2", KindH2, "H2"},
		{"O", KindO, "O"},
		{"H2o", KindH2o, "H2o"},
	}
	for _, c := range cases {
		if string(c.kind) != c.want {
			t.Errorf("%s: kind = %q, want %q", c.name, c.kind, c.want)
		}
	}
}

func TestAtom_InitialState(t *testing.T) {
	col := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	a := newAtom(KindO, 100, 200, 12.5, &stepMovement{}, col)

	if a.Kind() != KindO {
		t.Errorf("Kind() = %v, want %v", a.Kind(), KindO)
	}
	if a.X() != 100 {
		t.Errorf("X() = %v, want 100", a.X())
	}
	if a.Y() != 200 {
		t.Errorf("Y() = %v, want 200", a.Y())
	}
	if a.Radius() != 12.5 {
		t.Errorf("Radius() = %v, want 12.5", a.Radius())
	}
	if a.Color() != col {
		t.Errorf("Color() = %v, want %v", a.Color(), col)
	}
	if a.IsDead() {
		t.Error("expected not dead initially")
	}
}

func TestAtom_UpdateAdvancesPosition(t *testing.T) {
	a := newAtom(KindH, 5, 5, 1, &stepMovement{dx: 3, dy: -2}, color.RGBA{})
	a.Update()
	if a.X() != 8 || a.Y() != 3 {
		t.Errorf("after Update: X=%v Y=%v, want X=8 Y=3", a.X(), a.Y())
	}
}

func TestAtom_MarkDead(t *testing.T) {
	a := newAtom(KindH2, 0, 0, 1, &stepMovement{}, color.RGBA{})
	if a.IsDead() {
		t.Fatal("expected not dead initially")
	}
	a.MarkDead()
	if !a.IsDead() {
		t.Error("expected dead after MarkDead")
	}
}

func TestAtom_UpdateWithRandomWalk(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	movement := behavior.NewRandomWalk(bounds, 20, random)
	a := newAtom(KindH, 100, 100, 10, movement, color.RGBA{})

	a.Update()
	if a.X() == 100 && a.Y() == 100 {
		t.Error("expected position to change after Update with RandomWalk")
	}
}
