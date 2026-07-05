package bench

import (
	"strings"
	"testing"
	"time"
)

func sampleResult(name, label string, frames []float64, updates []float64, counts map[string]int) ScenarioResult {
	return ScenarioResult{Name: name, Label: label, FrameMs: frames, UpdateMs: updates, Counts: counts}
}

func TestBuildScenarioTable(t *testing.T) {
	entries := []ScenarioStats{
		Summarize(sampleResult("default", "default", []float64{16, 17, 18}, []float64{1, 1.2, 0.9}, map[string]int{"H": 10, "H2o": 5})),
	}
	table := BuildScenarioTable(entries)

	if !strings.Contains(table, "シナリオ") || !strings.Contains(table, "平均") || !strings.Contains(table, "最大") {
		t.Errorf("table header missing expected columns: %s", table)
	}
	if !strings.Contains(table, "default") {
		t.Errorf("table missing scenario label: %s", table)
	}
	if !strings.Contains(table, "計") { // 終了時粒子数の合計表記
		t.Errorf("table missing total particle count marker: %s", table)
	}
}

func TestBuildScenarioTable_NoUpdateSamplesShowsNA(t *testing.T) {
	entries := []ScenarioStats{
		Summarize(sampleResult("default", "default", []float64{16, 17}, nil, map[string]int{"H": 1})),
	}
	table := BuildScenarioTable(entries)
	if !strings.Contains(table, "N/A") {
		t.Errorf("table should show N/A when there are no update samples: %s", table)
	}
}

func TestBuildComparisonTable(t *testing.T) {
	current := Summarize(sampleResult("500", "h=500, o=500", []float64{100, 100, 100}, nil, nil))
	compare := Summarize(sampleResult("500", "h=500, o=500", []float64{300, 300, 300}, nil, nil))

	entries := []ComparisonEntry{
		NewComparisonEntry(current, compare),
	}
	table := BuildComparisonTable(entries)

	if !strings.Contains(table, "差分") {
		t.Errorf("comparison table should include a diff%% column: %s", table)
	}
	if !strings.Contains(table, "3.0x faster") {
		t.Errorf("comparison table should include the judged ratio label: %s", table)
	}
}

func TestBuildReport_WithoutCompare(t *testing.T) {
	env := Environment{
		OS:          "darwin",
		Arch:        "arm64",
		GoVersion:   "go1.24.4",
		GeneratedAt: time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC),
		WarmupMs:    3000,
	}
	current := []ScenarioResult{
		sampleResult("default", "デフォルト (h=30, o=50)", []float64{16, 17, 18}, []float64{1, 1, 1}, map[string]int{"H": 3}),
	}

	report := BuildReport(env, current, nil)

	if !strings.Contains(report.Markdown, "# Mizu-go ベンチマークレポート") {
		t.Errorf("markdown missing title: %s", report.Markdown)
	}
	if !strings.Contains(report.Markdown, "darwin/arm64") {
		t.Errorf("markdown missing OS/arch: %s", report.Markdown)
	}
	if !strings.Contains(report.Markdown, "go1.24.4") {
		t.Errorf("markdown missing go version: %s", report.Markdown)
	}
	if strings.Contains(report.Markdown, "## 比較") {
		t.Errorf("markdown should not contain a comparison section without --compare: %s", report.Markdown)
	}
	if report.ConsoleSummary == "" {
		t.Errorf("ConsoleSummary should not be empty")
	}
}

func TestBuildReport_WithCompare(t *testing.T) {
	env := Environment{
		OS: "darwin", Arch: "arm64", GoVersion: "go1.24.4",
		GeneratedAt: time.Now(),
		WarmupMs:    3000,
		CompareRef:  "main",
	}
	current := []ScenarioResult{
		sampleResult("500", "h=500, o=500", []float64{100, 100, 100}, nil, map[string]int{"H": 1}),
	}
	compare := []ScenarioResult{
		sampleResult("500", "h=500, o=500", []float64{300, 300, 300}, nil, map[string]int{"H": 1}),
	}

	report := BuildReport(env, current, compare)

	if !strings.Contains(report.Markdown, "## 比較") {
		t.Errorf("markdown should contain a comparison section: %s", report.Markdown)
	}
	if !strings.Contains(report.Markdown, "main") {
		t.Errorf("markdown should mention the compare ref: %s", report.Markdown)
	}
}
