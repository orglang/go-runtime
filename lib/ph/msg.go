package ph

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type Kind string

const (
	ID  = Kind("id")
	Sym = Kind("sym")
)

var kindRequired = []validation.Rule{
	validation.Required,
	validation.In(ID, Sym),
}

type Msg struct {
	K   Kind   `json:"kind"`
	ID  string `json:"id,omitempty"`
	Sym string `json:"sym,omitempty"`
}

func (dto Msg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.K, kindRequired...),
		validation.Field(&dto.ID, id.RequiredWhen(dto.K == ID)...),
		validation.Field(&dto.Sym, sym.ReqiredWhen(dto.K == Sym)...),
	)
}

func MsgFromPH(ph ADT) string {
	return string(ph)
}

func MsgToPH(dto string) (ADT, error) {
	return ADT(dto), nil
}
