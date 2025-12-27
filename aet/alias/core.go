package alias

import (
	"orglang/orglang/avt/data"
	"orglang/orglang/avt/id"
	"orglang/orglang/avt/rn"
	"orglang/orglang/avt/sym"
)

type Root struct {
	ID id.ADT
	RN rn.ADT
	QN sym.ADT
}

type Repo interface {
	Insert(data.Source, Root) error
}
