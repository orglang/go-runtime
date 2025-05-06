package role

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/state"
)

type SpecView struct {
	NS   string `form:"ns" json:"ns"`
	Name string `form:"name" json:"name"`
}

func (dto SpecView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.NS, sym.Required...),
		validation.Field(&dto.Name, sym.Required...),
	)
}

type RefView struct {
	RoleID string `form:"id" json:"id" param:"id"`
	RoleRN int64  `form:"rev" json:"rev"`
	Title  string `form:"name" json:"title"`
}

func (dto RefView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.RoleID, id.Required...),
		validation.Field(&dto.RoleRN, rn.Optional...),
		validation.Field(&dto.Title, sym.Required...),
	)
}

type RootView struct {
	RoleID string        `json:"id"`
	Title  string        `json:"title"`
	State  state.SpecMsg `json:"state"`
}

type SnapView struct {
	RoleID string        `json:"id"`
	RoleRN int64         `json:"rev"`
	Title  string        `json:"title"`
	State  state.SpecMsg `json:"state"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Msg.*
var (
	ViewFromRef  func(Ref) RefView
	ViewToRef    func(RefView) (Ref, error)
	ViewFromRefs func([]Ref) []RefView
	ViewToRefs   func([]RefView) ([]Ref, error)
	ViewFromSnap func(Snap) SnapView
)
