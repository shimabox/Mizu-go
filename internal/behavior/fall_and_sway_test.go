package behavior

import (
	"math"
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func TestFallAndSway_FallsBySizeTimes0Point1EachStep(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	fall := NewFallAndSway(bounds, 20, core.NewSeededRandom(42))

	y := 100.0
	for i := 0; i < 10; i++ {
		_, ny := fall.Next(100, y)
		want := y + 20*0.1
		if math.Abs(ny-want) > 1e-10 {
			t.Fatalf("iteration %d: y = %v, want %v", i, ny, want)
		}
		y = ny
	}
}

func TestFallAndSway_SwayIsAtMostOnePerStep(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	fall := NewFallAndSway(bounds, 20, core.NewSeededRandom(42))

	x, y := 100.0, 0.0
	for i := 0; i < 100; i++ {
		nx, ny := fall.Next(x, y)
		if math.Abs(nx-x) > 1 {
			t.Fatalf("iteration %d: sway = %v, want <= 1", i, math.Abs(nx-x))
		}
		x, y = nx, ny
	}
}

func TestFallAndSway_HasLandedAtScreenBottom(t *testing.T) {
	const sh = 600.0
	bounds := core.NewBounds(800, sh)
	fall := NewFallAndSway(bounds, 20, core.NewSeededRandom(42))

	if fall.HasLanded(sh - 1) {
		t.Errorf("HasLanded(sh-1) = true, want false")
	}
	if !fall.HasLanded(sh) {
		t.Errorf("HasLanded(sh) = false, want true")
	}
	if !fall.HasLanded(sh + 1) {
		t.Errorf("HasLanded(sh+1) = false, want true")
	}
}
