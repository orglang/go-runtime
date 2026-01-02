package expalias

import (
	"orglang/orglang/lib/sd"
)

type Repo interface {
	Insert(sd.Source, Root) error
}

type rootDS struct {
	ID  string
	RN  int64
	Sym string
}
