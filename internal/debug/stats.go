package debug

import (
	"fmt"
	"strings"
)

// FixedKinds は stats オーバーレイのテキストで常に表示される、粒子 kind 名の
// 固定順序であり、Mizu-ts の src/main.ts:66 の allKinds に対応する。
// H, H2, O, H2o は常に表示され、まだ出現していない kind もカウント 0 として
// 表示される。そのため、フレームが変わってもオーバーレイの行数・レイアウトは
// 変化しない(porting-plan §5.6)。
var FixedKinds = []string{"H", "H2", "O", "H2o"}

// frameWindow は FPS 移動平均のウィンドウサイズであり、Mizu-ts の
// src/debug/StatsOverlay.ts の 60 フレーム上限に対応する。
const frameWindow = 60

// Stats はフレーム所要時間の計測値を蓄積し、kind 別の粒子数と合わせて
// オーバーレイの表示テキストとして描画する。Mizu-ts の
// src/debug/StatsOverlay.ts に対応するが DOM は持たない。このパッケージは
// ebiten(および描画全般)への依存を持たず、レンダー層から渡される
// float64 の所要時間・カウントのみで動作する。
type Stats struct {
	// frameIntervals は記録済みフレーム間隔(ミリ秒)の移動ウィンドウであり、
	// 古いものが先頭、frameWindow 件を上限とする。
	frameIntervals []float64
	// frameInterval は直近に記録されたフレーム間隔(ms)。
	frameInterval float64
	// updateTime は直近に記録された Update() の所要時間(ms)。
	updateTime float64
}

// NewStats は何も記録されていない状態の Stats を返す。最初の RecordFrame
// 呼び出しまで FPS は 0 を返し、StatsOverlay.ts の初期状態
// (frameIntervals: [], lastTimestamp: null)に対応する。
func NewStats() *Stats {
	return &Stats{}
}

// RecordFrame は前フレームからの経過時間である intervalMs(ミリ秒)を
// FPS 移動平均のウィンドウに追加し、StatsOverlay.ts の frame() に対応する。
// 直近の frameWindow(60)件のみが保持され、ウィンドウが満杯になると
// それより古いものは破棄される。
func (s *Stats) RecordFrame(intervalMs float64) {
	s.frameIntervals = append(s.frameIntervals, intervalMs)
	if len(s.frameIntervals) > frameWindow {
		s.frameIntervals = s.frameIntervals[1:]
	}
	s.frameInterval = intervalMs
}

// RecordUpdate は直近の Update() 実行時間である durationMs(ミリ秒)を
// 記録し、StatsOverlay.ts の setFrameTime() に対応する。
func (s *Stats) RecordUpdate(durationMs float64) {
	s.updateTime = durationMs
}

// FPS は記録されたフレーム間隔の移動平均を fps(1 秒あたりのフレーム数)に
// 変換して返す。まだフレームが 1 件も記録されていない場合は 0 を返す
// (StatsOverlay.ts の getFPS() に対応)。
func (s *Stats) FPS() float64 {
	if len(s.frameIntervals) == 0 {
		return 0
	}

	var total float64
	for _, v := range s.frameIntervals {
		total += v
	}
	avg := total / float64(len(s.frameIntervals))
	if avg == 0 {
		return 0
	}
	return 1000 / avg
}

// Text はオーバーレイの表示テキスト全体を、1 行に 1 指標ずつ描画する。
// 順に FPS、直近のフレーム間隔、直近の Update() 所要時間、Ebitengine 自身の
// actualFPS/actualTPS(このパッケージが ebiten に依存しないよう呼び出し側から
// 渡される)、FixedKinds の順序による各 kind(counts に存在しなければ 0 とし、
// フレームが変わっても行数が変化しないようにする)、そして固定 kind 全体の
// 合計、という構成になる。すべての数値は小数点以下ちょうど 1 桁で
// フォーマットされ、StatsOverlay.ts の render()(toFixed(1))に対応する。
func (s *Stats) Text(counts map[string]int, actualFPS, actualTPS float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "FPS: %.1f\n", s.FPS())
	fmt.Fprintf(&b, "Frame: %.1fms\n", s.frameInterval)
	fmt.Fprintf(&b, "Update: %.1fms\n", s.updateTime)
	fmt.Fprintf(&b, "ActualFPS: %.1f ActualTPS: %.1f", actualFPS, actualTPS)

	total := 0
	for _, kind := range FixedKinds {
		count := counts[kind]
		total += count
		fmt.Fprintf(&b, "\n%s: %d", kind, count)
	}
	fmt.Fprintf(&b, "\nTotal: %d", total)

	return b.String()
}
