package behavior

import (
	"math"

	"github.com/shimabox/Mizu-go/internal/core"
)

// FallAndSway は H2o の「落下しながら揺れる」動きで、Mizu-ts の
// FallAndSway.ts から移植したもの。
type FallAndSway struct {
	bounds *core.Bounds
	size   float64 // 水滴の直径
	random core.Random
}

var _ MovementBehavior = (*FallAndSway)(nil)

// NewFallAndSway は指定サイズ(直径)の水滴用の FallAndSway を返す。
// bounds は現在の画面サイズへの共有参照であり、HasLanded は呼ばれる
// たびに bounds.H() を読むため、配線し直すことなくリサイズを反映できる。
func NewFallAndSway(bounds *core.Bounds, size float64, random core.Random) *FallAndSway {
	return &FallAndSway{bounds: bounds, size: size, random: random}
}

// Next は cos((y+dx)/100) で水平方向に揺れる(dx は最大5単位のランダム
// 性)とともに、垂直方向には size*0.1 ずつ落下する。
func (f *FallAndSway) Next(x, y float64) (nx, ny float64) {
	dx := f.random.Next() * 5
	nx = x + math.Cos((y+dx)/100)
	ny = y + f.size*0.1
	return nx, ny
}

// HasLanded は y が画面の底に到達したかどうかを返す。
func (f *FallAndSway) HasLanded(y float64) bool {
	return y >= f.bounds.H()
}
