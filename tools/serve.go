//go:build ignore

// serve.go は `make serve` から `go run tools/serve.go` として実行される
// 開発用の静的ファイルサーバーで、web/(wasm ビルドの出力先)を :8080
// でホストする。以前は python3 の http.server を使っていたが、python3 が
// 無い環境でも「Go が入っていれば動く」ように Go 標準ライブラリのみで
// 実装し直した。net/http は .wasm を正しい Content-Type
// (application/wasm)で配信するため、WebAssembly.instantiateStreaming
// の MIME チェックもそのまま通る。
//
// リポジトリのルートから実行すること(web/ を相対パスで参照する)。
package main

import (
	"log"
	"net/http"
)

func main() {
	const addr = ":8080"
	log.Printf("serving ./web on http://localhost%s (Ctrl+C で停止)", addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir("web"))))
}
