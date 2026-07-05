package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// prepareCompareBuild は ref を一時的な git worktree にチェックアウト
// し、そこに(必要であれば現在の作業ツリーの)cmd/bench・internal/bench
// をコピーしたうえで `go build` し、計測用バイナリのパスと worktree の
// パス(env の CompareCommit 表示用)、そして必ず呼び出すべき cleanup
// 関数を返す。
//
// cmd/bench・internal/bench を「無ければコピー」するのは、ref がこの
// ベンチツール自体を含まない古いコミットであっても、現在のベンチ
// ハーネスで計測できるようにするためである(Mizu-ts の worktree.mjs の
// コメントにある「現在のベンチスクリプトで旧コードを測る」という考え方
// と同じ)。ref が既に自身のベンチツールを含む場合はそれをそのまま使う。
//
// エラーが発生しても cleanup は必ず有効な関数を返す(呼び出し側は
// defer cleanup() でよい)。
func prepareCompareBuild(ref, repoRoot string) (binaryPath, worktreePath string, cleanup func(), err error) {
	suffix, err := randomHex(4)
	if err != nil {
		return "", "", func() {}, fmt.Errorf("generating temp suffix: %w", err)
	}
	worktreePath = filepath.Join(os.TempDir(), fmt.Sprintf("mizu-go-bench-%d-%s", time.Now().UnixNano(), suffix))

	cleanup = func() {
		removeCompareWorktree(repoRoot, worktreePath)
	}

	if out, err := exec.Command("git", "-C", repoRoot, "worktree", "add", "--detach", worktreePath, ref).CombinedOutput(); err != nil {
		return "", worktreePath, cleanup, fmt.Errorf("git worktree add %s %s failed: %w\n%s", worktreePath, ref, err, out)
	}

	worktreeCmdBench := filepath.Join(worktreePath, "cmd", "bench")
	worktreeInternalBench := filepath.Join(worktreePath, "internal", "bench")
	if !dirExists(worktreeCmdBench) || !dirExists(worktreeInternalBench) {
		if err := copyDir(filepath.Join(repoRoot, "cmd", "bench"), worktreeCmdBench); err != nil {
			return "", worktreePath, cleanup, fmt.Errorf("copying cmd/bench into worktree: %w", err)
		}
		if err := copyDir(filepath.Join(repoRoot, "internal", "bench"), worktreeInternalBench); err != nil {
			return "", worktreePath, cleanup, fmt.Errorf("copying internal/bench into worktree: %w", err)
		}
	}

	binaryPath = filepath.Join(worktreePath, "bench-compare")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/bench")
	buildCmd.Dir = worktreePath
	if out, err := buildCmd.CombinedOutput(); err != nil {
		return "", worktreePath, cleanup, fmt.Errorf(
			"ref %q のビルドに失敗しました(cmd/bench が要求する API が %[1]q 時点に存在しない可能性があります):\n%s\n%w",
			ref, out, err,
		)
	}

	return binaryPath, worktreePath, cleanup, nil
}

// removeCompareWorktree は prepareCompareBuild が作成した worktree を
// 削除する。`git worktree remove` に失敗した場合は、ディレクトリの直接
// 削除 + `git worktree prune` にフォールバックする(いずれもベスト
// エフォート)。
func removeCompareWorktree(repoRoot, worktreePath string) {
	if worktreePath == "" {
		return
	}
	if out, err := exec.Command("git", "-C", repoRoot, "worktree", "remove", "--force", worktreePath).CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "[bench] git worktree remove に失敗したため手動クリーンアップにフォールバックします: %v\n%s\n", err, out)
		_ = os.RemoveAll(worktreePath)
		_, _ = exec.Command("git", "-C", repoRoot, "worktree", "prune").CombinedOutput()
	}
}

// dirExists は path がディレクトリとして存在するかどうかを返す。
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// copyDir は src 配下のファイル・ディレクトリ構造をそのまま dst に
// コピーする(シンボリックリンクは対象外。cmd/bench・internal/bench は
// 通常のソースファイルのみのため十分)。
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
