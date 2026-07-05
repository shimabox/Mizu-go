package render

import (
	"testing"
)

// dropletGradientColor、dropletPixelColor、newDropletImage は
// (*ebiten.Image を使わず image/color のみを扱う)純粋な関数であり、
// H2o の水滴スプライト(porting-plan §5.2、Mizu-ts の
// DropletRenderer.ts を反映)の放射状グラデーションの計算は、render
// パッケージの他の部分と違って通常の `go test` で実行できる。

func TestDropletGradientColor_AtStop0(t *testing.T) {
	// t=0 は DropletRenderer.ts の最初のグラデーションストップを
	// そのまま再現していなければならない: rgba(255, 255, 255, 0.6)、
	// アルファ乗算済み。
	c := dropletGradientColor(0)

	wantA := uint8(0.6 * 255) // 153
	if c.A != wantA {
		t.Errorf("A = %v, want %v", c.A, wantA)
	}
	if c.R != wantA || c.G != wantA || c.B != wantA {
		t.Errorf("RGB = (%v,%v,%v), want (%v,%v,%v) (premultiplied white at stop0)", c.R, c.G, c.B, wantA, wantA, wantA)
	}
}

func TestDropletGradientColor_AtStop1(t *testing.T) {
	// t=1 は DropletRenderer.ts の 2 つ目のグラデーションストップを
	// そのまま再現していなければならない: rgba(0, 127, 255, 1)。これは
	// すでに不透明なので、アルファ乗算は何もしないのと同じになる。
	c := dropletGradientColor(1)

	if c.R != 0 || c.G != 127 || c.B != 255 || c.A != 255 {
		t.Errorf("color = (%v,%v,%v,%v), want (0,127,255,255)", c.R, c.G, c.B, c.A)
	}
}

func TestDropletGradientColor_ClampsBeyondStop1(t *testing.T) {
	c := dropletGradientColor(1.5)
	want := dropletGradientColor(1)
	if c != want {
		t.Errorf("dropletGradientColor(1.5) = %+v, want clamped to dropletGradientColor(1) = %+v", c, want)
	}
}

func TestDropletGradientColor_ClampsBelowStop0(t *testing.T) {
	c := dropletGradientColor(-0.5)
	want := dropletGradientColor(0)
	if c != want {
		t.Errorf("dropletGradientColor(-0.5) = %+v, want clamped to dropletGradientColor(0) = %+v", c, want)
	}
}

func TestDropletPixelColor_OutsideCircleIsTransparent(t *testing.T) {
	const size = 64.0
	// (size, size) は原点を中心とする半径 size/2 の円から十分に外れて
	// いる。
	c := dropletPixelColor(size, size, size)
	if c.A != 0 {
		t.Errorf("outside-circle alpha = %v, want 0 (fully transparent)", c.A)
	}
}

func TestDropletPixelColor_InsideCircleIsNotTransparent(t *testing.T) {
	const size = 64.0
	r := size / 2
	// 円の縁のすぐ内側で、グラデーション中心の偏り(下の
	// "lighter near gradient center" テストを参照)からは離れた点なので、
	// 完全に不透明な外側のストップと同じか、それに近い値になるはず。
	c := dropletPixelColor(r*0.99, 0, size)
	if c.A == 0 {
		t.Fatalf("point inside circle must not be transparent")
	}
}

func TestDropletPixelColor_GradientCenterSideIsLighterThanOppositeSide(t *testing.T) {
	// TS の DropletRenderer は、グラデーション中心を円の中心から
	// (-size*0.4, -size*0.4) だけずらしており、ハイライトを左上寄りに
	// 偏らせている。円のその側にある点は、円の中心を挟んで反対側
	// (右下)に反転した点よりも明るく(stop0 に近く、より白く、より
	// 透明で)なければならない。
	const size = 64.0
	nearCenter := dropletPixelColor(-10, -10, size)
	farSide := dropletPixelColor(10, 10, size)

	if nearCenter.A == 0 || farSide.A == 0 {
		t.Fatalf("both sample points must be inside the circle (nearCenter.A=%v, farSide.A=%v)", nearCenter.A, farSide.A)
	}
	// "Lighter"(より明るい) = R が大きい(#007fff の R=0 から白の
	// R=255 に近づく方向)、かつアルファが小さい(0.6 ストップ対 1.0
	// ストップ)こと。
	if nearCenter.R <= farSide.R {
		t.Errorf("gradient-center-side point R = %v, opposite-side point R = %v; want the near side strictly lighter (higher R)", nearCenter.R, farSide.R)
	}
	if nearCenter.A >= farSide.A {
		t.Errorf("gradient-center-side point A = %v, opposite-side point A = %v; want the near side strictly more transparent (lower A)", nearCenter.A, farSide.A)
	}
}

func TestNewDropletImage_CenterPixelOpaqueEnough(t *testing.T) {
	img := newDropletImage(64)
	b := img.Bounds()
	if b.Dx() != 64 || b.Dy() != 64 {
		t.Fatalf("image size = %vx%v, want 64x64", b.Dx(), b.Dy())
	}
	cx, cy := b.Dx()/2, b.Dy()/2
	c := img.RGBAAt(cx, cy)
	if c.A == 0 {
		t.Errorf("center pixel is fully transparent, want inside the droplet circle")
	}
}

// maxH2oDiameterCSS と maxExpectedDeviceScaleFactor は、
// DropletBaseDiameter(sprites.go)が本当に「常に縮小のみで済む」大きさ
// になっているかどうかを検証するための、既知の worst case を表す定数
// である。
//
//   - maxH2oDiameterCSS = (10 + 18) * 1.2 = 33.6 は
//     internal/particle/factory.go の CreateH2o/scale() が生成しうる
//     直径の実際の最大値(CSS px)。
//   - maxExpectedDeviceScaleFactor = 3 は、高密度 Retina 相当の
//     DeviceScaleFactor として本パッケージが想定する上限
//     (porting-plan §5.3 の HiDPI 対応、game.go の dsf を参照)。
const (
	maxH2oDiameterCSS            = 33.6
	maxExpectedDeviceScaleFactor = 3.0
)

func TestDropletBaseDiameter_CoversWorstCaseDevicePixelSize(t *testing.T) {
	// DropletBaseDiameter は常に「スプライトの実ピクセル寸法 >= 実際に
	// 描画される直径(device px)」を満たさなければならない。これが
	// 崩れると GeoM のスケールが 1 を超え、ぼやけ・ジャギーが再発する。
	worstCaseDevicePixels := maxH2oDiameterCSS * maxExpectedDeviceScaleFactor
	if float64(DropletBaseDiameter) < worstCaseDevicePixels {
		t.Errorf(
			"DropletBaseDiameter = %v, want >= %v (= maxH2oDiameterCSS(%v) * maxExpectedDeviceScaleFactor(%v)) to avoid upscaling",
			DropletBaseDiameter, worstCaseDevicePixels, maxH2oDiameterCSS, maxExpectedDeviceScaleFactor,
		)
	}
}

func TestNewDropletImage_CornersAreTransparent(t *testing.T) {
	img := newDropletImage(64)
	b := img.Bounds()
	corners := [][2]int{
		{b.Min.X, b.Min.Y},
		{b.Max.X - 1, b.Min.Y},
		{b.Min.X, b.Max.Y - 1},
		{b.Max.X - 1, b.Max.Y - 1},
	}
	for _, c := range corners {
		got := img.RGBAAt(c[0], c[1])
		if got.A != 0 {
			t.Errorf("corner pixel (%v,%v) alpha = %v, want 0 (outside circle)", c[0], c[1], got.A)
		}
	}
}
