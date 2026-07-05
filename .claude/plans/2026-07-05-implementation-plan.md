# Mizu-go 実装プラン(オーケストレーション)

作成日: 2026-07-05
前提: `2026-07-05-porting-plan.md`(移植プラン)承認済みであること

## 進め方

- オーケストレーター(Fable)は**実装しない**。各フェーズをサブエージェントに委譲し、
  受け入れ条件の検証・レビュー・次フェーズへの指示のみ行う
- 各タスクにはサブエージェントへ渡す仕様(移植プラン該当節 + TS 版参照ファイル)を明記する
- モデル選定基準:
  - **Haiku**: 定型・機械的(scaffolding、CI 定義、README、定数移植)
  - **Sonnet**: 通常の実装(ロジック移植 + テスト、WASM シェル)
  - **Opus/Fable**: 判断が絡む難所(Ebitengine 描画層の設計実装、性能チューニング)— 原則
    Sonnet で試み、レビューで問題があれば引き上げる
- 各フェーズ完了時にオーケストレーターが `go build ./... && go vet ./... && go test ./...` と
  受け入れ条件を確認してから次へ進む
- コミットは**意味のある塊**で行う(TDD サイクルごとの細切れは不要。フェーズ内でも論理的な
  まとまり — 例: 「SpatialGrid + テスト」「衝突検出 2 実装 + プロパティテスト」— で分ける)。
  メッセージは**日本語**でわかりやすく: 何を・なぜが 1 行目で伝わる形。git init は Phase 0

## 開発スタイル: テストファースト(TDD)— 必須

全実装フェーズは **red → green のリズム**で進める。実装エージェントへの指示に必ず含めること:

1. **red**: まずテストを書き、`go test` で**失敗することを確認**する
   (コンパイルエラーも red のうち。最小のスタブで「意図した理由で落ちる」状態にする)
2. **green**: テストを通す最小の実装を書き、`go test` 通過を確認する
3. 振る舞い 1 つごとにこのサイクルを回す(テストを全部書いてから実装をまとめて書く、はしない)
4. **迷ったらテストを書きながら検討する** — 設計判断に迷いが生じたら、先に「期待する振る舞い」を
   テストとして表現してみて、書きやすい方の設計を採る
5. リファクタはテスト green を保ったまま行う

TS 版のテスト(155 件)が仕様の一次情報なので、TDD と相性がよい:
**「TS 版テストの移植」がそのまま red の起点になる**。TS 版テスト → Go テストに書き換え →
失敗確認 → 実装、の順で進めること。

例外は描画層の一部(Phase 5 のウィンドウ表示・見た目)のみ。ここも純粋ロジック
(スプライト生成のピクセル値、GeoM 計算、バケツ順序など)はテストファーストで書き、
テスト不能な部分(実ウィンドウの目視)だけ受け入れ確認に委ねる。

## 依存関係と並列度

```
Phase 0 (scaffolding)
  └→ Phase 1 (core: Particle/Random/Bounds)
       ├→ Phase 2a (behavior)   ─┐ 並列実行可
       ├→ Phase 2b (physics)    ─┤
       └→ Phase 2c (reaction)   ─┘
            └→ Phase 3 (particle + factory)
                 └→ Phase 4 (sim: World/Simulator + 統合テスト)
                      └→ Phase 5 (render: Ebitengine 描画層 + main)
                           ├→ Phase 6a (overlay + フラグ)  ─┐ 並列実行可
                           └→ Phase 6b (WASM + Pages)      ─┘
                                └→ Phase 7 (検証・ベンチ・README)
```

## フェーズ定義

### Phase 0: Scaffolding — **Haiku**

- git init、`go mod init github.com/shimabox/Mizu-go`(go 1.24)
- ディレクトリ骨格(移植プラン §4.1)、.gitignore、ebiten v2.9.x を go get
- GitHub Actions: `go build/vet/test`(ロジック層のみ、ubuntu-latest。ebiten のビルドに必要な
  apt パッケージは公式 install ドキュメント記載のものを導入)
- 受け入れ: `go build ./...` が通る(空パッケージで可)

### Phase 1: core パッケージ — **Sonnet**

- `internal/core/particle.go`: Particle インターフェース(Kind/X/Y/Radius/Update/IsDead/MarkDead。
  描画は render 層に分離するため TS 版の render(ctx) は持たせない — 描画層が kind で引く)
- `internal/core/random.go`: `Random` interface(`Next() float64`)+ 実行時実装(rand/v2)+
  テスト用シード付き実装(`rand.New(rand.NewPCG(seed, seed))` ラッパー)
- `internal/core/bounds.go`: 共有画面境界(リサイズ追従用、移植プラン §5.4)
- テスト: Random の範囲・決定性
- 受け入れ: go test 通過

### Phase 2a: behavior — **Sonnet**(2b, 2c と並列)

- RandomWalk / FallAndSway を移植プラン §5.1 の数式・定数どおりに実装
- TS 版テスト(RandomWalk 11 件 + FallAndSway 3 件相当)を移植。期待値はシード付き乱数で再計算。
  **ラップアラウンド境界値(「ちょうど」は折り返さず「+1」で折り返す)は乱数非依存なので TS 版と同値で検証**
- 参照: `Mizu-ts/src/core/behaviors/`, `Mizu-ts/tests/core/behaviors/`

### Phase 2b: physics — **Sonnet**(並列)

- SpatialGrid(セルサイズ 48、クランプ、バケツ使い回し、`Resize(w,h)` 追加)、
  GridCollisionDetector(正準順序化・重複除外)、BruteForceCollisionDetector(参照実装)
- テスト移植: SpatialGrid 10 件 + Grid 7 件 + BruteForce 5 件 + **プロパティテスト(複数シード ×
  500 粒子で Grid と BruteForce の結果完全一致)**
- 参照: `Mizu-ts/src/physics/`, `Mizu-ts/tests/physics/`

### Phase 2c: reaction — **Sonnet**(並列)

- Rule インターフェース、Registry(両順登録・reactiveKinds)、HHFusion、OxidationToWater
- Factory はまだ無いので、粒子生成は関数注入(TS 版と同じく Factory のメソッドを受け取る形)
- テスト移植: Registry 8 件 + HHFusion 4 件 + OxidationToWater 5 件(個数収支の厳密検証)
- 参照: `Mizu-ts/src/reactions/`, `Mizu-ts/tests/reactions/`

### Phase 3: particle + factory — **Sonnet**

- H/H2/O/H2o の薄い struct(H2o のみ着地で MarkDead)、ParticleFactory
- サイズ計算: テキスト幅は render 層完成前のため「幅計測関数」を注入する設計にし、
  テストではスタブ、実行時は text.Measure を渡す
- 2 種類のスケール関数の非対称(粒子数 1.0/1.2/1.5、サイズ 1.0/1.2)を忠実に実装
- テスト移植: particles 9 件 + ParticleFactory 6 件相当
- 参照: `Mizu-ts/src/particles/`, `Mizu-ts/tests/particles/`

### Phase 4: sim(World / Simulator)— **Sonnet**

- World(Add/Count/All/Sweep)、Simulator の Update パイプライン
  (update → reactive 抽出 → 衝突検出 → 反応適用(dead skip)→ sweep)
- kind 別バケツ(初出順)は描画順情報として Simulator が提供(`Buckets()` 等)
- テスト移植: World 4 件 + MizuSimulator 13 件相当(収支・パイプライン順序・スケール)
- 受け入れ: **ここまでで ebiten 非依存のロジック層が完全動作**(CI で全テスト通過)
- 参照: `Mizu-ts/src/simulator/`, `Mizu-ts/tests/simulator/`

### Phase 5: render + main — **Sonnet**(要注意フェーズ。レビューは Fable が厳密に行い、
設計判断で詰まったら Opus に引き上げ)

- スプライト事前生成(移植プラン §5.2):
  - H/O/H2: 白文字グリフ + グレー影画像(text/v2、goregular、H2 は下付き合成)
  - H2o: 放射状グラデーション画像(CPU 生成で可。中心オフセット size×0.4、白 0.6 → #007fff)
- Game 実装: Update()/Draw()/Layout()、リサイズ追従(§5.4)、ColorScale ティント、
  GeoM 拡縮、kind 別バケツ順描画、背景白
- `cmd/mizu/main.go`: CLI フラグ(-h, -o, -m)、DI 組み立て、RunGame
- 受け入れ: `go run ./cmd/mizu` でウィンドウが開き、H/O が漂い、H2 → H2o 生成 → 落下消滅が
  目視確認できる(オーケストレーターがローカルで起動確認)
- 参照: 移植プラン §5.2/§5.4/§5.5、`Mizu-ts/src/core/renderers/`

### Phase 6a: 計測オーバーレイ — **Sonnet**(6b と並列)

- FPS 60 フレーム移動平均 / Update 実行時間 / kind 別カウント / Total、ActualFPS/TPS 併記
- 左上、黒半透明背景 + aqua 文字(移植プラン §5.6)
- 参照: `Mizu-ts/src/debug/StatsOverlay.ts`

### Phase 6b: WASM + GitHub Pages — **Sonnet**(並列)

- `web/index.html`(wasm_exec.js + instantiateStreaming)、URL クエリ読み(syscall/js)、
  Makefile or スクリプトで wasm ビルド、GitHub Actions で Pages デプロイ
- 受け入れ: ローカルで `python3 -m http.server` 等により wasm 版が動作
- 参照: 移植プラン §7

### Phase 7: 検証・仕上げ — **Haiku**(README)+ **Fable**(検証)

- オーケストレーターが最終検証: デスクトップ起動、`-h 1000 -o 1000` 負荷時の FPS 確認、
  TS 版 demo.gif との見た目比較、/code-review
- README.md(TS 版準拠: 概要、デモ URL、起動方法、フラグ表、開発コマンド)
- 最終コミット

## サブエージェントへの共通指示テンプレート

各実装エージェントには必ず以下を渡す:

1. 担当フェーズの本プラン該当節 + 移植プラン該当節(仕様の一次情報)
2. TS 版参照ファイルのパス(`~/shimabox/github/Mizu-ts/src/...` と `tests/...`)— 数式・定数・
   期待挙動は TS 版が正
3. **TDD 必須**: red → green のリズム(上記「開発スタイル」節を全文引用して渡す)。
   TS 版テストの移植を red の起点にすること。報告時に「どのテストを先に書き、失敗を確認したか」を
   含めさせる
4. 制約: 依存方向は core 向き一方向 / ebiten 依存は render・cmd のみ / 描画メソッドで状態変更禁止
5. 完了条件: `go build ./... && go vet ./... && go test ./...` 通過 + フェーズ受け入れ条件
6. gofmt 済みであること。過剰な抽象化をしない(TS 版の薄い合成スタイルを踏襲)

## 見積もり

- 直列クリティカルパス: Phase 0 → 1 → (2 並列) → 3 → 4 → 5 → (6 並列) → 7
- リスク集中点は Phase 5(描画層)。ここだけは受け入れ確認を目視で必ず行う
