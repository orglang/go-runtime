package typeexp

import (
	"database/sql"
	"fmt"

	"golang.org/x/exp/maps"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/qualsym"

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
		choices := make(map[qualsym.ADT]ExpRec, len(spec.Zs))
		for lab, st := range spec.Zs {
			choices[lab] = ConvertSpecToRec(st)
		}
		return WithRec{ExpID: identity.New(), Zs: choices}
	case PlusSpec:
		choices := make(map[qualsym.ADT]ExpRec, len(spec.Zs))
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
		choices := make(map[qualsym.ADT]ExpSpec, len(rec.Zs))
		for lab, rec := range rec.Zs {
			choices[lab] = ConvertRecToSpec(rec)
		}
		return WithSpec{Zs: choices}
	case PlusRec:
		choices := make(map[qualsym.ADT]ExpSpec, len(rec.Zs))
		for lab, st := range rec.Zs {
			choices[lab] = ConvertRecToSpec(st)
		}
		return PlusSpec{Zs: choices}
	default:
		panic(ErrRecTypeUnexpected(rec))
	}
}

func MsgFromTermSpec(s ExpSpec) typeexp.ExpSpecME {
	switch spec := s.(type) {
	case OneSpec:
		return typeexp.ExpSpecME{K: typeexp.OneExp}
	case LinkSpec:
		return typeexp.ExpSpecME{
			K:    typeexp.LinkExp,
			Link: &typeexp.LinkSpecME{TypeQN: qualsym.ConvertToString(spec.TypeQN)}}
	case TensorSpec:
		return typeexp.ExpSpecME{
			K: typeexp.TensorExp,
			Tensor: &typeexp.ProdSpecME{
				ValES:  MsgFromTermSpec(spec.Y),
				ContES: MsgFromTermSpec(spec.Z),
			},
		}
	case LolliSpec:
		return typeexp.ExpSpecME{
			K: typeexp.LolliExp,
			Lolli: &typeexp.ProdSpecME{
				ValES:  MsgFromTermSpec(spec.Y),
				ContES: MsgFromTermSpec(spec.Z),
			},
		}
	case WithSpec:
		choices := make([]typeexp.ChoiceSpecME, len(spec.Zs))
		for i, l := range maps.Keys(spec.Zs) {
			choices[i] = typeexp.ChoiceSpecME{Label: string(l), ContES: MsgFromTermSpec(spec.Zs[l])}
		}
		return typeexp.ExpSpecME{K: typeexp.WithExp, With: &typeexp.SumSpecME{Choices: choices}}
	case PlusSpec:
		choices := make([]typeexp.ChoiceSpecME, len(spec.Zs))
		for i, l := range maps.Keys(spec.Zs) {
			choices[i] = typeexp.ChoiceSpecME{Label: string(l), ContES: MsgFromTermSpec(spec.Zs[l])}
		}
		return typeexp.ExpSpecME{K: typeexp.PlusExp, Plus: &typeexp.SumSpecME{Choices: choices}}
	default:
		panic(ErrSpecTypeUnexpected(s))
	}
}

func MsgToTermSpec(dto typeexp.ExpSpecME) (ExpSpec, error) {
	switch dto.K {
	case typeexp.OneExp:
		return OneSpec{}, nil
	case typeexp.LinkExp:
		roleQN, err := qualsym.ConvertFromString(dto.Link.TypeQN)
		if err != nil {
			return nil, err
		}
		return LinkSpec{TypeQN: roleQN}, nil
	case typeexp.TensorExp:
		v, err := MsgToTermSpec(dto.Tensor.ValES)
		if err != nil {
			return nil, err
		}
		s, err := MsgToTermSpec(dto.Tensor.ContES)
		if err != nil {
			return nil, err
		}
		return TensorSpec{Y: v, Z: s}, nil
	case typeexp.LolliExp:
		v, err := MsgToTermSpec(dto.Lolli.ValES)
		if err != nil {
			return nil, err
		}
		s, err := MsgToTermSpec(dto.Lolli.ContES)
		if err != nil {
			return nil, err
		}
		return LolliSpec{Y: v, Z: s}, nil
	case typeexp.PlusExp:
		choices := make(map[qualsym.ADT]ExpSpec, len(dto.Plus.Choices))
		for _, ch := range dto.Plus.Choices {
			choice, err := MsgToTermSpec(ch.ContES)
			if err != nil {
				return nil, err
			}
			choices[qualsym.ADT(ch.Label)] = choice
		}
		return PlusSpec{Zs: choices}, nil
	case typeexp.WithExp:
		choices := make(map[qualsym.ADT]ExpSpec, len(dto.With.Choices))
		for _, ch := range dto.With.Choices {
			choice, err := MsgToTermSpec(ch.ContES)
			if err != nil {
				return nil, err
			}
			choices[qualsym.ADT(ch.Label)] = choice
		}
		return WithSpec{Zs: choices}, nil
	default:
		panic(typeexp.ErrKindUnexpected(dto.K))
	}
}

func MsgFromTermRef(r ExpRef) typeexp.ExpRefME {
	ident := r.Ident().String()
	switch r.(type) {
	case OneRef, OneRec:
		return typeexp.ExpRefME{K: typeexp.OneExp, ExpID: ident}
	case LinkRef, LinkRec:
		return typeexp.ExpRefME{K: typeexp.LinkExp, ExpID: ident}
	case TensorRef, TensorRec:
		return typeexp.ExpRefME{K: typeexp.TensorExp, ExpID: ident}
	case LolliRef, LolliRec:
		return typeexp.ExpRefME{K: typeexp.LolliExp, ExpID: ident}
	case PlusRef, PlusRec:
		return typeexp.ExpRefME{K: typeexp.PlusExp, ExpID: ident}
	case WithRef, WithRec:
		return typeexp.ExpRefME{K: typeexp.WithExp, ExpID: ident}
	default:
		panic(ErrRefTypeUnexpected(r))
	}
}

func MsgToTermRef(dto typeexp.ExpRefME) (ExpRef, error) {
	rid, err := identity.ConvertFromString(dto.ExpID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case typeexp.OneExp:
		return OneRef{rid}, nil
	case typeexp.LinkExp:
		return LinkRef{rid}, nil
	case typeexp.TensorExp:
		return TensorRef{rid}, nil
	case typeexp.LolliExp:
		return LolliRef{rid}, nil
	case typeexp.PlusExp:
		return PlusRef{rid}, nil
	case typeexp.WithExp:
		return WithRef{rid}, nil
	default:
		panic(typeexp.ErrKindUnexpected(dto.K))
	}
}

func DataFromTermRef(ref ExpRef) *ExpRefDS {
	if ref == nil {
		return nil
	}
	rid := ref.Ident().String()
	switch ref.(type) {
	case OneRef, OneRec:
		return &ExpRefDS{K: oneExp, ExpID: rid}
	case LinkRef, LinkRec:
		return &ExpRefDS{K: linkExp, ExpID: rid}
	case TensorRef, TensorRec:
		return &ExpRefDS{K: tensorExp, ExpID: rid}
	case LolliRef, LolliRec:
		return &ExpRefDS{K: lolliExp, ExpID: rid}
	case PlusRef, PlusRec:
		return &ExpRefDS{K: plusExp, ExpID: rid}
	case WithRef, WithRec:
		return &ExpRefDS{K: withExp, ExpID: rid}
	default:
		panic(ErrRefTypeUnexpected(ref))
	}
}

func DataToTermRef(dto *ExpRefDS) (ExpRef, error) {
	if dto == nil {
		return nil, nil
	}
	rid, err := identity.ConvertFromString(dto.ExpID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case oneExp:
		return OneRef{rid}, nil
	case linkExp:
		return LinkRef{rid}, nil
	case tensorExp:
		return TensorRef{rid}, nil
	case lolliExp:
		return LolliRef{rid}, nil
	case plusExp:
		return PlusRef{rid}, nil
	case withExp:
		return WithRef{rid}, nil
	default:
		panic(errUnexpectedKind(dto.K))
	}
}

func DataToTermRec(dto *expRecDS) (ExpRec, error) {
	states := make(map[string]stateDS, len(dto.States))
	for _, dto := range dto.States {
		states[dto.ExpID] = dto
	}
	return statesToTermRec(states, states[dto.ExpID])
}

func DataFromTermRec(root ExpRec) *expRecDS {
	if root == nil {
		return nil
	}
	dto := &expRecDS{
		ExpID:  root.Ident().String(),
		States: nil,
	}
	statesFromTermRec("", root, dto)
	return dto
}

func statesToTermRec(states map[string]stateDS, st stateDS) (ExpRec, error) {
	stID, err := identity.ConvertFromString(st.ExpID)
	if err != nil {
		return nil, err
	}
	switch st.K {
	case oneExp:
		return OneRec{ExpID: stID}, nil
	case linkExp:
		roleQN, err := qualsym.ConvertFromString(st.Spec.Link)
		if err != nil {
			return nil, err
		}
		return LinkRec{ExpID: stID, TypeQN: roleQN}, nil
	case tensorExp:
		b, err := statesToTermRec(states, states[st.Spec.Tensor.ValES])
		if err != nil {
			return nil, err
		}
		c, err := statesToTermRec(states, states[st.Spec.Tensor.ContES])
		if err != nil {
			return nil, err
		}
		return TensorRec{ExpID: stID, Y: b, Z: c}, nil
	case lolliExp:
		y, err := statesToTermRec(states, states[st.Spec.Lolli.ValES])
		if err != nil {
			return nil, err
		}
		z, err := statesToTermRec(states, states[st.Spec.Lolli.ContES])
		if err != nil {
			return nil, err
		}
		return LolliRec{ExpID: stID, Y: y, Z: z}, nil
	case plusExp:
		choices := make(map[qualsym.ADT]ExpRec, len(st.Spec.Plus))
		for _, ch := range st.Spec.Plus {
			choice, err := statesToTermRec(states, states[ch.ContES])
			if err != nil {
				return nil, err
			}
			choices[qualsym.ADT(ch.Lab)] = choice
		}
		return PlusRec{ExpID: stID, Zs: choices}, nil
	case withExp:
		choices := make(map[qualsym.ADT]ExpRec, len(st.Spec.With))
		for _, ch := range st.Spec.With {
			choice, err := statesToTermRec(states, states[ch.ContES])
			if err != nil {
				return nil, err
			}
			choices[qualsym.ADT(ch.Lab)] = choice
		}
		return WithRec{ExpID: stID, Zs: choices}, nil
	default:
		panic(errUnexpectedKind(st.K))
	}
}

func statesFromTermRec(from string, r ExpRec, dto *expRecDS) (string, error) {
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
				Link: qualsym.ConvertToString(root.TypeQN),
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case TensorRec:
		val, err := statesFromTermRec(stID, root.Y, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromTermRec(stID, root.Z, dto)
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
		val, err := statesFromTermRec(stID, root.Y, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromTermRec(stID, root.Z, dto)
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
			cont, err := statesFromTermRec(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumDS{string(label), cont})
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
			cont, err := statesFromTermRec(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumDS{string(label), cont})
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
