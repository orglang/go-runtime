package role

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/state"
)

type SpecMsg struct {
	RoleQN string        `json:"qn"`
	State  state.SpecMsg `json:"state"`
}

func (dto SpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.RoleQN, sym.Required...),
		validation.Field(&dto.State, validation.Required),
	)
}

type IdentMsg struct {
	ID string `json:"id" param:"id"`
}

func (dto IdentMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ID, id.Required...),
	)
}

type RefMsg struct {
	RoleID string `json:"id" param:"id"`
	RoleRN int64  `json:"rev" query:"rev"`
	Title  string `json:"title"`
}

func (dto RefMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.RoleID, id.Required...),
		validation.Field(&dto.RoleRN, rn.Optional...),
	)
}

type SnapMsg struct {
	RoleID string        `json:"id" param:"id"`
	RoleRN int64         `json:"rev" query:"rev"`
	Title  string        `json:"title"`
	RoleQN string        `json:"qn"`
	State  state.SpecMsg `json:"state"`
}

func (dto SnapMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.RoleID, id.Required...),
		validation.Field(&dto.RoleRN, rn.Optional...),
		validation.Field(&dto.State, validation.Required),
	)
}

type RootMsg struct {
	RoleID  string        `json:"id" param:"id"`
	RoleRN  int64         `json:"rev"`
	Title   string        `json:"title"`
	StateID string        `json:"state_id"`
	State   state.SpecMsg `json:"state"`
	Parts   []RefMsg      `json:"parts"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Msg.*
var (
	MsgFromSpec func(Spec) SpecMsg
	MsgToSpec   func(SpecMsg) (Spec, error)
	MsgFromRef  func(Ref) RefMsg
	MsgToRef    func(RefMsg) (Ref, error)
	MsgFromRefs func([]Ref) []RefMsg
	MsgToRefs   func([]RefMsg) ([]Ref, error)
	// MsgFromRoot func(Root) RootMsg
	MsgToRoot func(RootMsg) (Impl, error)
	// MsgFromRoots func([]Root) []RootMsg
	MsgToRoots   func([]RootMsg) ([]Impl, error)
	MsgFromSnap  func(Snap) SnapMsg
	MsgToSnap    func(SnapMsg) (Snap, error)
	MsgFromSnaps func([]Snap) []SnapMsg
	MsgToSnaps   func([]SnapMsg) ([]Snap, error)
)
