package expctx

import (
	"orglang/orglang/adt/qualsym"
)

type ADT []BindClaim

type BindClaim struct {
	// binding placeholder (aka variable name)
	BindPH qualsym.ADT
	// type qualified name (aka variable type)
	TypeQN qualsym.ADT
}

type BindOffer struct {
}

func IndexBy[K comparable, V any](getKey func(V) K, vals []V) map[K]V {
	indexed := make(map[K]V)
	for _, val := range vals {
		indexed[getKey(val)] = val
	}
	return indexed
}
