package syndec

import (
	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/qualsym"
	"orglang/go-runtime/adt/revnum"
)

type DecRec struct {
	DecID identity.ADT
	DecRN revnum.ADT
	DecQN qualsym.ADT
}
