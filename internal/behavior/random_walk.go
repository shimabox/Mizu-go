package behavior

import (
	"math"

	"github.com/shimabox/Mizu-go/internal/core"
)

// RandomWalk のデフォルトのチューニング定数。Mizu-ts の RandomWalk.ts
// から移植したもの(H / H2 / O はこの動きを共有する)。
const (
	defaultSpeedFactor = 0.075
	defaultMaxSpeed    = 1.05
)

// RandomWalk は H・H2・O 粒子に共通のランダムウォーク的な動き。毎ステップ
// ランダムな方向へ加速し、速度を maxSpeed でクランプし、画面端を折り
// 返す(ラップアラウンド)。
type RandomWalk struct {
	bounds      *core.Bounds
	bodySize    float64
	random      core.Random
	speedFactor float64
	maxSpeed    float64

	vx, vy float64
}

var _ MovementBehavior = (*RandomWalk)(nil)

// NewRandomWalk は Mizu-ts のデフォルト値に合わせ、デフォルトの
// speed factor(0.075)と最大速度(1.05)を持つ RandomWalk を返す。
//
// bounds は現在の画面サイズへの共有参照であり、RandomWalk は Next が
// 呼ばれるたびに bounds.W()/H() を読むため、配線し直すことなくリサイズ
// を反映できる。bodySize は粒子の幅・高さで、ラップアラウンドの余白
// 計算に使う(TS 版の bodySize に対応)。
func NewRandomWalk(bounds *core.Bounds, bodySize float64, random core.Random) *RandomWalk {
	return NewRandomWalkWithParams(bounds, bodySize, random, defaultSpeedFactor, defaultMaxSpeed)
}

// NewRandomWalkWithParams は NewRandomWalk と同様だが speedFactor と
// maxSpeed を上書きできる。主な用途は、テストで速度を凍結
// (speedFactor=0)してラップアラウンドの境界ロジックだけを切り出して
// 検証できるようにするため。
func NewRandomWalkWithParams(bounds *core.Bounds, bodySize float64, random core.Random, speedFactor, maxSpeed float64) *RandomWalk {
	return &RandomWalk{
		bounds:      bounds,
		bodySize:    bodySize,
		random:      random,
		speedFactor: speedFactor,
		maxSpeed:    maxSpeed,
	}
}

// Next はランダムな角度の加速度で内部速度を進め、maxSpeed でクランプ
// したうえで (x, y) に適用し、結果を現在の画面範囲でラップアラウンド
// させる。
func (r *RandomWalk) Next(x, y float64) (nx, ny float64) {
	angle := 2 * math.Pi * r.random.Next()
	r.vx += r.speedFactor * math.Cos(angle)
	r.vy += r.speedFactor * math.Sin(angle)

	speed := math.Sqrt(r.vx*r.vx + r.vy*r.vy)
	if speed > r.maxSpeed {
		r.vx = (r.vx / speed) * r.maxSpeed
		r.vy = (r.vy / speed) * r.maxSpeed
	}

	nx = x + r.vx
	ny = y + r.vy

	w := r.bodySize
	h := r.bodySize
	sw := r.bounds.W()
	sh := r.bounds.H()

	if nx > sw+w/2 {
		nx = -(w / 2)
	}
	if nx+w < 0 {
		nx = sw + w/2
	}
	if ny > sh+h/2 {
		ny = -(h / 2)
	}
	if ny+h < 0 {
		ny = sh + h/2
	}

	return nx, ny
}
