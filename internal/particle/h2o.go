package particle

import (
	"image/color"

	"github.com/shimabox/Mizu-go/internal/behavior"
	"github.com/shimabox/Mizu-go/internal/core"
)

// h2oColor は H2o の水滴に使われる固定の色。実際の描画はレンダー層
// (Phase 5) が描く放射状グラデーションのため、この値は描画には使われない
// が、h2o が他の粒子種別と同様に Colored を満たすようにするために存在する。
// #007fff はグラデーションの外側の色と一致する (porting-plan §5.2)。
var h2oColor = color.RGBA{R: 0x00, G: 0x7f, B: 0xff, A: 0xff}

// h2o は水滴の粒子。*behavior.FallAndSway によって落下しながら揺れ、着地
// すると自身を死亡としてマークする点が、画面をラップするだけの H、H2、O
// とは異なる (Mizu-ts の H2o.ts に対応する)。
type h2o struct {
	x, y     float64
	radius   float64
	movement *behavior.FallAndSway
	dead     bool
}

var (
	_ core.Particle = (*h2o)(nil)
	_ Colored       = (*h2o)(nil)
)

// newH2o は h2o の水滴を生成する。movement の具体的な型が (MovementBehavior
// インターフェースではなく) *behavior.FallAndSway なのは、Update が
// HasLanded を必要とし、それは FallAndSway だけが公開しているため。
func newH2o(x, y, radius float64, movement *behavior.FallAndSway) *h2o {
	return &h2o{x: x, y: y, radius: radius, movement: movement}
}

func (p *h2o) Kind() core.Kind   { return KindH2o }
func (p *h2o) X() float64        { return p.x }
func (p *h2o) Y() float64        { return p.y }
func (p *h2o) Radius() float64   { return p.radius }
func (p *h2o) Color() color.RGBA { return h2oColor }

// Update は位置を進めたのち、着地していれば水滴を死亡としてマークする。
// この順序 (まず移動し、その後新しい y を確認する) は Mizu-ts の
// H2o.ts:33-41 と厳密に一致する。
func (p *h2o) Update() {
	p.x, p.y = p.movement.Next(p.x, p.y)

	if p.movement.HasLanded(p.y) {
		p.MarkDead()
	}
}

func (p *h2o) IsDead() bool { return p.dead }
func (p *h2o) MarkDead()    { p.dead = true }
