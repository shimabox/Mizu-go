package bench

import "testing"

// Mizu-ts の scripts/bench/scenarios.mjs 相当のテスト。

func TestScenarioOrder(t *testing.T) {
	want := []string{"default", "500", "1000", "3000"}
	if len(ScenarioOrder) != len(want) {
		t.Fatalf("ScenarioOrder = %v, want %v", ScenarioOrder, want)
	}
	for i, name := range want {
		if ScenarioOrder[i] != name {
			t.Errorf("ScenarioOrder[%d] = %q, want %q", i, ScenarioOrder[i], name)
		}
	}
}

func TestDefaultFramesFor(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"default", 300},
		{"500", 300},
		{"1000", 300},
		{"3000", 60},
	}
	for _, tt := range tests {
		if got := DefaultFramesFor(tt.name); got != tt.want {
			t.Errorf("DefaultFramesFor(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestResolveScenarios(t *testing.T) {
	got, err := ResolveScenarios([]string{"default", "500"})
	if err != nil {
		t.Fatalf("ResolveScenarios error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].Name != "default" || got[0].H != 30 || got[0].O != 50 {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].Name != "500" || got[1].H != 500 || got[1].O != 500 {
		t.Errorf("got[1] = %+v", got[1])
	}

	if _, err := ResolveScenarios([]string{"unknown"}); err == nil {
		t.Errorf("ResolveScenarios([unknown]) error = nil, want error")
	}
}

func TestScenarioDefsCoverAllOrderedNames(t *testing.T) {
	for _, name := range ScenarioOrder {
		s, ok := ScenarioDefs[name]
		if !ok {
			t.Fatalf("ScenarioDefs[%q] missing", name)
		}
		if s.Label == "" {
			t.Errorf("ScenarioDefs[%q].Label is empty", name)
		}
	}
}
