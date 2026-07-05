package particle

import (
	"image/color"

	"github.com/shimabox/Mizu-go/internal/behavior"
	"github.com/shimabox/Mizu-go/internal/core"
)

// MeasureFunc は指定されたフォントサイズにおけるテキストのピクセル幅を計測
// する。Mizu-ts の ParticleFactory.ts が使う Canvas の ctx.measureText に
// 相当する Go 版の関数。実行時にはレンダー層が text/v2 の Measure を注入
// し、テストでは決定的なスタブを注入する。
type MeasureFunc func(text string, fontSize float64) float64

// measureKey は measureTextWidth のキャッシュキーであり、(text, fontSize)
// の組。Mizu-ts の `${text}@${fontSize}` というキャッシュキー
// (ParticleFactory.ts:100-117) に対応する。
type measureKey struct {
	text     string
	fontSize float64
}

// Factory は H、H2、O、H2o の各粒子を生成し、Mizu-ts の ParticleFactory.ts
// と同じ方法で movement・size・color・乱数を組み立てる。measureTextWidth
// のキャッシュを保持しており、アプリのライフタイム中に登場する
// (text, fontSize) の組は数えるほどしかないため、それぞれ一度だけ計測すれば
// 済む。
type Factory struct {
	bounds  *core.Bounds
	random  core.Random
	measure MeasureFunc

	measureCache map[measureKey]float64
}

// NewFactory は bounds の範囲内に粒子を配置する Factory を返す。位置・色・
// サイズには random を、テキストベースの粒子 (H、H2、O) のサイズ決定には
// measure を用いる。bounds は生成呼び出しのたびに読み取られるため、リサイズ
// はそれ以降に生成される粒子のサイズスケール (porting plan の §5.2/§5.4)
// に反映されるが、既存の粒子には影響しない。
func NewFactory(bounds *core.Bounds, random core.Random, measure MeasureFunc) *Factory {
	return &Factory{
		bounds:       bounds,
		random:       random,
		measure:      measure,
		measureCache: make(map[measureKey]float64),
	}
}

// CreateH は (x, y) に H 原子を生成する。
func (f *Factory) CreateH(x, y float64) core.Particle {
	size := f.measureTextWidth("H")
	return newAtom(KindH, x, y, size/2, f.newRandomWalk(size), f.randomColor())
}

// CreateHAtRandom は画面内のランダムな位置に H 原子を生成する。
func (f *Factory) CreateHAtRandom() core.Particle {
	return f.CreateH(f.randomX(), f.randomY())
}

// CreateH2 は (x, y) に H2 分子を生成する。そのサイズは "H2" という文字列
// 全体から計測され、ParticleFactory.ts の createH2 と一致する (下付き文字
// の描画に関する詳細は Phase 5 の関心事であり、このパッケージの関心事では
// ない)。
func (f *Factory) CreateH2(x, y float64) core.Particle {
	size := f.measureTextWidth("H2")
	return newAtom(KindH2, x, y, size/2, f.newRandomWalk(size), f.randomColor())
}

// CreateO は (x, y) に O 原子を生成する。
func (f *Factory) CreateO(x, y float64) core.Particle {
	size := f.measureTextWidth("O")
	return newAtom(KindO, x, y, size/2, f.newRandomWalk(size), f.randomColor())
}

// CreateOAtRandom は画面内のランダムな位置に O 原子を生成する。
func (f *Factory) CreateOAtRandom() core.Particle {
	return f.CreateO(f.randomX(), f.randomY())
}

// CreateH2o は (x, y) に H2o の水滴を生成する。その直径は
// (random*10 + 18) * scale であり、ParticleFactory.ts の createH2o と
// 一致する。
func (f *Factory) CreateH2o(x, y float64) core.Particle {
	size := (f.random.Next()*10 + 18) * f.scale()
	movement := behavior.NewFallAndSway(f.bounds, size, f.random)
	return newH2o(x, y, size/2, movement)
}

func (f *Factory) newRandomWalk(bodySize float64) *behavior.RandomWalk {
	return behavior.NewRandomWalk(f.bounds, bodySize, f.random)
}

// randomX と randomY は、*現在の* bounds を基準にした画面内のランダムな
// 座標を生成する (ParticleFactory.ts の `this.sw * this.random.next()`
// に対応する)。
func (f *Factory) randomX() float64 { return f.bounds.W() * f.random.Next() }
func (f *Factory) randomY() float64 { return f.bounds.H() * f.random.Next() }

// randomColor は各チャンネルにつき random.Next() を 1 回呼び出すことで、
// 一様ランダムな不透明 RGB カラーを選ぶ。これは Mizu-ts の
// `#${random.next().toString(16).slice(-6)}` という 16 進文字列による
// ハックを意図的に置き換えたものである (忠実な移植ではなく、porting-plan
// の §5.2 で math/rand/v2 への移行によって可能になった意図的な簡略化として
// 言及されている)。
func (f *Factory) randomColor() color.RGBA {
	return color.RGBA{
		R: uint8(f.random.Next() * 256),
		G: uint8(f.random.Next() * 256),
		B: uint8(f.random.Next() * 256),
		A: 0xff,
	}
}

// baseFontSize は 24 * scale であり、ParticleFactory.ts の
// baseFontSize() と一致する。
func (f *Factory) baseFontSize() float64 {
	return 24 * f.scale()
}

// scale は粒子の *サイズ* のスケールである (シミュレータ側が持つ粒子の
// *個数* のスケールとは別物で、それは internal/sim の責務でありこの
// パッケージの責務ではない — porting-plan の §5.1 でこの非対称性が言及
// されている)。
func (f *Factory) scale() float64 {
	if f.bounds.W() < 768 {
		return 1.0
	}
	return 1.2
}

// measureTextWidth は現在の基準フォントサイズでテキストを計測し、結果を
// (text, fontSize) の組ごとにキャッシュする。
func (f *Factory) measureTextWidth(text string) float64 {
	key := measureKey{text: text, fontSize: f.baseFontSize()}
	if w, ok := f.measureCache[key]; ok {
		return w
	}
	w := f.measure(text, key.fontSize)
	f.measureCache[key] = w
	return w
}
