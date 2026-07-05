// Package bench は、ベンチマークツール用の ebiten 非依存な純粋関数群
// (統計計算・シナリオ定義・結果の JSON シリアライズ・Markdown レポート
// 生成)を提供する。ここに置く関数は I/O やブラウザ・ウィンドウ操作を
// 一切行わないため、`go test ./internal/bench/...` だけで検証できる
// (Ebitengine 自体はヘッドレス環境では実行できないため、実際の計測は
// cmd/bench が担う)。
//
// これは Mizu-ts の scripts/bench/stats.mjs 他の Go 移植である
// (tests/bench/stats.test.ts が移植の一次情報)。
package bench
