// devshot は目視検証用の開発ツール。cmd/mizu と同じ配線でシミュレー
// ションを動かし、指定 tick 経過時点の画面をオフスクリーンから PNG に
// 書き出して終了する。
//
// Ebitengine にはヘッドレスモードがないため、CI や自動化された環境で
// 「描画結果を画像として確認する」手段がこれにあたる(実ウィンドウは
// 開くが、スクリーンショット権限は不要)。描画変更の before/after 比較
// (例: PR #1 のジャギー修正)や、負荷シナリオでのオーバーレイ数値の
// 確認などに使う。
//
// 使い方:
//
//	go run ./cmd/devshot -out shot.png                     # 既定(30/50)を 300 tick 後に撮影
//	go run ./cmd/devshot -out shot.png -m -h 500 -o 500    # 負荷 + 計測オーバーレイ
//	go run ./cmd/devshot -out shot.png -ticks 900          # 15 秒後の様子
package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/particle"
	"github.com/shimabox/Mizu-go/internal/physics"
	"github.com/shimabox/Mizu-go/internal/reaction"
	"github.com/shimabox/Mizu-go/internal/render"
	"github.com/shimabox/Mizu-go/internal/sim"
)

// shotGame は render.Game をラップし、毎フレームをオフスクリーンに
// 描画して、指定 tick に達したら PNG を書き出して終了する。
type shotGame struct {
	inner     *render.Game
	logicalW  int
	logicalH  int
	shotTicks int
	path      string

	offscreen  *ebiten.Image
	outW, outH int
	ticks      int
	done       bool
}

func (s *shotGame) Update() error {
	if err := s.inner.Update(); err != nil {
		return err
	}
	s.ticks++
	if s.done {
		return ebiten.Termination
	}
	return nil
}

func (s *shotGame) Draw(screen *ebiten.Image) {
	s.inner.Draw(s.offscreen)
	screen.DrawImage(s.offscreen, nil)

	if s.ticks < s.shotTicks || s.done {
		return
	}
	s.done = true

	px := make([]byte, 4*s.outW*s.outH)
	s.offscreen.ReadPixels(px)
	img := &image.RGBA{Pix: px, Stride: 4 * s.outW, Rect: image.Rect(0, 0, s.outW, s.outH)}
	f, err := os.Create(s.path)
	if err != nil {
		log.Printf("devshot: %v", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Printf("devshot: %v", err)
		return
	}
	log.Printf("devshot: wrote %s (%dx%d)", s.path, s.outW, s.outH)
}

// Layout は常に固定の論理サイズを inner に渡す。inner.Layout は
// DeviceScaleFactor を掛けた実解像度を返すことがあるため、オフスク
// リーンはその戻り値のサイズで確保する(サイズが変わったら作り直す)。
func (s *shotGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := s.inner.Layout(s.logicalW, s.logicalH)
	if s.offscreen == nil || s.outW != w || s.outH != h {
		s.outW, s.outH = w, h
		s.offscreen = ebiten.NewImage(w, h)
	}
	return w, h
}

func main() {
	outFlag := flag.String("out", "devshot.png", "出力 PNG のパス")
	hFlag := flag.Int("h", 30, "H 原子の初期数(粒子数スケール適用前)")
	oFlag := flag.Int("o", 50, "O 原子の初期数(粒子数スケール適用前)")
	measureFlag := flag.Bool("m", false, "計測オーバーレイを表示する")
	ticksFlag := flag.Int("ticks", 300, "撮影する tick(60 tick = 約 1 秒)")
	widthFlag := flag.Int("width", 1280, "論理解像度の幅(CSS px)")
	heightFlag := flag.Int("height", 720, "論理解像度の高さ(CSS px)")
	flag.Parse()

	// 以下の配線は cmd/mizu/main.go と同じ順序。ルール登録は
	// reactiveKinds のスナップショットのため NewSimulator より先に行う。
	bounds := core.NewBounds(float64(*widthFlag), float64(*heightFlag))
	random := core.NewRandom()

	fontSource, err := render.NewFontSource()
	if err != nil {
		log.Fatalf("devshot: loading font: %v", err)
	}
	faces := render.NewFaceCache(fontSource)
	factory := particle.NewFactory(bounds, random, faces.Measure)

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

	grid := physics.NewSpatialGrid(bounds.W(), bounds.H(), physics.DefaultCellSize)
	detector := physics.NewGridCollisionDetector(grid)
	world := sim.NewWorld()
	simulator := sim.NewSimulator(bounds, world, factory, registry, detector, grid)

	scale := simulator.CountScale()
	simulator.Init(int(float64(*hFlag)*scale), int(float64(*oFlag)*scale))

	sprites := render.NewSprites(faces)
	game := render.NewGame(simulator, sprites, faces, *measureFlag)

	shot := &shotGame{
		inner:     game,
		logicalW:  *widthFlag,
		logicalH:  *heightFlag,
		shotTicks: *ticksFlag,
		path:      *outFlag,
	}

	ebiten.SetWindowTitle("Mizu-go devshot")
	ebiten.SetWindowSize(*widthFlag, *heightFlag)
	if err := ebiten.RunGame(shot); err != nil {
		log.Fatalf("devshot: %v", err)
	}
}
