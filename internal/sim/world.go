package sim

import "github.com/shimabox/Mizu-go/internal/core"

// World は粒子の追加・削除と kind 別のカウントを管理し、Mizu-ts の
// World.ts に対応する。
type World struct {
	particles []core.Particle
}

// NewWorld は空の World を返す。
func NewWorld() *World {
	return &World{}
}

// Add は p を world に追加する。
func (w *World) Add(p core.Particle) {
	w.particles = append(w.particles, p)
}

// Count は現在保持している、生死を問わず指定した kind の粒子数を返す
// (単純な線形走査であり、World.ts の count() と同じ)。
func (w *World) Count(kind core.Kind) int {
	n := 0
	for _, p := range w.particles {
		if p.Kind() == kind {
			n++
		}
	}
	return n
}

// All は呼び出し側が読み取り専用で使うために現在の粒子を返す。返される
// スライスは World の内部配列を共有しているため、呼び出し側はこれを
// 変更したり、Sweep 呼び出しをまたいで保持したりしてはならない。
func (w *World) All() []core.Particle {
	return w.particles
}

// Sweep は死んだ粒子をすべて一度の走査で取り除き、生存粒子の相対的な順序
// を保つ(World.ts の `filter((p) => !p.isDead())` に対応するが、新しい
// スライスを確保するのではなくその場で詰め直す)。
func (w *World) Sweep() {
	survivors := w.particles[:0]
	for _, p := range w.particles {
		if !p.IsDead() {
			survivors = append(survivors, p)
		}
	}
	w.particles = survivors
}
