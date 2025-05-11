package def

import (
	"fmt"

	"golang.org/x/exp/maps"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"
)

type TypeSpecMsg struct {
	TypeQN string      `json:"qn"`
	TypeTS TermSpecMsg `json:"state"`
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
	TypeID string      `json:"id" param:"id"`
	TypeRN int64       `json:"rev" query:"rev"`
	Title  string      `json:"title"`
	TypeQN string      `json:"qn"`
	TypeTS TermSpecMsg `json:"state"`
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

type TermSpecMsg struct {
	K      TermKind     `json:"kind"`
	Link   *LinkSpecMsg `json:"link,omitempty"`
	Tensor *ProdSpecMsg `json:"tensor,omitempty"`
	Lolli  *ProdSpecMsg `json:"lolli,omitempty"`
	Plus   *SumSpecMsg  `json:"plus,omitempty"`
	With   *SumSpecMsg  `json:"with,omitempty"`
}

func (dto TermSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.K, kindRequired...),
		validation.Field(&dto.Link, validation.Required.When(dto.K == LinkKind), validation.Skip.When(dto.K != LinkKind)),
		validation.Field(&dto.Tensor, validation.Required.When(dto.K == TensorKind), validation.Skip.When(dto.K != TensorKind)),
		validation.Field(&dto.Lolli, validation.Required.When(dto.K == LolliKind), validation.Skip.When(dto.K != LolliKind)),
		validation.Field(&dto.Plus, validation.Required.When(dto.K == PlusKind), validation.Skip.When(dto.K != PlusKind)),
		validation.Field(&dto.With, validation.Required.When(dto.K == WithKind), validation.Skip.When(dto.K != WithKind)),
	)
}

type LinkSpecMsg struct {
	QN string `json:"qn"`
}

func (dto LinkSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.QN, sym.Required...),
	)
}

type ProdSpecMsg struct {
	Value TermSpecMsg `json:"value"`
	Cont  TermSpecMsg `json:"cont"`
}

func (dto ProdSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Value, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type SumSpecMsg struct {
	Choices []ChoiceSpecMsg `json:"choices"`
}

func (dto SumSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Choices,
			validation.Required,
			validation.Length(1, 10),
			validation.Each(validation.Required),
		),
	)
}

type ChoiceSpecMsg struct {
	Label string      `json:"label"`
	Cont  TermSpecMsg `json:"cont"`
}

func (dto ChoiceSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Label, core.NameRequired...),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type TermRefMsg struct {
	ID string   `json:"id" param:"id"`
	K  TermKind `json:"kind"`
}

func (dto TermRefMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ID, id.Required...),
		validation.Field(&dto.K, kindRequired...),
	)
}

type TermKind string

const (
	OneKind    = TermKind("one")
	LinkKind   = TermKind("link")
	TensorKind = TermKind("tensor")
	LolliKind  = TermKind("lolli")
	PlusKind   = TermKind("plus")
	WithKind   = TermKind("with")
)

var kindRequired = []validation.Rule{
	validation.Required,
	validation.In(OneKind, LinkKind, TensorKind, LolliKind, PlusKind, WithKind),
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend Msg.*
var (
	MsgFromTermRefs func([]TermRef) []TermRefMsg
	MsgToTermRefs   func([]TermRefMsg) ([]TermRef, error)
)

func MsgFromTermSpec(s TermSpec) TermSpecMsg {
	switch spec := s.(type) {
	case OneSpec:
		return TermSpecMsg{K: OneKind}
	case LinkSpec:
		return TermSpecMsg{
			K:    LinkKind,
			Link: &LinkSpecMsg{QN: sym.ConvertToString(spec.TypeQN)}}
	case TensorSpec:
		return TermSpecMsg{
			K: TensorKind,
			Tensor: &ProdSpecMsg{
				Value: MsgFromTermSpec(spec.B),
				Cont:  MsgFromTermSpec(spec.C),
			},
		}
	case LolliSpec:
		return TermSpecMsg{
			K: LolliKind,
			Lolli: &ProdSpecMsg{
				Value: MsgFromTermSpec(spec.Y),
				Cont:  MsgFromTermSpec(spec.Z),
			},
		}
	case WithSpec:
		choices := make([]ChoiceSpecMsg, len(spec.Choices))
		for i, l := range maps.Keys(spec.Choices) {
			choices[i] = ChoiceSpecMsg{Label: string(l), Cont: MsgFromTermSpec(spec.Choices[l])}
		}
		return TermSpecMsg{K: WithKind, With: &SumSpecMsg{Choices: choices}}
	case PlusSpec:
		choices := make([]ChoiceSpecMsg, len(spec.Choices))
		for i, l := range maps.Keys(spec.Choices) {
			choices[i] = ChoiceSpecMsg{Label: string(l), Cont: MsgFromTermSpec(spec.Choices[l])}
		}
		return TermSpecMsg{K: PlusKind, Plus: &SumSpecMsg{Choices: choices}}
	default:
		panic(ErrSpecTypeUnexpected(s))
	}
}

func MsgToTermSpec(dto TermSpecMsg) (TermSpec, error) {
	switch dto.K {
	case OneKind:
		return OneSpec{}, nil
	case LinkKind:
		roleQN, err := sym.ConvertFromString(dto.Link.QN)
		if err != nil {
			return nil, err
		}
		return LinkSpec{TypeQN: roleQN}, nil
	case TensorKind:
		v, err := MsgToTermSpec(dto.Tensor.Value)
		if err != nil {
			return nil, err
		}
		s, err := MsgToTermSpec(dto.Tensor.Cont)
		if err != nil {
			return nil, err
		}
		return TensorSpec{B: v, C: s}, nil
	case LolliKind:
		v, err := MsgToTermSpec(dto.Lolli.Value)
		if err != nil {
			return nil, err
		}
		s, err := MsgToTermSpec(dto.Lolli.Cont)
		if err != nil {
			return nil, err
		}
		return LolliSpec{Y: v, Z: s}, nil
	case PlusKind:
		choices := make(map[sym.ADT]TermSpec, len(dto.Plus.Choices))
		for _, ch := range dto.Plus.Choices {
			choice, err := MsgToTermSpec(ch.Cont)
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Label)] = choice
		}
		return PlusSpec{Choices: choices}, nil
	case WithKind:
		choices := make(map[sym.ADT]TermSpec, len(dto.With.Choices))
		for _, ch := range dto.With.Choices {
			choice, err := MsgToTermSpec(ch.Cont)
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Label)] = choice
		}
		return WithSpec{Choices: choices}, nil
	default:
		panic(errKindUnexpected(dto.K))
	}
}

func MsgFromTermRef(r TermRef) TermRefMsg {
	ident := r.Ident().String()
	switch r.(type) {
	case OneRef, OneRec:
		return TermRefMsg{K: OneKind, ID: ident}
	case LinkRef, LinkRec:
		return TermRefMsg{K: LinkKind, ID: ident}
	case TensorRef, TensorRec:
		return TermRefMsg{K: TensorKind, ID: ident}
	case LolliRef, LolliRec:
		return TermRefMsg{K: LolliKind, ID: ident}
	case PlusRef, PlusRec:
		return TermRefMsg{K: PlusKind, ID: ident}
	case WithRef, WithRec:
		return TermRefMsg{K: WithKind, ID: ident}
	default:
		panic(ErrRefTypeUnexpected(r))
	}
}

func MsgToTermRef(dto TermRefMsg) (TermRef, error) {
	rid, err := id.ConvertFromString(dto.ID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case OneKind:
		return OneRef{rid}, nil
	case LinkKind:
		return LinkRef{rid}, nil
	case TensorKind:
		return TensorRef{rid}, nil
	case LolliKind:
		return LolliRef{rid}, nil
	case PlusKind:
		return PlusRef{rid}, nil
	case WithKind:
		return WithRef{rid}, nil
	default:
		panic(errKindUnexpected(dto.K))
	}
}

func errKindUnexpected(got TermKind) error {
	return fmt.Errorf("kind unexpected: %v", got)
}
