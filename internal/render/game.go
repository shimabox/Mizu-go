package render

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/debug"
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

// overlayFontSize は、統計オーバーレイのフォントサイズ(px 単位)で
// あり、Mizu-ts の StatsOverlay.ts の fontSize: '12px' に一致する
// (porting-plan §5.6)。
const overlayFontSize = 12.0

// overlayPadX, overlayPadY は、統計オーバーレイのパディングであり、
// Mizu-ts の StatsOverlay.ts の padding: '3px 5px'(CSS の
// ショートハンド: 縦 3px、横 5px)に一致する。
const (
	overlayPadX = 5.0
	overlayPadY = 3.0
)

// overlayBackground と overlayTextColor は、Mizu-ts の
// StatsOverlay.ts の backgroundColor: 'rgba(0, 0, 0, 0.5)' および
// color: 'aqua'(#00ffff)に一致する。
var (
	overlayBackground = color.RGBA{R: 0, G: 0, B: 0, A: 128}
	overlayTextColor  = color.RGBA{R: 0, G: 255, B: 255, A: 255}
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

	// 以下のフィールドは measure が true のときのみ使われる
	// (porting-plan §5.6)。

	// stats は、オーバーレイの FPS・フレーム時間・更新時間の計測値を
	// 蓄積し、その表示用テキストを生成する。internal/debug を参照。
	stats *debug.Stats
	// overlayFace と overlayLineSpacing は、オーバーレイのテキストを
	// 計測・描画するために使われる(drawOverlay を参照)。
	overlayFace        *text.GoTextFace
	overlayLineSpacing float64
	// lastDrawTime は前回の Draw 呼び出しの実時刻であり、
	// stats.RecordFrame に渡すフレーム間隔の計算に使われる。ゼロ値は
	// 「まだ前フレームがない」ことを表し(StatsOverlay.ts の frame()
	// の lastTimestamp === null チェックを反映したもの)、これにより
	// 最初のフレームでとんでもなく大きい間隔値を記録してしまうことを
	// 防いでいる。
	lastDrawTime time.Time
}

var _ ebiten.Game = (*Game)(nil)

// NewGame は、sim を駆動し sprites を使って描画する Game を返す。
// measure が true の場合、統計オーバーレイ(porting-plan §5.6)が
// 有効になる: そのテキスト用 face の構築に faces が使われる(measure
// が false のときは faces は nil でよい)。
func NewGame(s *sim.Simulator, sprites *Sprites, faces *FaceCache, measure bool) *Game {
	g := &Game{sim: s, sprites: sprites, measure: measure}
	if measure {
		g.stats = debug.NewStats()
		g.overlayFace = faces.Face(overlayFontSize)
		m := g.overlayFace.Metrics()
		g.overlayLineSpacing = m.HAscent + m.HDescent + m.HLineGap
	}
	return g
}

// Update はシミュレーションを 1 ティック進める。何も描画してはならない
// (porting-plan §5.5)。measure が有効な場合は、Simulator.Update の
// 時間も計測し、統計オーバーレイ用にその所要時間を記録する。
func (g *Game) Update() error {
	if !g.measure {
		g.sim.Update()
		return nil
	}

	start := time.Now()
	g.sim.Update()
	g.stats.RecordUpdate(float64(time.Since(start)) / float64(time.Millisecond))
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

	buckets := g.sim.Buckets()
	for _, b := range buckets {
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
		// 以下で更新されるのは、オーバーレイ自身の計測状態(g.stats
		// 内のフレーム間隔の移動平均リング)のみである — これはシミュ
		// レーションの状態ではないので、「Draw はシミュレーションの
		// 状態を変更してはならない」というルール(porting-plan §5.5)
		// には違反しない。Draw はレンダリングされるフレームごとに
		// ちょうど 1 回呼ばれ、これは StatsOverlay.ts の rAF 駆動の
		// frame() が計測するのと同じ周期である。
		now := time.Now()
		if !g.lastDrawTime.IsZero() {
			interval := float64(now.Sub(g.lastDrawTime)) / float64(time.Millisecond)
			g.stats.RecordFrame(interval)
		}
		g.lastDrawTime = now

		g.drawOverlay(screen, buckets)
	}
}

// drawOverlay は計測用オーバーレイ(porting-plan §5.6)を描画する:
// 黒でアルファ約 50% の背景矩形の上に、水色のテキストで FPS、フレーム
// 間隔、Update の所要時間、Ebitengine 自身の ActualFPS/ActualTPS、
// 各粒子種類ごとのカウント(固定順、0 埋め — debug.FixedKinds を
// 参照)、そして合計を並べて表示する。buckets/g.sim から読み取るのみ
// で、シミュレーションの状態は変更しない。
func (g *Game) drawOverlay(screen *ebiten.Image, buckets []sim.Bucket) {
	counts := make(map[string]int, len(buckets))
	for _, b := range buckets {
		counts[string(b.Kind)] = len(b.Particles)
	}

	statsText := g.stats.Text(counts, ebiten.ActualFPS(), ebiten.ActualTPS())

	w, h := text.Measure(statsText, g.overlayFace, g.overlayLineSpacing)
	vector.FillRect(
		screen,
		0, 0,
		float32(w+2*overlayPadX), float32(h+2*overlayPadY),
		overlayBackground, false,
	)

	opts := &text.DrawOptions{}
	opts.LineSpacing = g.overlayLineSpacing
	opts.GeoM.Translate(overlayPadX, overlayPadY)
	opts.ColorScale.ScaleWithColor(overlayTextColor)
	text.Draw(screen, statsText, g.overlayFace, opts)
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
