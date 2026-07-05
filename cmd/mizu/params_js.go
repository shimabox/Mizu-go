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
