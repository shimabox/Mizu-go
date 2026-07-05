package reaction

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func TestOxidationToWater_PairIsOH2(t *testing.T) {
	rule := &OxidationToWater{}

	a, b := rule.Pair()
	if a != core.Kind("O") || b != core.Kind("H2") {
		t.Fatalf("Pair() = (%v, %v), want (O, H2)", a, b)
	}
}

func newOxidationRule(newO, newH, newH2o core.Particle) *OxidationToWater {
	return &OxidationToWater{
		NewOAtRandom: func() core.Particle { return newO },
		NewHAtRandom: func() core.Particle { return newH },
		NewH2o:       func(x, y float64) core.Particle { return newH2o },
	}
}

func TestOxidationToWater_BalanceIsO0H2Minus1HPlus1H2oPlus1(t *testing.T) {
	o := newFakeParticle(core.Kind("O"), 100, 100, 12)
	h2 := newFakeParticle(core.Kind("H2"), 105, 105, 12)
	newO := newFakeParticle(core.Kind("O"), 1, 2, 12)
	newH := newFakeParticle(core.Kind("H"), 3, 4, 12)
	newH2o := newFakeParticle(core.Kind("H2o"), o.X(), o.Y(), 18)

	rule := newOxidationRule(newO, newH, newH2o)

	result := rule.React(o, h2)

	balance := func(kind core.Kind) int {
		count := func(ps []core.Particle) int {
			n := 0
			for _, p := range ps {
				if p.Kind() == kind {
					n++
				}
			}
			return n
		}
		return count(result.Produced) - count(result.Consumed)
	}

	if got := balance(core.Kind("O")); got != 0 {
		t.Fatalf("O balance = %d, want 0", got)
	}
	if got := balance(core.Kind("H2")); got != -1 {
		t.Fatalf("H2 balance = %d, want -1", got)
	}
	if got := balance(core.Kind("H")); got != 1 {
		t.Fatalf("H balance = %d, want 1", got)
	}
	if got := balance(core.Kind("H2o")); got != 1 {
		t.Fatalf("H2o balance = %d, want 1", got)
	}
}

func TestOxidationToWater_H2oCreatedAtOCoords(t *testing.T) {
	o := newFakeParticle(core.Kind("O"), 200, 300, 12)
	h2 := newFakeParticle(core.Kind("H2"), 205, 305, 12)

	var gotX, gotY float64
	rule := &OxidationToWater{
		NewOAtRandom: func() core.Particle { return newFakeParticle(core.Kind("O"), 1, 2, 12) },
		NewHAtRandom: func() core.Particle { return newFakeParticle(core.Kind("H"), 3, 4, 12) },
		NewH2o: func(x, y float64) core.Particle {
			gotX, gotY = x, y
			return newFakeParticle(core.Kind("H2o"), x, y, 18)
		},
	}

	result := rule.React(o, h2)

	if gotX != 200 || gotY != 300 {
		t.Fatalf("NewH2o called with (%v, %v), want (200, 300)", gotX, gotY)
	}

	var h2o core.Particle
	for _, p := range result.Produced {
		if p.Kind() == core.Kind("H2o") {
			h2o = p
		}
	}
	if h2o == nil {
		t.Fatal("Produced missing H2o")
	}
	if h2o.X() != 200 || h2o.Y() != 300 {
		t.Fatalf("H2o coords = (%v, %v), want O's coords (200, 300)", h2o.X(), h2o.Y())
	}
}

func TestOxidationToWater_ArgumentOrderH2ThenODoesNotMatter(t *testing.T) {
	o := newFakeParticle(core.Kind("O"), 200, 300, 12)
	h2 := newFakeParticle(core.Kind("H2"), 205, 305, 12)

	var gotX, gotY float64
	rule := &OxidationToWater{
		NewOAtRandom: func() core.Particle { return newFakeParticle(core.Kind("O"), 1, 2, 12) },
		NewHAtRandom: func() core.Particle { return newFakeParticle(core.Kind("H"), 3, 4, 12) },
		NewH2o: func(x, y float64) core.Particle {
			gotX, gotY = x, y
			return newFakeParticle(core.Kind("H2o"), x, y, 18)
		},
	}

	// Arguments passed as (h2, o) instead of (o, h2).
	result := rule.React(h2, o)

	foundO, foundH2 := false, false
	for _, p := range result.Consumed {
		if p == core.Particle(o) {
			foundO = true
		}
		if p == core.Particle(h2) {
			foundH2 = true
		}
	}
	if !foundO || !foundH2 {
		t.Fatalf("Consumed = %v, want to contain both o and h2", result.Consumed)
	}

	if gotX != 200 || gotY != 300 {
		t.Fatalf("NewH2o called with (%v, %v), want O's coords (200, 300)", gotX, gotY)
	}

	var h2o core.Particle
	for _, p := range result.Produced {
		if p.Kind() == core.Kind("H2o") {
			h2o = p
		}
	}
	if h2o == nil || h2o.X() != 200 || h2o.Y() != 300 {
		t.Fatalf("H2o = %v, want coords (200, 300)", h2o)
	}
}

func TestOxidationToWater_ProducesNewOAndNewHFromInjectedFactories(t *testing.T) {
	o := newFakeParticle(core.Kind("O"), 100, 100, 12)
	h2 := newFakeParticle(core.Kind("H2"), 105, 105, 12)
	newO := newFakeParticle(core.Kind("O"), 1, 2, 12)
	newH := newFakeParticle(core.Kind("H"), 3, 4, 12)
	newH2o := newFakeParticle(core.Kind("H2o"), o.X(), o.Y(), 18)

	rule := newOxidationRule(newO, newH, newH2o)

	result := rule.React(o, h2)

	if len(result.Produced) != 3 {
		t.Fatalf("len(Produced) = %d, want 3", len(result.Produced))
	}

	var gotO, gotH core.Particle
	for _, p := range result.Produced {
		switch p.Kind() {
		case core.Kind("O"):
			gotO = p
		case core.Kind("H"):
			gotH = p
		}
	}

	if gotO != core.Particle(newO) {
		t.Fatalf("produced O = %v, want the particle returned by NewOAtRandom", gotO)
	}
	if gotH != core.Particle(newH) {
		t.Fatalf("produced H = %v, want the particle returned by NewHAtRandom", gotH)
	}
}
