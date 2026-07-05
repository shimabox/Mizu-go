package debug

import (
	"math"
	"regexp"
	"strings"
	"testing"
)

func TestFPS_NoFramesRecorded(t *testing.T) {
	s := NewStats()

	if got := s.FPS(); got != 0 {
		t.Errorf("FPS() = %v, want 0", got)
	}
	if text := s.Text(nil, 0, 0); !strings.Contains(text, "FPS: 0.0") {
		t.Errorf("Text() = %q, want to contain %q", text, "FPS: 0.0")
	}
}

func TestFPS_SixtyFPS(t *testing.T) {
	s := NewStats()
	for i := 0; i < 60; i++ {
		s.RecordFrame(16.67)
	}

	if got := s.FPS(); math.Abs(got-60) > 1 {
		t.Errorf("FPS() = %v, want ~60", got)
	}
}

func TestFPS_ThirtyFPS(t *testing.T) {
	s := NewStats()
	for i := 0; i < 60; i++ {
		s.RecordFrame(33.33)
	}

	if got := s.FPS(); math.Abs(got-30) > 1 {
		t.Errorf("FPS() = %v, want ~30", got)
	}
}

func TestFPS_MovingAverageWindow(t *testing.T) {
	s := NewStats()
	// 最初の 50 フレームは 10ms(100fps)。
	for i := 0; i < 50; i++ {
		s.RecordFrame(10)
	}
	// 次の 20 フレームは 50ms(20fps) - ウィンドウが 60 件を超えると、
	// 古い 10ms のフレームから追い出される。
	for i := 0; i < 20; i++ {
		s.RecordFrame(50)
	}

	got := s.FPS()
	if got <= 20 {
		t.Errorf("FPS() = %v, want > 20 (should be pulled up by the earlier 10ms frames)", got)
	}
	if got >= 100 {
		t.Errorf("FPS() = %v, want < 100 (should be pulled down by the later 50ms frames)", got)
	}
}

func TestFPS_WindowCappedAtSixty(t *testing.T) {
	s := NewStats()
	for i := 0; i < 100; i++ {
		s.RecordFrame(16.67)
	}

	if got := s.FPS(); math.Abs(got-60) > 1 {
		t.Errorf("FPS() = %v, want ~60 (only the most recent 60 frames should count)", got)
	}
}

func TestText_FrameInterval_OneDecimal(t *testing.T) {
	s := NewStats()
	s.RecordFrame(16.58)

	text := s.Text(nil, 0, 0)
	if !strings.Contains(text, "Frame: 16.6ms") {
		t.Errorf("Text() = %q, want to contain %q", text, "Frame: 16.6ms")
	}
}

func TestText_FrameInterval_NotFloored(t *testing.T) {
	s := NewStats()
	s.RecordFrame(0.58)

	text := s.Text(nil, 0, 0)
	re := regexp.MustCompile(`Frame: ([\d.]+)ms`)
	m := re.FindStringSubmatch(text)
	if m == nil {
		t.Fatalf("Text() = %q, no Frame: match found", text)
	}
	if m[1] == "0.0" {
		t.Errorf("Frame value = %q, want > 0 (0.58ms must not floor to 0)", m[1])
	}
}

func TestText_FrameInterval_RoundedToOneDecimal(t *testing.T) {
	s := NewStats()
	s.RecordFrame(16.667)

	text := s.Text(nil, 0, 0)
	re := regexp.MustCompile(`Frame: ([\d.]+)ms`)
	m := re.FindStringSubmatch(text)
	if m == nil {
		t.Fatalf("Text() = %q, no Frame: match found", text)
	}
	if !regexp.MustCompile(`^\d+\.\d$`).MatchString(m[1]) {
		t.Errorf("Frame value = %q, want exactly one decimal digit", m[1])
	}
}

func TestText_UpdateTime(t *testing.T) {
	s := NewStats()
	s.RecordUpdate(1.34)

	text := s.Text(nil, 0, 0)
	if !strings.Contains(text, "Update: 1.3ms") {
		t.Errorf("Text() = %q, want to contain %q", text, "Update: 1.3ms")
	}
}

func TestText_UpdateTime_NotFloored(t *testing.T) {
	s := NewStats()
	s.RecordUpdate(0.58)

	text := s.Text(nil, 0, 0)
	if !strings.Contains(text, "Update: 0.6ms") {
		t.Errorf("Text() = %q, want to contain %q", text, "Update: 0.6ms")
	}
}

func TestText_FrameAndUpdate_AreIndependent(t *testing.T) {
	s := NewStats()
	// rAF-equivalent interval: 80ms, but Update() itself only took 69.2ms.
	s.RecordFrame(80)
	s.RecordUpdate(69.2)

	text := s.Text(nil, 0, 0)
	if !strings.Contains(text, "Frame: 80.0ms") {
		t.Errorf("Text() = %q, want to contain %q", text, "Frame: 80.0ms")
	}
	if !strings.Contains(text, "Update: 69.2ms") {
		t.Errorf("Text() = %q, want to contain %q", text, "Update: 69.2ms")
	}
}

func TestText_Counts_FixedOrderZeroFilled(t *testing.T) {
	s := NewStats()
	counts := map[string]int{"H": 30, "O": 50} // H2, H2o intentionally absent

	text := s.Text(counts, 0, 0)
	for _, want := range []string{"H: 30", "H2: 0", "O: 50", "H2o: 0", "Total: 80"} {
		if !strings.Contains(text, want) {
			t.Errorf("Text() = %q, want to contain %q", text, want)
		}
	}
}

func TestText_Counts_AllKindsAndTotal(t *testing.T) {
	s := NewStats()
	counts := map[string]int{"H": 30, "H2": 5, "O": 50, "H2o": 100}

	text := s.Text(counts, 0, 0)
	if !strings.Contains(text, "Total: 185") {
		t.Errorf("Text() = %q, want to contain %q", text, "Total: 185")
	}

	hIdx := strings.Index(text, "H: 30")
	h2Idx := strings.Index(text, "H2: 5")
	oIdx := strings.Index(text, "O: 50")
	h2oIdx := strings.Index(text, "H2o: 100")
	if !(hIdx < h2Idx && h2Idx < oIdx && oIdx < h2oIdx) {
		t.Errorf("kind lines out of fixed order: H=%d H2=%d O=%d H2o=%d", hIdx, h2Idx, oIdx, h2oIdx)
	}
}

func TestText_LineOrderAndCount(t *testing.T) {
	s := NewStats()
	s.RecordFrame(16.67)
	s.RecordUpdate(1.3)
	counts := map[string]int{"H": 10, "O": 20} // H2, H2o zero-filled

	text := s.Text(counts, 59.9, 60.0)
	lines := strings.Split(text, "\n")

	wantPrefixes := []string{"FPS:", "Frame:", "Update:", "ActualFPS:", "H:", "H2:", "O:", "H2o:", "Total:"}
	if len(lines) != len(wantPrefixes) {
		t.Fatalf("got %d lines, want %d\ntext:\n%s", len(lines), len(wantPrefixes), text)
	}
	for i, prefix := range wantPrefixes {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("line %d = %q, want prefix %q", i, lines[i], prefix)
		}
	}
}

func TestText_ActualFPSAndTPS(t *testing.T) {
	s := NewStats()

	text := s.Text(nil, 59.9, 60.0)
	if !strings.Contains(text, "ActualFPS: 59.9") {
		t.Errorf("Text() = %q, want to contain %q", text, "ActualFPS: 59.9")
	}
	if !strings.Contains(text, "ActualTPS: 60.0") {
		t.Errorf("Text() = %q, want to contain %q", text, "ActualTPS: 60.0")
	}
}
