package def

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"
)

type TypeSpecView struct {
	NS   string `form:"ns" json:"ns"`
	Name string `form:"name" json:"name"`
}

func (dto TypeSpecView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.NS, sym.Required...),
		validation.Field(&dto.Name, sym.Required...),
	)
}

type TypeRefView struct {
	RoleID string `form:"id" json:"id" param:"id"`
	RoleRN int64  `form:"rev" json:"rev"`
	Title  string `form:"name" json:"title"`
}

func (dto TypeRefView) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.RoleID, id.Required...),
		validation.Field(&dto.RoleRN, rn.Optional...),
		validation.Field(&dto.Title, sym.Required...),
	)
}

type TypeSnapView struct {
	RoleID string      `json:"id"`
	RoleRN int64       `json:"rev"`
	Title  string      `json:"title"`
	State  TermSpecMsg `json:"state"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/app/type/def:Msg.*
var (
	ViewFromTypeRef  func(TypeRef) TypeRefView
	ViewToTypeRef    func(TypeRefView) (TypeRef, error)
	ViewFromTypeRefs func([]TypeRef) []TypeRefView
	ViewToTypeRefs   func([]TypeRefView) ([]TypeRef, error)
	ViewFromTypeSnap func(TypeSnap) TypeSnapView
)
