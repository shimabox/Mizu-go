package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// spriteSupersample は、H/O/H2 のテキストスプライトを実際の想定描画
// サイズより何倍大きい解像度でラスタライズしておくかを表す供給倍率
// である(Ebitengine 公式 examples/highdpi と同じ定石: スプライトは
// 常に「縮小方向」にのみ拡大縮小し、拡大は絶対に発生させない)。
//
// 本体フォントの実際の最大サイズは BaseFontSize * 1.2 = 28.8 CSS px
// (internal/particle/factory.go の scale()、幅 768px 以上の画面)。
// DeviceScaleFactor が最大 3(高密度 Retina 相当)になっても
// 28.8 * 3 = 86.4 device px までは常に縮小のみで済むよう、
// BaseFontSize(24) の 4 倍 = 96px でラスタライズする。
const spriteSupersample = 4.0

// DropletBaseDiameter は、H2o の水滴スプライトを生成する際のピクセル
// 直径である(porting-plan §5.2 に従い、画面上のどの水滴よりも大きい
// サイズにすることで、常に縮小のみで済み拡大は発生しないようにして
// いる)。
//
// CreateH2o が生成する直径の実際の最大値は (10+18) * 1.2 = 33.6 CSS px
// (internal/particle/factory.go の scale() と CreateH2o を参照)。
// DeviceScaleFactor が最大 3(高密度 Retina 相当)になっても
// 33.6 * 3 = 100.8 device px までは常に縮小のみで済むよう、余裕を
// 持って 128px にしている。実際の水滴の半径の想定される上限について
// は physics.MaxParticleRadius も参照。
const DropletBaseDiameter = 128

// spriteAAPadding は、グリフをラスタライズする前に、厳密に計測した
// 寸法の周囲に加える小さな余白(px 単位)である。これにより
// text.Draw のアンチエイリアスされた縁が、画像の境界でクリップされる
// ことなく描画できる余地を確保する。
const spriteAAPadding = 2

// Sprites は、このパッケージが毎フレーム描画元として使う、起動時に
// あらかじめ生成された画像を保持する(porting-plan §5.2: 「毎フレーム
// text.Draw や vector 描画はしない」)。H、O、H2 は白で描画されて
// おり、描画層が粒子ごとに ColorScale で色を付けられるようになって
// いる。Droplet はすでにグラデーションの色が焼き込まれている。
type Sprites struct {
	H, O, H2 *ebiten.Image
	Droplet  *ebiten.Image
}

// NewSprites は、このパッケージが描画するすべてのスプライトを、faces
// を使ったテキストの計測・描画によって一度だけ生成する。これは
// ebiten.RunGame が始まる前に呼び出さなければならない(ebiten.NewImage
// は RunGame 開始前でも使用できる)。
func NewSprites(faces *FaceCache) *Sprites {
	bodyFace := faces.Face(BaseFontSize * spriteSupersample)
	return &Sprites{
		H:       newGlyphSprite(bodyFace, "H"),
		O:       newGlyphSprite(bodyFace, "O"),
		H2:      newH2Sprite(faces),
		Droplet: ebiten.NewImageFromImage(newDropletImage(DropletBaseDiameter)),
	}
}

// newGlyphSprite は、s(単一のグリフ、"H" または "O")を白で、計測した
// 寸法ぴったりに切り詰めて描画する。center/middle 揃えにすることで、
// 自身の画像中心が粒子位置の基準点になる(CenterOffset を参照)。
func newGlyphSprite(face *text.GoTextFace, s string) *ebiten.Image {
	w, h := text.Measure(s, face, 0)
	imgW := int(math.Ceil(w)) + spriteAAPadding
	imgH := int(math.Ceil(h)) + spriteAAPadding
	if imgW < 1 {
		imgW = 1
	}
	if imgH < 1 {
		imgH = 1
	}

	img := ebiten.NewImage(imgW, imgH)
	drawGlyph(img, s, face, float64(imgW)/2, float64(imgH)/2)
	return img
}

// newH2Sprite は、"H" 本体(BaseFontSize)と "2" の subscript
// (SubscriptFontSize)を白で合成描画する。NewH2Layout を使って
// Mizu-ts の SubscriptTextRenderer.ts:31-34 の相対配置を再現しつつ、
// 粒子位置の基準点を合成画像自身の中心に保つ。spriteSupersample 倍の
// フォントサイズで計測・描画するため、NewH2Layout の subOffsetScale
// にも同じ倍率を渡し、本体に対する subscript の相対位置が供給倍率に
// よって崩れないようにする。
func newH2Sprite(faces *FaceCache) *ebiten.Image {
	bodyFace := faces.Face(BaseFontSize * spriteSupersample)
	subFace := faces.Face(SubscriptFontSize * spriteSupersample)

	bodyW, bodyH := text.Measure("H", bodyFace, 0)
	subW, subH := text.Measure("2", subFace, 0)
	layout := NewH2Layout(bodyW, bodyH, subW, subH, spriteSupersample)

	img := ebiten.NewImage(int(math.Ceil(layout.CanvasW)), int(math.Ceil(layout.CanvasH)))
	drawGlyph(img, "H", bodyFace, layout.BodyX, layout.BodyY)
	drawGlyph(img, "2", subFace, layout.SubX, layout.SubY)
	return img
}

// drawGlyph は、s を白で、dst 上の (x, y) を中心に描画する。
func drawGlyph(dst *ebiten.Image, s string, face *text.GoTextFace, x, y float64) {
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.PrimaryAlign = text.AlignCenter
	opts.SecondaryAlign = text.AlignCenter
	opts.ColorScale.ScaleWithColor(color.White)
	text.Draw(dst, s, face, opts)
}
