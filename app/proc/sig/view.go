package sig

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/app/proc/bnd"
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
	SigID string `form:"sig_id" json:"sig_id" param:"id"`
	SigRN int64  `form:"sig_rn" json:"sig_rn"`
	Title string `form:"name" json:"title"`
}

func (dto RefView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.SigID, id.Required...),
		validation.Field(&dto.Title, sym.Required...),
	)
}

type RootView struct {
	X     bnd.SpecMsg   `json:"x"`
	SigID string        `json:"sig_id"`
	Ys    []bnd.SpecMsg `json:"ys"`
	Title string        `json:"title"`
	SigRN int64         `json:"sig_rn"`
}

type SnapView struct {
	SigID string `json:"sig_id"`
	SigRN int64  `json:"sig_rn"`
	Title string `json:"title"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Msg.*
// goverter:extend Msg.*
var (
	ViewFromRef  func(Ref) RefView
	ViewToRef    func(RefView) (Ref, error)
	ViewFromRefs func([]Ref) []RefView
	ViewToRefs   func([]RefView) ([]Ref, error)
	ViewFromSnap func(Snap) SnapView
	ViewFromRoot func(Impl) RootView
)
