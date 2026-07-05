//go:build !js

package main

// applyURLParams は非 js ビルドでは何もしない。URL クエリパラメータ
// はブラウザ上でのみ意味を持つ(wasm/js 向けの実装は params_js.go を
// 参照。Mizu-ts の src/main.ts:12-21 を反映)。
func applyURLParams(h, o *int, m *bool) {}

// initialWindowSize は非 js ビルドでは起動時ウィンドウサイズの定数を
// そのまま返す。デスクトップでは SetWindowSize でこの定数どおりの
// ウィンドウを実際に作るため、bounds と実ウィンドウは最初から一致して
// おり、js のような実サイズ取得(params_js.go を参照)は不要である。
func initialWindowSize() (w, h float64) {
	return initialWindowWidth, initialWindowHeight
}
