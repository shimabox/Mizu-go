// Command bench は Mizu-go のベンチマークツールであり、Mizu-ts の
// `npm run bench` に相当する。cmd/mizu と同じ DI で組んだ Simulator を
// ebiten.RunGame で走らせ、ウォームアップ後のフレーム間隔(壁時計)と
// Simulator.Update() の実行時間をシナリオごとに収集して Markdown
// レポートを書き出す。
//
// ebiten.RunGame は 1 プロセスにつき 1 回しか呼び出せないため、複数
// シナリオを 1 回の起動で計測することはできない。そのため、このコマンド
// には 2 つの動作モードがある:
//
//   - オーケストレータモード(引数なし、または -scenarios/-compare 等):
//     自分自身(os.Executable())をシナリオごとにサブプロセスとして
//     起動し、-run-one で 1 シナリオずつ計測させ、JSON 結果を集めて
//     レポートを組み立てる。
//   - -run-one <scenario> -json <path>(内部用): 単一シナリオを
//     ebiten.RunGame で計測し、結果を JSON として path に書き出して
//     終了する。
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// viewportWidth/viewportHeight は、Mizu-ts のベンチツールが使う
// 1280x800 のビューポートに合わせた、ベンチ計測時に固定するウィンドウ
// サイズ(porting-plan の Go 版設計)。Simulator.CountScale() は幅のみを
// 見るため(768 未満で 1.0、1280 未満で 1.2、それ以外で 1.5)、この値は
// cmd/mizu の既定ウィンドウ幅(1280)と同じスケール(1.5)になる。
const (
	viewportWidth  = 1280
	viewportHeight = 800
)

func main() {
	scenariosFlag := flag.String("scenarios", "default,500,1000", "計測するシナリオ(カンマ区切り。選択肢: default,500,1000,3000)")
	framesFlag := flag.Int("frames", 0, "シナリオごとに収集するフレーム数(既定: 300。3000 シナリオのみ既定 60。0 は「既定値を使う」の意味)")
	warmupFlag := flag.Int("warmup", 3000, "計測前のウォームアップ時間(ms)")
	outFlag := flag.String("out", "", "レポート出力先(既定: bench-reports/report-<YYYYMMDD-HHmmss>.md)")
	compareFlag := flag.String("compare", "", "指定した git ref を worktree に展開し、現在の作業ツリーと A/B 比較する")

	runOneFlag := flag.String("run-one", "", "内部用: 単一シナリオを計測して -json に結果を書き出す")
	jsonFlag := flag.String("json", "", "内部用: -run-one の結果を書き出す JSON パス")

	flag.Parse()

	if *runOneFlag != "" {
		if *jsonFlag == "" {
			log.Fatal("bench: -run-one には -json <path> が必須です")
		}
		if err := runOne(*runOneFlag, *framesFlag, *warmupFlag, *jsonFlag); err != nil {
			log.Fatalf("bench: シナリオ %q の計測に失敗しました: %v", *runOneFlag, err)
		}
		return
	}

	if err := orchestrate(orchestrateOptions{
		scenariosCSV: *scenariosFlag,
		framesFlag:   *framesFlag,
		warmupMs:     *warmupFlag,
		out:          *outFlag,
		compareRef:   *compareFlag,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "bench: %v\n", err)
		os.Exit(1)
	}
}
