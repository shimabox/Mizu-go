.PHONY: wasm run test serve

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
