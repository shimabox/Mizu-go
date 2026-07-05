package sim

import (
	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/particle"
	"github.com/shimabox/Mizu-go/internal/physics"
	"github.com/shimabox/Mizu-go/internal/reaction"
)

// Bucket は同じ Kind の粒子をまとめたもので、レンダー層が kind 単位で
// 描画できるようにする(Mizu-ts の MizuSimulator.ts の renderBuckets に
// 対応。Simulator.Buckets を参照)。
type Bucket struct {
	Kind      core.Kind
	Particles []core.Particle
}

// Simulator はフレームパイプライン(update -> collide -> react ->
// sweep)を駆動し、Mizu-ts の MizuSimulator.ts に対応する。TS 版と異なり、
// 描画はこのパッケージの責務ではない(porting-plan §5.5)。Simulator が
// 公開するのは Update(パイプライン本体)と Buckets(レンダー層が描画する
// ための粒子のグループ化)のみである。
//
// Simulator は具体的な粒子 kind については何も知らない。粒子の生成は
// factory 経由、反応は registry 経由で行うのみであり、MizuSimulator.ts と
// 同じ関心の分離になっている。
type Simulator struct {
	bounds            *core.Bounds
	world             *World
	factory           *particle.Factory
	registry          *reaction.Registry
	collisionDetector physics.CollisionDetector
	grid              *physics.SpatialGrid

	// reactiveKinds は構築時に一度だけ計算される(TS コンストラクタの
	// `this.reactiveKinds = registry.reactiveKinds()` に対応)。
	// Registry.ReactiveKinds はライブビューではなくスナップショットの
	// コピーを返すため、ルールはすべて NewSimulator 呼び出し前に
	// 登録されていなければならない。
	reactiveKinds map[core.Kind]struct{}

	// reactiveBuf は Update 呼び出しをまたいで再利用され、毎フレーム
	// reactive kind 用スライスを再確保しないようにする(SpatialGrid の
	// バケツ再利用と同じ発想)。
	reactiveBuf []core.Particle

	// buckets と bucketIndex は Buckets 呼び出しをまたいで再利用される。
	// バケツのスライスは再確保ではなく長さを 0 に戻すことでクリアされ、
	// bucketIndex は各 kind のスロットを記憶することで、一度出現した
	// kind のバケツは(空になっても)残り続ける。これは
	// MizuSimulator.ts の renderBuckets の Map セマンティクスと一致する。
	buckets     []Bucket
	bucketIndex map[core.Kind]int
}

// NewSimulator は Simulator が必要とするコンポーネントを結び付ける。
// grid は collisionDetector が *physics.GridCollisionDetector である
// 場合に、その内部で使われているのと同じ *physics.SpatialGrid である。
// Simulator は Resize 時に grid を bounds と同期させるためだけにこれを
// 利用する。
func NewSimulator(
	bounds *core.Bounds,
	world *World,
	factory *particle.Factory,
	registry *reaction.Registry,
	collisionDetector physics.CollisionDetector,
	grid *physics.SpatialGrid,
) *Simulator {
	return &Simulator{
		bounds:            bounds,
		world:             world,
		factory:           factory,
		registry:          registry,
		collisionDetector: collisionDetector,
		grid:              grid,
		reactiveKinds:     registry.ReactiveKinds(),
		bucketIndex:       make(map[core.Kind]int),
	}
}

// Init は画面内のランダムな位置に hCount 個の H 原子と oCount 個の O 原子を
// 生成し、world に追加する。MizuSimulator.ts の init に対応する。
func (s *Simulator) Init(hCount, oCount int) {
	for i := 0; i < hCount; i++ {
		s.world.Add(s.factory.CreateHAtRandom())
	}
	for i := 0; i < oCount; i++ {
		s.world.Add(s.factory.CreateOAtRandom())
	}
}

// Update は 1 フレーム分を進める。すべての粒子を更新し、reactive な
// kind のみを対象に衝突を検出し、該当する反応ルールを適用し、死んだ粒子を
// sweep する。MizuSimulator.ts の renderFrame のステップ 1-4 に対応する
// (ステップ 5 の描画はレンダー層の仕事であり、Buckets を参照)。
func (s *Simulator) Update() {
	// 1. すべての粒子の位置・状態を更新する。
	for _, p := range s.world.All() {
		p.Update()
	}

	// 2. reactive な kind のみを対象に衝突判定を行う。登録済みルールに
	// 一度も現れない kind(H2o など)は除外される。これらは大量に
	// 蓄積しがちで、除外しなければグリッドへの挿入・近傍列挙を
	// 支配してしまうためである。
	reactive := s.reactiveBuf[:0]
	for _, p := range s.world.All() {
		if _, ok := s.reactiveKinds[p.Kind()]; ok {
			reactive = append(reactive, p)
		}
	}
	s.reactiveBuf = reactive

	pairs := s.collisionDetector.FindHitPairs(reactive)

	// 3. 反応を適用する。
	for _, pair := range pairs {
		a, b := pair.A, pair.B
		if a.IsDead() || b.IsDead() {
			continue // 同一フレーム内で同じ粒子が二重に反応するのを防ぐ
		}
		rule, ok := s.registry.Find(a.Kind(), b.Kind())
		if !ok {
			continue
		}
		result := rule.React(a, b)
		for _, c := range result.Consumed {
			c.MarkDead()
		}
		for _, p := range result.Produced {
			s.world.Add(p)
		}
	}

	// 4. 死んだ粒子を一度の走査でまとめて回収する。
	s.world.Sweep()
}

// Buckets は world にある現在の粒子を Kind ごとにグループ化し、各 kind が
// 最初に現れた順序(この呼び出し内だけでなく Simulator のライフタイム
// 全体を通じての順序)で返す。MizuSimulator.ts の renderBuckets による
// グループ化(renderFrame のステップ 5)に対応する。返されるスライスと
// その Particles スライスは呼び出しをまたいで再利用されるため、呼び出し側
// は次の Buckets/Update 呼び出し以降までこれらを保持してはならない。
func (s *Simulator) Buckets() []Bucket {
	for i := range s.buckets {
		s.buckets[i].Particles = s.buckets[i].Particles[:0]
	}
	for _, p := range s.world.All() {
		k := p.Kind()
		idx, ok := s.bucketIndex[k]
		if !ok {
			s.buckets = append(s.buckets, Bucket{Kind: k})
			idx = len(s.buckets) - 1
			s.bucketIndex[k] = idx
		}
		s.buckets[idx].Particles = append(s.buckets[idx].Particles, p)
	}
	return s.buckets
}

// CountScale は現在の bounds の幅に応じた粒子数のスケールを返し、
// MizuSimulator.ts の getScale() に対応する。これは particle.Factory の
// サイズスケール(1.0/1.2)とは別物であり、porting-plan §5.1 では
// この非対称性が意図的なものであると明記されている。
func (s *Simulator) CountScale() float64 {
	w := s.bounds.W()
	if w < 768 {
		return 1.0
	}
	if w < 1280 {
		return 1.2
	}
	return 1.5
}

// Resize は共有の bounds を更新し、新しい画面サイズに合わせて衝突用グリッド
// を再構築することで、両者をウィンドウのリサイズに同期させる
// (porting-plan §5.4。TS 版のキャンバスサイズは起動時に固定されるため、
// 対応する機能は存在しない)。
func (s *Simulator) Resize(w, h float64) {
	s.bounds.Set(w, h)
	s.grid.Resize(w, h)
}
