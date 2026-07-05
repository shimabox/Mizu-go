package core

// Bounds は現在の画面サイズを共有するための参照。behavior・factory・
// simulator はいずれも同じ *Bounds を保持しており、ウィンドウリサイズ
// (Set 経由で反映される)が配線し直すことなくどこからでも観測できる。
type Bounds struct {
	w, h float64
}

// NewBounds は指定した幅・高さで初期化した Bounds を返す。
func NewBounds(w, h float64) *Bounds {
	return &Bounds{w: w, h: h}
}

// W は現在の幅を返す。
func (b *Bounds) W() float64 {
	return b.w
}

// H は現在の高さを返す。
func (b *Bounds) H() float64 {
	return b.h
}

// Set は幅と高さを更新する(ウィンドウリサイズ時などに使う)。
func (b *Bounds) Set(w, h float64) {
	b.w = w
	b.h = h
}
