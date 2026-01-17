package uniqsym

import (
	"orglang/go-runtime/adt/symbol"
)

type ADT struct {
	sym symbol.ADT
	ns  *ADT
}

func New(name symbol.ADT) ADT {
	return ADT{name, nil}
}

func (space ADT) New(name symbol.ADT) ADT {
	return ADT{name, &space}
}

// symbol
func (adt ADT) Sym() symbol.ADT {
	return adt.sym
}

// namespace
func (adt ADT) NS() ADT {
	if adt.ns == nil {
		return empty
	}
	return *adt.ns
}

func (a ADT) Equal(b ADT) bool {
	if a.sym == b.sym && a.ns == b.ns {
		return true
	}
	if a.ns == nil || b.ns == nil {
		return false
	}
	return a.ns.Equal(*b.ns)
}

var (
	empty ADT
)
