package bench

import (
	"encoding/json"
	"testing"
)

func TestScenarioResultJSONRoundTrip(t *testing.T) {
	want := ScenarioResult{
		Name:     "500",
		Label:    "h=500, o=500",
		FrameMs:  []float64{16.0, 17.0, 18.0},
		UpdateMs: []float64{1.0, 1.2, 0.9},
		Counts:   map[string]int{"H": 10, "H2": 2, "O": 3, "H2o": 40},
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got ScenarioResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Name != want.Name || got.Label != want.Label {
		t.Errorf("got = %+v, want %+v", got, want)
	}
	if len(got.FrameMs) != 3 || got.FrameMs[1] != 17.0 {
		t.Errorf("FrameMs round-trip failed: %v", got.FrameMs)
	}
	if len(got.UpdateMs) != 3 {
		t.Errorf("UpdateMs round-trip failed: %v", got.UpdateMs)
	}
	if got.Counts["H2o"] != 40 {
		t.Errorf("Counts round-trip failed: %v", got.Counts)
	}
}

func TestSummarize(t *testing.T) {
	r := ScenarioResult{
		Name:     "default",
		Label:    "デフォルト (h=30, o=50)",
		FrameMs:  []float64{10, 20, 30, 40},
		UpdateMs: []float64{1, 2, 3, 4},
		Counts:   map[string]int{"H": 5, "H2": 1, "O": 2, "H2o": 3},
	}

	entry := Summarize(r)

	if entry.Label != r.Label {
		t.Errorf("Label = %q, want %q", entry.Label, r.Label)
	}
	if entry.MeanMs != Mean(r.FrameMs) {
		t.Errorf("MeanMs = %v, want %v", entry.MeanMs, Mean(r.FrameMs))
	}
	if entry.MedianMs != Median(r.FrameMs) {
		t.Errorf("MedianMs = %v, want %v", entry.MedianMs, Median(r.FrameMs))
	}
	if entry.P95Ms != P95(r.FrameMs) {
		t.Errorf("P95Ms = %v, want %v", entry.P95Ms, P95(r.FrameMs))
	}
	if entry.MaxMs != Max(r.FrameMs) {
		t.Errorf("MaxMs = %v, want %v", entry.MaxMs, Max(r.FrameMs))
	}
	if entry.FPS != EstimateFPS(Mean(r.FrameMs)) {
		t.Errorf("FPS = %v, want %v", entry.FPS, EstimateFPS(Mean(r.FrameMs)))
	}
	if entry.UpdateMeanMs == nil || *entry.UpdateMeanMs != Mean(r.UpdateMs) {
		t.Errorf("UpdateMeanMs = %v, want %v", entry.UpdateMeanMs, Mean(r.UpdateMs))
	}
	total := 0
	for _, c := range r.Counts {
		total += c
	}
	if entry.TotalCount() != total {
		t.Errorf("TotalCount() = %d, want %d", entry.TotalCount(), total)
	}
	if entry.Frames != len(r.FrameMs) {
		t.Errorf("Frames = %d, want %d", entry.Frames, len(r.FrameMs))
	}

	wantJudge := JudgePerformance(entry.MeanMs, entry.MedianMs, entry.P95Ms)
	if entry.Judge != wantJudge {
		t.Errorf("Judge = %+v, want %+v", entry.Judge, wantJudge)
	}
}

func TestSummarize_NoUpdateSamplesYieldsNilUpdateStats(t *testing.T) {
	r := ScenarioResult{
		Name:    "default",
		Label:   "デフォルト (h=30, o=50)",
		FrameMs: []float64{10, 20},
	}
	entry := Summarize(r)
	if entry.UpdateMeanMs != nil || entry.UpdateMedianMs != nil || entry.UpdateP95Ms != nil {
		t.Errorf("update stats should be nil when there are no samples: %+v", entry)
	}
}
