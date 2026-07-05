package physics

import "github.com/shimabox/Mizu-go/internal/core"

// GridCollisionDetector は衝突検出に SpatialGrid(一様グリッド)を使う。
//
// これが実行時に動くもので、BruteForceCollisionDetector(O(n^2))の
// 代わりを務める。SpatialGrid.cellSize が最大衝突距離以上である限り、
// 隣接する9セルだけを走査すれば衝突候補を漏れなく列挙できる
// (SpatialGrid のコメントを参照)。密度が極端でない限り、平均的な計算量
// はおおよそ O(N) になる。
type GridCollisionDetector struct {
	grid *SpatialGrid
}

// NewGridCollisionDetector は grid を使う GridCollisionDetector を返す。
func NewGridCollisionDetector(grid *SpatialGrid) *GridCollisionDetector {
	return &GridCollisionDetector{grid: grid}
}

// FindHitPairs は CollisionDetector を実装する。
func (d *GridCollisionDetector) FindHitPairs(particles []core.Particle) []Pair {
	d.grid.Clear()
	for _, p := range particles {
		d.grid.Insert(p)
	}

	// 重複ペアや自己ペア((a,b) と (b,a))を避けるため、元のスライスに
	// おける各粒子のインデックスを「正規の順序」として使う(必ず
	// インデックスの小さいほうの粒子を先に置く)。
	indexOf := make(map[core.Particle]int, len(particles))
	for i, p := range particles {
		indexOf[p] = i
	}

	var pairs []Pair
	for _, p := range particles {
		pIndex, ok := indexOf[p]
		if !ok {
			continue
		}
		for _, q := range d.grid.Neighbors(p) {
			if q == p {
				continue // 自分自身を除外
			}
			qIndex, ok := indexOf[q]
			if !ok || qIndex <= pIndex {
				continue // 重複ペア((a,b)/(b,a))を除外
			}
			if gridIsHit(p, q) {
				pairs = append(pairs, Pair{A: p, B: q})
			}
		}
	}
	return pairs
}

func gridIsHit(a, b core.Particle) bool {
	dx := b.X() - a.X()
	dy := b.Y() - a.Y()
	hitDistance := a.Radius() + b.Radius()
	// math.Sqrt を避けるため、距離の2乗と半径合計の2乗を比較する。
	return dx*dx+dy*dy < hitDistance*hitDistance
}
