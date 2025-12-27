package exec

import (
	"database/sql"
	"fmt"

	procdef "orglang/orglang/aat/proc/def"
)

type modData struct {
	Locks []lockData
	Bnds  []bndData
	Steps []SemRecData
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
// goverter:extend orglang/orglang/avt/id:Convert.*
// goverter:extend orglang/orglang/avt/rn:Convert.*
// goverter:extend orglang/orglang/aat/proc/def:Data.*
var (
	DataFromMod func(Mod) (modData, error)
	DataFromBnd func(Bnd) bndData
)

type SemRecData struct {
	ID  string              `db:"id"`
	K   semKind             `db:"kind"`
	PID sql.NullString      `db:"pid"`
	VID sql.NullString      `db:"vid"`
	TR  procdef.TermRecData `db:"spec"`
}

type semKind int

const (
	nonsem = semKind(iota)
	msgKind
	svcKind
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend data.*
var (
	DataToSemRecs   func([]SemRecData) ([]SemRec, error)
	DataFromSemRecs func([]SemRec) ([]SemRecData, error)
)

func dataFromSemRec(r SemRec) (SemRecData, error) {
	if r == nil {
		return SemRecData{}, nil
	}
	switch rec := r.(type) {
	case MsgRec:
		msgVal, err := procdef.DataFromTermRec(rec.Val)
		if err != nil {
			return SemRecData{}, err
		}
		return SemRecData{
			K:  msgKind,
			TR: msgVal,
		}, nil
	case SvcRec:
		svcCont, err := procdef.DataFromTermRec(rec.Cont)
		if err != nil {
			return SemRecData{}, err
		}
		return SemRecData{
			K:  svcKind,
			TR: svcCont,
		}, nil
	default:
		panic(ErrRootTypeUnexpected(rec))
	}
}

func dataToSemRec(dto SemRecData) (SemRec, error) {
	var nilData SemRecData
	if dto == nilData {
		return nil, nil
	}
	switch dto.K {
	case msgKind:
		val, err := procdef.DataToTermRec(dto.TR)
		if err != nil {
			return nil, err
		}
		return MsgRec{Val: val}, nil
	case svcKind:
		cont, err := procdef.DataToTermRec(dto.TR)
		if err != nil {
			return nil, err
		}
		return SvcRec{Cont: cont}, nil
	default:
		panic(errUnexpectedStepKind(dto.K))
	}
}

func errUnexpectedStepKind(k semKind) error {
	return fmt.Errorf("unexpected step kind: %v", k)
}
