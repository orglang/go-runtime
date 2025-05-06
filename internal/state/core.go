package state

import (
	"fmt"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/pol"
	"smecalculus/rolevod/lib/sym"
)

// for external readability
type ID = id.ADT

type Spec interface {
	spec()
}

type OneSpec struct{}

func (OneSpec) spec() {}

// aka TpName
type LinkSpec struct {
	RoleQN sym.ADT
}

func (LinkSpec) spec() {}

type TensorSpec struct {
	B Spec
	C Spec
}

func (TensorSpec) spec() {}

type LolliSpec struct {
	Y Spec
	Z Spec
}

func (LolliSpec) spec() {}

// aka Internal Choice
type PlusSpec struct {
	Choices map[sym.ADT]Spec
}

func (PlusSpec) spec() {}

// aka External Choice
type WithSpec struct {
	Choices map[sym.ADT]Spec
}

func (WithSpec) spec() {}

type UpSpec struct {
	X Spec
}

func (UpSpec) spec() {}

type DownSpec struct {
	X Spec
}

func (DownSpec) spec() {}

type Ref interface {
	id.Identifiable
}

type OneRef struct {
	ID id.ADT
}

func (r OneRef) Ident() id.ADT { return r.ID }

type LinkRef struct {
	ID id.ADT
}

func (r LinkRef) Ident() id.ADT { return r.ID }

type PlusRef struct {
	ID id.ADT
}

func (r PlusRef) Ident() id.ADT { return r.ID }

type WithRef struct {
	ID id.ADT
}

func (r WithRef) Ident() id.ADT { return r.ID }

type TensorRef struct {
	ID id.ADT
}

func (r TensorRef) Ident() id.ADT { return r.ID }

type LolliRef struct {
	ID id.ADT
}

func (r LolliRef) Ident() id.ADT { return r.ID }

type UpRef struct {
	ID id.ADT
}

func (r UpRef) Ident() id.ADT { return r.ID }

type DownRef struct {
	ID id.ADT
}

func (r DownRef) Ident() id.ADT { return r.ID }

// aka Stype
type Root interface {
	id.Identifiable
	pol.Polarizable
}

type Prod interface {
	Next() id.ADT
}

type Sum interface {
	Next(sym.ADT) id.ADT
}

type OneRoot struct {
	ID id.ADT
}

// type EP struct {

// }

func (OneRoot) spec() {}

func (r OneRoot) Ident() id.ADT { return r.ID }

func (r OneRoot) Pol() pol.ADT { return pol.Pos }

// aka TpName
type LinkRoot struct {
	ID     id.ADT
	RoleQN sym.ADT
}

func (LinkRoot) spec() {}

func (r LinkRoot) Ident() id.ADT { return r.ID }

func (LinkRoot) Pol() pol.ADT { return pol.Zero }

// aka Internal Choice
type PlusRoot struct {
	ID      id.ADT
	Choices map[sym.ADT]Root
}

func (PlusRoot) spec() {}

func (r PlusRoot) Ident() id.ADT { return r.ID }

func (r PlusRoot) Next(l sym.ADT) id.ADT { return r.Choices[l].Ident() }

func (PlusRoot) Pol() pol.ADT { return pol.Pos }

// aka External Choice
type WithRoot struct {
	ID      id.ADT
	Choices map[sym.ADT]Root
}

func (WithRoot) spec() {}

func (r WithRoot) Ident() id.ADT { return r.ID }

func (r WithRoot) Next(l sym.ADT) id.ADT { return r.Choices[l].Ident() }

func (WithRoot) Pol() pol.ADT { return pol.Neg }

type TensorRoot struct {
	ID id.ADT
	B  Root // value
	C  Root // cont
}

func (TensorRoot) spec() {}

func (r TensorRoot) Ident() id.ADT { return r.ID }

func (r TensorRoot) Next() id.ADT { return r.C.Ident() }

func (TensorRoot) Pol() pol.ADT { return pol.Pos }

type LolliRoot struct {
	ID id.ADT
	Y  Root // value
	Z  Root // cont
}

func (LolliRoot) spec() {}

func (r LolliRoot) Ident() id.ADT { return r.ID }

func (r LolliRoot) Next() id.ADT { return r.Z.Ident() }

func (LolliRoot) Pol() pol.ADT { return pol.Neg }

type UpRoot struct {
	ID id.ADT
	A  Root
}

func (UpRoot) spec() {}

func (r UpRoot) Ident() id.ADT { return r.ID }

func (r UpRoot) Pol() pol.ADT { return pol.Zero }

type DownRoot struct {
	ID id.ADT
	A  Root
}

func (DownRoot) spec() {}

func (r DownRoot) Ident() id.ADT { return r.ID }

func (r DownRoot) Pol() pol.ADT { return pol.Zero }

type Context struct {
	Assets map[sym.ADT]Root
	Liabs  map[sym.ADT]Root
}

// Endpoint aka ChanTp
type EP struct {
	Z sym.ADT
	C Root
}

type Repo interface {
	Insert(data.Source, Root) error
	SelectAll(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Root, error)
	SelectByIDs(data.Source, []id.ADT) ([]Root, error)
	SelectEnv(data.Source, []id.ADT) (map[id.ADT]Root, error)
}

func ConvertSpecToRoot(s Spec) Root {
	if s == nil {
		return nil
	}
	switch spec := s.(type) {
	case OneSpec:
		return OneRoot{ID: id.New()}
	case LinkSpec:
		return LinkRoot{ID: id.New(), RoleQN: spec.RoleQN}
	case TensorSpec:
		return TensorRoot{
			ID: id.New(),
			B:  ConvertSpecToRoot(spec.B),
			C:  ConvertSpecToRoot(spec.C),
		}
	case LolliSpec:
		return LolliRoot{
			ID: id.New(),
			Y:  ConvertSpecToRoot(spec.Y),
			Z:  ConvertSpecToRoot(spec.Z),
		}
	case WithSpec:
		choices := make(map[sym.ADT]Root, len(spec.Choices))
		for lab, st := range spec.Choices {
			choices[lab] = ConvertSpecToRoot(st)
		}
		return WithRoot{ID: id.New(), Choices: choices}
	case PlusSpec:
		choices := make(map[sym.ADT]Root, len(spec.Choices))
		for lab, st := range spec.Choices {
			choices[lab] = ConvertSpecToRoot(st)
		}
		return PlusRoot{ID: id.New(), Choices: choices}
	default:
		panic(ErrSpecTypeUnexpected(spec))
	}
}

func ConvertRootToSpec(r Root) Spec {
	if r == nil {
		return nil
	}
	switch root := r.(type) {
	case OneRoot:
		return OneSpec{}
	case LinkRoot:
		return LinkSpec{RoleQN: root.RoleQN}
	case TensorRoot:
		return TensorSpec{
			B: ConvertRootToSpec(root.B),
			C: ConvertRootToSpec(root.C),
		}
	case LolliRoot:
		return LolliSpec{
			Y: ConvertRootToSpec(root.Y),
			Z: ConvertRootToSpec(root.Z),
		}
	case WithRoot:
		choices := make(map[sym.ADT]Spec, len(root.Choices))
		for lab, st := range root.Choices {
			choices[lab] = ConvertRootToSpec(st)
		}
		return WithSpec{Choices: choices}
	case PlusRoot:
		choices := make(map[sym.ADT]Spec, len(root.Choices))
		for lab, st := range root.Choices {
			choices[lab] = ConvertRootToSpec(st)
		}
		return PlusSpec{Choices: choices}
	default:
		panic(ErrRootTypeUnexpected(root))
	}
}

func CheckRef(got, want ID) error {
	if got != want {
		return fmt.Errorf("state mismatch: want %+v, got %+v", want, got)
	}
	return nil
}

// aka eqtp
func CheckSpec(got, want Spec) error {
	switch wantSt := want.(type) {
	case OneSpec:
		_, ok := got.(OneSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		return nil
	case TensorSpec:
		gotSt, ok := got.(TensorSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		err := CheckSpec(gotSt.B, wantSt.B)
		if err != nil {
			return err
		}
		return CheckSpec(gotSt.C, wantSt.C)
	case LolliSpec:
		gotSt, ok := got.(LolliSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		err := CheckSpec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckSpec(gotSt.Z, wantSt.Z)
	case PlusSpec:
		gotSt, ok := got.(PlusSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckSpec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	case WithSpec:
		gotSt, ok := got.(WithSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckSpec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		panic(ErrSpecTypeUnexpected(want))
	}
}

// aka eqtp
func CheckRoot(got, want Root) error {
	switch wantSt := want.(type) {
	case OneRoot:
		_, ok := got.(OneRoot)
		if !ok {
			return ErrRootTypeMismatch(got, want)
		}
		return nil
	case TensorRoot:
		gotSt, ok := got.(TensorRoot)
		if !ok {
			return ErrRootTypeMismatch(got, want)
		}
		err := CheckRoot(gotSt.B, wantSt.B)
		if err != nil {
			return err
		}
		return CheckRoot(gotSt.C, wantSt.C)
	case LolliRoot:
		gotSt, ok := got.(LolliRoot)
		if !ok {
			return ErrRootTypeMismatch(got, want)
		}
		err := CheckRoot(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckRoot(gotSt.Z, wantSt.Z)
	case PlusRoot:
		gotSt, ok := got.(PlusRoot)
		if !ok {
			return ErrRootTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckRoot(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	case WithRoot:
		gotSt, ok := got.(WithRoot)
		if !ok {
			return ErrRootTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckRoot(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		panic(ErrRootTypeUnexpected(want))
	}
}

func ErrSpecTypeUnexpected(got Spec) error {
	return fmt.Errorf("spec type unexpected: %T", got)
}

func ErrRefTypeUnexpected(got Ref) error {
	return fmt.Errorf("ref type unexpected: %T", got)
}

func ErrDoesNotExist(want ID) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrMissingInEnv(want ID) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func ErrMissingInCfg(want ID) error {
	return fmt.Errorf("root missing in cfg: %v", want)
}

func ErrMissingInCtx(want sym.ADT) error {
	return fmt.Errorf("root missing in ctx: %v", want)
}

func ErrRootTypeUnexpected(got Root) error {
	return fmt.Errorf("root type unexpected: %T", got)
}

func ErrSpecTypeMismatch(got, want Spec) error {
	return fmt.Errorf("spec type mismatch: want %T, got %T", want, got)
}

func ErrRootTypeMismatch(got, want Root) error {
	return fmt.Errorf("root type mismatch: want %T, got %T", want, got)
}
