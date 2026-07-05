package reaction

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

// stubRule は Registry の検索の挙動を確認するためだけに使う最小限の Rule
// であり、その React の結果はこれらのテストとは無関係。
type stubRule struct {
	a, b core.Kind
}

func (r stubRule) Pair() (core.Kind, core.Kind) { return r.a, r.b }

func (r stubRule) React(a, b core.Particle) Result { return Result{} }

func TestRegistry_FindsRuleInBothOrders(t *testing.T) {
	registry := NewRegistry()
	rule := stubRule{a: core.Kind("O"), b: core.Kind("H2")}
	registry.Register(rule)

	got, ok := registry.Find(core.Kind("O"), core.Kind("H2"))
	if !ok || got != rule {
		t.Fatalf("Find(O, H2) = %v, %v; want %v, true", got, ok, rule)
	}

	got, ok = registry.Find(core.Kind("H2"), core.Kind("O"))
	if !ok || got != rule {
		t.Fatalf("Find(H2, O) = %v, %v; want %v, true", got, ok, rule)
	}
}

func TestRegistry_FindsSameKindPair(t *testing.T) {
	registry := NewRegistry()
	rule := stubRule{a: core.Kind("H"), b: core.Kind("H")}
	registry.Register(rule)

	got, ok := registry.Find(core.Kind("H"), core.Kind("H"))
	if !ok || got != rule {
		t.Fatalf("Find(H, H) = %v, %v; want %v, true", got, ok, rule)
	}
}

func TestRegistry_UnregisteredPairNotFound(t *testing.T) {
	registry := NewRegistry()
	registry.Register(stubRule{a: core.Kind("H"), b: core.Kind("H")})

	if _, ok := registry.Find(core.Kind("H"), core.Kind("O")); ok {
		t.Fatal("Find(H, O) should not be found")
	}
	if _, ok := registry.Find(core.Kind("H2o"), core.Kind("H2o")); ok {
		t.Fatal("Find(H2o, H2o) should not be found")
	}
}

func TestRegistry_MultipleRulesEachFindable(t *testing.T) {
	registry := NewRegistry()
	hh := stubRule{a: core.Kind("H"), b: core.Kind("H")}
	oh2 := stubRule{a: core.Kind("O"), b: core.Kind("H2")}
	registry.Register(hh)
	registry.Register(oh2)

	if got, ok := registry.Find(core.Kind("H"), core.Kind("H")); !ok || got != hh {
		t.Fatalf("Find(H, H) = %v, %v; want %v, true", got, ok, hh)
	}
	if got, ok := registry.Find(core.Kind("H2"), core.Kind("O")); !ok || got != oh2 {
		t.Fatalf("Find(H2, O) = %v, %v; want %v, true", got, ok, oh2)
	}
}

func TestRegistry_ReactiveKindsContainsRegisteredPairKinds(t *testing.T) {
	registry := NewRegistry()
	registry.Register(stubRule{a: core.Kind("H"), b: core.Kind("H")})
	registry.Register(stubRule{a: core.Kind("O"), b: core.Kind("H2")})

	kinds := registry.ReactiveKinds()
	for _, k := range []core.Kind{"H", "O", "H2"} {
		if _, ok := kinds[k]; !ok {
			t.Fatalf("ReactiveKinds() missing %v", k)
		}
	}
	if len(kinds) != 3 {
		t.Fatalf("ReactiveKinds() size = %d, want 3", len(kinds))
	}
}

func TestRegistry_ReactiveKindsExcludesUnreferencedKind(t *testing.T) {
	registry := NewRegistry()
	registry.Register(stubRule{a: core.Kind("H"), b: core.Kind("H")})
	registry.Register(stubRule{a: core.Kind("O"), b: core.Kind("H2")})

	if _, ok := registry.ReactiveKinds()[core.Kind("H2o")]; ok {
		t.Fatal("ReactiveKinds() should not contain H2o")
	}
}

func TestRegistry_ReactiveKindsEmptyWhenNoRulesRegistered(t *testing.T) {
	registry := NewRegistry()
	if got := len(registry.ReactiveKinds()); got != 0 {
		t.Fatalf("ReactiveKinds() size = %d, want 0", got)
	}
}

func TestRegistry_ReactiveKindsReflectsLaterRegister(t *testing.T) {
	registry := NewRegistry()
	if _, ok := registry.ReactiveKinds()[core.Kind("O3")]; ok {
		t.Fatal("ReactiveKinds() should not contain O3 before it is registered")
	}

	registry.Register(stubRule{a: core.Kind("O"), b: core.Kind("O3")})

	if _, ok := registry.ReactiveKinds()[core.Kind("O3")]; !ok {
		t.Fatal("ReactiveKinds() should contain O3 after it is registered")
	}
}
