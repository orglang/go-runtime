package typeexp

import (
	"database/sql"
	"fmt"

	"golang.org/x/exp/maps"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/uniqsym"

	"github.com/orglang/go-sdk/adt/typeexp"
)

func ConvertSpecToRec(s ExpSpec) ExpRec {
	if s == nil {
		return nil
	}
	switch spec := s.(type) {
	case OneSpec:
		return OneRec{ExpID: identity.New()}
	case LinkSpec:
		return LinkRec{ExpID: identity.New(), TypeQN: spec.TypeQN}
	case TensorSpec:
		return TensorRec{
			ExpID: identity.New(),
			Y:     ConvertSpecToRec(spec.Y),
			Z:     ConvertSpecToRec(spec.Z),
		}
	case LolliSpec:
		return LolliRec{
			ExpID: identity.New(),
			Y:     ConvertSpecToRec(spec.Y),
			Z:     ConvertSpecToRec(spec.Z),
		}
	case WithSpec:
		choices := make(map[uniqsym.ADT]ExpRec, len(spec.Zs))
		for lab, st := range spec.Zs {
			choices[lab] = ConvertSpecToRec(st)
		}
		return WithRec{ExpID: identity.New(), Zs: choices}
	case PlusSpec:
		choices := make(map[uniqsym.ADT]ExpRec, len(spec.Zs))
		for lab, rec := range spec.Zs {
			choices[lab] = ConvertSpecToRec(rec)
		}
		return PlusRec{ExpID: identity.New(), Zs: choices}
	default:
		panic(ErrSpecTypeUnexpected(spec))
	}
}

func ConvertRecToSpec(r ExpRec) ExpSpec {
	if r == nil {
		return nil
	}
	switch rec := r.(type) {
	case OneRec:
		return OneSpec{}
	case LinkRec:
		return LinkSpec{TypeQN: rec.TypeQN}
	case TensorRec:
		return TensorSpec{
			Y: ConvertRecToSpec(rec.Y),
			Z: ConvertRecToSpec(rec.Z),
		}
	case LolliRec:
		return LolliSpec{
			Y: ConvertRecToSpec(rec.Y),
			Z: ConvertRecToSpec(rec.Z),
		}
	case WithRec:
		choices := make(map[uniqsym.ADT]ExpSpec, len(rec.Zs))
		for lab, rec := range rec.Zs {
			choices[lab] = ConvertRecToSpec(rec)
		}
		return WithSpec{Zs: choices}
	case PlusRec:
		choices := make(map[uniqsym.ADT]ExpSpec, len(rec.Zs))
		for lab, st := range rec.Zs {
			choices[lab] = ConvertRecToSpec(st)
		}
		return PlusSpec{Zs: choices}
	default:
		panic(ErrRecTypeUnexpected(rec))
	}
}

func MsgFromExpSpec(s ExpSpec) typeexp.ExpSpec {
	switch spec := s.(type) {
	case OneSpec:
		return typeexp.ExpSpec{K: typeexp.One}
	case LinkSpec:
		return typeexp.ExpSpec{
			K:    typeexp.Link,
			Link: &typeexp.LinkSpec{TypeQN: uniqsym.ConvertToString(spec.TypeQN)}}
	case TensorSpec:
		return typeexp.ExpSpec{
			K: typeexp.Tensor,
			Tensor: &typeexp.ProdSpec{
				ValES:  MsgFromExpSpec(spec.Y),
				ContES: MsgFromExpSpec(spec.Z),
			},
		}
	case LolliSpec:
		return typeexp.ExpSpec{
			K: typeexp.Lolli,
			Lolli: &typeexp.ProdSpec{
				ValES:  MsgFromExpSpec(spec.Y),
				ContES: MsgFromExpSpec(spec.Z),
			},
		}
	case WithSpec:
		choices := make([]typeexp.ChoiceSpec, len(spec.Zs))
		for i, l := range maps.Keys(spec.Zs) {
			choices[i] = typeexp.ChoiceSpec{
				LabQN:  uniqsym.ConvertToString(l),
				ContES: MsgFromExpSpec(spec.Zs[l]),
			}
		}
		return typeexp.ExpSpec{K: typeexp.With, With: &typeexp.SumSpec{Choices: choices}}
	case PlusSpec:
		choices := make([]typeexp.ChoiceSpec, len(spec.Zs))
		for i, l := range maps.Keys(spec.Zs) {
			choices[i] = typeexp.ChoiceSpec{
				LabQN:  uniqsym.ConvertToString(l),
				ContES: MsgFromExpSpec(spec.Zs[l]),
			}
		}
		return typeexp.ExpSpec{K: typeexp.Plus, Plus: &typeexp.SumSpec{Choices: choices}}
	default:
		panic(ErrSpecTypeUnexpected(s))
	}
}

func MsgToExpSpec(dto typeexp.ExpSpec) (ExpSpec, error) {
	switch dto.K {
	case typeexp.One:
		return OneSpec{}, nil
	case typeexp.Link:
		typeQN, err := uniqsym.ConvertFromString(dto.Link.TypeQN)
		if err != nil {
			return nil, err
		}
		return LinkSpec{TypeQN: typeQN}, nil
	case typeexp.Tensor:
		valES, err := MsgToExpSpec(dto.Tensor.ValES)
		if err != nil {
			return nil, err
		}
		contES, err := MsgToExpSpec(dto.Tensor.ContES)
		if err != nil {
			return nil, err
		}
		return TensorSpec{Y: valES, Z: contES}, nil
	case typeexp.Lolli:
		valES, err := MsgToExpSpec(dto.Lolli.ValES)
		if err != nil {
			return nil, err
		}
		contES, err := MsgToExpSpec(dto.Lolli.ContES)
		if err != nil {
			return nil, err
		}
		return LolliSpec{Y: valES, Z: contES}, nil
	case typeexp.Plus:
		choices := make(map[uniqsym.ADT]ExpSpec, len(dto.Plus.Choices))
		for _, ch := range dto.Plus.Choices {
			choice, err := MsgToExpSpec(ch.ContES)
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(ch.LabQN)
			if err != nil {
				return nil, err
			}
			choices[label] = choice
		}
		return PlusSpec{Zs: choices}, nil
	case typeexp.With:
		choices := make(map[uniqsym.ADT]ExpSpec, len(dto.With.Choices))
		for _, ch := range dto.With.Choices {
			choice, err := MsgToExpSpec(ch.ContES)
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(ch.LabQN)
			if err != nil {
				return nil, err
			}
			choices[label] = choice
		}
		return WithSpec{Zs: choices}, nil
	default:
		panic(typeexp.ErrKindUnexpected(dto.K))
	}
}

func MsgFromExpRef(r ExpRef) typeexp.ExpRef {
	ident := r.Ident().String()
	switch r.(type) {
	case OneRef, OneRec:
		return typeexp.ExpRef{K: typeexp.One, ExpID: ident}
	case LinkRef, LinkRec:
		return typeexp.ExpRef{K: typeexp.Link, ExpID: ident}
	case TensorRef, TensorRec:
		return typeexp.ExpRef{K: typeexp.Tensor, ExpID: ident}
	case LolliRef, LolliRec:
		return typeexp.ExpRef{K: typeexp.Lolli, ExpID: ident}
	case PlusRef, PlusRec:
		return typeexp.ExpRef{K: typeexp.Plus, ExpID: ident}
	case WithRef, WithRec:
		return typeexp.ExpRef{K: typeexp.With, ExpID: ident}
	default:
		panic(ErrRefTypeUnexpected(r))
	}
}

func MsgToExpRef(dto typeexp.ExpRef) (ExpRef, error) {
	expID, err := identity.ConvertFromString(dto.ExpID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case typeexp.One:
		return OneRef{expID}, nil
	case typeexp.Link:
		return LinkRef{expID}, nil
	case typeexp.Tensor:
		return TensorRef{expID}, nil
	case typeexp.Lolli:
		return LolliRef{expID}, nil
	case typeexp.Plus:
		return PlusRef{expID}, nil
	case typeexp.With:
		return WithRef{expID}, nil
	default:
		panic(typeexp.ErrKindUnexpected(dto.K))
	}
}

func DataFromExpRef(ref ExpRef) *ExpRefDS {
	if ref == nil {
		return nil
	}
	expID := ref.Ident().String()
	switch ref.(type) {
	case OneRef, OneRec:
		return &ExpRefDS{K: oneExp, ExpID: expID}
	case LinkRef, LinkRec:
		return &ExpRefDS{K: linkExp, ExpID: expID}
	case TensorRef, TensorRec:
		return &ExpRefDS{K: tensorExp, ExpID: expID}
	case LolliRef, LolliRec:
		return &ExpRefDS{K: lolliExp, ExpID: expID}
	case PlusRef, PlusRec:
		return &ExpRefDS{K: plusExp, ExpID: expID}
	case WithRef, WithRec:
		return &ExpRefDS{K: withExp, ExpID: expID}
	default:
		panic(ErrRefTypeUnexpected(ref))
	}
}

func DataToExpRef(dto *ExpRefDS) (ExpRef, error) {
	if dto == nil {
		return nil, nil
	}
	expID, err := identity.ConvertFromString(dto.ExpID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case oneExp:
		return OneRef{expID}, nil
	case linkExp:
		return LinkRef{expID}, nil
	case tensorExp:
		return TensorRef{expID}, nil
	case lolliExp:
		return LolliRef{expID}, nil
	case plusExp:
		return PlusRef{expID}, nil
	case withExp:
		return WithRef{expID}, nil
	default:
		panic(errUnexpectedKind(dto.K))
	}
}

func DataToExpRec(dto *expRecDS) (ExpRec, error) {
	states := make(map[string]stateDS, len(dto.States))
	for _, dto := range dto.States {
		states[dto.ExpID] = dto
	}
	return statesToExpRec(states, states[dto.ExpID])
}

func DataFromExpRec(rec ExpRec) *expRecDS {
	if rec == nil {
		return nil
	}
	dto := &expRecDS{
		ExpID:  rec.Ident().String(),
		States: nil,
	}
	statesFromExpRec("", rec, dto)
	return dto
}

func statesToExpRec(states map[string]stateDS, st stateDS) (ExpRec, error) {
	stID, err := identity.ConvertFromString(st.ExpID)
	if err != nil {
		return nil, err
	}
	switch st.K {
	case oneExp:
		return OneRec{ExpID: stID}, nil
	case linkExp:
		roleQN, err := uniqsym.ConvertFromString(st.Spec.Link)
		if err != nil {
			return nil, err
		}
		return LinkRec{ExpID: stID, TypeQN: roleQN}, nil
	case tensorExp:
		b, err := statesToExpRec(states, states[st.Spec.Tensor.ValES])
		if err != nil {
			return nil, err
		}
		c, err := statesToExpRec(states, states[st.Spec.Tensor.ContES])
		if err != nil {
			return nil, err
		}
		return TensorRec{ExpID: stID, Y: b, Z: c}, nil
	case lolliExp:
		y, err := statesToExpRec(states, states[st.Spec.Lolli.ValES])
		if err != nil {
			return nil, err
		}
		z, err := statesToExpRec(states, states[st.Spec.Lolli.ContES])
		if err != nil {
			return nil, err
		}
		return LolliRec{ExpID: stID, Y: y, Z: z}, nil
	case plusExp:
		choices := make(map[uniqsym.ADT]ExpRec, len(st.Spec.Plus))
		for _, ch := range st.Spec.Plus {
			choice, err := statesToExpRec(states, states[ch.ContES])
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(ch.Lab)
			if err != nil {
				return nil, err
			}
			choices[label] = choice
		}
		return PlusRec{ExpID: stID, Zs: choices}, nil
	case withExp:
		choices := make(map[uniqsym.ADT]ExpRec, len(st.Spec.With))
		for _, ch := range st.Spec.With {
			choice, err := statesToExpRec(states, states[ch.ContES])
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(ch.Lab)
			if err != nil {
				return nil, err
			}
			choices[label] = choice
		}
		return WithRec{ExpID: stID, Zs: choices}, nil
	default:
		panic(errUnexpectedKind(st.K))
	}
}

func statesFromExpRec(from string, r ExpRec, dto *expRecDS) (string, error) {
	var fromID sql.NullString
	if len(from) > 0 {
		fromID = sql.NullString{String: from, Valid: true}
	}
	stID := r.Ident().String()
	switch root := r.(type) {
	case OneRec:
		st := stateDS{ExpID: stID, K: oneExp, FromID: fromID}
		dto.States = append(dto.States, st)
		return stID, nil
	case LinkRec:
		st := stateDS{
			ExpID:  stID,
			K:      linkExp,
			FromID: fromID,
			Spec: expSpecDS{
				Link: uniqsym.ConvertToString(root.TypeQN),
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case TensorRec:
		val, err := statesFromExpRec(stID, root.Y, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromExpRec(stID, root.Z, dto)
		if err != nil {
			return "", err
		}
		st := stateDS{
			ExpID:  stID,
			K:      tensorExp,
			FromID: fromID,
			Spec: expSpecDS{
				Tensor: &prodDS{val, cont},
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case LolliRec:
		val, err := statesFromExpRec(stID, root.Y, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromExpRec(stID, root.Z, dto)
		if err != nil {
			return "", err
		}
		st := stateDS{
			ExpID:  stID,
			K:      lolliExp,
			FromID: fromID,
			Spec: expSpecDS{
				Lolli: &prodDS{val, cont},
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case PlusRec:
		var choices []sumDS
		for label, choice := range root.Zs {
			cont, err := statesFromExpRec(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumDS{uniqsym.ConvertToString(label), cont})
		}
		st := stateDS{
			ExpID:  stID,
			K:      plusExp,
			FromID: fromID,
			Spec:   expSpecDS{Plus: choices},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case WithRec:
		var choices []sumDS
		for label, choice := range root.Zs {
			cont, err := statesFromExpRec(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumDS{uniqsym.ConvertToString(label), cont})
		}
		st := stateDS{
			ExpID:  stID,
			K:      withExp,
			FromID: fromID,
			Spec:   expSpecDS{With: choices},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	default:
		panic(ErrRecTypeUnexpected(r))
	}
}

func errUnexpectedKind(k expKindDS) error {
	return fmt.Errorf("unexpected kind %q", k)
}
