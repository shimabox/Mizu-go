package physics

import (
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func TestSpatialGrid_AssignsExactBoundaryCoordinateToCorrectCell(t *testing.T) {
	// cellSize=10。probe のセルは floor(25/10)=2 なので、neighbors は
	// col/row の 1..3 にまたがる。
	grid := NewSpatialGrid(100, 100, 10)
	probe := newFakeParticle("H", 25, 25, 1)

	justBelowBoundary := newFakeParticle("H", 9.999, 25, 1) // col 0 (範囲外)
	onBoundary := newFakeParticle("H", 10, 25, 1)           // col 1 (範囲内: ちょうど境界は繰り上がる)

	grid.Insert(justBelowBoundary)
	grid.Insert(onBoundary)

	neighbors := grid.Neighbors(probe)
	if containsParticle(neighbors, justBelowBoundary) {
		t.Fatal("expected neighbors not to contain justBelowBoundary")
	}
	if !containsParticle(neighbors, onBoundary) {
		t.Fatal("expected neighbors to contain onBoundary")
	}
}

func TestSpatialGrid_ClampsOutOfScreenCoordinatesToLastCell(t *testing.T) {
	grid := NewSpatialGrid(100, 100, 10) // cols=rows=10、最後のセルのインデックスは9
	probe := newFakeParticle("H", 95, 95, 1)
	atEdge := newFakeParticle("H", 100, 100, 1)
	farOutside := newFakeParticle("H", 1000, 1000, 1)

	grid.Insert(atEdge)
	grid.Insert(farOutside)

	neighbors := grid.Neighbors(probe)
	if !containsParticle(neighbors, atEdge) {
		t.Fatal("expected neighbors to contain atEdge")
	}
	if !containsParticle(neighbors, farOutside) {
		t.Fatal("expected neighbors to contain farOutside")
	}
}

func TestSpatialGrid_ClampsNegativeCoordinatesToFirstCell(t *testing.T) {
	grid := NewSpatialGrid(100, 100, 10)
	probe := newFakeParticle("H", 5, 5, 1)
	negative := newFakeParticle("H", -50, -50, 1)
	farNegative := newFakeParticle("H", -1000, -1000, 1)
	distantProbe := newFakeParticle("H", 95, 95, 1)

	grid.Insert(negative)
	grid.Insert(farNegative)

	if !containsParticle(grid.Neighbors(probe), negative) {
		t.Fatal("expected neighbors(probe) to contain negative")
	}
	if !containsParticle(grid.Neighbors(probe), farNegative) {
		t.Fatal("expected neighbors(probe) to contain farNegative")
	}
	if containsParticle(grid.Neighbors(distantProbe), negative) {
		t.Fatal("expected neighbors(distantProbe) not to contain negative")
	}
	if containsParticle(grid.Neighbors(distantProbe), farNegative) {
		t.Fatal("expected neighbors(distantProbe) not to contain farNegative")
	}
}

func TestSpatialGrid_NeighborsReturnsSelfCellAndAdjacent8Cells(t *testing.T) {
	grid := NewSpatialGrid(300, 300, 10)
	center := newFakeParticle("H", 105, 105, 1) // col=row=10

	withinRange := []*fakeParticle{
		newFakeParticle("H", 95, 95, 1),   // col=9,row=9
		newFakeParticle("H", 115, 115, 1), // col=11,row=11
		newFakeParticle("H", 105, 95, 1),  // col=10,row=9
	}
	outOfRange := []*fakeParticle{
		newFakeParticle("H", 200, 200, 1), // col=20,row=20
		newFakeParticle("H", 5, 5, 1),     // col=0,row=0
	}

	grid.Insert(center)
	for _, p := range withinRange {
		grid.Insert(p)
	}
	for _, p := range outOfRange {
		grid.Insert(p)
	}

	neighbors := grid.Neighbors(center)
	if !containsParticle(neighbors, center) {
		t.Fatal("expected neighbors to contain center (self cell includes itself)")
	}
	for _, p := range withinRange {
		if !containsParticle(neighbors, p) {
			t.Fatalf("expected neighbors to contain %v", p)
		}
	}
	for _, p := range outOfRange {
		if containsParticle(neighbors, p) {
			t.Fatalf("expected neighbors not to contain %v", p)
		}
	}
}

func TestSpatialGrid_ClearEmptiesAllCells(t *testing.T) {
	grid := NewSpatialGrid(100, 100, 10)
	p := newFakeParticle("H", 50, 50, 1)
	grid.Insert(p)
	if !containsParticle(grid.Neighbors(p), p) {
		t.Fatal("expected neighbors to contain p before Clear")
	}

	grid.Clear()
	if got := grid.Neighbors(p); len(got) != 0 {
		t.Fatalf("expected empty neighbors after Clear, got %v", got)
	}
}

func TestSpatialGrid_InsertAfterClearDoesNotLeakOldContent(t *testing.T) {
	grid := NewSpatialGrid(100, 100, 10)
	first := newFakeParticle("H", 50, 50, 1)
	second := newFakeParticle("H", 50, 50, 1)

	grid.Insert(first)
	grid.Clear()
	grid.Insert(second)

	neighbors := grid.Neighbors(second)
	if !containsParticle(neighbors, second) {
		t.Fatal("expected neighbors to contain second")
	}
	if containsParticle(neighbors, first) {
		t.Fatal("expected neighbors not to contain first")
	}
}

func TestSpatialGrid_TwoParticlesAcrossCellBoundaryAreMutualNeighbors(t *testing.T) {
	grid := NewSpatialGrid(100, 100, 10)
	// x=9.9 は col 0、x=10.1 は col 1: 隣接しているが別々のセル。
	a := newFakeParticle("H", 9.9, 50, 5)
	b := newFakeParticle("H", 10.1, 50, 5)

	grid.Insert(a)
	grid.Insert(b)

	if !containsParticle(grid.Neighbors(a), b) {
		t.Fatal("expected neighbors(a) to contain b")
	}
	if !containsParticle(grid.Neighbors(b), a) {
		t.Fatal("expected neighbors(b) to contain a")
	}
}

func TestNewSpatialGrid_PanicsWhenCellSizeIsNotPositive(t *testing.T) {
	assertPanics(t, func() { NewSpatialGrid(100, 100, 0) })
	assertPanics(t, func() { NewSpatialGrid(100, 100, -5) })
}

func TestSpatialGrid_CellSizeMustBeAtLeastMaxCollisionDistance(t *testing.T) {
	if MaxCollisionDistance != MaxParticleRadius*2 {
		t.Fatalf("MaxCollisionDistance = %v, want %v", MaxCollisionDistance, MaxParticleRadius*2)
	}
	if DefaultCellSize != MaxCollisionDistance {
		t.Fatalf("DefaultCellSize = %v, want %v (the 3x3 neighbor-scan invariant requires this)", DefaultCellSize, MaxCollisionDistance)
	}
}

// TestSpatialGrid_ExceedingMaxParticleRadiusBreaksTheInvariant_Canary は、
// MaxParticleRadius に記載されている前提「衝突しうる任意の2粒子は必ず
// 同じセルか隣接セルに存在する」を検証する Go 側のカナリアテスト。
// 粒子の半径が MaxParticleRadius を超えて大きくなると、この前提が破られ
// うる。すると、グリッドベースの検出器は、brute-force のスキャンなら
// 見つけられるはずの衝突を、気づかないうちに見逃してしまう。このテスト
// はその失敗モードを固定して、不変条件が破られた瞬間にすぐ気づけるように
// している。
//
// (TS 版のカナリアは代わりに、ParticleFactory が MaxParticleRadius を
// 超える半径を決して生成しないことを検証している。このパッケージは
// internal/particle に依存してはならないため、カナリアのその半分は、
// internal/particle が存在するようになったら、そちらの独自のテストに
// 属することになる。)
func TestSpatialGrid_ExceedingMaxParticleRadiusBreaksTheInvariant_Canary(t *testing.T) {
	grid := NewSpatialGrid(400, 400, DefaultCellSize)

	// 丸々2セル分離れている(互いの3x3近傍には入らない)2つの粒子だが、
	// 半径の合計が MaxCollisionDistance を大きく超えているため、幾何学的
	// には依然として重なり合っている。
	oversizedRadius := MaxParticleRadius * 3.0
	a := newFakeParticle("H", 0, 0, oversizedRadius)
	b := newFakeParticle("H", DefaultCellSize*2, 0, oversizedRadius)

	dx := b.X() - a.X()
	hitDistance := a.Radius() + b.Radius()
	if !(dx*dx < hitDistance*hitDistance) {
		t.Fatal("test setup invalid: particles must actually overlap")
	}

	grid.Clear()
	grid.Insert(a)
	grid.Insert(b)

	if containsParticle(grid.Neighbors(a), b) {
		t.Fatal("expected the grid to miss b as a neighbor of a once radii exceed MaxParticleRadius (that's the point of the invariant)")
	}
}

func TestSpatialGrid_ResizeRebuildsGridForNewDimensions(t *testing.T) {
	grid := NewSpatialGrid(100, 100, 10)
	p := newFakeParticle("H", 50, 50, 1)
	grid.Insert(p)

	grid.Resize(20, 20)

	// Resize はグリッドを再構築するため、それ以前に挿入していた粒子は
	// 消えている。
	if got := grid.Neighbors(p); len(got) != 0 {
		t.Fatalf("expected Resize to clear previous contents, got %v", got)
	}

	// 新しい寸法でもグリッドは正しく動作しなければならない。新しい範囲を
	// 大きく超えた粒子は、以前と同様に最後のセルにクランプされるべき。
	probe := newFakeParticle("H", 15, 15, 1)
	farOutside := newFakeParticle("H", 1000, 1000, 1)
	grid.Insert(farOutside)
	if !containsParticle(grid.Neighbors(probe), farOutside) {
		t.Fatal("expected farOutside to clamp into probe's neighborhood after Resize")
	}
}

func containsParticle(particles []core.Particle, target core.Particle) bool {
	for _, p := range particles {
		if p == target {
			return true
		}
	}
	return false
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected a panic, got none")
		}
	}()
	fn()
}
