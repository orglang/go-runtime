package def

import (
	"database/sql"
	"fmt"
	id "smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type typeRefData struct {
	TypeID string `db:"role_id"`
	TypeRN int64  `db:"rev"`
	Title  string `db:"title"`
}

type typeRecData struct {
	TypeID string `db:"role_id"`
	Title  string `db:"title"`
	TermID string `db:"state_id"`
	TypeRN int64  `db:"rev"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/lib/rn:Convert.*
// goverter:extend smecalculus/rolevod/app/type/def:Data.*
var (
	DataToTypeRef    func(typeRefData) (TypeRef, error)
	DataFromTypeRef  func(TypeRef) (typeRefData, error)
	DataToTypeRefs   func([]typeRefData) ([]TypeRef, error)
	DataFromTypeRefs func([]TypeRef) ([]typeRefData, error)
	DataToTypeRec    func(typeRecData) (TypeRec, error)
	DataFromTypeRec  func(TypeRec) (typeRecData, error)
	DataToTypeRecs   func([]typeRecData) ([]TypeRec, error)
	DataFromTypeRecs func([]TypeRec) ([]typeRecData, error)
)

type termKind int

const (
	nonterm termKind = iota
	oneKind
	linkKind
	tensorKind
	lolliKind
	plusKind
	withKind
)

type TermRefData struct {
	ID string   `db:"id" json:"id"`
	K  termKind `db:"kind" json:"kind"`
}

type termRecData struct {
	ID     string
	States []stateData
}

type stateData struct {
	ID     string         `db:"id"`
	K      termKind       `db:"kind"`
	FromID sql.NullString `db:"from_id"`
	Spec   specData       `db:"spec"`
}

type specData struct {
	Link   string    `json:"link,omitempty"`
	Tensor *prodData `json:"tensor,omitempty"`
	Lolli  *prodData `json:"lolli,omitempty"`
	Plus   []sumData `json:"plus,omitempty"`
	With   []sumData `json:"with,omitempty"`
}

type prodData struct {
	Val  string `json:"on"`
	Cont string `json:"to"`
}

type sumData struct {
	Lab  string `json:"on"`
	Cont string `json:"to"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend data.*
// goverter:extend DataToTermRef
// goverter:extend DataFromTermRef
var (
	DataToTermRefs    func([]*TermRefData) ([]TermRef, error)
	DataFromTermRefs  func([]TermRef) []*TermRefData
	DataToTermRoots   func([]*termRecData) ([]TermRec, error)
	DataFromTermRoots func([]TermRec) []*termRecData
)

func DataFromTermRef(ref TermRef) *TermRefData {
	if ref == nil {
		return nil
	}
	rid := ref.Ident().String()
	switch ref.(type) {
	case OneRef, OneRec:
		return &TermRefData{K: oneKind, ID: rid}
	case LinkRef, LinkRec:
		return &TermRefData{K: linkKind, ID: rid}
	case TensorRef, TensorRec:
		return &TermRefData{K: tensorKind, ID: rid}
	case LolliRef, LolliRec:
		return &TermRefData{K: lolliKind, ID: rid}
	case PlusRef, PlusRec:
		return &TermRefData{K: plusKind, ID: rid}
	case WithRef, WithRec:
		return &TermRefData{K: withKind, ID: rid}
	default:
		panic(ErrRefTypeUnexpected(ref))
	}
}

func DataToTermRef(dto *TermRefData) (TermRef, error) {
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

func dataToTermRec(dto *termRecData) (TermRec, error) {
	states := make(map[string]stateData, len(dto.States))
	for _, dto := range dto.States {
		states[dto.ID] = dto
	}
	return statesToTermRec(states, states[dto.ID])
}

func dataFromTermRec(root TermRec) *termRecData {
	if root == nil {
		return nil
	}
	dto := &termRecData{
		ID:     root.Ident().String(),
		States: nil,
	}
	statesFromTermRec("", root, dto)
	return dto
}

func statesToTermRec(states map[string]stateData, st stateData) (TermRec, error) {
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
		return TensorRec{TermID: stID, B: b, C: c}, nil
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
		return PlusRec{TermID: stID, Choices: choices}, nil
	case withKind:
		choices := make(map[sym.ADT]TermRec, len(st.Spec.With))
		for _, ch := range st.Spec.With {
			choice, err := statesToTermRec(states, states[ch.Cont])
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Lab)] = choice
		}
		return WithRec{TermID: stID, Choices: choices}, nil
	default:
		panic(errUnexpectedKind(st.K))
	}
}

func statesFromTermRec(from string, r TermRec, dto *termRecData) (string, error) {
	var fromID sql.NullString
	if len(from) > 0 {
		fromID = sql.NullString{String: from, Valid: true}
	}
	stID := r.Ident().String()
	switch root := r.(type) {
	case OneRec:
		st := stateData{ID: stID, K: oneKind, FromID: fromID}
		dto.States = append(dto.States, st)
		return stID, nil
	case LinkRec:
		st := stateData{
			ID:     stID,
			K:      linkKind,
			FromID: fromID,
			Spec: specData{
				Link: sym.ConvertToString(root.TypeQN),
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case TensorRec:
		val, err := statesFromTermRec(stID, root.B, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromTermRec(stID, root.C, dto)
		if err != nil {
			return "", err
		}
		st := stateData{
			ID:     stID,
			K:      tensorKind,
			FromID: fromID,
			Spec: specData{
				Tensor: &prodData{val, cont},
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
		st := stateData{
			ID:     stID,
			K:      lolliKind,
			FromID: fromID,
			Spec: specData{
				Lolli: &prodData{val, cont},
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case PlusRec:
		var choices []sumData
		for label, choice := range root.Choices {
			cont, err := statesFromTermRec(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumData{string(label), cont})
		}
		st := stateData{
			ID:     stID,
			K:      plusKind,
			FromID: fromID,
			Spec:   specData{Plus: choices},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case WithRec:
		var choices []sumData
		for label, choice := range root.Choices {
			cont, err := statesFromTermRec(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumData{string(label), cont})
		}
		st := stateData{
			ID:     stID,
			K:      withKind,
			FromID: fromID,
			Spec:   specData{With: choices},
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
