package particle

import (
	"image/color"

	"github.com/shimabox/Mizu-go/internal/behavior"
	"github.com/shimabox/Mizu-go/internal/core"
)

// このパッケージが生成する粒子の種別定数。Mizu-ts の ParticleKind リテラル
// ('H' | 'H2' | 'O' | 'H2o') に対応する。
const (
	KindH   core.Kind = "H"
	KindH2  core.Kind = "H2"
	KindO   core.Kind = "O"
	KindH2o core.Kind = "H2o"
)

// Colored は自身の描画色を持つ粒子が実装するインターフェース。Mizu-ts では
// レンダラーが粒子ごとのランダムな色を保持していたのに対し、Go 版のレンダー
// 層は Kind によって粒子の描画方法を決定するため、色は粒子自身が保持する。
// レンダー層 (Phase 5) はこのインターフェースに型アサーションを行う。
type Colored interface {
	Color() color.RGBA
}

// atom は H、H2、O 粒子が共有する薄い構造体。Mizu-ts の H.ts、H2.ts、O.ts は
// kind 以外まったく同一の実装のため、単一の構造体と kind ごとのコンストラク
// タ (factory.go を参照) があれば十分であり、H/H2/O をそれぞれ別の型として
// 定義する必要はない。
type atom struct {
	kind     core.Kind
	x, y     float64
	radius   float64
	movement behavior.MovementBehavior
	dead     bool
	color    color.RGBA
}

var (
	_ core.Particle = (*atom)(nil)
	_ Colored       = (*atom)(nil)
)

// newAtom は指定された kind の atom を生成する。movement を外部から注入する
// ことで、H、H2、O はいずれも RandomWalk を共有しつつ、その実装を意識せずに
// 済む。
func newAtom(kind core.Kind, x, y, radius float64, movement behavior.MovementBehavior, col color.RGBA) *atom {
	return &atom{
		kind:     kind,
		x:        x,
		y:        y,
		radius:   radius,
		movement: movement,
		color:    col,
	}
}

func (a *atom) Kind() core.Kind   { return a.kind }
func (a *atom) X() float64        { return a.x }
func (a *atom) Y() float64        { return a.y }
func (a *atom) Radius() float64   { return a.radius }
func (a *atom) Color() color.RGBA { return a.color }

// Update は注入された movement behavior によって位置を進める。描画は一切
// 行ってはならない (Mizu-ts の H/H2/O の update() に対応する)。
func (a *atom) Update() {
	a.x, a.y = a.movement.Next(a.x, a.y)
}

func (a *atom) IsDead() bool { return a.dead }
func (a *atom) MarkDead()    { a.dead = true }
