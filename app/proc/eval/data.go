package eval

import (
	"smecalculus/rolevod/app/proc/def"
)

type modData struct {
	Locks []lockData
	Bnds  []bndData
	Steps []def.SemRecData
}

type lockData struct {
	PoolID string
	PoolRN int64
}

type bndData struct {
	ProcID  string
	ChnlPH  string
	ChnlID  string
	StateID string
	PoolRN  int64
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/app/proc/def:Data.*
var (
	DataFromMod func(Mod) (modData, error)
	DataFromBnd func(Bnd) bndData
)
