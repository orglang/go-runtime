package def

import (
	"database/sql"
	"fmt"

	"golang.org/x/exp/maps"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
)

func ConvertSpecToRec(s TermSpec) TermRec {
	if s == nil {
		return nil
	}
	switch spec := s.(type) {
	case OneSpec:
		return OneRec{TermID: id.New()}
	case LinkSpec:
		return LinkRec{TermID: id.New(), TypeQN: spec.TypeQN}
	case TensorSpec:
		return TensorRec{
			TermID: id.New(),
			Y:      ConvertSpecToRec(spec.Y),
			Z:      ConvertSpecToRec(spec.Z),
		}
	case LolliSpec:
		return LolliRec{
			TermID: id.New(),
			Y:      ConvertSpecToRec(spec.Y),
			Z:      ConvertSpecToRec(spec.Z),
		}
	case WithSpec:
		choices := make(map[sym.ADT]TermRec, len(spec.Zs))
		for lab, st := range spec.Zs {
			choices[lab] = ConvertSpecToRec(st)
		}
		return WithRec{TermID: id.New(), Zs: choices}
	case PlusSpec:
		choices := make(map[sym.ADT]TermRec, len(spec.Zs))
		for lab, rec := range spec.Zs {
			choices[lab] = ConvertSpecToRec(rec)
		}
		return PlusRec{TermID: id.New(), Zs: choices}
	default:
		panic(ErrSpecTypeUnexpected(spec))
	}
}

func ConvertRecToSpec(r TermRec) TermSpec {
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
		choices := make(map[sym.ADT]TermSpec, len(rec.Zs))
		for lab, rec := range rec.Zs {
			choices[lab] = ConvertRecToSpec(rec)
		}
		return WithSpec{Zs: choices}
	case PlusRec:
		choices := make(map[sym.ADT]TermSpec, len(rec.Zs))
		for lab, st := range rec.Zs {
			choices[lab] = ConvertRecToSpec(st)
		}
		return PlusSpec{Zs: choices}
	default:
		panic(ErrRecTypeUnexpected(rec))
	}
}

func MsgFromTermSpec(s TermSpec) TermSpecME {
	switch spec := s.(type) {
	case OneSpec:
		return TermSpecME{K: OneKind}
	case LinkSpec:
		return TermSpecME{
			K:    LinkKind,
			Link: &LinkSpecME{QN: sym.ConvertToString(spec.TypeQN)}}
	case TensorSpec:
		return TermSpecME{
			K: TensorKind,
			Tensor: &ProdSpecME{
				Value: MsgFromTermSpec(spec.Y),
				Cont:  MsgFromTermSpec(spec.Z),
			},
		}
	case LolliSpec:
		return TermSpecME{
			K: LolliKind,
			Lolli: &ProdSpecME{
				Value: MsgFromTermSpec(spec.Y),
				Cont:  MsgFromTermSpec(spec.Z),
			},
		}
	case WithSpec:
		choices := make([]ChoiceSpecME, len(spec.Zs))
		for i, l := range maps.Keys(spec.Zs) {
			choices[i] = ChoiceSpecME{Label: string(l), Cont: MsgFromTermSpec(spec.Zs[l])}
		}
		return TermSpecME{K: WithKind, With: &SumSpecME{Choices: choices}}
	case PlusSpec:
		choices := make([]ChoiceSpecME, len(spec.Zs))
		for i, l := range maps.Keys(spec.Zs) {
			choices[i] = ChoiceSpecME{Label: string(l), Cont: MsgFromTermSpec(spec.Zs[l])}
		}
		return TermSpecME{K: PlusKind, Plus: &SumSpecME{Choices: choices}}
	default:
		panic(ErrSpecTypeUnexpected(s))
	}
}

func MsgToTermSpec(dto TermSpecME) (TermSpec, error) {
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
		return TensorSpec{Y: v, Z: s}, nil
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
		return PlusSpec{Zs: choices}, nil
	case WithKind:
		choices := make(map[sym.ADT]TermSpec, len(dto.With.Choices))
		for _, ch := range dto.With.Choices {
			choice, err := MsgToTermSpec(ch.Cont)
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Label)] = choice
		}
		return WithSpec{Zs: choices}, nil
	default:
		panic(errKindUnexpected(dto.K))
	}
}

func MsgFromTermRef(r TermRef) TermRefME {
	ident := r.Ident().String()
	switch r.(type) {
	case OneRef, OneRec:
		return TermRefME{K: OneKind, ID: ident}
	case LinkRef, LinkRec:
		return TermRefME{K: LinkKind, ID: ident}
	case TensorRef, TensorRec:
		return TermRefME{K: TensorKind, ID: ident}
	case LolliRef, LolliRec:
		return TermRefME{K: LolliKind, ID: ident}
	case PlusRef, PlusRec:
		return TermRefME{K: PlusKind, ID: ident}
	case WithRef, WithRec:
		return TermRefME{K: WithKind, ID: ident}
	default:
		panic(ErrRefTypeUnexpected(r))
	}
}

func MsgToTermRef(dto TermRefME) (TermRef, error) {
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

func DataFromTermRef(ref TermRef) *TermRefDS {
	if ref == nil {
		return nil
	}
	rid := ref.Ident().String()
	switch ref.(type) {
	case OneRef, OneRec:
		return &TermRefDS{K: oneKind, ID: rid}
	case LinkRef, LinkRec:
		return &TermRefDS{K: linkKind, ID: rid}
	case TensorRef, TensorRec:
		return &TermRefDS{K: tensorKind, ID: rid}
	case LolliRef, LolliRec:
		return &TermRefDS{K: lolliKind, ID: rid}
	case PlusRef, PlusRec:
		return &TermRefDS{K: plusKind, ID: rid}
	case WithRef, WithRec:
		return &TermRefDS{K: withKind, ID: rid}
	default:
		panic(ErrRefTypeUnexpected(ref))
	}
}

func DataToTermRef(dto *TermRefDS) (TermRef, error) {
	if dto == nil {
		return nil, nil
	}
	rid, err := id.ConvertFromString(dto.ID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case oneKind:
		return OneRef{rid}, nil
	case linkKind:
		return LinkRef{rid}, nil
	case tensorKind:
		return TensorRef{rid}, nil
	case lolliKind:
		return LolliRef{rid}, nil
	case plusKind:
		return PlusRef{rid}, nil
	case withKind:
		return WithRef{rid}, nil
	default:
		panic(errUnexpectedKind(dto.K))
	}
}

func dataToTermRec(dto *termRecDS) (TermRec, error) {
	states := make(map[string]stateDS, len(dto.States))
	for _, dto := range dto.States {
		states[dto.ID] = dto
	}
	return statesToTermRec(states, states[dto.ID])
}

func dataFromTermRec(root TermRec) *termRecDS {
	if root == nil {
		return nil
	}
	dto := &termRecDS{
		ID:     root.Ident().String(),
		States: nil,
	}
	statesFromTermRec("", root, dto)
	return dto
}

func statesToTermRec(states map[string]stateDS, st stateDS) (TermRec, error) {
	stID, err := id.ConvertFromString(st.ID)
	if err != nil {
		return nil, err
	}
	switch st.K {
	case oneKind:
		return OneRec{TermID: stID}, nil
	case linkKind:
		roleQN, err := sym.ConvertFromString(st.Spec.Link)
		if err != nil {
			return nil, err
		}
		return LinkRec{TermID: stID, TypeQN: roleQN}, nil
	case tensorKind:
		b, err := statesToTermRec(states, states[st.Spec.Tensor.Val])
		if err != nil {
			return nil, err
		}
		c, err := statesToTermRec(states, states[st.Spec.Tensor.Cont])
		if err != nil {
			return nil, err
		}
		return TensorRec{TermID: stID, Y: b, Z: c}, nil
	case lolliKind:
		y, err := statesToTermRec(states, states[st.Spec.Lolli.Val])
		if err != nil {
			return nil, err
		}
		z, err := statesToTermRec(states, states[st.Spec.Lolli.Cont])
		if err != nil {
			return nil, err
		}
		return LolliRec{TermID: stID, Y: y, Z: z}, nil
	case plusKind:
		choices := make(map[sym.ADT]TermRec, len(st.Spec.Plus))
		for _, ch := range st.Spec.Plus {
			choice, err := statesToTermRec(states, states[ch.Cont])
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Lab)] = choice
		}
		return PlusRec{TermID: stID, Zs: choices}, nil
	case withKind:
		choices := make(map[sym.ADT]TermRec, len(st.Spec.With))
		for _, ch := range st.Spec.With {
			choice, err := statesToTermRec(states, states[ch.Cont])
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Lab)] = choice
		}
		return WithRec{TermID: stID, Zs: choices}, nil
	default:
		panic(errUnexpectedKind(st.K))
	}
}

func statesFromTermRec(from string, r TermRec, dto *termRecDS) (string, error) {
	var fromID sql.NullString
	if len(from) > 0 {
		fromID = sql.NullString{String: from, Valid: true}
	}
	stID := r.Ident().String()
	switch root := r.(type) {
	case OneRec:
		st := stateDS{ID: stID, K: oneKind, FromID: fromID}
		dto.States = append(dto.States, st)
		return stID, nil
	case LinkRec:
		st := stateDS{
			ID:     stID,
			K:      linkKind,
			FromID: fromID,
			Spec: specDS{
				Link: sym.ConvertToString(root.TypeQN),
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
			ID:     stID,
			K:      tensorKind,
			FromID: fromID,
			Spec: specDS{
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
			ID:     stID,
			K:      lolliKind,
			FromID: fromID,
			Spec: specDS{
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
			ID:     stID,
			K:      plusKind,
			FromID: fromID,
			Spec:   specDS{Plus: choices},
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
			ID:     stID,
			K:      withKind,
			FromID: fromID,
			Spec:   specDS{With: choices},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	default:
		panic(ErrRecTypeUnexpected(r))
	}
}

func errUnexpectedKind(k termKind) error {
	return fmt.Errorf("unexpected kind %q", k)
}
