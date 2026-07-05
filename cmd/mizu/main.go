// Command mizu は、Mizu-go の粒子シミュレーションを Ebitengine の
// ウィンドウ内で実行する。これは Mizu-ts の src/main.ts のエントリー
// ポイントを反映したものである(porting-plan §4.2)。
package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/particle"
	"github.com/shimabox/Mizu-go/internal/physics"
	"github.com/shimabox/Mizu-go/internal/reaction"
	"github.com/shimabox/Mizu-go/internal/render"
	"github.com/shimabox/Mizu-go/internal/sim"
)

// initialWindowWidth と initialWindowHeight は起動時のウィンドウ
// サイズであり、その後ウィンドウはリサイズ可能である(porting-plan
// §5.4)。
const (
	initialWindowWidth  = 1280
	initialWindowHeight = 720
)

func main() {
	hFlag := flag.Int("h", 30, "initial H atom count (before particle-count scaling)")
	oFlag := flag.Int("o", 50, "initial O atom count (before particle-count scaling)")
	measureFlag := flag.Bool("m", false, "enable the measurement overlay (FPS/frame time/particle counts)")
	flag.Parse()

	// 1. 共有境界。behavior/factory/simulator から読み取られ、
	// リサイズ時に更新される(porting-plan §5.4)。
	bounds := core.NewBounds(initialWindowWidth, initialWindowHeight)

	// 2. 乱数。
	random := core.NewRandom()

	// 3. フォントと計測、続いて粒子ファクトリ(Mizu-ts の
	// main.ts:31-33: `new ParticleFactory(...)` を反映)。
	fontSource, err := render.NewFontSource()
	if err != nil {
		log.Fatalf("mizu: loading font: %v", err)
	}
	faces := render.NewFaceCache(fontSource)
	factory := particle.NewFactory(bounds, random, faces.Measure)

	// 4. 反応レジストリ: ReactiveKinds のスナップショットにすべての
	// 反応対象種類が含まれるよう、ルールは NewSimulator が呼ばれる前
	// に登録する(手順 6 のコメントを参照)。
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

	// 5. 衝突判定グリッド/検出器。
	grid := physics.NewSpatialGrid(bounds.W(), bounds.H(), physics.DefaultCellSize)
	detector := physics.NewGridCollisionDetector(grid)

	// 6. Simulator。重要: NewSimulator は構築時に
	// registry.ReactiveKinds() のスナップショットを取るので、上記の
	// Register 呼び出しはすべてそれより先に行わなければならない
	// (internal/sim/simulator.go の NewSimulator の doc コメントを
	// 参照)。
	world := sim.NewWorld()
	simulator := sim.NewSimulator(bounds, world, factory, registry, detector, grid)

	// 7. 初期粒子数。シミュレーターの粒子数スケールで拡大縮小される
	// (Mizu-ts の main.ts:47-48: `simulator.init(hCount * scale, oCount
	// * scale)` を反映。int への切り捨ては、Init の int 引数に小数の
	// 積が渡されたときの JS の暗黙的な挙動に合わせたもの)。
	scale := simulator.CountScale()
	simulator.Init(int(float64(*hFlag)*scale), int(float64(*oFlag)*scale))

	// 8. スプライトを RunGame の開始前に一度だけ生成し(porting-plan
	// §5.2)、続いて Game とウィンドウを用意する。faces は -m の計測
	// オーバーレイのテキストにも再利用される(porting-plan §5.6)。
	sprites := render.NewSprites(faces)
	game := render.NewGame(simulator, sprites, faces, *measureFlag)

	ebiten.SetWindowTitle("Mizu-go")
	ebiten.SetWindowSize(initialWindowWidth, initialWindowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("mizu: %v", err)
	}
}
