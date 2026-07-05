package behavior

import (
	"math"
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

func TestRandomWalk_DeterministicWithSameSeed(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	walk1 := NewRandomWalk(bounds, 20, core.NewSeededRandom(42))
	walk2 := NewRandomWalk(bounds, 20, core.NewSeededRandom(42))

	x1, y1 := 400.0, 300.0
	x2, y2 := 400.0, 300.0
	for i := 0; i < 100; i++ {
		x1, y1 = walk1.Next(x1, y1)
		x2, y2 = walk2.Next(x2, y2)
		if x1 != x2 || y1 != y2 {
			t.Fatalf("iteration %d: got (%v,%v) and (%v,%v), want equal", i, x1, y1, x2, y2)
		}
	}
}

func TestRandomWalk_PositionUpdates(t *testing.T) {
	bounds := core.NewBounds(800, 600)
	walk := NewRandomWalk(bounds, 20, core.NewSeededRandom(42))
	nx, ny := walk.Next(400, 300)
	if nx == 400 && ny == 300 {
		t.Fatalf("expected position to change, got unchanged (%v,%v)", nx, ny)
	}
}

func TestRandomWalk_SpeedNeverExceedsMaxSpeed(t *testing.T) {
	// ラップアラウンドが発生しないほど十分に大きな画面にすることで、
	// 速度クランプの検証をラップアラウンドのロジックから切り離す
	// (RandomWalk.test.ts に対応)。
	bounds := core.NewBounds(100000, 100000)
	walk := NewRandomWalk(bounds, 20, core.NewSeededRandom(7))

	x, y := 50000.0, 50000.0
	for i := 0; i < 500; i++ {
		nx, ny := walk.Next(x, y)
		dx, dy := nx-x, ny-y
		moved := math.Sqrt(dx*dx + dy*dy)
		if moved > 1.05+1e-9 {
			t.Fatalf("iteration %d: moved %v, want <= 1.05", i, moved)
		}
		x, y = nx, ny
	}
}

// TestRandomWalk_WraparoundBoundaries は RandomWalk.test.ts の
// "画面端のラップアラウンド(境界値)" describe ブロックに対応する。
// speedFactor=0 で速度をゼロに凍結することで境界値の比較だけを検証
// できるようにしており、「ちょうど端」では折り返してはならず、
// 「端 + 1」では折り返さなければならない。
func TestRandomWalk_WraparoundBoundaries(t *testing.T) {
	const (
		sw       = 800.0
		sh       = 600.0
		bodySize = 20.0
	)
	stillWalk := func() *RandomWalk {
		bounds := core.NewBounds(sw, sh)
		return NewRandomWalkWithParams(bounds, bodySize, core.NewSeededRandom(1), 0, defaultMaxSpeed)
	}

	tests := []struct {
		name   string
		x, y   float64
		wantX  float64
		wantY  float64
		checkX bool
		checkY bool
	}{
		{name: "右端ちょうど(sw+w/2)では折り返さない", x: sw + bodySize/2, y: 300, wantX: sw + bodySize/2, checkX: true},
		{name: "右端+1で左側(-w/2)へ折り返す", x: sw + bodySize/2 + 1, y: 300, wantX: -(bodySize / 2), checkX: true},
		{name: "左端ちょうど(x+w=0)では折り返さない", x: -bodySize, y: 300, wantX: -bodySize, checkX: true},
		{name: "左端-1で右側(sw+w/2)へ折り返す", x: -bodySize - 1, y: 300, wantX: sw + bodySize/2, checkX: true},
		{name: "下端ちょうど(sh+h/2)では折り返さない", x: 400, y: sh + bodySize/2, wantY: sh + bodySize/2, checkY: true},
		{name: "下端+1で上側(-h/2)へ折り返す", x: 400, y: sh + bodySize/2 + 1, wantY: -(bodySize / 2), checkY: true},
		{name: "上端ちょうど(y+h=0)では折り返さない", x: 400, y: -bodySize, wantY: -bodySize, checkY: true},
		{name: "上端-1で下側(sh+h/2)へ折り返す", x: 400, y: -bodySize - 1, wantY: sh + bodySize/2, checkY: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nx, ny := stillWalk().Next(tt.x, tt.y)
			if tt.checkX && nx != tt.wantX {
				t.Errorf("x = %v, want %v", nx, tt.wantX)
			}
			if tt.checkY && ny != tt.wantY {
				t.Errorf("y = %v, want %v", ny, tt.wantY)
			}
		})
	}
}
