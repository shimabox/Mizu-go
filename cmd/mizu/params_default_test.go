//go:build !js

package main

import "testing"

// TestInitialWindowSize_DesktopReturnsConstants は、非 js ビルドの
// initialWindowSize が SetWindowSize に渡す定数と同じ値を返すことを
// 確認する。これにより、bounds の初期値と実際に作られるウィンドウの
// サイズが一致するというデスクトップの前提(params_default.go の doc
// コメントを参照)が崩れていないことを保証する。
func TestInitialWindowSize_DesktopReturnsConstants(t *testing.T) {
	w, h := initialWindowSize()
	if w != initialWindowWidth || h != initialWindowHeight {
		t.Errorf("initialWindowSize() = (%v, %v), want (%v, %v)",
			w, h, float64(initialWindowWidth), float64(initialWindowHeight))
	}
}
