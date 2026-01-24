package procbind

import (
	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

type BindSpec struct {
	// channel placeholder (aka variable name)
	ChnlPH symbol.ADT
	// type qualified name (aka variable type)
	TypeQN uniqsym.ADT
}

type BindRec struct {
	// процес, в рамках которого связка
	ExecRef uniqref.ADT
	Kind    BindKind
	ChnlPH  symbol.ADT
	ChnlID  identity.ADT
	ExpID   identity.ADT
	PoolRN  revnum.ADT
}

type BindKind int

const (
	nonBind = BindKind(iota)
	ProviderBind
	ClientBind
)

func IndexBy[K comparable, V any](getKey func(V) K, vals []V) map[K]V {
	indexed := make(map[K]V)
	for _, val := range vals {
		indexed[getKey(val)] = val
	}
	return indexed
}
