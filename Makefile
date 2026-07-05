.PHONY: wasm run test serve bench

# wasm は cmd/mizu を GOOS=js/GOARCH=wasm 向けにビルドし、対応する
# wasm_exec.js グルースクリプトを web/ にコピーする(porting-plan §7)。
wasm:
	GOOS=js GOARCH=wasm go build -o web/mizu.wasm ./cmd/mizu
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/

# run はデスクトップ向け Ebitengine ビルドを起動する。
run:
	go run ./cmd/mizu

# test はテストスイート全体を実行する。
test:
	go test ./...

# serve はブラウザで wasm ビルドを手動確認できるよう、web/ を
# :8080 でローカルにホストする(事前に `make wasm` を実行しておくこと)。
serve:
	cd web && python3 -m http.server 8080

# bench はベンチマークツール(cmd/bench)を既定オプションで実行する。
# 実ウィンドウが必要なため、SSH 先やディスプレイのない CI では動かない
# (README.md の「Benchmark Tool」節を参照)。追加オプションを渡す場合は
# `go run ./cmd/bench -scenarios default,500 -frames 60` のように直接
# 呼び出すこと。
bench:
	go run ./cmd/bench
