package behavior

// MovementBehavior は粒子の現在位置から次の位置を計算する。実装は
// 呼び出しをまたいで内部状態(速度など)を保持してよい。
type MovementBehavior interface {
	// Next は現在位置を受け取り、次の (x, y) 位置を返す。
	Next(x, y float64) (nx, ny float64)
}
