package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shimabox/Mizu-go/internal/bench"
)

// orchestrateOptions は orchestrate に渡すコマンドライン引数の集合。
type orchestrateOptions struct {
	scenariosCSV string
	framesFlag   int
	warmupMs     int
	out          string
	compareRef   string
}

// orchestrate はオーケストレータモードの本体。ebiten.RunGame は 1
// プロセスにつき 1 回しか呼べないため、シナリオごとに自分自身
// (os.Executable())を `-run-one` でサブプロセス起動して計測し、JSON
// 結果を集めて Markdown レポートを組み立てる。-compare が指定された
// 場合は、対象 git ref を一時 worktree に展開してビルドしたバイナリ
// でも同じシナリオを計測する。
func orchestrate(opts orchestrateOptions) error {
	names := splitCSV(opts.scenariosCSV)
	scenarios, err := bench.ResolveScenarios(names)
	if err != nil {
		return err
	}

	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving own executable path: %w", err)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	current, err := measureAll(selfPath, "", scenarios, opts.framesFlag, opts.warmupMs, "")
	if err != nil {
		return err
	}

	env := bench.Environment{
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoVersion:   runtime.Version(),
		GeneratedAt: time.Now(),
		WarmupMs:    opts.warmupMs,
	}
	env.CurrentCommit = gitDescribe(repoRoot)

	var compare []bench.ScenarioResult
	if opts.compareRef != "" {
		env.CompareRef = opts.compareRef

		fmt.Fprintf(os.Stderr, "[bench] 比較対象 %q 用の worktree を準備します...\n", opts.compareRef)
		compareBinary, compareWorktree, cleanup, err := prepareCompareBuild(opts.compareRef, repoRoot)
		defer cleanup()
		if err != nil {
			return err
		}
		env.CompareCommit = gitDescribe(compareWorktree)

		compare, err = measureAll(compareBinary, opts.compareRef, scenarios, opts.framesFlag, opts.warmupMs, compareWorktree)
		if err != nil {
			return err
		}
	}

	report := bench.BuildReport(env, current, compare)

	outPath := opts.out
	if outPath == "" {
		outPath = filepath.Join("bench-reports", fmt.Sprintf("report-%s.md", timestamp(time.Now())))
	}
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(repoRoot, outPath)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating report directory: %w", err)
	}
	if err := os.WriteFile(outPath, []byte(report.Markdown+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	fmt.Println(report.ConsoleSummary)
	fmt.Fprintf(os.Stderr, "\n[bench] レポートを書き出しました: %s\n", outPath)
	return nil
}

// measureAll は scenarios を binaryPath(-run-one 対応バイナリ)で順に
// 計測する。label はログ表示用の接尾辞(比較対象の計測には ref 名を
// 付ける。現在の計測では空文字列)。dir を指定するとサブプロセスの
// 作業ディレクトリをそこに固定する(compare 用 worktree 向け)。
func measureAll(binaryPath, label string, scenarios []bench.Scenario, framesFlag, warmupMs int, dir string) ([]bench.ScenarioResult, error) {
	results := make([]bench.ScenarioResult, 0, len(scenarios))
	for _, scenario := range scenarios {
		frames := framesFlag
		if frames <= 0 {
			frames = bench.DefaultFramesFor(scenario.Name)
		}

		tag := scenario.Name
		if label != "" {
			tag = fmt.Sprintf("%s (compare @ %s)", scenario.Name, label)
		}
		fmt.Fprintf(os.Stderr, "scenario %s: warmup...\n", tag)

		result, err := runSubprocess(binaryPath, scenario.Name, frames, warmupMs, dir)
		if err != nil {
			return nil, fmt.Errorf("scenario %s: %w", scenario.Name, err)
		}

		fmt.Fprintf(os.Stderr, "scenario %s: sampling %d frames... done (mean %.1fms)\n", tag, frames, bench.Mean(result.FrameMs))
		results = append(results, result)
	}
	return results, nil
}

// runSubprocess は binaryPath を `-run-one <scenario> -json <tmpfile>
// -frames <frames> -warmup <warmupMs>` で起動し、書き出された JSON を
// 読み込んで bench.ScenarioResult として返す。サブプロセスの標準エラー
// はそのまま親プロセスの標準エラーに透過する。
func runSubprocess(binaryPath, scenario string, frames, warmupMs int, dir string) (bench.ScenarioResult, error) {
	tmp, err := os.CreateTemp("", "mizu-go-bench-*.json")
	if err != nil {
		return bench.ScenarioResult{}, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command(binaryPath,
		"-run-one", scenario,
		"-json", tmpPath,
		"-frames", strconv.Itoa(frames),
		"-warmup", strconv.Itoa(warmupMs),
	)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr

	if err := cmd.Run(); err != nil {
		return bench.ScenarioResult{}, fmt.Errorf("running %s -run-one %s: %w", binaryPath, scenario, err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return bench.ScenarioResult{}, fmt.Errorf("reading result JSON: %w", err)
	}
	var result bench.ScenarioResult
	if err := json.Unmarshal(data, &result); err != nil {
		return bench.ScenarioResult{}, fmt.Errorf("parsing result JSON: %w", err)
	}
	return result, nil
}

// splitCSV はカンマ区切り文字列を、前後の空白を落とし空要素を除いた
// スライスに変換する。
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// timestamp は "report-<YYYYMMDD-HHmmss>.md" の既定出力パスに使う
// タイムスタンプ文字列を返す。
func timestamp(t time.Time) string {
	return t.Format("20060102-150405")
}

// findRepoRoot はリポジトリのルートディレクトリを返す
// (`git rev-parse --show-toplevel`)。失敗した場合はカレントディレクトリ
// にフォールバックする。
func findRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		if root := strings.TrimSpace(string(out)); root != "" {
			return root, nil
		}
	}
	return os.Getwd()
}

// gitDescribe は dir における簡潔なブランチ名@短縮コミットハッシュを
// 返す(レポートのヘッダに使う参考情報)。取得できない場合は空文字列を
// 返す(その情報がベストエフォートであることを示す)。
func gitDescribe(dir string) string {
	branch, err := runGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	commit, err := runGit(dir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s@%s", branch, commit)
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
