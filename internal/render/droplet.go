package render

import (
	"image"
	"image/color"
	"math"
)

// dropletGradientOffsetFactor と dropletGradientRadiusFactor は
// Mizu-ts の DropletRenderer.ts を反映したもの。グラデーションの焦点は
// 円の中心から size*0.4 だけずらしてあり、その半径は size/2 + size*0.4
// になる。
const dropletGradientOffsetFactor = 0.4

// dropletGradientColor は、水滴の放射状グラデーション上で距離 t の位置
// にある(アルファ乗算済みの)色を返す。これは DropletRenderer.ts の
// 2 つのカラーストップを反映している。
//
//	stop 0: rgba(255, 255, 255, 0.6)
//	stop 1: rgba(0, 127, 255, 1)
//
// t は [0, 1] にクランプされる。これは canvas のグラデーションが定義
// 範囲外でも同様に振る舞うことに合わせたものである。
func dropletGradientColor(t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	const (
		r0, g0, b0, a0 = 255.0, 255.0, 255.0, 0.6
		r1, g1, b1, a1 = 0.0, 127.0, 255.0, 1.0
	)

	r := r0 + (r1-r0)*t
	g := g0 + (g1-g0)*t
	b := b0 + (b1-b0)*t
	a := a0 + (a1-a0)*t

	// color.RGBA は(color.NRGBA と異なり)アルファ乗算済みの形式なので、
	// 上で計算した素の r/g/b 成分は、格納する前にアルファを掛けて
	// スケーリングする必要がある。
	return color.RGBA{
		R: uint8(math.Round(r * a)),
		G: uint8(math.Round(g * a)),
		B: uint8(math.Round(b * a)),
		A: uint8(math.Round(a * 255)),
	}
}

// dropletPixelColor は、直径 size の水滴について、円の中心からの
// オフセット (dx, dy) にある水滴スプライトのピクセル色を返す。これは
// DropletRenderer.ts の createRadialGradient + arc(0, 2*pi) + fill を
// 反映したもの。半径 size/2 の円の外側にあるピクセルは完全に透明で、
// 内側にあるピクセルは(円の中心を基準に)(-size*0.4, -size*0.4) を
// 中心とし、半径 size/2 + size*0.4 の放射状グラデーションに従う。
func dropletPixelColor(dx, dy, size float64) color.RGBA {
	radius := size / 2
	if dx*dx+dy*dy > radius*radius {
		return color.RGBA{} // 完全に透明
	}

	offset := size * dropletGradientOffsetFactor
	gdx, gdy := dx+offset, dy+offset
	gradientRadius := radius + offset

	t := math.Hypot(gdx, gdy) / gradientRadius
	return dropletGradientColor(t)
}

// newDropletImage は、H2o の水滴グラデーションを含む diameter x
// diameter の *image.RGBA を描画する。これは起動時に一度だけ生成され、
// 後は各粒子の実際のサイズに合わせて GeoM でスケーリングされる
// (porting-plan §5.2)。ピクセル中心の規約により、円の中心は
// (diameter/2, diameter/2) に置かれる。
func newDropletImage(diameter int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, diameter, diameter))
	cx := float64(diameter) / 2
	cy := float64(diameter) / 2
	size := float64(diameter)

	for y := 0; y < diameter; y++ {
		for x := 0; x < diameter; x++ {
			// 円の縁で少しアンチエイリアスが効くよう、ピクセル中心で
			// サンプリングする。
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			img.SetRGBA(x, y, dropletPixelColor(dx, dy, size))
		}
	}
	return img
}
