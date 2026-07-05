package core

import "testing"

func TestNewBounds_InitialValues(t *testing.T) {
	b := NewBounds(800, 600)
	if b.W() != 800 {
		t.Errorf("W() = %v, want 800", b.W())
	}
	if b.H() != 600 {
		t.Errorf("H() = %v, want 600", b.H())
	}
}

func TestBounds_SetUpdatesValues(t *testing.T) {
	b := NewBounds(800, 600)
	b.Set(1024, 768)
	if b.W() != 1024 {
		t.Errorf("W() = %v, want 1024", b.W())
	}
	if b.H() != 768 {
		t.Errorf("H() = %v, want 768", b.H())
	}
}

func TestBounds_SharedReferenceSeesUpdates(t *testing.T) {
	b := NewBounds(100, 100)

	// 複数の利用者(例えば behavior・factory・simulator)が同じ *Bounds を
	// 保持していて、そのうちの1つがウィンドウリサイズでサイズを変更する
	// 状況を再現する。
	consumerW := func() float64 { return b.W() }

	b.Set(200, 150)

	if got := consumerW(); got != 200 {
		t.Errorf("consumer sees W() = %v, want 200 (shared reference must observe Set)", got)
	}
}
