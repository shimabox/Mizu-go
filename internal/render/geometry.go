package render

import "math"

// UniformScale は、本来の(起動時に生成されたままの、未スケーリングの)
// ピクセルサイズが spriteSize であるスプライトを、targetSize(粒子の
// 実際の直径、core.Particle.Radius()*2)で描画するために掛けるべき
// 倍率を返す。テキストおよび水滴のスプライトはいずれも、アスペクト比
// を保つため常に両軸に同じ倍率で拡大縮小される(porting-plan §5.2:
// スプライトは一度だけ生成され、以降は GeoM によるリサイズのみで、
// 再描画は行わない)。
func UniformScale(targetSize, spriteSize float64) float64 {
	if spriteSize == 0 {
		return 0
	}
	return targetSize / spriteSize
}

// CenterOffset は、spriteW x spriteH のスプライトを scale で均一に
// 拡大縮小して描画する際、そのスプライト自身の幾何学的中心がちょうど
// (x, y) に来るようにするために必要な GeoM の translate 値
// (スクリーン座標系でのスプライト左上角の位置)を返す。これは TS の
// レンダラーにおける textAlign='center'/textBaseline='middle' の
// 規約を反映したものである。このパッケージが生成するすべてのスプライト
// は「粒子中心」を自身の画像中心に置くので、この 1 つの関数だけで
// それら全部を中央基準で配置できる。
//
// dx, dy は拡大縮小の*後に*適用される、さらなるスクリーン空間上の
// ずらし量である(例えば porting-plan §5.2 の TextRenderer.ts の
// shadowOffsetX/Y による 1px の影オフセット)。Mizu-ts の canvas の
// 影オフセットはフォントサイズに応じて拡大縮小されない一定の生の
// ピクセル量なので、dx/dy は拡大縮小されない値のまま渡され、そのまま
// 加算される — 影を伴わない描画では 0, 0 を渡す。
func CenterOffset(spriteW, spriteH, scale, x, y, dx, dy float64) (tx, ty float64) {
	tx = x - (spriteW/2)*scale + dx
	ty = y - (spriteH/2)*scale + dy
	return tx, ty
}

// H2Layout は、H2 スプライトを合成する際に "H" 本体と "2" の
// subscript のグリフをどこに配置するか、および両方を収めつつ粒子中心
// の原点をキャンバスの幾何学的中心にちょうど一致させるために必要な
// キャンバスサイズを表す(CenterOffset を参照)。
type H2Layout struct {
	CanvasW, CanvasH float64
	BodyX, BodyY     float64 // 本体グリフの中心(キャンバス座標系)
	SubX, SubY       float64 // subscript グリフの中心(キャンバス座標系)
}

// h2SubOffsetX と h2SubOffsetY は、粒子の原点を基準とした subscript
// の位置であり、Mizu-ts の SubscriptTextRenderer.ts(fillText(subscript,
// x+12, y+3))を反映している。
const (
	h2SubOffsetX = 12.0
	h2SubOffsetY = 3.0
)

// NewH2Layout は、"H" 本体(bodyW x bodyH、本体フォントサイズ)と
// "2" の subscript(subW x subH、subscript フォントサイズ)を合成
// するためのレイアウトを計算し、Mizu-ts の
// SubscriptTextRenderer.ts:31-34 の相対オフセットを再現する: 本体は
// x-bodyWidth/6 の位置に(y はそのまま)、subscript は x+12, y+3 の
// 位置に、いずれも自身の位置を中心として center/middle 揃えで描画
// される。
func NewH2Layout(bodyW, bodyH, subW, subH float64) H2Layout {
	bodyOffsetX := -bodyW / 6.0

	left := math.Min(bodyOffsetX-bodyW/2, h2SubOffsetX-subW/2)
	right := math.Max(bodyOffsetX+bodyW/2, h2SubOffsetX+subW/2)
	top := math.Min(0-bodyH/2, h2SubOffsetY-subH/2)
	bottom := math.Max(0+bodyH/2, h2SubOffsetY+subH/2)

	// halfW/halfH は、キャンバスが原点を中心に対称であり続ける(つまり
	// 原点がキャンバスの幾何学的中心にちょうど一致する)よう、原点から
	// 端までの距離を*両側とも*包含するようにしなければならない。
	halfW := math.Max(math.Abs(left), right)
	halfH := math.Max(math.Abs(top), bottom)

	// 各辺に +1: 小さな AA 用の余白マージンを加えたうえで 2 倍にし、
	// キャンバスサイズが常にちょうど偶数ピクセルになるようにする
	// (原点がちょうど中心に来るために必要。
	// TestNewH2Layout_CanvasIsSymmetricAroundOrigin を参照)。
	canvasW := (math.Ceil(halfW) + 1) * 2
	canvasH := (math.Ceil(halfH) + 1) * 2

	cx := canvasW / 2
	cy := canvasH / 2

	return H2Layout{
		CanvasW: canvasW,
		CanvasH: canvasH,
		BodyX:   cx + bodyOffsetX,
		BodyY:   cy,
		SubX:    cx + h2SubOffsetX,
		SubY:    cy + h2SubOffsetY,
	}
}
