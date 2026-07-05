package reaction

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func TestHHFusion_PairIsHH(t *testing.T) {
	rule := &HHFusion{}

	a, b := rule.Pair()
	if a != core.Kind("H") || b != core.Kind("H") {
		t.Fatalf("Pair() = (%v, %v), want (H, H)", a, b)
	}
}

func TestHHFusion_CollidingPairIsConsumed(t *testing.T) {
	a := newFakeParticle(core.Kind("H"), 100, 100, 12)
	b := newFakeParticle(core.Kind("H"), 105, 105, 12)
	newH := newFakeParticle(core.Kind("H"), 42, 7, 12)

	rule := &HHFusion{
		NewH2:        func(x, y float64) core.Particle { return newFakeParticle(core.Kind("H2"), x, y, 12) },
		NewHAtRandom: func() core.Particle { return newH },
	}

	result := rule.React(a, b)

	if len(result.Consumed) != 2 || result.Consumed[0] != core.Particle(a) || result.Consumed[1] != core.Particle(b) {
		t.Fatalf("Consumed = %v, want [a, b]", result.Consumed)
	}
}

func TestHHFusion_ProducesH2AtBCoordsAndNewH(t *testing.T) {
	a := newFakeParticle(core.Kind("H"), 100, 100, 12)
	b := newFakeParticle(core.Kind("H"), 105, 105, 12)
	newH := newFakeParticle(core.Kind("H"), 42, 7, 12)

	var h2CalledWithX, h2CalledWithY float64
	rule := &HHFusion{
		NewH2: func(x, y float64) core.Particle {
			h2CalledWithX, h2CalledWithY = x, y
			return newFakeParticle(core.Kind("H2"), x, y, 12)
		},
		NewHAtRandom: func() core.Particle { return newH },
	}

	result := rule.React(a, b)

	if len(result.Produced) != 2 {
		t.Fatalf("len(Produced) = %d, want 2", len(result.Produced))
	}

	var h2, h core.Particle
	for _, p := range result.Produced {
		switch p.Kind() {
		case core.Kind("H2"):
			h2 = p
		case core.Kind("H"):
			h = p
		}
	}

	if h2 == nil {
		t.Fatal("Produced missing H2")
	}
	if h2CalledWithX != b.X() || h2CalledWithY != b.Y() {
		t.Fatalf("NewH2 called with (%v, %v), want b's coords (%v, %v)", h2CalledWithX, h2CalledWithY, b.X(), b.Y())
	}
	if h2.X() != b.X() || h2.Y() != b.Y() {
		t.Fatalf("H2 coords = (%v, %v), want b's coords (%v, %v)", h2.X(), h2.Y(), b.X(), b.Y())
	}

	if h == nil {
		t.Fatal("Produced missing H")
	}
	if h != core.Particle(newH) {
		t.Fatalf("H = %v, want the particle returned by NewHAtRandom", h)
	}
}

func TestHHFusion_BalanceIsHMinus1H2Plus1(t *testing.T) {
	a := newFakeParticle(core.Kind("H"), 100, 100, 12)
	b := newFakeParticle(core.Kind("H"), 105, 105, 12)
	newH := newFakeParticle(core.Kind("H"), 42, 7, 12)

	rule := &HHFusion{
		NewH2:        func(x, y float64) core.Particle { return newFakeParticle(core.Kind("H2"), x, y, 12) },
		NewHAtRandom: func() core.Particle { return newH },
	}

	result := rule.React(a, b)

	countKind := func(ps []core.Particle, kind core.Kind) int {
		n := 0
		for _, p := range ps {
			if p.Kind() == kind {
				n++
			}
		}
		return n
	}

	consumedH := countKind(result.Consumed, core.Kind("H"))
	producedH := countKind(result.Produced, core.Kind("H"))
	consumedH2 := countKind(result.Consumed, core.Kind("H2"))
	producedH2 := countKind(result.Produced, core.Kind("H2"))

	if producedH-consumedH != -1 {
		t.Fatalf("H balance = %d, want -1", producedH-consumedH)
	}
	if producedH2-consumedH2 != 1 {
		t.Fatalf("H2 balance = %d, want 1", producedH2-consumedH2)
	}
}
