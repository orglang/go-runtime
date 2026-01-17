package termctx

import (
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/uniqsym"
)

type BindClaim struct {
	// binding placeholder (aka variable name)
	BindPH symbol.ADT
	// type qualified name (aka variable type)
	TypeQN uniqsym.ADT
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
