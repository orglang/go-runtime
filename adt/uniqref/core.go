package uniqref

import (
	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/revnum"
)

type ADT struct {
	ID identity.ADT
	RN revnum.ADT
}
