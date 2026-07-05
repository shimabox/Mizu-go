package bench

import "fmt"

// Scenario はベンチ対象の 1 シナリオ(初期 H/O 原子数)を表す。Go の
// Simulator.Init(hCount, oCount) にそのまま渡す想定であり、シミュ
// レーター自身が CountScale() でスケールする(cmd/mizu/main.go 参照)。
type Scenario struct {
	Name  string
	Label string
	H     int
	O     int
}

// ScenarioOrder は Mizu-ts の scripts/bench/scenarios.mjs の
// SCENARIO_ORDER に対応するシナリオ名の既定順序。
var ScenarioOrder = []string{"default", "500", "1000", "3000"}

// ScenarioDefs は Mizu-ts の SCENARIO_DEFS に対応するシナリオ定義。
// 「default」シナリオは cmd/mizu 自身の既定値(h=30, o=50。実際の個数は
// Simulator.CountScale() でスケールされる)に依存し、負荷シナリオは
// 明示的な H/O 値を指定する。
var ScenarioDefs = map[string]Scenario{
	"default": {Name: "default", Label: "デフォルト (h=30, o=50)", H: 30, O: 50},
	"500":     {Name: "500", Label: "h=500, o=500", H: 500, O: 500},
	"1000":    {Name: "1000", Label: "h=1000, o=1000", H: 1000, O: 1000},
	"3000":    {Name: "3000", Label: "h=3000, o=3000", H: 3000, O: 3000},
}

// DefaultFramesFor は、3000 シナリオのみ既定フレーム数を 60 にし、それ
// 以外は 300 にする(Mizu-ts の defaultFramesFor に対応)。
func DefaultFramesFor(scenarioName string) int {
	if scenarioName == "3000" {
		return 60
	}
	return 300
}

// ResolveScenarios はシナリオ名の並びをシナリオ定義の並びに変換する。
// 未知の名前が含まれる場合はエラーを返す(Mizu-ts の resolveScenarios
// に対応)。
func ResolveScenarios(names []string) ([]Scenario, error) {
	scenarios := make([]Scenario, 0, len(names))
	for _, name := range names {
		def, ok := ScenarioDefs[name]
		if !ok {
			return nil, fmt.Errorf("unknown scenario %q (available: %v)", name, ScenarioOrder)
		}
		scenarios = append(scenarios, def)
	}
	return scenarios, nil
}
