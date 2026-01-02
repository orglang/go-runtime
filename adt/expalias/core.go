package expalias

import (
	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/revnum"
)

type Root struct {
	ID identity.ADT
	RN revnum.ADT
	QN qualsym.ADT
}
