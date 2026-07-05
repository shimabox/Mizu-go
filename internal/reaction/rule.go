package reaction

import "github.com/shimabox/Mizu-go/internal/core"

// Result は Rule が粒子のペアに反応した結果である: 消滅する粒子
// (Consumed) と、反応の結果として新たに生じる粒子 (Produced)。
type Result struct {
	Consumed []core.Particle
	Produced []core.Particle
}

// Rule は特定の、順序を問わない粒子種別のペアを反応させて Result にする。
type Rule interface {
	// Pair はこの rule が反応させる 2 つの kind を返す。Registry は引数の
	// 順序に関係なくペアをマッチさせる。
	Pair() (core.Kind, core.Kind)

	// React は衝突したペアに反応を適用し、消費された粒子と生成された粒子
	// を返す。呼び出し側は a と b をどちらの順序で渡してもよい。どちらが
	// どちらかを区別する必要がある実装は、Kind によって判別しなければ
	// ならない (OxidationToWater を参照)。
	React(a, b core.Particle) Result
}
