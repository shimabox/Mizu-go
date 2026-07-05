package physics

import (
	"fmt"
	"math"

	"github.com/shimabox/Mizu-go/internal/core"
)

// MaxParticleRadius は、粒子が取り得ると想定する最大半径(px)。
//
// 根拠(ParticleFactory の生成ロジックから移植したもの。参考として実測値も
// 添える):
//   - フォントサイズは `24 * getScale()`(getScale は 1.0 か 1.2)なので、
//     上限は 28.8px。
//   - 最も幅の広いテキストは(下付き文字を含む)"H2" で、フォントサイズ
//     28.8px のとき(node canvas、sans-serif で計測)幅は約36.8px ->
//     半径は約18.4px。
//   - H2o(水滴)の直径は `(random*10 + 18) * getScale()` なので、上限は
//     `28 * 1.2 = 33.6px` -> 半径16.8px。
//   - 実測での最大半径は約18.4px。ブラウザやフォントの違いによるばらつき
//     を吸収するため安全マージンを加え、24px を「想定最大半径」とする。
//
// この前提は spatial_grid_test.go のカナリアテスト
// ("exceeding MaxParticleRadius breaks the invariant")で継続的に検証
// されており、(internal/particle が存在するようになったら)そちらでも
// 実際の factory の出力に対して検証すべきである。
const MaxParticleRadius = 24.0

// MaxCollisionDistance は2つの粒子が衝突しうる最大距離(両方が
// MaxParticleRadius のときの半径の合計)。cellSize は少なくともこの値以上
// でなければならない。
//
// これにより「衝突しうる任意の2粒子は必ず同じセルか隣接セル(3x3)内に
// 収まる」という不変条件が成り立つ。したがって、あるセルとその8近傍を
// 走査すれば、衝突候補を漏れなく列挙できることが保証される(cellSize
// 未満しか離れていない2点は、各軸のセルインデックスの差が高々1に
// なるため)。
const MaxCollisionDistance = MaxParticleRadius * 2

// DefaultCellSize は、特に選ぶ理由がない限り CollisionDetector とその
// 仲間が使うセルサイズ。
const DefaultCellSize = MaxCollisionDistance

// SpatialGrid は空間分割のための一様グリッド。
//
//   - バケツ(bucket)のスライスは(NewSpatialGrid で)一度だけ確保され、
//     使い回される。Clear は各バケツの長さを 0 に戻すだけで、裏の配列を
//     再確保することはない。
//   - 画面外の座標(負の値、または幅・高さ以上の値)は、最も近い端の
//     セルにクランプされる。
type SpatialGrid struct {
	width, height float64
	cellSize      float64
	cols, rows    int
	buckets       [][]core.Particle
}

// NewSpatialGrid は [0,width) x [0,height) を覆い、cellSize x cellSize
// のセルに分割された SpatialGrid を返す。cellSize <= 0 の場合は panic
// する。
func NewSpatialGrid(width, height, cellSize float64) *SpatialGrid {
	if cellSize <= 0 {
		panic(fmt.Sprintf("physics: cellSize must be greater than 0, got %v", cellSize))
	}

	g := &SpatialGrid{}
	g.rebuild(width, height, cellSize)
	return g
}

// Resize は新しい画面サイズに合わせてグリッドを再構築し、それまでに
// 挿入されていた粒子は破棄する。TS 版に対応するものはなく、simulator が
// ウィンドウのリサイズに合わせてグリッドを同期させ続けられるように
// 存在する。
func (g *SpatialGrid) Resize(width, height float64) {
	g.rebuild(width, height, g.cellSize)
}

func (g *SpatialGrid) rebuild(width, height, cellSize float64) {
	g.width = width
	g.height = height
	g.cellSize = cellSize
	g.cols = int(math.Max(1, math.Ceil(width/cellSize)))
	g.rows = int(math.Max(1, math.Ceil(height/cellSize)))

	g.buckets = make([][]core.Particle, g.cols*g.rows)
}

// Clear は裏の配列を再確保することなく、全てのバケツを空にする。
func (g *SpatialGrid) Clear() {
	for i := range g.buckets {
		g.buckets[i] = g.buckets[i][:0]
	}
}

// Insert は p を、その現在位置に対応するバケツに配置する。
func (g *SpatialGrid) Insert(p core.Particle) {
	idx := g.cellIndex(p.X(), p.Y())
	g.buckets[idx] = append(g.buckets[idx], p)
}

// Neighbors は p 自身を含め、p のセルとその8個の隣接セル(3x3)に含まれる
// 粒子を返す。
func (g *SpatialGrid) Neighbors(p core.Particle) []core.Particle {
	col := g.clampCol(int(math.Floor(p.X() / g.cellSize)))
	row := g.clampRow(int(math.Floor(p.Y() / g.cellSize)))

	var result []core.Particle
	for dy := -1; dy <= 1; dy++ {
		ny := row + dy
		if ny < 0 || ny >= g.rows {
			continue
		}
		for dx := -1; dx <= 1; dx++ {
			nx := col + dx
			if nx < 0 || nx >= g.cols {
				continue
			}
			result = append(result, g.buckets[ny*g.cols+nx]...)
		}
	}
	return result
}

func (g *SpatialGrid) cellIndex(x, y float64) int {
	col := g.clampCol(int(math.Floor(x / g.cellSize)))
	row := g.clampRow(int(math.Floor(y / g.cellSize)))
	return row*g.cols + col
}

func (g *SpatialGrid) clampCol(col int) int {
	return clampInt(col, 0, g.cols-1)
}

func (g *SpatialGrid) clampRow(row int) int {
	return clampInt(row, 0, g.rows-1)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
