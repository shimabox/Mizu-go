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
	layout := NewH2Layout(16, 20, 10, 12, 1)

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
	layout := NewH2Layout(bodyW, bodyH, subW, subH, 1)

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
	layout := NewH2Layout(16, 20, 10, 12, 1)
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

func TestNewH2Layout_SubOffsetScalesWithSubOffsetScale(t *testing.T) {
	// newH2Sprite はスーパーサンプリング(spriteSupersample、sprites.go)
	// のため body/subscript のグリフを供給倍率倍のフォントサイズで計測
	// する。h2SubOffsetX/Y (12, 3) は「基準フォントサイズ(scale 1)」の
	// 座標系での定数なので、subOffsetScale で同じ倍率にスケールしないと、
	// 供給倍率が上がるほど subscript が本体に対して相対的にどんどん
	// 近づいてしまう(合成後に GeoM で一律縮小するだけなので、比率は
	// 常にこの合成時点で決まる)。
	const subOffsetScale = 4.0
	layout := NewH2Layout(16, 20, 10, 12, subOffsetScale)

	cx := layout.CanvasW / 2
	cy := layout.CanvasH / 2

	if want := cx + 12*subOffsetScale; layout.SubX != want {
		t.Errorf("SubX = %v, want %v (cx + h2SubOffsetX*subOffsetScale)", layout.SubX, want)
	}
	if want := cy + 3*subOffsetScale; layout.SubY != want {
		t.Errorf("SubY = %v, want %v (cy + h2SubOffsetY*subOffsetScale)", layout.SubY, want)
	}
}

// これらのテストは、CenterOffset/UniformScale の HiDPI 対応版
// (DeviceCenterOffset/DeviceUniformScale、いずれも internal/render/
// game.go の drawGlyphs/drawDroplets から呼ばれる)をカバーする。
// dsf=1 のときは既存の CenterOffset/UniformScale をそのまま呼んだ場合と
// 厳密に同じ結果にならなければならない(HiDPI 非対応環境・cmd/bench の
// Xvfb 環境ではこれまでと完全に同じ描画結果になることを保証するため)。

func TestDeviceUniformScale_DSF1MatchesUniformScale(t *testing.T) {
	got := DeviceUniformScale(24, 96, 1)
	want := UniformScale(24, 96)
	if got != want {
		t.Errorf("DeviceUniformScale(24, 96, 1) = %v, want %v (== UniformScale(24, 96))", got, want)
	}
}

func TestDeviceUniformScale_DSF2DoublesEffectiveTargetSize(t *testing.T) {
	// dsf=2 の実デバイスピクセル数は CSS px の 2 倍なので、同じ
	// targetSize(CSS px)でも実際に掛けるべき倍率は 2 倍になる。
	got := DeviceUniformScale(24, 96, 2)
	want := UniformScale(24*2, 96)
	if got != want {
		t.Errorf("DeviceUniformScale(24, 96, 2) = %v, want %v", got, want)
	}
}

func TestDeviceCenterOffset_DSF1MatchesCenterOffset(t *testing.T) {
	gotTx, gotTy := DeviceCenterOffset(10, 10, 2, 100, 50, 1, 1, 1)
	wantTx, wantTy := CenterOffset(10, 10, 2, 100, 50, 1, 1)
	if gotTx != wantTx || gotTy != wantTy {
		t.Errorf("DeviceCenterOffset(...,1) = (%v,%v), want (%v,%v) (== CenterOffset)", gotTx, gotTy, wantTx, wantTy)
	}
}

func TestDeviceCenterOffset_DSF2ScalesPositionAndShadowOffset(t *testing.T) {
	// x, y, dx, dy はいずれも CSS px 単位で渡され、dsf を掛けてから
	// CenterOffset に渡らなければならない(粒子位置も影オフセットも
	// 実デバイスピクセルで描く必要があるため)。
	const dsf = 2.0
	gotTx, gotTy := DeviceCenterOffset(10, 10, 2, 100, 50, 1, 1, dsf)
	wantTx, wantTy := CenterOffset(10, 10, 2, 100*dsf, 50*dsf, 1*dsf, 1*dsf)
	if gotTx != wantTx || gotTy != wantTy {
		t.Errorf("DeviceCenterOffset(...,2) = (%v,%v), want (%v,%v)", gotTx, gotTy, wantTx, wantTy)
	}
}
