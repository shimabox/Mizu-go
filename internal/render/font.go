package render

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// BaseFontSize と SubscriptFontSize は、スプライトを生成する際の
// フォントサイズ(scale 1.0)であり、Mizu-ts の baseFontSize() /
// SubscriptTextRenderer.ts の subscript フォントサイズに一致する
// (porting-plan §5.1、§5.2: 本体は 24 * scale、subscript は 18 *
// scale。スプライトは常に scale 1.0 で生成され、後から GeoM で
// リサイズされる)。
const (
	BaseFontSize      = 24.0
	SubscriptFontSize = 18.0
)

// NewFontSource は golang.org/x/image/font/gofont/goregular(Go 版に
// おける Mizu-ts の `sans-serif` の代替、porting-plan §5.2)を
// text/v2 の face source として読み込む。
func NewFontSource() (*text.GoTextFaceSource, error) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, fmt.Errorf("render: loading embedded font: %w", err)
	}
	return source, nil
}

// FaceCache は、フォントサイズごとに 1 つの *text.GoTextFace を
// 遅延生成してキャッシュする。単一の GoTextFaceSource を共有すること
// で、グリフのラスタライズ(これ自体は GoTextFaceSource がキャッシュ
// する)がこのパッケージで使うすべてのサイズ・種類にわたって共有
// される。これは Mizu-ts の ParticleFactory の measureTextWidth
// キャッシュを精神的に踏襲したものである。
type FaceCache struct {
	source *text.GoTextFaceSource
	faces  map[float64]*text.GoTextFace
}

// NewFaceCache は source を裏付けとする FaceCache を返す。
func NewFaceCache(source *text.GoTextFaceSource) *FaceCache {
	return &FaceCache{source: source, faces: make(map[float64]*text.GoTextFace)}
}

// Face は指定サイズの(キャッシュ済みの)face を返す。
func (c *FaceCache) Face(size float64) *text.GoTextFace {
	if f, ok := c.faces[size]; ok {
		return f
	}
	f := &text.GoTextFace{Source: c.source, Size: size}
	c.faces[size] = f
	return f
}

// Measure は text/v2 の Measure を使って particle.MeasureFunc を
// 実装する。起動時に particle.NewFactory へ注入され(cmd/mizu/main.go)、
// ロジック層が ebiten/text への依存を持たずに済むようにしている。
func (c *FaceCache) Measure(txt string, fontSize float64) float64 {
	w, _ := text.Measure(txt, c.Face(fontSize), 0)
	return w
}
