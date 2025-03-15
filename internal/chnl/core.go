package chnl

import (
	"fmt"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/rev"
	"smecalculus/rolevod/lib/sym"
)

// for external readability
type ID = id.ADT
type Key = string

// aka ChanTp
type Spec struct {
	ChnlPH ph.ADT
	RoleQN sym.ADT
}

// aka Z
type Ref struct {
	ID    id.ADT
	Title string
}

type Root struct {
	ID      id.ADT
	Title   string
	StateID *id.ADT
	PoolID  *id.ADT
	Revs    []rev.ADT
}

type Repo interface {
	Insert(data.Source, Root) error
	SelectRefs(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Root, error)
	SelectByIDs(data.Source, []id.ADT) ([]Root, error)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/ak:Convert.*
var (
	ConvertRootToRef func(Root) Ref
)

const (
	stateRev = 0
	poolRev  = 1
)

func ErrDoesNotExist(want ID) error {
	return fmt.Errorf("channel doesn't exist: %v", want)
}

func ErrMissingInCfg(want ph.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCtx(want ph.ADT) error {
	return fmt.Errorf("channel missing in ctx: %v", want)
}

func ErrAlreadyClosed(got ID) error {
	return fmt.Errorf("channel already closed: %v", got)
}

func ErrNotAnID(got ph.ADT) error {
	return fmt.Errorf("not a channel id: %v", got)
}

func errOptimisticUpdate(got rev.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}
