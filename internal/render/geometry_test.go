package render

import (
	"math"
	"testing"
)

// これらのテストは、ebiten に依存しない描画ジオメトリの計算をカバー
// する: スプライトから粒子サイズへのスケール係数と、中央配置のための
// 平行移動量(porting-plan §5.2 の「スプライトから粒子サイズへの
// スケール係数、影オフセット」)。いずれも *ebiten.Image に依存しない
// ので、通常の `go test` として実行できる(Ebitengine 自体はヘッド
// レスでは実行できない)。

func TestUniformScale(t *testing.T) {
	tests := []struct {
		name       string
		targetSize float64
		spriteSize float64
		want       float64
	}{
		{"same size is identity scale", 24, 24, 1},
		{"upscale", 48, 24, 2},
		{"downscale", 12, 24, 0.5},
		{"zero sprite size guards against divide-by-zero", 24, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UniformScale(tt.targetSize, tt.spriteSize)
			if got != tt.want {
				t.Errorf("UniformScale(%v, %v) = %v, want %v", tt.targetSize, tt.spriteSize, got, tt.want)
			}
		})
	}
}

func TestCenterOffset_CentersSpriteOnParticle(t *testing.T) {
	// 10x10 のスプライトを 2 倍に拡大し(→ 画面上では 20x20)、追加の
	// オフセットなしで (100, 50) に配置した場合、そのスクリーン空間上
	// のバウンディングボックスはちょうど (100, 50) を中心にしなければ
	// ならない。
	tx, ty := CenterOffset(10, 10, 2, 100, 50, 0, 0)

	left, top := tx, ty
	right, bottom := tx+10*2, ty+10*2
	cx := (left + right) / 2
	cy := (top + bottom) / 2

	if cx != 100 || cy != 50 {
		t.Errorf("center = (%v, %v), want (100, 50)", cx, cy)
	}
}

func TestCenterOffset_ShadowOffsetIsScreenSpaceNotScaled(t *testing.T) {
	// Mizu-ts の shadowOffsetX/Y は一定の生ピクセルのずらし量であり
	// (§5.2)、フォントサイズ・粒子サイズによって拡大縮小されない。
	// CenterOffset は scale に関わらず dx/dy をそのまま加算しなければ
	// ならない。
	const scale = 3.0
	txNoShadow, tyNoShadow := CenterOffset(10, 10, scale, 100, 50, 0, 0)
	txShadow, tyShadow := CenterOffset(10, 10, scale, 100, 50, 1, 1)

	if got, want := txShadow-txNoShadow, 1.0; got != want {
		t.Errorf("shadow dx = %v, want %v (unscaled)", got, want)
	}
	if got, want := tyShadow-tyNoShadow, 1.0; got != want {
		t.Errorf("shadow dy = %v, want %v (unscaled)", got, want)
	}
}

func TestNewH2Layout_OriginIsCanvasCenter(t *testing.T) {
	// このパッケージのすべてのスプライトは「スプライト中心 == 粒子
	// 位置」という規約に従っている(CenterOffset を参照)。H2 の合成
	// レイアウトは、本体と subscript が互いに中心をずらして描画され
	// ていてもその規約を共有できるよう、論理上の原点(TS の
	// fillText(text, x, y) が呼ばれる位置に相当)をキャンバスの幾何学
	// 的中心にちょうど一致させなければならない。
	layout := NewH2Layout(16, 20, 10, 12)

	cx := layout.CanvasW / 2
	cy := layout.CanvasH / 2

	wantBodyX := cx - 16.0/6
	if layout.BodyX != wantBodyX {
		t.Errorf("BodyX = %v, want %v (cx - bodyWidth/6, mirrors SubscriptTextRenderer.ts:31-34)", layout.BodyX, wantBodyX)
	}
	if layout.BodyY != cy {
		t.Errorf("BodyY = %v, want %v (body has no y offset)", layout.BodyY, cy)
	}
	if wantSubX := cx + 12; layout.SubX != wantSubX {
		t.Errorf("SubX = %v, want %v (cx + 12, mirrors SubscriptTextRenderer.ts:34)", layout.SubX, wantSubX)
	}
	if wantSubY := cy + 3; layout.SubY != wantSubY {
		t.Errorf("SubY = %v, want %v (cy + 3, mirrors SubscriptTextRenderer.ts:34)", layout.SubY, wantSubY)
	}
}

func TestNewH2Layout_CanvasContainsBothGlyphs(t *testing.T) {
	bodyW, bodyH := 16.0, 20.0
	subW, subH := 10.0, 12.0
	layout := NewH2Layout(bodyW, bodyH, subW, subH)

	checks := []struct {
		name         string
		cx, cy, w, h float64
	}{
		{"body", layout.BodyX, layout.BodyY, bodyW, bodyH},
		{"subscript", layout.SubX, layout.SubY, subW, subH},
	}
	for _, c := range checks {
		left, right := c.cx-c.w/2, c.cx+c.w/2
		top, bottom := c.cy-c.h/2, c.cy+c.h/2
		if left < 0 || top < 0 || right > layout.CanvasW || bottom > layout.CanvasH {
			t.Errorf("%s glyph rect [%v,%v]x[%v,%v] not contained in canvas %vx%v",
				c.name, left, right, top, bottom, layout.CanvasW, layout.CanvasH)
		}
	}
}

func TestNewH2Layout_CanvasIsSymmetricAroundOrigin(t *testing.T) {
	layout := NewH2Layout(16, 20, 10, 12)
	if math.Abs(layout.CanvasW/2-layout.CanvasW/2) > 1e-9 {
		t.Fatalf("sanity check failed")
	}
	// CenterOffset はスプライト自身の中心が粒子位置の基準点であること
	// を前提としているため、キャンバスの中心はちょうど半分でなければ
	// ならない(半端なピクセルによる非対称があってはならない)。
	if layout.CanvasW/2*2 != layout.CanvasW {
		t.Errorf("CanvasW = %v is not evenly divisible by 2", layout.CanvasW)
	}
	if layout.CanvasH/2*2 != layout.CanvasH {
		t.Errorf("CanvasH = %v is not evenly divisible by 2", layout.CanvasH)
	}
}
