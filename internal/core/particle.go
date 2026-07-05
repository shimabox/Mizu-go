package core

// Kind は粒子の種類(例: "H", "H2", "O", "H2o")を識別する。TS 版の
// ParticleKind に倣って単純な string にしてあり、このパッケージを変更
// せずに新しい種類を追加できるようにしている。
type Kind string

// Particle は全ての粒子種類が共有する振る舞い。描画は意図的にこの
// インターフェースに含めていない。TS 版の render(ctx) とは異なり、Go
// の render 層は Kind をもとに粒子の描き方を引くようになっており、
// このパッケージを描画関連の関心事から切り離している。
type Particle interface {
	Kind() Kind

	X() float64
	Y() float64
	Radius() float64

	// Update は位置・状態を進める。描画を行ってはならない。
	Update()

	// IsDead は World がこの粒子を回収すべきかどうかを返す。
	IsDead() bool
	// MarkDead はこの粒子を回収対象としてマークする。
	MarkDead()
}
