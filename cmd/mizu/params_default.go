//go:build !js

package main

// applyURLParams は非 js ビルドでは何もしない。URL クエリパラメータ
// はブラウザ上でのみ意味を持つ(wasm/js 向けの実装は params_js.go を
// 参照。Mizu-ts の src/main.ts:12-21 を反映)。
func applyURLParams(h, o *int, m *bool) {}
