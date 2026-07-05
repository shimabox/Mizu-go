//go:build js

package main

import (
	"net/url"
	"strconv"
	"strings"
	"syscall/js"
)

// applyURLParams は window.location.search を読み取り、h、o、m を
// その場で上書きする。これは Mizu-ts の src/main.ts:12-21 と同じ意味
// 付けに従う(getSafeNumber: クエリの生の値に対する Number.parseInt。
// 結果が NaN の場合は既存のデフォルト値にフォールバックする。m=1 で
// 計測モードを有効化)。パースに失敗した値はそのまま変更されない
// ので、呼び出し側はこれを呼ぶ前に h/o/m にフラグのデフォルト値を
// あらかじめセットしておくこと。
func applyURLParams(h, o *int, m *bool) {
	search := js.Global().Get("location").Get("search").String()
	values, err := url.ParseQuery(strings.TrimPrefix(search, "?"))
	if err != nil {
		return
	}

	if v := values.Get("h"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*h = n
		}
	}
	if v := values.Get("o"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*o = n
		}
	}
	if values.Get("m") == "1" {
		*m = true
	}
}

// initialWindowSize は js ビルドでは window.innerWidth/innerHeight
// (CSS ピクセル。Ebitengine が Layout に渡す device-independent pixels
// と同じ単位)を返す。
//
// 初期粒子の生成(Simulator.Init)は RunGame(最初の Layout)より前に
// 行われるため、js ではここで実ウィンドウサイズを使って bounds を初期化
// しないと、スマホなど幅 768px 未満の画面で粒子のサイズスケール
// (1.2 扱い)と個数スケール(1.5 扱い)が誤り、最初の Layout 以降に
// 生成される粒子とサイズの異なる初期粒子が混在してしまう(Mizu-ts は
// 起動時に実ウィンドウサイズでキャンバスを作ってから粒子を生成するため、
// この問題が存在しない)。
//
// 取得値が 0 以下などの異常値の場合は、デスクトップと同じ
// initialWindowWidth/initialWindowHeight にフォールバックする。
func initialWindowSize() (w, h float64) {
	window := js.Global().Get("window")
	w = window.Get("innerWidth").Float()
	h = window.Get("innerHeight").Float()
	if w <= 0 || h <= 0 {
		return initialWindowWidth, initialWindowHeight
	}
	return w, h
}
