package bench

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Mean は算術平均を返す。空スライスの場合は 0 を返す。
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// sortedCopy は values を破壊しない昇順ソート済みコピーを返す
// (Median/Percentile が呼び出し側のスライスを変更しないようにするため)。
func sortedCopy(values []float64) []float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	return sorted
}

// Median は中央値を返す。空スライスの場合は 0 を返す。要素数が偶数の
// 場合は中央 2 値の平均を返す。引数のスライスは変更しない。
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := sortedCopy(values)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// Percentile は nearest-rank 法によるパーセンタイルを返す(単純で決定的)。
// p は 0〜100。空スライスの場合は 0 を返す。
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := sortedCopy(values)
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx > len(sorted)-1 {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// P95 は Percentile(values, 95) のショートハンド。
func P95(values []float64) float64 {
	return Percentile(values, 95)
}

// Max は最大値を返す。空スライスの場合は 0 を返す(TS 版には存在しない
// が、Go 版の詳細表が frame ms の最大値も表示するために追加した)。
func Max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// EstimateFPS は平均フレーム時間(ms)から概算 FPS を求める。meanMs が
// 0 以下の場合は 0 を返す。
func EstimateFPS(meanMs float64) float64 {
	if meanMs > 0 {
		return 1000 / meanMs
	}
	return 0
}

// refreshBoundJitterMs は、リフレッシュレート律速とみなす
// (p95 - 中央値) ジッタ上限(ms)。
const refreshBoundJitterMs = 2.0

// Judgement はフレーム時間の統計から導かれるシナリオの性能判定。
type Judgement struct {
	Emoji      string
	Label      string
	VsyncBound bool
}

// JudgePerformance はフレーム時間の統計からシナリオの性能を判定する。
//   - 平均 17.5ms 以下: 「60fps維持」(✅)。さらに (p95Ms - medianMs) が
//     refreshBoundJitterMs 未満(= フレーム時間が一定値に張り付いている)
//     なら、ディスプレイのリフレッシュレートに律速されている可能性が
//     高いため、「表示リフレッシュレート律速の可能性」の注記と
//     VsyncBound フラグを付ける。60Hz(16.7ms)前提の閾値判定はしない
//     (75Hz なら 13.3ms 等で張り付くため)。
//   - 平均 33.4ms 以下: 「30〜60fps」(⚠️)
//   - それ以外: 「30fps未満」(❌)
//
// medianMs/p95Ms が meanMs と同じ値の場合はジッタ 0 とみなす(張り付き
// 扱い)。TS 版では medianMs/p95Ms は省略可能なデフォルト引数だったが、
// Go には default 引数がないため、呼び出し側が明示的に meanMs を渡す
// ことで同じ挙動を再現できる。
func JudgePerformance(meanMs, medianMs, p95Ms float64) Judgement {
	if meanMs <= 17.5 {
		vsyncBound := p95Ms-medianMs < refreshBoundJitterMs
		label := "60fps維持"
		if vsyncBound {
			label = "60fps維持(表示リフレッシュレート律速の可能性)"
		}
		return Judgement{Emoji: "✅", Label: label, VsyncBound: vsyncBound}
	}
	if meanMs <= 33.4 {
		return Judgement{Emoji: "⚠️", Label: "30〜60fps"}
	}
	return Judgement{Emoji: "❌", Label: "30fps未満"}
}

// equivalenceBandPct は「≒同等」とみなす差分の上限(±%)。
const equivalenceBandPct = 5.0

// RatioJudgement は --compare の比較結果。Ratio は「双方ディスプレイ
// 同期律速」で比較不能な場合のみ nil になる(TS 版の ratio: null に
// 対応)。
type RatioJudgement struct {
	Ratio *float64
	Emoji string
	Label string
}

// RatioOptions は JudgeRatio のオプション引数。
type RatioOptions struct {
	// BothVsyncBound が true の場合、current・compare 双方がディスプレイ
	// 同期律速と判定されているとみなし、倍率に意味がないため
	// 「比較不能」を返す。
	BothVsyncBound bool
}

// JudgeRatio は比較対象(--compare ref)の平均フレーム時間と現在の平均
// フレーム時間を比べ、人が読みやすい倍率ラベルを生成する。
//   - currentMs か compareMs が 0 以下の場合は測定不能として「N/A」を
//     返す(BothVsyncBound より優先される)。
//   - opts.BothVsyncBound が true の場合、フレーム時間はどちらもリフ
//     レッシュレートの上限に張り付いているだけで倍率に意味がないため
//     「➖ 比較不能」を返す(Ratio は nil)。
//   - 差が ±5% 以内なら「➖ ≒同等(符号付き差分%)」。差分% は
//     (currentMs - compareMs) / compareMs で、正 = 現在の方が遅い。
//   - ±5% を超えた場合のみ ratio = compareMs / currentMs で、
//     ratio >= 1 なら ✅ "Nx faster"、ratio < 1 なら ❌ "Nx slower"。
func JudgeRatio(compareMs, currentMs float64, opts RatioOptions) RatioJudgement {
	if currentMs <= 0 || compareMs <= 0 {
		zero := 0.0
		return RatioJudgement{Ratio: &zero, Emoji: "❌", Label: "N/A"}
	}
	if opts.BothVsyncBound {
		return RatioJudgement{Ratio: nil, Emoji: "➖", Label: "比較不能(双方ディスプレイ同期律速)"}
	}

	diffPct := (currentMs - compareMs) / compareMs * 100
	ratio := compareMs / currentMs

	if math.Abs(diffPct) <= equivalenceBandPct {
		sign := ""
		if diffPct >= 0 {
			sign = "+"
		}
		return RatioJudgement{Ratio: &ratio, Emoji: "➖", Label: fmt.Sprintf("≒同等(%s%.1f%%)", sign, diffPct)}
	}
	if ratio >= 1 {
		return RatioJudgement{Ratio: &ratio, Emoji: "✅", Label: fmt.Sprintf("%.1fx faster", ratio)}
	}
	return RatioJudgement{Ratio: &ratio, Emoji: "❌", Label: fmt.Sprintf("%.1fx slower", 1/ratio)}
}

// RenderMarkdownTable はヘッダ配列と行セル配列から GitHub Flavored
// Markdown のテーブルを生成する。
func RenderMarkdownTable(headers []string, rows [][]string) string {
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, "| "+strings.Join(headers, " | ")+" |")

	sep := make([]string, len(headers))
	for i := range sep {
		sep[i] = "---"
	}
	lines = append(lines, "| "+strings.Join(sep, " | ")+" |")

	for _, row := range rows {
		lines = append(lines, "| "+strings.Join(row, " | ")+" |")
	}
	return strings.Join(lines, "\n")
}
