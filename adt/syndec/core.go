package syndec

import (
	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/uniqsym"
)

type DecRec struct {
	DecID identity.ADT
	DecRN revnum.ADT
	DecQN uniqsym.ADT
}
