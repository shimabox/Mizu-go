// Package core はシミュレーション全体で共有される基礎的な型を提供する。
// 具体的には、Particle インターフェース、テストで behavior を決定的に
// するための Random 抽象、そして behavior/factory/simulator が参照する
// 共有の画面サイズである Bounds(ウィンドウリサイズがあっても互いに
// ずれないようにするためのもの)を含む。
//
// このパッケージは他の internal パッケージにも Ebitengine にも依存し
// ない。モジュール内のどこからでも安全に import できる状態を保つ必要
// がある。
package core
