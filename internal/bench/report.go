package bench

import (
	"fmt"
	"strings"
	"time"
)

// fixedKindOrder は「終了時粒子数」セルを常に同じ順序で表示するための
// 固定順序。internal/debug.FixedKinds と同じ並びだが、internal/debug は
// レンダー層寄りのパッケージのため、ここでは独立した定数として持つ
// (internal/bench は ebiten 非依存という制約を保つため、internal/debug
// への依存自体は問題ないが、キーワードの重複を避けたい単純な理由で
// 独自に定義している)。
var fixedKindOrder = []string{"H", "H2", "O", "H2o"}

// formatCounts は kind 別カウントを "H:10 H2:2 O:3 H2o:5 (計20)" の
// ような 1 セル分の文字列に整形する。fixedKindOrder にない kind が
// counts に含まれる場合も末尾に追加する(将来 kind が増えても取りこぼ
// さないため)。
func formatCounts(counts map[string]int) string {
	seen := make(map[string]bool, len(counts))
	parts := make([]string, 0, len(counts)+1)
	total := 0

	appendKind := func(kind string) {
		count := counts[kind]
		total += count
		parts = append(parts, fmt.Sprintf("%s:%d", kind, count))
		seen[kind] = true
	}

	for _, kind := range fixedKindOrder {
		appendKind(kind)
	}
	for kind := range counts {
		if !seen[kind] {
			appendKind(kind)
		}
	}

	return fmt.Sprintf("%s (計%d)", strings.Join(parts, " "), total)
}

// formatOptionalMs は *float64 を小数点以下 2 桁で整形する。nil の
// 場合は "N/A" を返す(Update ms のサンプルが 1 件もないシナリオ用)。
func formatOptionalMs(v *float64) string {
	if v == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.2f", *v)
}

// BuildScenarioTable は、シナリオごとの frame ms (mean/median/p95/max)・
// 換算 FPS・update ms (mean/median/p95)・終了時粒子数・判定・計測
// フレーム数をまとめた表を生成する。TS 版の buildSummaryTable /
// buildDetailTable を 1 本化し、列構成を拡張したもの(porting-plan の
// Go 版設計に合わせた意図的な変更)。
func BuildScenarioTable(entries []ScenarioStats) string {
	headers := []string{
		"シナリオ",
		"平均 (ms)", "中央値 (ms)", "p95 (ms)", "最大 (ms)", "FPS",
		"Update 平均 (ms)", "Update 中央値 (ms)", "Update p95 (ms)",
		"終了時粒子数",
		"判定",
		"計測フレーム数",
	}
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			e.Label,
			fmt.Sprintf("%.2f", e.MeanMs),
			fmt.Sprintf("%.2f", e.MedianMs),
			fmt.Sprintf("%.2f", e.P95Ms),
			fmt.Sprintf("%.2f", e.MaxMs),
			fmt.Sprintf("%.1f", e.FPS),
			formatOptionalMs(e.UpdateMeanMs),
			formatOptionalMs(e.UpdateMedianMs),
			formatOptionalMs(e.UpdateP95Ms),
			formatCounts(e.Counts),
			fmt.Sprintf("%s %s", e.Judge.Emoji, e.Judge.Label),
			fmt.Sprintf("%d", e.Frames),
		})
	}
	return RenderMarkdownTable(headers, rows)
}

// ComparisonEntry は --compare 比較表の 1 行分。
type ComparisonEntry struct {
	Label     string
	CompareMs float64
	CurrentMs float64
	DiffPct   float64
	Ratio     RatioJudgement
}

// NewComparisonEntry は current/compare 双方の ScenarioStats から
// ComparisonEntry を組み立てる。双方が vsyncBound(ディスプレイ同期律速)
// と判定されている場合は倍率を「比較不能」とする(TS 版の
// bothVsyncBound と同じ考え方)。
func NewComparisonEntry(current, compare ScenarioStats) ComparisonEntry {
	bothVsyncBound := current.Judge.VsyncBound && compare.Judge.VsyncBound
	diffPct := 0.0
	if compare.MeanMs > 0 {
		diffPct = (current.MeanMs - compare.MeanMs) / compare.MeanMs * 100
	}
	return ComparisonEntry{
		Label:     current.Label,
		CompareMs: compare.MeanMs,
		CurrentMs: current.MeanMs,
		DiffPct:   diffPct,
		Ratio:     JudgeRatio(compare.MeanMs, current.MeanMs, RatioOptions{BothVsyncBound: bothVsyncBound}),
	}
}

// BuildComparisonTable は --compare 比較表(現在 vs 比較対象、差分%、
// 判定ラベル)を生成する。
func BuildComparisonTable(entries []ComparisonEntry) string {
	headers := []string{"シナリオ", "比較対象 (ms)", "現在 (ms)", "差分", "判定"}
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		sign := ""
		if e.DiffPct >= 0 {
			sign = "+"
		}
		rows = append(rows, []string{
			e.Label,
			fmt.Sprintf("%.2f", e.CompareMs),
			fmt.Sprintf("%.2f", e.CurrentMs),
			fmt.Sprintf("%s%.1f%%", sign, e.DiffPct),
			fmt.Sprintf("%s %s", e.Ratio.Emoji, e.Ratio.Label),
		})
	}
	return RenderMarkdownTable(headers, rows)
}

// Environment はレポートのヘッダに載せる実行環境情報。
type Environment struct {
	OS, Arch, GoVersion string
	GeneratedAt         time.Time
	WarmupMs            int

	// CurrentCommit は現在の作業ツリーの識別情報(例: "main@abcdef0")。
	// 取得できない場合は空文字列でよい(その行自体を省略する)。
	CurrentCommit string
	// CompareRef は --compare に指定された git ref。比較を行わない
	// 場合は空文字列。
	CompareRef string
	// CompareCommit は CompareRef を解決したコミット情報(取得できた
	// 場合のみ)。
	CompareCommit string
}

// Report は生成された Markdown 全文と、コンソール向けの短いサマリ。
type Report struct {
	Markdown       string
	ConsoleSummary string
}

// readingGuide は Mizu-ts の report.mjs の READING_GUIDE を Go 版の
// 用語(rAF ではなく壁時計フレーム間隔、renderFrame ではなく
// Simulator.Update)に合わせて訳したもの。
const readingGuide = `## レポートの読み方

- **フレーム時間の一定値への張り付き**: ディスプレイのリフレッシュレート(60Hz なら 16.7ms、75Hz なら 13.3ms など)の上限で頭打ちになっている状態です。それ以上は速くなりません。
- **Frame と Update の乖離**: ` + "`平均 (ms)`" + `(フレーム間隔)が ` + "`Update 平均 (ms)`" + `(Simulator.Update の実行時間)より大きい場合、その差分は描画・GC 等のコストです。
- **粒子数列も見ること**: H2o の滞留数がフレーム時間を支配するため、終了時粒子数は毎回確認してください。
- **絶対値の比較は不可**: 環境ノイズがあるため、別日に計測した絶対値同士を比較しないでください。比較する場合は必ず ` + "`-compare`" + ` で同一セッション・同一マシン状態で計測してください。`

// BuildReport は Markdown レポート全文とコンソール向けサマリを組み立
// てる。current と compare は同じ並び(シナリオ順)であることを前提と
// する。compare が nil の場合は比較セクションを省略する。
func BuildReport(env Environment, current []ScenarioResult, compare []ScenarioResult) Report {
	var b strings.Builder

	fmt.Fprintln(&b, "# Mizu-go ベンチマークレポート")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- 実行日時: %s\n", env.GeneratedAt.Format("2006-01-02 15:04:05"))
	if env.CurrentCommit != "" {
		fmt.Fprintf(&b, "- 現在: `%s`\n", env.CurrentCommit)
	}
	if env.CompareRef != "" {
		if env.CompareCommit != "" {
			fmt.Fprintf(&b, "- 比較対象: `%s`(`%s`)\n", env.CompareRef, env.CompareCommit)
		} else {
			fmt.Fprintf(&b, "- 比較対象: `%s`\n", env.CompareRef)
		}
	}
	fmt.Fprintf(&b, "- OS: %s/%s\n", env.OS, env.Arch)
	fmt.Fprintf(&b, "- Go: %s\n", env.GoVersion)
	fmt.Fprintf(&b, "- プロトコル: warmup %dms\n", env.WarmupMs)
	fmt.Fprintln(&b)

	currentEntries := make([]ScenarioStats, 0, len(current))
	for _, r := range current {
		currentEntries = append(currentEntries, Summarize(r))
	}

	fmt.Fprintln(&b, "## 結果")
	fmt.Fprintln(&b)
	scenarioTable := BuildScenarioTable(currentEntries)
	fmt.Fprintln(&b, scenarioTable)
	fmt.Fprintln(&b)

	consoleSummary := "結果:\n" + scenarioTable

	if compare != nil {
		compareEntries := make([]ScenarioStats, 0, len(compare))
		for _, r := range compare {
			compareEntries = append(compareEntries, Summarize(r))
		}

		fmt.Fprintln(&b, "## 結果(比較対象)")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, BuildScenarioTable(compareEntries))
		fmt.Fprintln(&b)

		fmt.Fprintln(&b, "## 比較(-compare)")
		fmt.Fprintln(&b)
		n := len(currentEntries)
		if len(compareEntries) < n {
			n = len(compareEntries)
		}
		comparisonEntries := make([]ComparisonEntry, 0, n)
		for i := 0; i < n; i++ {
			comparisonEntries = append(comparisonEntries, NewComparisonEntry(currentEntries[i], compareEntries[i]))
		}
		fmt.Fprintln(&b, BuildComparisonTable(comparisonEntries))
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, readingGuide)

	return Report{Markdown: b.String(), ConsoleSummary: consoleSummary}
}
