package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shimabox/Mizu-go/internal/bench"
	"github.com/shimabox/Mizu-go/internal/core"
	"github.com/shimabox/Mizu-go/internal/particle"
	"github.com/shimabox/Mizu-go/internal/physics"
	"github.com/shimabox/Mizu-go/internal/reaction"
	"github.com/shimabox/Mizu-go/internal/render"
	"github.com/shimabox/Mizu-go/internal/sim"
)

// runOne は単一シナリオを ebiten.RunGame で計測し、結果を jsonPath に
// JSON として書き出す(オーケストレータプロセスがサブプロセスとして
// 起動する内部モード)。
func runOne(scenarioName string, framesOverride int, warmupMs int, jsonPath string) error {
	scenarios, err := bench.ResolveScenarios([]string{scenarioName})
	if err != nil {
		return err
	}
	scenario := scenarios[0]

	frames := framesOverride
	if frames <= 0 {
		frames = bench.DefaultFramesFor(scenario.Name)
	}

	result, err := measureScenario(scenario, frames, time.Duration(warmupMs)*time.Millisecond)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling result: %w", err)
	}
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", jsonPath, err)
	}
	return nil
}

// measureScenario は、cmd/mizu/main.go と同じ手順で Simulator を組み立て
// (bounds は viewportWidth x viewportHeight で固定)、それを
// measuringGame でラップして ebiten.RunGame で走らせる。vsync を無効化し
// TPS を SyncWithFPS にすることで、TPS が FPS と 1:1 になり、Update が
// 描画のスループット限界で連続実行される(porting-plan の Go 版設計)。
func measureScenario(scenario bench.Scenario, frames int, warmup time.Duration) (bench.ScenarioResult, error) {
	bounds := core.NewBounds(viewportWidth, viewportHeight)
	random := core.NewRandom()

	fontSource, err := render.NewFontSource()
	if err != nil {
		return bench.ScenarioResult{}, fmt.Errorf("loading font: %w", err)
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
	simulator.Init(int(float64(scenario.H)*scale), int(float64(scenario.O)*scale))

	sprites := render.NewSprites(faces)
	// measure=false: オーバーレイの描画・計測コストは実運用と異なる
	// オーバーヘッドになるため、計測にはこのファイル自身の
	// measuringGame を使う。
	innerGame := render.NewGame(simulator, sprites, nil, false)

	mg := newMeasuringGame(innerGame, simulator, warmup, frames)

	ebiten.SetVsyncEnabled(false)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetWindowTitle(fmt.Sprintf("Mizu-go bench: %s", scenario.Name))
	ebiten.SetWindowSize(viewportWidth, viewportHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(mg); err != nil {
		return bench.ScenarioResult{}, fmt.Errorf("running scenario %s: %w", scenario.Name, err)
	}

	return bench.ScenarioResult{
		Name:     scenario.Name,
		Label:    scenario.Label,
		FrameMs:  mg.frameSamples,
		UpdateMs: mg.updateSamples,
		Counts:   mg.counts,
	}, nil
}

// measuringGame は render.Game をラップし、ウォームアップ後の N フレーム
// 分について、フレーム間隔(壁時計、ms)と Simulator.Update() の実行
// 時間(ms)を収集する。実際の描画は inner.Draw に委譲するため、
// フレーム間隔には実運用と同じ描画コストが含まれる。
type measuringGame struct {
	inner *render.Game
	sim   *sim.Simulator

	warmup       time.Duration
	targetFrames int
	start        time.Time
	lastFrame    time.Time

	frameSamples  []float64
	updateSamples []float64
	counts        map[string]int
}

func newMeasuringGame(inner *render.Game, s *sim.Simulator, warmup time.Duration, targetFrames int) *measuringGame {
	return &measuringGame{
		inner:        inner,
		sim:          s,
		warmup:       warmup,
		targetFrames: targetFrames,
		start:        time.Now(),
	}
}

// Update は、既に必要なフレーム数のサンプルを集め終えていれば
// ebiten.Termination を返して計測を終了させる。それ以外は inner.Update
// (= Simulator.Update())の実行時間を計測し、ウォームアップ経過後のみ
// サンプルに追加する。
func (m *measuringGame) Update() error {
	if len(m.frameSamples) >= m.targetFrames {
		return ebiten.Termination
	}

	start := time.Now()
	if err := m.inner.Update(); err != nil {
		return err
	}
	durationMs := float64(time.Since(start)) / float64(time.Millisecond)

	if time.Since(m.start) >= m.warmup && len(m.updateSamples) < m.targetFrames {
		m.updateSamples = append(m.updateSamples, durationMs)
	}
	return nil
}

// Draw は実際の描画を inner.Draw に委譲したうえで、ウォームアップ経過後
// のフレーム間隔(壁時計)をサンプリングし、Simulator の現在の kind 別
// 粒子数も記録する(最後に記録されたものがレポートの「終了時粒子数」
// になる)。
func (m *measuringGame) Draw(screen *ebiten.Image) {
	m.inner.Draw(screen)

	now := time.Now()
	if time.Since(m.start) >= m.warmup {
		if !m.lastFrame.IsZero() && len(m.frameSamples) < m.targetFrames {
			interval := float64(now.Sub(m.lastFrame)) / float64(time.Millisecond)
			m.frameSamples = append(m.frameSamples, interval)
		}
		m.lastFrame = now
	}

	buckets := m.sim.Buckets()
	counts := make(map[string]int, len(buckets))
	for _, b := range buckets {
		counts[string(b.Kind)] = len(b.Particles)
	}
	m.counts = counts
}

// Layout は inner.Layout(= render.Game.Layout)にそのまま委譲する。
// ウィンドウを固定サイズ・リサイズ不可にしているため、outsideWidth/
// outsideHeight は常に viewportWidth/viewportHeight になる。
//
// 重要: render.Game.Layout は HiDPI 対応のため outside(CSS px)に
// DeviceScaleFactor を掛けた実デバイスピクセルを返す。ここでその
// 戻り値をそのまま返すことで、Ebitengine が Draw に渡すオフスクリーン
// (screen)も inner.Layout の戻り値のサイズで確保され、Retina 上でも
// 描画がはみ出したり切れたりしない。sim.Resize には inner.Layout 内で
// 従来どおり CSS px(= viewportWidth/viewportHeight)が渡るため、
// 計測セマンティクス(シナリオの論理解像度・粒子スケール)は変わら
// ない。
func (m *measuringGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return m.inner.Layout(outsideWidth, outsideHeight)
}

var _ ebiten.Game = (*measuringGame)(nil)
