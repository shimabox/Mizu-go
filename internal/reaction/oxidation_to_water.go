package reaction

import "github.com/shimabox/Mizu-go/internal/core"

// OxidationToWater は O + H2 -> H2o を実装する: 酸素の座標に H2o 粒子が
// 生成され、酸素はランダムな位置に新しい O として置き換えられ (差し引き
// 0)、H2 は消滅し、新しい H が追加される。差し引きの効果: O ±0、H2 -1、
// H +1、H2o +1。
//
// 粒子の生成は外部から注入される。これは internal/particle (Phase 3) が
// 粒子の構築を担当しており、このパッケージは何をどこに生成するかだけを
// 決めるためである。
type OxidationToWater struct {
	// NewOAtRandom は画面内のランダムな位置に O 粒子を生成する。
	NewOAtRandom func() core.Particle
	// NewHAtRandom は画面内のランダムな位置に H 粒子を生成する。
	NewHAtRandom func() core.Particle
	// NewH2o は指定された座標に H2o 粒子を生成する。
	NewH2o func(x, y float64) core.Particle
}

// Pair は (O, H2) を返す。
func (r *OxidationToWater) Pair() (core.Kind, core.Kind) {
	return core.Kind("O"), core.Kind("H2")
}

// React は Kind によって a/b を判別し (引数の順序は問わない)、O と H2 を
// 消費して、新しい O、新しい H、そして元の O の座標に H2o を生成する。
func (r *OxidationToWater) React(a, b core.Particle) Result {
	o, h2 := a, b
	if a.Kind() != core.Kind("O") {
		o, h2 = b, a
	}

	return Result{
		Consumed: []core.Particle{o, h2},
		Produced: []core.Particle{
			r.NewOAtRandom(),
			r.NewHAtRandom(),
			r.NewH2o(o.X(), o.Y()),
		},
	}
}
