package core

import "math/rand/v2"

// Random は乱数生成のためのインターフェース。テストでは SeededRandom
// を注入することで behavior を決定的にでき、実行時には NewRandom を
// 使う。
type Random interface {
	// Next は [0, 1) の範囲の乱数値を返す。
	Next() float64
}

// mathRandom は math/rand/v2 を使った実行時の実装。
type mathRandom struct{}

// NewRandom は math/rand/v2 のデフォルトソースを使う Random を返す。
func NewRandom() Random {
	return mathRandom{}
}

func (mathRandom) Next() float64 {
	return rand.Float64()
}

// seededRandom はテスト用の決定的な実装。
type seededRandom struct {
	r *rand.Rand
}

// NewSeededRandom は同じシードに対して同じ乱数列を生成する Random を
// 返す。
func NewSeededRandom(seed uint64) Random {
	return &seededRandom{r: rand.New(rand.NewPCG(seed, seed))}
}

func (s *seededRandom) Next() float64 {
	return s.r.Float64()
}
