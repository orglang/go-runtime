package state

import (
	"database/sql"
	"fmt"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type kind int

const (
	nonst kind = iota
	one
	link
	tensor
	lolli
	plus
	with
)

type RefData struct {
	ID string `db:"id" json:"id"`
	K  kind   `db:"kind" json:"kind"`
}

type rootData struct {
	ID     string
	States []stateData
}

type stateData struct {
	ID     string         `db:"id"`
	K      kind           `db:"kind"`
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
// goverter:extend DataToRef
// goverter:extend DataFromRef
var (
	DataToRefs    func([]*RefData) ([]Ref, error)
	DataFromRefs  func([]Ref) []*RefData
	DataToRoots   func([]*rootData) ([]Root, error)
	DataFromRoots func([]Root) []*rootData
)

func DataFromRef(ref Ref) *RefData {
	if ref == nil {
		return nil
	}
	rid := ref.Ident().String()
	switch ref.(type) {
	case OneRef, OneRoot:
		return &RefData{K: one, ID: rid}
	case LinkRef, LinkRoot:
		return &RefData{K: link, ID: rid}
	case TensorRef, TensorRoot:
		return &RefData{K: tensor, ID: rid}
	case LolliRef, LolliRoot:
		return &RefData{K: lolli, ID: rid}
	case PlusRef, PlusRoot:
		return &RefData{K: plus, ID: rid}
	case WithRef, WithRoot:
		return &RefData{K: with, ID: rid}
	default:
		panic(ErrRefTypeUnexpected(ref))
	}
}

func DataToRef(dto *RefData) (Ref, error) {
	if dto == nil {
		return nil, nil
	}
	rid, err := id.ConvertFromString(dto.ID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case one:
		return OneRef{rid}, nil
	case link:
		return LinkRef{rid}, nil
	case tensor:
		return TensorRef{rid}, nil
	case lolli:
		return LolliRef{rid}, nil
	case plus:
		return PlusRef{rid}, nil
	case with:
		return WithRef{rid}, nil
	default:
		panic(errUnexpectedKind(dto.K))
	}
}

func dataToRoot(dto *rootData) (Root, error) {
	states := make(map[string]stateData, len(dto.States))
	for _, dto := range dto.States {
		states[dto.ID] = dto
	}
	return statesToRoot(states, states[dto.ID])
}

func dataFromRoot(root Root) *rootData {
	if root == nil {
		return nil
	}
	dto := &rootData{
		ID:     root.Ident().String(),
		States: nil,
	}
	statesFromRoot("", root, dto)
	return dto
}

func statesToRoot(states map[string]stateData, st stateData) (Root, error) {
	stID, err := id.ConvertFromString(st.ID)
	if err != nil {
		return nil, err
	}
	switch st.K {
	case one:
		return OneRoot{ID: stID}, nil
	case link:
		roleQN, err := sym.ConvertFromString(st.Spec.Link)
		if err != nil {
			return nil, err
		}
		return LinkRoot{ID: stID, RoleQN: roleQN}, nil
	case tensor:
		b, err := statesToRoot(states, states[st.Spec.Tensor.Val])
		if err != nil {
			return nil, err
		}
		c, err := statesToRoot(states, states[st.Spec.Tensor.Cont])
		if err != nil {
			return nil, err
		}
		return TensorRoot{ID: stID, B: b, C: c}, nil
	case lolli:
		y, err := statesToRoot(states, states[st.Spec.Lolli.Val])
		if err != nil {
			return nil, err
		}
		z, err := statesToRoot(states, states[st.Spec.Lolli.Cont])
		if err != nil {
			return nil, err
		}
		return LolliRoot{ID: stID, Y: y, Z: z}, nil
	case plus:
		choices := make(map[sym.ADT]Root, len(st.Spec.Plus))
		for _, ch := range st.Spec.Plus {
			choice, err := statesToRoot(states, states[ch.Cont])
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Lab)] = choice
		}
		return PlusRoot{ID: stID, Choices: choices}, nil
	case with:
		choices := make(map[sym.ADT]Root, len(st.Spec.With))
		for _, ch := range st.Spec.With {
			choice, err := statesToRoot(states, states[ch.Cont])
			if err != nil {
				return nil, err
			}
			choices[sym.ADT(ch.Lab)] = choice
		}
		return WithRoot{ID: stID, Choices: choices}, nil
	default:
		panic(errUnexpectedKind(st.K))
	}
}

func statesFromRoot(from string, r Root, dto *rootData) (string, error) {
	var fromID sql.NullString
	if len(from) > 0 {
		fromID = sql.NullString{String: from, Valid: true}
	}
	stID := r.Ident().String()
	switch root := r.(type) {
	case OneRoot:
		st := stateData{ID: stID, K: one, FromID: fromID}
		dto.States = append(dto.States, st)
		return stID, nil
	case LinkRoot:
		st := stateData{
			ID:     stID,
			K:      link,
			FromID: fromID,
			Spec: specData{
				Link: sym.ConvertToString(root.RoleQN),
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case TensorRoot:
		val, err := statesFromRoot(stID, root.B, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromRoot(stID, root.C, dto)
		if err != nil {
			return "", err
		}
		st := stateData{
			ID:     stID,
			K:      tensor,
			FromID: fromID,
			Spec: specData{
				Tensor: &prodData{val, cont},
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case LolliRoot:
		val, err := statesFromRoot(stID, root.Y, dto)
		if err != nil {
			return "", err
		}
		cont, err := statesFromRoot(stID, root.Z, dto)
		if err != nil {
			return "", err
		}
		st := stateData{
			ID:     stID,
			K:      lolli,
			FromID: fromID,
			Spec: specData{
				Lolli: &prodData{val, cont},
			},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case PlusRoot:
		var choices []sumData
		for label, choice := range root.Choices {
			cont, err := statesFromRoot(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumData{string(label), cont})
		}
		st := stateData{
			ID:     stID,
			K:      plus,
			FromID: fromID,
			Spec:   specData{Plus: choices},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	case WithRoot:
		var choices []sumData
		for label, choice := range root.Choices {
			cont, err := statesFromRoot(stID, choice, dto)
			if err != nil {
				return "", err
			}
			choices = append(choices, sumData{string(label), cont})
		}
		st := stateData{
			ID:     stID,
			K:      with,
			FromID: fromID,
			Spec:   specData{With: choices},
		}
		dto.States = append(dto.States, st)
		return stID, nil
	default:
		panic(ErrRootTypeUnexpected(r))
	}
}

func errUnexpectedKind(k kind) error {
	return fmt.Errorf("unexpected kind %q", k)
}
