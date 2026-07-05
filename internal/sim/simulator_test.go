package sim

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/particle"
	"github.com/shimabox/Mizu-go/internal/physics"
	"github.com/shimabox/Mizu-go/internal/reaction"
)

// testMeasure は text/v2 の Measure の代わりとなる、決定的で依存のない
// スタブである(internal/particle 自身のテストスタブに対応する)。幅は
// テキスト長とフォントサイズに比例するだけで、このテスト群にはそれで十分。
func testMeasure(text string, fontSize float64) float64 {
	return float64(len(text)) * fontSize / 2
}

// newTestSimulator は実物のコンポーネント(実際の Factory、両ルールを
// 備えた実際の Registry、実際の GridCollisionDetector)を組み合わせる。
// これは cmd/mizu が実行時に simulator を組み立てる方法に合わせている。
func newTestSimulator(w, h float64, seed uint64) (*Simulator, *World, *particle.Factory, *core.Bounds) {
	bounds := core.NewBounds(w, h)
	random := core.NewSeededRandom(seed)
	factory := particle.NewFactory(bounds, random, testMeasure)

	registry := reaction.NewRegistry()
	registry.Register(&reaction.HHFusion{
		NewH2:        factory.CreateH2,
		NewHAtRandom: factory.CreateHAtRandom,
	})
	registry.Register(&reaction.OxidationToWater{
		NewOAtRandom: factory.CreateOAtRandom,
		NewHAtRandom: factory.CreateHAtRandom,
		NewH2o:       factory.CreateH2o,
	})

	world := NewWorld()
	grid := physics.NewSpatialGrid(w, h, physics.DefaultCellSize)
	detector := physics.NewGridCollisionDetector(grid)

	sim := NewSimulator(bounds, world, factory, registry, detector, grid)
	return sim, world, factory, bounds
}

// TestSimulator_InitCreatesRequestedCounts は MizuSimulator.test.ts の
// "初期化時に指定された数の H と O が生成されること" に対応する。
func TestSimulator_InitCreatesRequestedCounts(t *testing.T) {
	sim, world, _, _ := newTestSimulator(800, 600, 42)
	sim.Init(10, 5)

	if got := world.Count(particle.KindH); got != 10 {
		t.Errorf("Count(H) = %d, want 10", got)
	}
	if got := world.Count(particle.KindO); got != 5 {
		t.Errorf("Count(O) = %d, want 5", got)
	}
	if got := world.Count(particle.KindH2); got != 0 {
		t.Errorf("Count(H2) = %d, want 0", got)
	}
	if got := world.Count(particle.KindH2o); got != 0 {
		t.Errorf("Count(H2o) = %d, want 0", got)
	}
}

// TestSimulator_InitPlacesParticlesOnScreen は "H と O がランダムな座標
// (画面内)で初期化されること" に対応する。
func TestSimulator_InitPlacesParticlesOnScreen(t *testing.T) {
	const sw, sh = 800.0, 600.0
	sim, world, _, _ := newTestSimulator(sw, sh, 42)
	sim.Init(5, 5)

	for _, p := range world.All() {
		if p.X() < 0 || p.X() > sw {
			t.Errorf("particle x = %v, want within [0, %v]", p.X(), sw)
		}
		if p.Y() < 0 || p.Y() > sh {
			t.Errorf("particle y = %v, want within [0, %v]", p.Y(), sh)
		}
	}
}

// TestSimulator_UpdateMovesParticles は "フレーム描画時に H と O が
// 移動すること" に対応する。
func TestSimulator_UpdateMovesParticles(t *testing.T) {
	sim, world, factory, _ := newTestSimulator(800, 600, 42)
	world.Add(factory.CreateH(100, 100))
	world.Add(factory.CreateO(400, 400))

	sim.Update()

	all := world.All()
	h, o := all[0], all[1]
	if h.X() == 100 && h.Y() == 100 {
		t.Errorf("H did not move")
	}
	if o.X() == 400 && o.Y() == 400 {
		t.Errorf("O did not move")
	}
}

// TestSimulator_HHFusion は "H 同士の衝突で H: -1, H2: +1 となること" に
// 対応する。
func TestSimulator_HHFusion(t *testing.T) {
	sim, world, factory, _ := newTestSimulator(800, 600, 42)
	world.Add(factory.CreateH(100, 100))
	world.Add(factory.CreateH(105, 105))

	sim.Update()

	if got := world.Count(particle.KindH); got != 1 {
		t.Errorf("Count(H) = %d, want 1", got)
	}
	if got := world.Count(particle.KindH2); got != 1 {
		t.Errorf("Count(H2) = %d, want 1", got)
	}
}

// TestSimulator_NoMultipleReactionsPerFrame は "同一フレームで同じ粒子が
// 多重反応しないこと(H 3 つが密集)" に対応する。
func TestSimulator_NoMultipleReactionsPerFrame(t *testing.T) {
	sim, world, factory, _ := newTestSimulator(800, 600, 42)
	world.Add(factory.CreateH(100, 100))
	world.Add(factory.CreateH(103, 103))
	world.Add(factory.CreateH(106, 106))

	sim.Update()

	if got := world.Count(particle.KindH); got != 2 {
		t.Errorf("Count(H) = %d, want 2", got)
	}
	if got := world.Count(particle.KindH2); got != 1 {
		t.Errorf("Count(H2) = %d, want 1", got)
	}
}

// TestSimulator_OxidationToWater は "O と H2 の衝突で O: ±0, H2: -1,
// H: +1, H2o: +1 となること" に対応する。
func TestSimulator_OxidationToWater(t *testing.T) {
	sim, world, factory, _ := newTestSimulator(800, 600, 42)
	world.Add(factory.CreateO(100, 100))
	world.Add(factory.CreateH2(105, 105))

	sim.Update()

	if got := world.Count(particle.KindO); got != 1 {
		t.Errorf("Count(O) = %d, want 1", got)
	}
	if got := world.Count(particle.KindH2); got != 0 {
		t.Errorf("Count(H2) = %d, want 0", got)
	}
	if got := world.Count(particle.KindH); got != 1 {
		t.Errorf("Count(H) = %d, want 1", got)
	}
	if got := world.Count(particle.KindH2o); got != 1 {
		t.Errorf("Count(H2o) = %d, want 1", got)
	}
}

// TestSimulator_H2oLandsAndDisappears は "H2o は落下し、画面下端に達した
// フレームの後で World から消えていること" に対応する。
func TestSimulator_H2oLandsAndDisappears(t *testing.T) {
	const sh = 600.0
	sim, world, factory, _ := newTestSimulator(800, sh, 42)
	world.Add(factory.CreateH2o(400, sh))

	sim.Update()

	if got := world.Count(particle.KindH2o); got != 0 {
		t.Errorf("Count(H2o) = %d, want 0", got)
	}
	if got := len(world.All()); got != 0 {
		t.Errorf("len(All()) = %d, want 0", got)
	}
}

// TestSimulator_H2oKeepsFallingWhileOnScreen は "H2o は画面内にいる間は
// 消えず、落下し続けること" に対応する。
func TestSimulator_H2oKeepsFallingWhileOnScreen(t *testing.T) {
	sim, world, factory, _ := newTestSimulator(800, 600, 42)
	world.Add(factory.CreateH2o(400, 100))

	sim.Update()

	if got := world.Count(particle.KindH2o); got != 1 {
		t.Errorf("Count(H2o) = %d, want 1", got)
	}
	if got := world.All()[0].Y(); got <= 100 {
		t.Errorf("H2o y = %v, want > 100", got)
	}
}

// TestSimulator_SmokeManyFrames は "多数フレーム実行してもエラーが出ない
// こと(スモーク)" に対応する。O の数はフレームをまたいで不変でなければ
// ならない。なぜなら OxidationToWater は消費した O を必ず新しい O で
// 置き換えるため(差し引きゼロ)、実際に何回反応が発生したかに
// かかわらず総数は変わらないからである。
func TestSimulator_SmokeManyFrames(t *testing.T) {
	sim, world, _, _ := newTestSimulator(800, 600, 42)
	sim.Init(10, 10)

	for i := 0; i < 100; i++ {
		sim.Update()
	}

	if got := world.Count(particle.KindO); got != 10 {
		t.Errorf("Count(O) after 100 frames = %d, want 10", got)
	}
}

// fakeCollisionDetector はチェックを依頼された粒子を記録する。
// MizuSimulator.test.ts のインライン fakeDetector に対応する。
type fakeCollisionDetector struct {
	seen [][]core.Particle
}

func (d *fakeCollisionDetector) FindHitPairs(particles []core.Particle) []physics.Pair {
	cp := make([]core.Particle, len(particles))
	copy(cp, particles)
	d.seen = append(d.seen, cp)
	return nil
}

// TestSimulator_H2oExcludedFromCollisionDetection は "反応ルールに関与
// しない kind(H2o)は衝突判定に渡されないこと" に対応する。
func TestSimulator_H2oExcludedFromCollisionDetection(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(42)
	factory := particle.NewFactory(bounds, random, testMeasure)

	registry := reaction.NewRegistry()
	registry.Register(&reaction.HHFusion{
		NewH2:        factory.CreateH2,
		NewHAtRandom: factory.CreateHAtRandom,
	})
	registry.Register(&reaction.OxidationToWater{
		NewOAtRandom: factory.CreateOAtRandom,
		NewHAtRandom: factory.CreateHAtRandom,
		NewH2o:       factory.CreateH2o,
	})

	world := NewWorld()
	grid := physics.NewSpatialGrid(800, 600, physics.DefaultCellSize)
	fake := &fakeCollisionDetector{}
	sim := NewSimulator(bounds, world, factory, registry, fake, grid)

	world.Add(factory.CreateH(100, 100))
	world.Add(factory.CreateO(300, 300))
	world.Add(factory.CreateH2(500, 100))
	world.Add(factory.CreateH2o(400, 100))

	sim.Update()

	if len(fake.seen) != 1 {
		t.Fatalf("FindHitPairs called %d times, want 1", len(fake.seen))
	}
	kinds := make(map[core.Kind]bool)
	for _, p := range fake.seen[0] {
		kinds[p.Kind()] = true
	}
	if !kinds[particle.KindH] || !kinds[particle.KindO] || !kinds[particle.KindH2] {
		t.Errorf("kinds seen = %v, want H, O, H2 present", kinds)
	}
	if kinds[particle.KindH2o] {
		t.Errorf("H2o must not be passed to collision detection")
	}
}

// TestSimulator_BucketsGroupByFirstSeenKind は "描画が kind ごとにまと
// まって実行されること(world 配列順で交互に混ざらない)" に対応する。
func TestSimulator_BucketsGroupByFirstSeenKind(t *testing.T) {
	sim, world, _, _ := newTestSimulator(800, 600, 42)

	// kind が交互になる順序(A,B,A,B,A)で追加する。
	world.Add(newFakeParticle("A", 10, 10, 5))
	world.Add(newFakeParticle("B", 100, 100, 5))
	world.Add(newFakeParticle("A", 200, 200, 5))
	world.Add(newFakeParticle("B", 300, 300, 5))
	world.Add(newFakeParticle("A", 400, 400, 5))

	sim.Update()
	buckets := sim.Buckets()

	if len(buckets) != 2 {
		t.Fatalf("len(Buckets()) = %d, want 2", len(buckets))
	}
	if buckets[0].Kind != "A" || len(buckets[0].Particles) != 3 {
		t.Errorf("buckets[0] = %+v, want kind A with 3 particles", buckets[0])
	}
	if buckets[1].Kind != "B" || len(buckets[1].Particles) != 2 {
		t.Errorf("buckets[1] = %+v, want kind B with 2 particles", buckets[1])
	}
}

// TestSimulator_CountScale は "画面サイズによって scale が正しい値を
// 返すこと"(getScale() のしきい値)に対応する。
func TestSimulator_CountScale(t *testing.T) {
	cases := []struct {
		width float64
		want  float64
	}{
		{767, 1.0},
		{768, 1.2},
		{1279, 1.2},
		{1280, 1.5},
	}
	for _, c := range cases {
		sim, _, _, _ := newTestSimulator(c.width, 600, 42)
		if got := sim.CountScale(); got != c.want {
			t.Errorf("CountScale() at width %v = %v, want %v", c.width, got, c.want)
		}
	}
}

// TestSimulator_ResizeUpdatesBoundsAndGrid は、Resize が共有の Bounds
// (CountScale と factory が生成する粒子から観測できる)と SpatialGrid
// (リサイズ後も衝突判定が動作し続けることから観測できる)の両方に
// 反映されることを確認する。これに直接対応する TS 版の機能はない
// (porting-plan §5.4 のリサイズ対応は Go 版のみの追加機能である)。
func TestSimulator_ResizeUpdatesBoundsAndGrid(t *testing.T) {
	sim, world, factory, bounds := newTestSimulator(800, 600, 42)

	sim.Resize(1000, 700)

	if bounds.W() != 1000 || bounds.H() != 700 {
		t.Errorf("bounds after Resize = (%v, %v), want (1000, 700)", bounds.W(), bounds.H())
	}
	if got := sim.CountScale(); got != 1.2 {
		t.Errorf("CountScale() after Resize = %v, want 1.2", got)
	}

	// リサイズ後に生成された粒子は新しい bounds 内に収まらなければならない。
	p := factory.CreateHAtRandom()
	if p.X() < 0 || p.X() > 1000 || p.Y() < 0 || p.Y() > 700 {
		t.Errorf("particle created after Resize = (%v, %v), want within [0,1000]x[0,700]", p.X(), p.Y())
	}

	// Resize 後もグリッドは正しく衝突を検出できなければならない
	// (grid.Resize が bounds を変えただけで buckets/cols/rows を
	// 再構築していない、といった不完全な状態を残していないことの証明)。
	world.Add(factory.CreateH(500, 500))
	world.Add(factory.CreateH(505, 505))
	sim.Update()

	if got := world.Count(particle.KindH2); got != 1 {
		t.Errorf("Count(H2) after resize+collision = %d, want 1", got)
	}
}
