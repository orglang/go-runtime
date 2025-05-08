package dec

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	def "smecalculus/rolevod/app/type/def"
)

type TypeSpecMsg struct {
	TypeQN string          `json:"qn"`
	TypeTS def.TermSpecMsg `json:"state"`
}

func (dto TypeSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeQN, sym.Required...),
		validation.Field(&dto.TypeTS, validation.Required),
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

type TypeRefMsg struct {
	TypeID string `json:"id" param:"id"`
	TypeRN int64  `json:"rev" query:"rev"`
	Title  string `json:"title"`
}

func (dto TypeRefMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeID, id.Required...),
		validation.Field(&dto.TypeRN, rn.Optional...),
	)
}

type TypeSnapMsg struct {
	TypeID string          `json:"id" param:"id"`
	TypeRN int64           `json:"rev" query:"rev"`
	Title  string          `json:"title"`
	TypeQN string          `json:"qn"`
	TypeTS def.TermSpecMsg `json:"state"`
}

func (dto TypeSnapMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.TypeID, id.Required...),
		validation.Field(&dto.TypeRN, rn.Optional...),
		validation.Field(&dto.TypeTS, validation.Required),
	)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/app/type/def:Msg.*
var (
	MsgFromTypeSpec  func(TypeSpec) TypeSpecMsg
	MsgToTypeSpec    func(TypeSpecMsg) (TypeSpec, error)
	MsgFromTypeRef   func(TypeRef) TypeRefMsg
	MsgToTypeRef     func(TypeRefMsg) (TypeRef, error)
	MsgFromTypeRefs  func([]TypeRef) []TypeRefMsg
	MsgToTypeRefs    func([]TypeRefMsg) ([]TypeRef, error)
	MsgFromTypeSnap  func(TypeSnap) TypeSnapMsg
	MsgToTypeSnap    func(TypeSnapMsg) (TypeSnap, error)
	MsgFromTypeSnaps func([]TypeSnap) []TypeSnapMsg
	MsgToTypeSnaps   func([]TypeSnapMsg) ([]TypeSnap, error)
)
