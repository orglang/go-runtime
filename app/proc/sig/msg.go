package sig

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/app/proc/bnd"
)

type SpecMsg struct {
	X     bnd.SpecMsg   `json:"x"`
	SigQN string        `json:"sig_qn"`
	Ys    []bnd.SpecMsg `json:"ys"`
}

func (dto SpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigQN, sym.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
	)
}

type IdentMsg struct {
	SigID string `json:"id" param:"id"`
}

func (dto IdentMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
	)
}

type RefMsg struct {
	SigID string `json:"id" param:"id"`
	Title string `json:"title"`
	SigRN int64  `json:"rev"`
}

func (dto RefMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
	)
}

type ImplMsg struct {
	X     bnd.SpecMsg   `json:"x"`
	SigID string        `json:"sig_id"`
	Ys    []bnd.SpecMsg `json:"ys"`
	Title string        `json:"title"`
	SigRN int64         `json:"rev"`
}

type SnapMsg struct {
	X     bnd.SpecMsg   `json:"x"`
	SigID string        `json:"sig_id"`
	Ys    []bnd.SpecMsg `json:"ys"`
	Title string        `json:"title"`
	SigRN int64         `json:"rev"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/app/role/root:Msg.*
// goverter:extend smecalculus/rolevod/internal/state:Msg.*
// goverter:extend smecalculus/rolevod/app/proc/bnd:Msg.*
var (
	MsgToID      func(string) (id.ADT, error)
	MsgFromID    func(id.ADT) string
	MsgToSpec    func(SpecMsg) (Spec, error)
	MsgFromSpec  func(Spec) SpecMsg
	MsgToRef     func(RefMsg) (Ref, error)
	MsgFromRef   func(Ref) RefMsg
	MsgToRoot    func(ImplMsg) (Impl, error)
	MsgFromRoot  func(Impl) ImplMsg
	MsgFromRoots func([]Impl) []ImplMsg
	MsgToSnap    func(SnapMsg) (Snap, error)
	MsgFromSnap  func(Snap) SnapMsg
)
