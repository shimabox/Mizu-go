package bench

import (
	"math"
	"strings"
	"testing"
)

// このファイルは Mizu-ts の tests/bench/stats.test.ts (32 件) を Go に
// 移植したものである。数値の丸め・境界条件は TS 版の期待値をそのまま
// 引き継いでいる。TS 版に存在しない BuildScenarioTable/BuildComparisonTable
// (Go 版で列構成を拡張したもの) 相当のテストは report_test.go に置く。

func TestMean(t *testing.T) {
	if got := Mean(nil); got != 0 {
		t.Errorf("Mean(nil) = %v, want 0", got)
	}
	if got := Mean([]float64{1, 2, 3}); got != 2 {
		t.Errorf("Mean([1,2,3]) = %v, want 2", got)
	}
	if got := Mean([]float64{10}); got != 10 {
		t.Errorf("Mean([10]) = %v, want 10", got)
	}
}

func TestMedian(t *testing.T) {
	if got := Median(nil); got != 0 {
		t.Errorf("Median(nil) = %v, want 0", got)
	}
	if got := Median([]float64{3, 1, 2}); got != 2 {
		t.Errorf("Median([3,1,2]) = %v, want 2 (奇数個は中央の値)", got)
	}
	if got := Median([]float64{1, 2, 3, 4}); got != 2.5 {
		t.Errorf("Median([1,2,3,4]) = %v, want 2.5 (偶数個は中央2値の平均)", got)
	}

	values := []float64{3, 1, 2}
	Median(values)
	if values[0] != 3 || values[1] != 1 || values[2] != 2 {
		t.Errorf("Median は元の配列を破壊してはならない: got %v", values)
	}
}

func TestPercentileAndP95(t *testing.T) {
	if got := Percentile(nil, 95); got != 0 {
		t.Errorf("Percentile(nil, 95) = %v, want 0", got)
	}
	if got := P95(nil); got != 0 {
		t.Errorf("P95(nil) = %v, want 0", got)
	}

	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i + 1) // 1..100
	}
	if got := P95(values); got != 95 {
		t.Errorf("P95(1..100) = %v, want 95 (nearest-rank 法)", got)
	}

	if got := Percentile([]float64{5}, 95); got != 5 {
		t.Errorf("Percentile([5], 95) = %v, want 5", got)
	}
	if got := Percentile([]float64{1, 2}, 50); got != 1 {
		t.Errorf("Percentile([1,2], 50) = %v, want 1 (範囲外にならない)", got)
	}
}

func TestEstimateFPS(t *testing.T) {
	if got := EstimateFPS(16.666); math.Round(got) != 60 {
		t.Errorf("EstimateFPS(16.666) = %v, want ≈60", got)
	}
	if got := EstimateFPS(0); got != 0 {
		t.Errorf("EstimateFPS(0) = %v, want 0", got)
	}
	if got := EstimateFPS(-1); got != 0 {
		t.Errorf("EstimateFPS(-1) = %v, want 0", got)
	}
}

func TestJudgePerformance(t *testing.T) {
	t.Run("60Hz で張り付き(16.7ms・低ジッタ)はリフレッシュレート律速の注記付き", func(t *testing.T) {
		got := JudgePerformance(16.7, 16.7, 17.5) // p95 - 中央値 = 0.8ms < 2ms
		if got.Emoji != "✅" {
			t.Errorf("Emoji = %q, want ✅", got.Emoji)
		}
		if got.Label != "60fps維持(表示リフレッシュレート律速の可能性)" {
			t.Errorf("Label = %q", got.Label)
		}
		if !got.VsyncBound {
			t.Errorf("VsyncBound = false, want true")
		}
	})

	t.Run("75Hz で張り付き(13.3ms・低ジッタ)も律速と判定される(60Hz 前提にしない)", func(t *testing.T) {
		got := JudgePerformance(13.34, 13.3, 14.3) // p95 - 中央値 = 1.0ms < 2ms
		if got.Emoji != "✅" {
			t.Errorf("Emoji = %q, want ✅", got.Emoji)
		}
		if !strings.Contains(got.Label, "表示リフレッシュレート律速の可能性") {
			t.Errorf("Label = %q, want に含む: 表示リフレッシュレート律速の可能性", got.Label)
		}
		if !got.VsyncBound {
			t.Errorf("VsyncBound = false, want true")
		}
	})

	t.Run("平均が速くてもジッタが大きい場合は張り付きとみなさない", func(t *testing.T) {
		got := JudgePerformance(10, 9, 15) // p95 - 中央値 = 6ms
		if got.Emoji != "✅" || got.Label != "60fps維持" || got.VsyncBound {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("ジッタがちょうど 2ms は張り付きとみなさない(境界)", func(t *testing.T) {
		got := JudgePerformance(13, 13, 15) // p95 - 中央値 = 2.0ms
		if got.VsyncBound {
			t.Errorf("VsyncBound = true, want false (境界は含まない)")
		}
		if got.Label != "60fps維持" {
			t.Errorf("Label = %q, want 60fps維持", got.Label)
		}
	})

	t.Run("30〜60fps 相当は warning", func(t *testing.T) {
		got := JudgePerformance(25, 25, 26)
		if got.Emoji != "⚠️" || got.VsyncBound {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("30fps 未満は NG", func(t *testing.T) {
		got := JudgePerformance(100, 100, 110)
		if got.Emoji != "❌" || got.VsyncBound {
			t.Errorf("got = %+v", got)
		}
	})
}

func TestJudgeRatio(t *testing.T) {
	t.Run("現在の方が速い場合は faster 表記", func(t *testing.T) {
		got := JudgeRatio(300, 100, RatioOptions{}) // compare=300ms, current=100ms -> 3x faster
		if got.Emoji != "✅" || got.Label != "3.0x faster" {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("現在の方が遅い場合は slower 表記", func(t *testing.T) {
		got := JudgeRatio(100, 200, RatioOptions{}) // current is 2x slower
		if got.Emoji != "❌" || got.Label != "2.0x slower" {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("同値は同等(+0.0%)", func(t *testing.T) {
		got := JudgeRatio(13.34, 13.34, RatioOptions{})
		if got.Emoji != "➖" || got.Label != "≒同等(+0.0%)" {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("差が +5% ちょうどは同等(境界は同等帯に含む)", func(t *testing.T) {
		got := JudgeRatio(100, 105, RatioOptions{}) // current が 5% 遅い
		if got.Emoji != "➖" || got.Label != "≒同等(+5.0%)" {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("差が -5% ちょうども同等(符号付きで表示)", func(t *testing.T) {
		got := JudgeRatio(100, 95, RatioOptions{}) // current が 5% 速い
		if got.Emoji != "➖" || got.Label != "≒同等(-5.0%)" {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("差が +5% を超えたら slower 表記(境界のすぐ外)", func(t *testing.T) {
		got := JudgeRatio(100, 105.2, RatioOptions{}) // +5.2%
		if got.Emoji != "❌" || !strings.Contains(got.Label, "slower") {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("差が -5% を超えたら faster 表記(境界のすぐ外)", func(t *testing.T) {
		got := JudgeRatio(100, 94.8, RatioOptions{}) // -5.2%
		if got.Emoji != "✅" || !strings.Contains(got.Label, "faster") {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("双方ディスプレイ同期律速なら比較不能", func(t *testing.T) {
		got := JudgeRatio(13.34, 13.34, RatioOptions{BothVsyncBound: true})
		if got.Emoji != "➖" || got.Label != "比較不能(双方ディスプレイ同期律速)" {
			t.Errorf("got = %+v", got)
		}
		if got.Ratio != nil {
			t.Errorf("Ratio = %v, want nil", *got.Ratio)
		}
	})

	t.Run("bothVsyncBound でも 0 以下の値は N/A が優先", func(t *testing.T) {
		got := JudgeRatio(0, 100, RatioOptions{BothVsyncBound: true})
		if got.Label != "N/A" {
			t.Errorf("Label = %q, want N/A", got.Label)
		}
	})

	t.Run("0 以下の値は N/A", func(t *testing.T) {
		if got := JudgeRatio(0, 100, RatioOptions{}); got.Label != "N/A" {
			t.Errorf("Label = %q, want N/A", got.Label)
		}
		if got := JudgeRatio(100, 0, RatioOptions{}); got.Label != "N/A" {
			t.Errorf("Label = %q, want N/A", got.Label)
		}
	})
}

func TestRenderMarkdownTable(t *testing.T) {
	got := RenderMarkdownTable([]string{"A", "B"}, [][]string{{"1", "2"}})
	want := "| A | B |\n| --- | --- |\n| 1 | 2 |"
	if got != want {
		t.Errorf("RenderMarkdownTable = %q, want %q", got, want)
	}
}

func TestMax(t *testing.T) {
	if got := Max(nil); got != 0 {
		t.Errorf("Max(nil) = %v, want 0", got)
	}
	if got := Max([]float64{3, 1, 2}); got != 3 {
		t.Errorf("Max([3,1,2]) = %v, want 3", got)
	}
}
