# Mizu-go 移植プラン — Mizu-ts を Go + Ebitengine で書き換える

作成日: 2026-07-05
ステータス: 承認待ち → 実装プランは `2026-07-05-implementation-plan.md` を参照

## 1. 目的と結論

[Mizu-ts](https://github.com/shimabox/Mizu-ts)(TypeScript + Canvas 2D の水生成シミュレーション)を
Go + [Ebitengine v2](https://github.com/hajimehoshi/ebiten) で書き換える。

**実現可能性: 高い。** 調査の結果、全要件が Ebitengine の得意領域内であることを確認した。

| 要件 | Ebitengine での実現手段 |
|---|---|
| テキスト "H" / "H2" / "O" の描画 | `text/v2`(GoTextFace)。下付き文字は小さい Size の Face + Y オフセットで実現 |
| 水滴(放射状グラデーション円) | 起動時に 1 枚の `*ebiten.Image` へ事前レンダリング → `DrawImage` で拡縮 |
| 数百〜数千パーティクルの毎フレーム描画 | 自動バッチング + テクスチャアトラス。公式が 1 万スプライト超の実績(examples/sprites) |
| FPS 計測オーバーレイ | `ebiten.ActualFPS()` / `ActualTPS()` + text/v2 で自作オーバーレイ |
| Web デモ(GitHub Pages) | `GOOS=js GOARCH=wasm go build`。静的ファイルのみで公開可。cgo 不要でクロスコンパイル可 |

主要な設計原則は 1 点: **テキスト・図形を毎フレーム直接描画せず、起動時に `*ebiten.Image` に
事前レンダリングしてスプライトとして `DrawImage` する**(バッチングが効き、描画コストが桁違いに下がる)。

## 2. 確定した方針(ユーザー決定事項)

| 項目 | 決定 | 影響 |
|---|---|---|
| ビルドターゲット | デスクトップ(macOS)+ WASM(GitHub Pages) | web/ シェルと GH Actions デプロイを含める |
| 乱数 | Go 標準 `math/rand/v2`(mulberry32 は移植しない) | テスト期待値は Go 版のシード付き乱数で再計算。TS 版とフレーム単位の同一挙動は求めない |
| ウィンドウリサイズ | 追従対応を **追加**(TS 版は起動時固定) | SpatialGrid の再構築・画面境界の動的参照が必要(§5.4) |
| テスト | ロジック層(behaviors/physics/reactions/simulator/world)をフル移植 | 描画系テストは Ebitengine 向けに再設計。CI は ebiten 非依存パッケージのみなら xvfb 不要 |

## 3. 環境・前提

- ローカル: Go 1.24.4 (darwin/arm64)、Xcode CLT 導入済み → **Ebitengine v2.9 系の要件(Go 1.24+、macOS では cgo/clang)を満たす**
- Ebitengine 最新安定版: v2.9.9(2026-03 時点)
- `~/shimabox/github/Mizu-go` は現状空。git 未初期化(`git init` から始める)
- macOS からデスクトップ向けクロスコンパイルは不可(cgo)。WASM は可

## 4. TS → Go 対応マップ

### 4.1 パッケージ構成

```
Mizu-go/
├── go.mod                      # module github.com/shimabox/Mizu-go, go 1.24
├── cmd/mizu/main.go            # DI 組み立て + ebiten.RunGame(エントリポイント)
├── internal/core/              # Particle インターフェース, Random インターフェース
├── internal/behavior/          # RandomWalk, FallAndSway        ← ebiten 非依存
├── internal/physics/           # SpatialGrid, GridCollisionDetector, BruteForceCollisionDetector
├── internal/reaction/          # Rule, Registry, HHFusion, OxidationToWater
├── internal/particle/          # H, H2, O, H2o, Factory
├── internal/sim/               # World, Simulator(フレームパイプライン) ← ebiten 非依存
├── internal/render/            # スプライト事前生成, 描画, Game 実装  ← ebiten 依存はここに集約
├── internal/debug/             # StatsOverlay
├── assets/fonts/               # go:embed するフォント(または x/image/font/gofont)
├── web/index.html              # WASM シェル(GitHub Pages 用)
└── .github/workflows/          # CI(test) + Pages デプロイ
```

TS 版の「依存は core 方向へ一方向・循環ゼロ」を維持する。**ebiten への依存は
`internal/render` と `cmd/mizu` に閉じ込め、ロジック層は通常の `go test` で回るようにする**
(Ebitengine にヘッドレスモードはないため。CI で描画テストが必要になったら xvfb を使う)。

### 4.2 ファイル対応表

| Mizu-ts | Mizu-go | 備考 |
|---|---|---|
| `src/main.ts` | `cmd/mizu/main.go` | URL パラメータ → CLI フラグ(`-h`, `-o`, `-m`)。WASM 時は `syscall/js` で URL クエリを読む |
| `src/core/Particle.ts` | `internal/core/particle.go` | インターフェースをそのまま移植 |
| `src/core/Random.ts` | `internal/core/random.go` | `Random` interface(`Next() float64` ∈ [0,1))。実行時 = `rand/v2`、テスト = `rand.New(rand.NewPCG(seed, seed))` |
| `src/core/behaviors/*` | `internal/behavior/*` | 数式は §5.1 の通り忠実移植 |
| `src/core/renderers/*` | `internal/render/sprites.go` ほか | Canvas API 直接呼び出し → 事前レンダリング方式に再設計(§5.2) |
| `src/physics/*` | `internal/physics/*` | 定数含め忠実移植。リサイズ対応で `Resize(w,h)` を追加 |
| `src/reactions/*` | `internal/reaction/*` | 収支ルール(§5.3)を厳密に維持 |
| `src/particles/*` | `internal/particle/*` | H/H2/O は同一構造の薄い struct。H2o のみ着地で死亡 |
| `src/simulator/*` | `internal/sim/*` | renderFrame の update 部を `Update()`、描画部を `Draw()` に分離(§5.5) |
| `src/debug/StatsOverlay.ts` | `internal/debug/overlay.go` | DOM → text/v2 + 半透明矩形で自作 |

## 5. 忠実に再現すべき仕様(Mizu-ts 分析結果)

### 5.1 挙動の数式・定数

**RandomWalk**(H / H2 / O):
```
角度 = 2π × random.Next()
vx += 0.075 × cos(角度);  vy += 0.075 × sin(角度)
speed = √(vx²+vy²);  speed > 1.05 なら 1.05 にクランプ
ラップアラウンド: x > sw + w/2 → -(w/2)、x + w < 0 → sw + w/2(y も同様)
境界値: 「ちょうど」は折り返さず「+1」で折り返す(テストで固定)
```

**FallAndSway**(H2o):
```
x' = x + cos((y + random.Next()×5) / 100)   // 横揺れ
y' = y + size × 0.1                          // 落下
y >= 画面高 で着地 → markDead
```

**主要定数**(TS 版の行番号は分析レポート参照):

| 項目 | 値 |
|---|---|
| デフォルト初期数 | H=30, O=50(×粒子数スケール) |
| 粒子数スケール | 幅<768→1.0 / 768–1279→1.2 / ≥1280→1.5 |
| 粒子サイズスケール | 幅<768→1.0 / ≥768→1.2 ※**粒子数とは別の関数。この非対称を維持** |
| baseFontSize | 24 × サイズスケール(下付き文字は 18 × サイズスケール) |
| H2o 直径 | (random×10 + 18) × サイズスケール |
| SpatialGrid | MAX_PARTICLE_RADIUS=24, セルサイズ=48(3×3 近傍探索の不変条件) |
| 衝突判定 | dx²+dy² < (rA+rB)²(sqrt 回避) |
| 半径 | テキスト実測幅 / 2(text.Measure で計測、kind 別にキャッシュ) |

### 5.2 見た目(描画仕様)

| 粒子 | TS 版 | Go 版での実現 |
|---|---|---|
| H, O | fillText、ランダム色、影(#888, offset 1,1, blur 1) | 白文字グリフを kind 別に 1 枚事前レンダリング → ColorScale でランダム色にティント。影はグレーの同型画像を (1,1) オフセットで先に描く(2 draw/粒子、バッチング有効) |
| H2 | 本体 24px + 下付き "2" 18px(x+12, y+3 オフセット) | 合成済みの 1 枚を事前レンダリング(本体 x−width/6、下付き x+12, y+3 を再現) |
| H2o | radial gradient(中心オフセット size×0.4、白 0.6 → #007fff)+青影 | 大きめの 1 枚を CPU で生成(`image.RGBA` にグラデーション計算)or Kage シェーダーで起動時に 1 回生成 → GeoM で拡縮 |
| 背景 | #fff | `screen.Fill(color.White)` |
| ランダム色 | `#` + random の 16 進下 6 桁(特殊な生成方式) | ランダム RGB 各 [0,256) に**意図的に置き換え**(rand/v2 採用に伴う簡素化。見た目は同等) |

フォント: TS 版は `sans-serif`。Go 版は `golang.org/x/image/font/gofont/goregular`(埋め込み可能、
ライセンス問題なし)を第一候補とする。文字幅が変わるため粒子半径も TS 版と厳密一致はしない(許容)。

### 5.3 反応ルール(個数収支を厳密維持)

| ルール | 反応 | consumed | produced | 収支 |
|---|---|---|---|---|
| HHFusion | H + H → H2 | 両方の H | H2(b の座標)+ H(ランダム座標に再生成) | H −1, H2 +1 |
| OxidationToWater | O + H2 → H2o | O と H2 | O(ランダム)+ H(ランダム)+ H2o(O の座標) | O ±0, H2 −1, H +1, H2o +1 |

- Registry は pair を両順で登録(順不同マッチ)
- `reactiveKinds()` に含まれない kind(H2o)は衝突判定から除外(性能上の要点)
- 同一フレーム多重反応防止: ペア処理時に `a.isDead() || b.isDead()` なら skip

### 5.4 リサイズ対応(TS 版からの追加改善)

- `ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)`
- `Layout(ow, oh)` で外形サイズをそのまま論理解像度として返す(可変)
- サイズ変化を検知したら: SpatialGrid を再構築(`Resize(w,h)`)、behavior が参照する画面境界を更新
- 画面境界は共有の `*Bounds` 構造体を DI し、behavior / factory / simulator が常に現在値を参照する設計にする
- スケール関数(粒子数・サイズ)は**生成時点**の幅で評価(既存粒子のサイズは変えない)

### 5.5 フレームパイプライン(Update / Draw 分離)

TS 版 `renderFrame()` は更新と描画が一体(rAF)。Ebitengine は TPS(既定 60)と FPS が分離しているため:

- `Update()`: ①全粒子 update → ②reactive kind のみ衝突検出 → ③反応適用(markDead / world.Add)→ ④sweep
- `Draw(screen)`: 背景 Fill → kind 別バケツ(初出順)で描画 → 計測モード時はオーバーレイ
- `Draw` 内で状態を変更しない(TS 版の設計原則③と同じ)

### 5.6 計測オーバーレイ(`-m` フラグ)

TS 版の表示項目を再現: FPS(60 フレーム移動平均)/ Frame 間隔 ms / Update 実行時間 ms /
kind 別粒子数 / Total。表示は左上、黒半透明背景 + aqua 文字、等幅フォント。
`ebiten.ActualFPS()` / `ActualTPS()` も併記する。

## 6. テスト戦略

- ロジック層(core/behavior/physics/reaction/particle/sim)は TS 版 155 テスト中の該当分をフル移植
  - 期待値は Go のシード付き乱数(`rand.NewPCG`)で再計算。**シードを固定すれば Go 内では完全決定的**
  - Grid vs BruteForce のプロパティテスト(複数シード × 500 粒子で完全一致)は必ず移植 — 衝突判定の正しさの要
  - RandomWalk のラップアラウンド境界値テスト(「ちょうど」vs「+1」)も移植
- 描画層: 事前レンダリング画像の生成はスモークテストに留める(Ebitengine はヘッドレス不可のため、
  CI で回すなら xvfb が必要 → まずはロジック層のみ CI 対象)
- 収支テスト(反応前後の kind 別個数)は simulator 統合テストとして移植

## 7. WASM / GitHub Pages

- `GOOS=js GOARCH=wasm go build -o web/mizu.wasm ./cmd/mizu`
- `wasm_exec.js` は `$(go env GOROOT)/lib/wasm/wasm_exec.js`(Go 1.24+)
- `web/index.html`: wasm_exec.js + instantiateStreaming の定型。URL クエリ(`?m=1&h=60&o=100`)は
  `syscall/js` で読んでフラグと同じ変数に流し込む
- GitHub Actions: push 時に wasm ビルド → Pages デプロイ(TS 版と同じ公開体験)

## 8. リスクと対処

| リスク | 対処 |
|---|---|
| vector パッケージは experimental(v2.9 で API 刷新歴あり) | 水滴は vector に依存せず CPU 生成 or Kage で事前画像化するため影響小 |
| 影(Canvas の shadow*)の直接対応なし | グレー/青のオフセット描画で近似。blur 1px は視認差が小さいため省略可(実装後に見た目比較) |
| フォント差で粒子サイズ・見た目が TS 版とずれる | text.Measure ベースの半径計算は同じロジックにする。見た目は demo.gif と目視比較 |
| 数千粒子時の性能 | 事前レンダリング + バッチングで TS 版(グラデーション毎フレーム描画)より有利な見込み。`-m` オーバーレイと ebitenginedebug タグで実測 |
| CI でのウィンドウ環境 | ロジック層は ebiten 非依存で通常 go test。描画層はローカル目視 + 必要なら xvfb |

## 9. 参考資料

- 調査済み: https://ebitengine.org/ja/ / https://ebitengine.org/ja/documents/ /
  https://ebitengine.org/ja/documents/shader.html / https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2
  (text/v2, vector, ebitenutil 含む) / performancetips / webassembly ドキュメント
- Mizu-ts 詳細分析: TS 版 `.claude/docs/architecture.md` と全ソースの分析結果は本プラン §5 に反映済み
