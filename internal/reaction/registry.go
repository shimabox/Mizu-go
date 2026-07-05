package reaction

import "github.com/shimabox/Mizu-go/internal/core"

// pairKey は順序付きの (kindA, kindB) の組を検索するためのキー。
type pairKey struct {
	a, b core.Kind
}

// Registry は粒子種別のペアを、それらを反応させる Rule に対応付ける。
// ペアは両方の順序で登録されるため、検索は順序に依存しない。これは TS 版
// の ReactionRegistry に対応する。
type Registry struct {
	rules map[pairKey]Rule
	kinds map[core.Kind]struct{}
}

// NewRegistry は空の Registry を返す。
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[pairKey]Rule),
		kinds: make(map[core.Kind]struct{}),
	}
}

// Register は rule をその Pair() の kind の両方の順序で追加し、両方の
// kind を反応可能な kind の集合にも追加する。
func (r *Registry) Register(rule Rule) {
	a, b := rule.Pair()
	r.rules[pairKey{a, b}] = rule
	r.rules[pairKey{b, a}] = rule
	r.kinds[a] = struct{}{}
	r.kinds[b] = struct{}{}
}

// Find は (a, b) の組について登録された rule を検索する。引数の順序には
// 関係なくマッチする。
func (r *Registry) Find(a, b core.Kind) (Rule, bool) {
	rule, ok := r.rules[pairKey{a, b}]
	return rule, ok
}

// ReactiveKinds は、登録済みのいずれかの rule のペアに登場する kind の
// 集合、つまり反応に関与しうる kind の集合を返す。どの rule のペアにも
// 登場しない kind (例えば H2o) は含まれず、これにより呼び出し側は衝突判定
// の際にそれらをスキップできる。返される集合は呼び出し時点での Registry
// の状態を反映するため、さらに Register を呼び出した後に呼び出せば、
// 新たに追加された kind も含まれる。
func (r *Registry) ReactiveKinds() map[core.Kind]struct{} {
	out := make(map[core.Kind]struct{}, len(r.kinds))
	for k := range r.kinds {
		out[k] = struct{}{}
	}
	return out
}
