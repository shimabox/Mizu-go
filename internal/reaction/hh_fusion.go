package reaction

import "github.com/shimabox/Mizu-go/internal/core"

// HHFusion は H + H -> H2 を実装する: ペアの一方 (b) はその場で H2 になり、
// もう一方 (a) はランダムな位置に新しい H として置き換えられる。これにより、
// H2 の融合が H 原子を単純に消滅させてしまうことを防いでいる。差し引きの
// 効果: H -1、H2 +1。
//
// 粒子の生成は外部から注入される。これは internal/particle (Phase 3) が
// 粒子の構築を担当しており、このパッケージは何をどこに生成するかだけを
// 決めるためである。
type HHFusion struct {
	// NewH2 は指定された座標に H2 粒子を生成する。
	NewH2 func(x, y float64) core.Particle
	// NewHAtRandom は画面内のランダムな位置に H 粒子を生成する。
	NewHAtRandom func() core.Particle
}

// Pair は (H, H) を返す。
func (r *HHFusion) Pair() (core.Kind, core.Kind) {
	return core.Kind("H"), core.Kind("H")
}

// React は a と b の両方を消費し、b の座標に H2 を、新たに配置した H を
// 生成する。
func (r *HHFusion) React(a, b core.Particle) Result {
	return Result{
		Consumed: []core.Particle{a, b},
		Produced: []core.Particle{
			r.NewH2(b.X(), b.Y()),
			r.NewHAtRandom(),
		},
	}
}
