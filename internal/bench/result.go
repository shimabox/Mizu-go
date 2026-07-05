package bench

// ScenarioResult は 1 シナリオの計測結果。cmd/bench の run-one サブ
// プロセスが JSON として書き出し、オーケストレータプロセス(および
// -compare の比較対象プロセス)がそれを読み込んでレポートを組み立てる
// (porting-plan の設計における "サブプロセス連携と -compare 用" の
// JSON シリアライズ)。
type ScenarioResult struct {
	// Name はシナリオ名("default"/"500"/"1000"/"3000")。
	Name string `json:"name"`
	// Label は Markdown レポートに表示する表示名。
	Label string `json:"label"`
	// FrameMs はウォームアップ後に収集したフレーム間隔(壁時計、ms)の
	// サンプル列。
	FrameMs []float64 `json:"frameMs"`
	// UpdateMs はウォームアップ後に収集した Simulator.Update() の
	// 実行時間(ms)のサンプル列。
	UpdateMs []float64 `json:"updateMs"`
	// Counts は計測終了時点での kind 別の生存粒子数。
	Counts map[string]int `json:"counts"`
}

// ScenarioStats は ScenarioResult から導出した、レポート表示用の統計値。
// ScenarioResult 自体は生サンプルのみを保持する(JSON 往復・サブ
// プロセス連携のため軽量に保つ)一方、この型は Mean/Median/P95/Max・
// 判定などレポートが必要とする値をまとめて保持する。
type ScenarioStats struct {
	Name  string
	Label string

	Frames   int
	MeanMs   float64
	MedianMs float64
	P95Ms    float64
	MaxMs    float64
	FPS      float64

	// UpdateMeanMs/UpdateMedianMs/UpdateP95Ms は Update ms のサンプルが
	// 1 件もない場合 nil になる("N/A" 表示のため)。
	UpdateMeanMs   *float64
	UpdateMedianMs *float64
	UpdateP95Ms    *float64

	Counts map[string]int

	Judge Judgement
}

// TotalCount は Counts の合計値を返す。
func (s ScenarioStats) TotalCount() int {
	total := 0
	for _, c := range s.Counts {
		total += c
	}
	return total
}

// Summarize は ScenarioResult の生サンプルから ScenarioStats を計算する。
func Summarize(r ScenarioResult) ScenarioStats {
	meanMs := Mean(r.FrameMs)
	medianMs := Median(r.FrameMs)
	p95Ms := P95(r.FrameMs)

	stats := ScenarioStats{
		Name:     r.Name,
		Label:    r.Label,
		Frames:   len(r.FrameMs),
		MeanMs:   meanMs,
		MedianMs: medianMs,
		P95Ms:    p95Ms,
		MaxMs:    Max(r.FrameMs),
		FPS:      EstimateFPS(meanMs),
		Counts:   r.Counts,
		Judge:    JudgePerformance(meanMs, medianMs, p95Ms),
	}

	if len(r.UpdateMs) > 0 {
		updateMean := Mean(r.UpdateMs)
		updateMedian := Median(r.UpdateMs)
		updateP95 := P95(r.UpdateMs)
		stats.UpdateMeanMs = &updateMean
		stats.UpdateMedianMs = &updateMedian
		stats.UpdateP95Ms = &updateP95
	}

	return stats
}
