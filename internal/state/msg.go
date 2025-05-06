package state

import (
	"fmt"

	"golang.org/x/exp/maps"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type SpecMsg struct {
	K      Kind     `json:"kind"`
	Link   *LinkMsg `json:"link,omitempty"`
	Tensor *ProdMsg `json:"tensor,omitempty"`
	Lolli  *ProdMsg `json:"lolli,omitempty"`
	Plus   *SumMsg  `json:"plus,omitempty"`
	With   *SumMsg  `json:"with,omitempty"`
}

func (dto SpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.K, kindRequired...),
		validation.Field(&dto.Link, validation.Required.When(dto.K == Link), validation.Skip.When(dto.K != Link)),
		validation.Field(&dto.Tensor, validation.Required.When(dto.K == Tensor), validation.Skip.When(dto.K != Tensor)),
		validation.Field(&dto.Lolli, validation.Required.When(dto.K == Lolli), validation.Skip.When(dto.K != Lolli)),
		validation.Field(&dto.Plus, validation.Required.When(dto.K == Plus), validation.Skip.When(dto.K != Plus)),
		validation.Field(&dto.With, validation.Required.When(dto.K == With), validation.Skip.When(dto.K != With)),
	)
}

type LinkMsg struct {
	QN string `json:"qn"`
}

func (dto LinkMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.QN, sym.Required...),
	)
}

type ProdMsg struct {
	Value SpecMsg `json:"value"`
	Cont  SpecMsg `json:"cont"`
}

func (dto ProdMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Value, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type SumMsg struct {
	Choices []ChoiceMsg `json:"choices"`
}

func (dto SumMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Choices,
			validation.Required,
			validation.Length(1, 10),
			validation.Each(validation.Required),
		),
	)
}

type ChoiceMsg struct {
	Label string  `json:"label"`
	Cont  SpecMsg `json:"cont"`
}

func (dto ChoiceMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Label, core.NameRequired...),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type RefMsg struct {
	ID string `json:"id" param:"id"`
	K  Kind   `json:"kind"`
}

func (dto RefMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.ID, id.Required...),
		validation.Field(&dto.K, kindRequired...),
	)
}

type Kind string

const (
	One    = Kind("one")
	Link   = Kind("link")
	Tensor = Kind("tensor")
	Lolli  = Kind("lolli")
	Plus   = Kind("plus")
	With   = Kind("with")
)

var kindRequired = []validation.Rule{
	validation.Required,
	validation.In(One, Link, Tensor, Lolli, Plus, With),
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend Msg.*
var (
	MsgFromRefs func([]Ref) []RefMsg
	MsgToRefs   func([]RefMsg) ([]Ref, error)
)

func MsgFromSpec(s Spec) SpecMsg {
	switch spec := s.(type) {
	case OneSpec:
		return SpecMsg{K: One}
	case LinkSpec:
		return SpecMsg{
			K:    Link,
			Link: &LinkMsg{QN: sym.ConvertToString(spec.RoleQN)}}
	case TensorSpec:
		return SpecMsg{
			K: Tensor,
			Tensor: &ProdMsg{
				Value: MsgFromSpec(spec.B),
				Cont:  MsgFromSpec(spec.C),
			},
		}
	case LolliSpec:
		return SpecMsg{
			K: Lolli,
			Lolli: &ProdMsg{
				Value: MsgFromSpec(spec.Y),
				Cont:  MsgFromSpec(spec.Z),
			},
		}
	case WithSpec:
		choices := make([]ChoiceMsg, len(spec.Choices))
		for i, l := range maps.Keys(spec.Choices) {
			choices[i] = ChoiceMsg{Label: string(l), Cont: MsgFromSpec(spec.Choices[l])}
		}
		return SpecMsg{K: With, With: &SumMsg{Choices: choices}}
	case PlusSpec:
		choices := make([]ChoiceMsg, len(spec.Choices))
		for i, l := range maps.Keys(spec.Choices) {
			choices[i] = ChoiceMsg{Label: string(l), Cont: MsgFromSpec(spec.Choices[l])}
		}
		return SpecMsg{K: Plus, Plus: &SumMsg{Choices: choices}}
	default:
		panic(ErrSpecTypeUnexpected(s))
	}
}

func MsgToSpec(dto SpecMsg) (Spec, error) {
	switch dto.K {
	case One:
		return OneSpec{}, nil
	case Link:
		roleQN, err := sym.ConvertFromString(dto.Link.QN)
		if err != nil {
			return nil, err
		}
		return LinkSpec{RoleQN: roleQN}, nil
	case Tensor:
		v, err := MsgToSpec(dto.Tensor.Value)
		if err != nil {
			return nil, err
		}
		s, err := MsgToSpec(dto.Tensor.Cont)
		if err != nil {
			return nil, err
		}
		return TensorSpec{B: v, C: s}, nil
	case Lolli:
		v, err := MsgToSpec(dto.Lolli.Value)
		if err != nil {
			return nil, err
		}
		s, err := MsgToSpec(dto.Lolli.Cont)
		if err != nil {
			return nil, err
		}
		return LolliSpec{Y: v, Z: s}, nil
	case Plus:
		choices := make(map[sym.ADT]Spec, len(dto.Plus.Choices))
		for _, ch := range dto.Plus.Choices {
			choice, err := MsgToSpec(ch.Cont)
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Label)] = choice
		}
		return PlusSpec{Choices: choices}, nil
	case With:
		choices := make(map[sym.ADT]Spec, len(dto.With.Choices))
		for _, ch := range dto.With.Choices {
			choice, err := MsgToSpec(ch.Cont)
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

func MsgFromRef(r Ref) RefMsg {
	ident := r.Ident().String()
	switch r.(type) {
	case OneRef, OneRoot:
		return RefMsg{K: One, ID: ident}
	case LinkRef, LinkRoot:
		return RefMsg{K: Link, ID: ident}
	case TensorRef, TensorRoot:
		return RefMsg{K: Tensor, ID: ident}
	case LolliRef, LolliRoot:
		return RefMsg{K: Lolli, ID: ident}
	case PlusRef, PlusRoot:
		return RefMsg{K: Plus, ID: ident}
	case WithRef, WithRoot:
		return RefMsg{K: With, ID: ident}
	default:
		panic(ErrRefTypeUnexpected(r))
	}
}

func MsgToRef(dto RefMsg) (Ref, error) {
	rid, err := id.ConvertFromString(dto.ID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case One:
		return OneRef{rid}, nil
	case Link:
		return LinkRef{rid}, nil
	case Tensor:
		return TensorRef{rid}, nil
	case Lolli:
		return LolliRef{rid}, nil
	case Plus:
		return PlusRef{rid}, nil
	case With:
		return WithRef{rid}, nil
	default:
		panic(errKindUnexpected(dto.K))
	}
}

func ErrPolarityUnexpected(got Root) error {
	return fmt.Errorf("root polarity unexpected: %v", got.Pol())
}

func ErrPolarityMismatch(a, b Root) error {
	return fmt.Errorf("root polarity mismatch: %v!=%v", a.Pol(), b.Pol())
}

func errKindUnexpected(got Kind) error {
	return fmt.Errorf("kind unexpected: %v", got)
}
