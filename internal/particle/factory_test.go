package particle

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

// newMeasureStub は text/v2 の Measure に代わる、決定的で依存関係のない
// スタブを返す。幅はテキストの長さとフォントサイズに比例するだけで、これら
// のテストにはそれで十分である (正確なグリフの計測値はレンダー層の関心事
// であり、このパッケージの関心事ではない)。calls は呼び出し回数を数える
// ため、キャッシュの挙動をアサートできる。
func newMeasureStub() (MeasureFunc, *int) {
	calls := 0
	return func(text string, fontSize float64) float64 {
		calls++
		return float64(len(text)) * fontSize / 2
	}, &calls
}

func TestFactory_CreateAtGivenCoordinates(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(42)
	measure, _ := newMeasureStub()
	f := NewFactory(bounds, random, measure)

	if h := f.CreateH(10, 20); h.Kind() != KindH || h.X() != 10 || h.Y() != 20 {
		t.Errorf("CreateH(10,20): kind=%v x=%v y=%v", h.Kind(), h.X(), h.Y())
	}
	if h2 := f.CreateH2(30, 40); h2.Kind() != KindH2 || h2.X() != 30 || h2.Y() != 40 {
		t.Errorf("CreateH2(30,40): kind=%v x=%v y=%v", h2.Kind(), h2.X(), h2.Y())
	}
	if o := f.CreateO(50, 60); o.Kind() != KindO || o.X() != 50 || o.Y() != 60 {
		t.Errorf("CreateO(50,60): kind=%v x=%v y=%v", o.Kind(), o.X(), o.Y())
	}
	if h2o := f.CreateH2o(70, 80); h2o.Kind() != KindH2o || h2o.X() != 70 || h2o.Y() != 80 {
		t.Errorf("CreateH2o(70,80): kind=%v x=%v y=%v", h2o.Kind(), h2o.X(), h2o.Y())
	}
}

func TestFactory_RadiusIsHalfMeasuredSize(t *testing.T) {
	bounds := core.NewBounds(800, 600) // 幅 >= 768 -> scale 1.2、fontSize 28.8
	random := core.NewSeededRandom(1)
	measure, _ := newMeasureStub()
	f := NewFactory(bounds, random, measure)

	const fontSize = 24 * 1.2
	const epsilon = 1e-9

	h := f.CreateH(0, 0)
	wantH := (1 * fontSize / 2) / 2
	if diff := h.Radius() - wantH; diff > epsilon || diff < -epsilon {
		t.Errorf("H radius = %v, want %v", h.Radius(), wantH)
	}

	h2 := f.CreateH2(0, 0)
	wantH2 := (2 * fontSize / 2) / 2 // "H2" という文字列全体として計測される
	if diff := h2.Radius() - wantH2; diff > epsilon || diff < -epsilon {
		t.Errorf("H2 radius = %v, want %v", h2.Radius(), wantH2)
	}

	o := f.CreateO(0, 0)
	wantO := (1 * fontSize / 2) / 2
	if diff := o.Radius() - wantO; diff > epsilon || diff < -epsilon {
		t.Errorf("O radius = %v, want %v", o.Radius(), wantO)
	}
}

func TestFactory_AtRandom_WithinBounds(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(3)
	measure, _ := newMeasureStub()
	f := NewFactory(bounds, random, measure)

	for i := 0; i < 20; i++ {
		h := f.CreateHAtRandom()
		if h.X() < 0 || h.X() >= 800 || h.Y() < 0 || h.Y() >= 600 {
			t.Fatalf("H at random out of bounds: x=%v y=%v", h.X(), h.Y())
		}

		o := f.CreateOAtRandom()
		if o.X() < 0 || o.X() >= 800 || o.Y() < 0 || o.Y() >= 600 {
			t.Fatalf("O at random out of bounds: x=%v y=%v", o.X(), o.Y())
		}
	}
}

func TestFactory_SizeScale_NarrowVsWide(t *testing.T) {
	measure, _ := newMeasureStub()

	narrow := NewFactory(core.NewBounds(767, 600), core.NewSeededRandom(1), measure)
	wide := NewFactory(core.NewBounds(768, 600), core.NewSeededRandom(1), measure)

	rNarrow := narrow.CreateH(0, 0).Radius()
	rWide := wide.CreateH(0, 0).Radius()

	if !(rWide > rNarrow) {
		t.Errorf("expected wide radius (%v) > narrow radius (%v)", rWide, rNarrow)
	}
}

func TestFactory_SizeScale_EvaluatedAtCreationTime(t *testing.T) {
	// scale は生成呼び出しのたびに、構築時にキャッシュされた値ではなく
	// *現在の* bounds の幅を使って再評価されなければならない (bounds は
	// ウィンドウのリサイズで変わりうる)。
	measure, _ := newMeasureStub()
	bounds := core.NewBounds(767, 600)
	f := NewFactory(bounds, core.NewSeededRandom(1), measure)

	rBefore := f.CreateH(0, 0).Radius()
	bounds.Set(768, 600)
	rAfter := f.CreateH(0, 0).Radius()

	if !(rAfter > rBefore) {
		t.Errorf("expected radius to grow after resize: before=%v after=%v", rBefore, rAfter)
	}
}

func TestFactory_H2o_SizeRange(t *testing.T) {
	bounds := core.NewBounds(800, 600) // scale 1.2
	random := core.NewSeededRandom(9)
	measure, _ := newMeasureStub()
	f := NewFactory(bounds, random, measure)

	for i := 0; i < 50; i++ {
		r := f.CreateH2o(100, 100).Radius()
		if r < 9*1.2 || r >= 14*1.2 {
			t.Fatalf("H2o radius out of range [9*1.2, 14*1.2): got %v", r)
		}
	}
}

func TestFactory_MeasureCache_HitsAfterFirstCall(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	measure, calls := newMeasureStub()
	f := NewFactory(bounds, random, measure)

	f.CreateH(0, 0)
	after1 := *calls
	if after1 == 0 {
		t.Fatal("expected measure to be called at least once for the first CreateH")
	}

	f.CreateH(10, 10)
	if after2 := *calls; after2 != after1 {
		t.Errorf("expected cache hit on 2nd CreateH: calls %d -> %d", after1, after2)
	}

	f.CreateH(20, 20)
	if after3 := *calls; after3 != after1 {
		t.Errorf("expected cache hit on 3rd CreateH: calls %d -> %d", after1, after3)
	}
}

func TestFactory_MeasureCache_DistinctKeysPerTextAndFontSize(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	random := core.NewSeededRandom(1)
	measure, calls := newMeasureStub()
	f := NewFactory(bounds, random, measure)

	f.CreateH(0, 0)  // "H" @ fontSize
	f.CreateO(0, 0)  // "O" @ fontSize (異なるテキストなのでキャッシュミスするはず)
	f.CreateH2(0, 0) // "H2" @ fontSize (異なるテキストなのでキャッシュミスするはず)

	if got := *calls; got != 3 {
		t.Errorf("expected 3 distinct measure calls (H, O, H2), got %d", got)
	}
}

func TestFactory_Determinism_SameSeedSamePositionAndRadius(t *testing.T) {
	measure, _ := newMeasureStub()
	a := NewFactory(core.NewBounds(800, 600), core.NewSeededRandom(7), measure).CreateHAtRandom()
	b := NewFactory(core.NewBounds(800, 600), core.NewSeededRandom(7), measure).CreateHAtRandom()

	if a.X() != b.X() || a.Y() != b.Y() || a.Radius() != b.Radius() {
		t.Errorf("expected deterministic output for same seed: a=(%v,%v,%v) b=(%v,%v,%v)",
			a.X(), a.Y(), a.Radius(), b.X(), b.Y(), b.Radius())
	}
}

func TestFactory_RandomWalkBodySizeUsesMeasuredWidth(t *testing.T) {
	// リグレッション防止: RandomWalk のラップアラウンドのマージンは、
	// (radius ではなく) 計測されたテキスト幅全体を bodySize として使わな
	// ければならない。これは Mizu-ts の ParticleFactory.ts
	// (`new RandomWalk(this.sw, this.sh, size, ...)`、ここで size は半分
	// にする前の計測幅であり、radius は size/2 として別に持つ) と一致す
	// る。右端をわずかに超えた位置に粒子を置き、ちょうど -(bodySize/2)
	// にラップされることを確認する。もし factory が bodySize として
	// radius を渡していたら、ラップ先が 2 倍ずれてしまう。
	bounds := core.NewBounds(100, 100)
	random := core.NewSeededRandom(1) // speedFactor のデフォルトでは初速が小さくなるため、代わりに bounds を使って大きなジャンプを強制する
	measure, _ := newMeasureStub()    // "H" @ fontSize 24 (scale 1.0) -> size=12, radius=6, bodySize は 12 であるべき

	f := NewFactory(bounds, random, measure)
	h := f.CreateH(200, 50) // すでに sw+radius を超えているため、速度の向きや大きさに関わらず最初の Update でラップが発生する

	h.Update()
	wantBodySize := 2 * h.Radius() // size = 2*radius = 12
	wantWrappedX := -(wantBodySize / 2)
	if h.X() != wantWrappedX {
		t.Errorf("expected wraparound to use bodySize=measured size (%v), got wrapped X=%v, want %v",
			wantBodySize, h.X(), wantWrappedX)
	}
}
