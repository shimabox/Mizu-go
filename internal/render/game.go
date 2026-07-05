package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/particle"
	"github.com/shimabox/Mizu-go/internal/sim"
)

// shadowOffset は、すべての粒子の落ち影に使われる生のスクリーン
// ピクセルのずらし量であり、Mizu-ts の shadowOffsetX/shadowOffsetY = 1
// (TextRenderer.ts、SubscriptTextRenderer.ts、DropletRenderer.ts)を
// 反映している。これは拡大縮小されない一定量である — CenterOffset の
// dx/dy の契約を参照。
const shadowOffset = 1.0

// textShadowTint は、Mizu-ts の TextRenderer.ts /
// SubscriptTextRenderer.ts の shadowColor '#888'(0x88 = 136)を、白い
// H/O/H2 グリフスプライトに適用する ColorScale の乗数として近似した
// ものである(porting-plan §5.2: 「白スプライト×0.53」)。
var textShadowTint = float32(0x88) / 255

// dropletShadowR, dropletShadowG, dropletShadowB は、Mizu-ts の
// DropletRenderer.ts の shadowColor '#007fff' を、(すでにグラデー
// ションで着色済みの)水滴スプライトに適用する ColorScale の乗数
// として近似したものである。
const (
	dropletShadowR = 0x00 / 255.0
	dropletShadowG = 0x7f / 255.0
	dropletShadowB = 0xff / 255.0
)

// Game は ebiten.Game を実装し、Simulator を駆動して、その粒子を
// あらかじめ描画済みの Sprites を使って種類ごとに描画する
// (porting-plan §5.2、§5.5)。Update はシミュレーションを進めるのみで、
// Draw が状態を変更することは決してない。
type Game struct {
	sim     *sim.Simulator
	sprites *Sprites
	measure bool

	// opts は、粒子ごとのアロケーションを避けるため、すべての
	// DrawImage 呼び出しで使い回される(使用前にリセットされる)。
	// これは porting plan の「DrawImageOptions は毎回 new せず使い
	// 回してアロケーションを抑える」を反映したものである。
	opts ebiten.DrawImageOptions

	lastOutsideW, lastOutsideH int
}

var _ ebiten.Game = (*Game)(nil)

// NewGame は、sim を駆動し sprites を使って描画する Game を返す。
// measure は統計オーバーレイを有効にするフラグだが、その配線は
// Phase 6a に先送りされているため(Draw 内の TODO を参照)、現時点では
// 受け取るだけで未使用である。
func NewGame(s *sim.Simulator, sprites *Sprites, measure bool) *Game {
	return &Game{sim: s, sprites: sprites, measure: measure}
}

// Update はシミュレーションを 1 ティック進める。何も描画してはならない
// (porting-plan §5.5)。
func (g *Game) Update() error {
	g.sim.Update()
	return nil
}

// Draw は現在のフレームを描画する: 白い背景を描き、続いて
// (Simulator.Buckets により初出順で)種類ごとにすべての粒子を、あら
// かじめ描画済みのスプライトへの 2 回の DrawImage 呼び出し(オフセット
// させた着色済みの影、続いて着色済みの本体)として描画することで、
// Ebitengine の自動バッチングが適用されるようにする。Draw はシミュ
// レーションの状態を変更してはならず(porting-plan §5.5)、g.sim から
// 読み取るのみである。
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	for _, b := range g.sim.Buckets() {
		switch b.Kind {
		case particle.KindH:
			g.drawGlyphs(screen, g.sprites.H, b.Particles)
		case particle.KindO:
			g.drawGlyphs(screen, g.sprites.O, b.Particles)
		case particle.KindH2:
			g.drawGlyphs(screen, g.sprites.H2, b.Particles)
		case particle.KindH2o:
			g.drawDroplets(screen, b.Particles)
		}
	}

	if g.measure {
		// TODO(Phase 6a): porting-plan §5.6 に記載された FPS/フレーム
		// 時間/粒子数の統計オーバーレイをここで描画する。
	}
}

// drawGlyphs は、particles 内のすべての粒子を sprite(Sprites.H/O/H2
// のいずれか)を使って描画する: (shadowOffset, shadowOffset) だけ
// オフセットしたグレー着色の影のコピーを描き、続いてその上に粒子自身
// の色で着色したものを描く。どちらも粒子の (X, Y) を中心とし、実際の
// 直径に合わせて拡大縮小される(porting-plan §5.2)。
func (g *Game) drawGlyphs(screen *ebiten.Image, sprite *ebiten.Image, particles []core.Particle) {
	bounds := sprite.Bounds()
	spriteW, spriteH := float64(bounds.Dx()), float64(bounds.Dy())

	for _, p := range particles {
		scale := UniformScale(p.Radius()*2, spriteW)

		// 影: グレー着色、(shadowOffset, shadowOffset) だけオフセット。
		tx, ty := CenterOffset(spriteW, spriteH, scale, p.X(), p.Y(), shadowOffset, shadowOffset)
		g.opts.GeoM.Reset()
		g.opts.GeoM.Scale(scale, scale)
		g.opts.GeoM.Translate(tx, ty)
		g.opts.ColorScale.Reset()
		g.opts.ColorScale.Scale(textShadowTint, textShadowTint, textShadowTint, 1)
		screen.DrawImage(sprite, &g.opts)

		// 本体: 粒子自身のランダムな色、オフセットなし。
		tx, ty = CenterOffset(spriteW, spriteH, scale, p.X(), p.Y(), 0, 0)
		g.opts.GeoM.Reset()
		g.opts.GeoM.Scale(scale, scale)
		g.opts.GeoM.Translate(tx, ty)
		g.opts.ColorScale.Reset()
		if col, ok := p.(particle.Colored); ok {
			g.opts.ColorScale.ScaleWithColor(col.Color())
		}
		screen.DrawImage(sprite, &g.opts)
	}
}

// drawDroplets は、Sprites.Droplet を使ってすべての H2o 粒子を描画
// する: (shadowOffset, shadowOffset) だけオフセットした #007fff 着色の
// 影のコピーを描き、続いてその上に着色なしの本体(グラデーション自身
// に焼き込まれた色)を描く。どちらも粒子の (X, Y) を中心とし、実際の
// 直径に合わせて拡大縮小される(porting-plan §5.2)。
func (g *Game) drawDroplets(screen *ebiten.Image, particles []core.Particle) {
	sprite := g.sprites.Droplet
	bounds := sprite.Bounds()
	spriteW, spriteH := float64(bounds.Dx()), float64(bounds.Dy())

	for _, p := range particles {
		scale := UniformScale(p.Radius()*2, spriteW)

		// 影: #007fff 着色、(shadowOffset, shadowOffset) だけオフセット。
		tx, ty := CenterOffset(spriteW, spriteH, scale, p.X(), p.Y(), shadowOffset, shadowOffset)
		g.opts.GeoM.Reset()
		g.opts.GeoM.Scale(scale, scale)
		g.opts.GeoM.Translate(tx, ty)
		g.opts.ColorScale.Reset()
		g.opts.ColorScale.Scale(dropletShadowR, dropletShadowG, dropletShadowB, 1)
		screen.DrawImage(sprite, &g.opts)

		// 本体: 着色なし、グラデーション自身に焼き込まれた色。
		tx, ty = CenterOffset(spriteW, spriteH, scale, p.X(), p.Y(), 0, 0)
		g.opts.GeoM.Reset()
		g.opts.GeoM.Scale(scale, scale)
		g.opts.GeoM.Translate(tx, ty)
		g.opts.ColorScale.Reset()
		screen.DrawImage(sprite, &g.opts)
	}
}

// Layout は、outside のサイズをそのまま論理解像度として返し
// (porting-plan §5.4: ウィンドウはリサイズ可能で、論理解像度はそれに
// 1:1 で追従する)、サイズが変化した場合はシミュレーターの共有境界と
// 衝突判定グリッドをそれに合わせてリサイズする。
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth != g.lastOutsideW || outsideHeight != g.lastOutsideH {
		g.lastOutsideW, g.lastOutsideH = outsideWidth, outsideHeight
		g.sim.Resize(float64(outsideWidth), float64(outsideHeight))
	}
	return outsideWidth, outsideHeight
}
