package def

import (
	"context"
	"fmt"
	"log/slog"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/pol"
	"orglang/orglang/avt/rn"
	"orglang/orglang/avt/sym"

	"orglang/orglang/aet/alias"
)

type API interface {
	Incept(sym.ADT) (TypeRef, error)
	Create(TypeSpec) (TypeSnap, error)
	Modify(TypeSnap) (TypeSnap, error)
	Retrieve(id.ADT) (TypeSnap, error)
	retrieveSnap(TypeRec) (TypeSnap, error)
	RetreiveRefs() ([]TypeRef, error)
}

type TypeSpec struct {
	TypeNS sym.ADT
	TypeSN sym.ADT
	TypeTS TermSpec
}

type TypeRef struct {
	TypeID id.ADT
	Title  string
	TypeRN rn.ADT
}

// aka TpDef
type TypeRec struct {
	TypeID id.ADT
	Title  string
	TermID id.ADT
	TypeRN rn.ADT
}

type TypeSnap struct {
	TypeID id.ADT
	Title  string
	TypeQN sym.ADT
	TypeTS TermSpec
	TypeRN rn.ADT
}

type TermSpec interface {
	spec()
}

type OneSpec struct{}

func (OneSpec) spec() {}

// aka TpName
type LinkSpec struct {
	TypeQN sym.ADT
}

func (LinkSpec) spec() {}

type TensorSpec struct {
	Y TermSpec // val to send
	Z TermSpec // cont
}

func (TensorSpec) spec() {}

type LolliSpec struct {
	Y TermSpec // val to receive
	Z TermSpec // cont
}

func (LolliSpec) spec() {}

// aka Internal Choice
type PlusSpec struct {
	Zs map[sym.ADT]TermSpec // conts
}

func (PlusSpec) spec() {}

// aka External Choice
type WithSpec struct {
	Zs map[sym.ADT]TermSpec // conts
}

func (WithSpec) spec() {}

type UpSpec struct {
	Z TermSpec // cont
}

func (UpSpec) spec() {}

type DownSpec struct {
	Z TermSpec // cont
}

func (DownSpec) spec() {}

type XactSpec struct {
	Zs map[sym.ADT]TermSpec // conts
}

func (XactSpec) spec() {}

type TermRef interface {
	id.Identifiable
}

type OneRef struct {
	TermID id.ADT
}

func (r OneRef) Ident() id.ADT { return r.TermID }

type LinkRef struct {
	TermID id.ADT
}

func (r LinkRef) Ident() id.ADT { return r.TermID }

type PlusRef struct {
	TermID id.ADT
}

func (r PlusRef) Ident() id.ADT { return r.TermID }

type WithRef struct {
	TermID id.ADT
}

func (r WithRef) Ident() id.ADT { return r.TermID }

type TensorRef struct {
	TermID id.ADT
}

func (r TensorRef) Ident() id.ADT { return r.TermID }

type LolliRef struct {
	TermID id.ADT
}

func (r LolliRef) Ident() id.ADT { return r.TermID }

type UpRef struct {
	TermID id.ADT
}

func (r UpRef) Ident() id.ADT { return r.TermID }

type DownRef struct {
	TermID id.ADT
}

func (r DownRef) Ident() id.ADT { return r.TermID }

// aka Stype
type TermRec interface {
	id.Identifiable
	pol.Polarizable
}

type ProdRec interface {
	Next() id.ADT
}

type SumRec interface {
	Next(sym.ADT) id.ADT
}

type OneRec struct {
	TermID id.ADT
}

func (OneRec) spec() {}

func (r OneRec) Ident() id.ADT { return r.TermID }

func (OneRec) Pol() pol.ADT { return pol.Pos }

// aka TpName
type LinkRec struct {
	TermID id.ADT
	TypeQN sym.ADT
}

func (LinkRec) spec() {}

func (r LinkRec) Ident() id.ADT { return r.TermID }

func (LinkRec) Pol() pol.ADT { return pol.Zero }

// aka Internal Choice
type PlusRec struct {
	TermID id.ADT
	Zs     map[sym.ADT]TermRec
}

func (PlusRec) spec() {}

func (r PlusRec) Ident() id.ADT { return r.TermID }

func (r PlusRec) Next(l sym.ADT) id.ADT { return r.Zs[l].Ident() }

func (PlusRec) Pol() pol.ADT { return pol.Pos }

// aka External Choice
type WithRec struct {
	TermID id.ADT
	Zs     map[sym.ADT]TermRec
}

func (WithRec) spec() {}

func (r WithRec) Ident() id.ADT { return r.TermID }

func (r WithRec) Next(l sym.ADT) id.ADT { return r.Zs[l].Ident() }

func (WithRec) Pol() pol.ADT { return pol.Neg }

type TensorRec struct {
	TermID id.ADT
	Y      TermRec
	Z      TermRec
}

func (TensorRec) spec() {}

func (r TensorRec) Ident() id.ADT { return r.TermID }

func (r TensorRec) Next() id.ADT { return r.Z.Ident() }

func (TensorRec) Pol() pol.ADT { return pol.Pos }

type LolliRec struct {
	TermID id.ADT
	Y      TermRec
	Z      TermRec
}

func (LolliRec) spec() {}

func (r LolliRec) Ident() id.ADT { return r.TermID }

func (r LolliRec) Next() id.ADT { return r.Z.Ident() }

func (LolliRec) Pol() pol.ADT { return pol.Neg }

type UpRec struct {
	TermID id.ADT
	Z      TermRec
}

func (UpRec) spec() {}

func (r UpRec) Ident() id.ADT { return r.TermID }

func (UpRec) Pol() pol.ADT { return pol.Zero }

type DownRec struct {
	TermID id.ADT
	Z      TermRec
}

func (DownRec) spec() {}

func (r DownRec) Ident() id.ADT { return r.TermID }

func (DownRec) Pol() pol.ADT { return pol.Zero }

type Context struct {
	Assets map[sym.ADT]TermRec
	Liabs  map[sym.ADT]TermRec
}

type service struct {
	types    Repo
	aliases  alias.Repo
	operator sd.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	types Repo,
	aliases alias.Repo,
	operator sd.Operator,
	l *slog.Logger,
) *service {
	return &service{types, aliases, operator, l}
}

func (s *service) Incept(qn sym.ADT) (_ TypeRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("roleQN", qn)
	s.log.Debug("inception started", qnAttr)
	newAlias := alias.Root{QN: qn, ID: id.New(), RN: rn.Initial()}
	newType := TypeRec{TypeID: newAlias.ID, TypeRN: newAlias.RN, Title: newAlias.QN.SN()}
	s.operator.Explicit(ctx, func(ds sd.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.types.InsertType(ds, newType)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return TypeRef{}, err
	}
	s.log.Debug("inception succeeded", qnAttr, slog.Any("roleID", newType.TypeID))
	return ConvertRecToRef(newType), nil
}

func (s *service) Create(spec TypeSpec) (_ TypeSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("typeQN", spec.TypeSN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newAlias := alias.Root{QN: spec.TypeSN, ID: id.New(), RN: rn.Initial()}
	newTerm := ConvertSpecToRec(spec.TypeTS)
	newType := TypeRec{
		TypeID: newAlias.ID,
		TypeRN: newAlias.RN,
		Title:  newAlias.QN.SN(),
		TermID: newTerm.Ident(),
	}
	s.operator.Explicit(ctx, func(ds sd.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.types.InsertTerm(ds, newTerm)
		if err != nil {
			return err
		}
		err = s.types.InsertType(ds, newType)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return TypeSnap{}, err
	}
	s.log.Debug("creation succeeded", qnAttr, slog.Any("typeID", newType.TypeID))
	return TypeSnap{
		TypeID: newType.TypeID,
		TypeRN: newType.TypeRN,
		Title:  newType.Title,
		TypeQN: newAlias.QN,
		TypeTS: ConvertRecToSpec(newTerm),
	}, nil
}

func (s *service) Modify(snap TypeSnap) (_ TypeSnap, err error) {
	ctx := context.Background()
	idAttr := slog.Any("typeID", snap.TypeID)
	s.log.Debug("modification started", idAttr)
	var rec TypeRec
	s.operator.Implicit(ctx, func(ds sd.Source) error {
		rec, err = s.types.SelectTypeRecByID(ds, snap.TypeID)
		return err
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, err
	}
	if snap.TypeRN != rec.TypeRN {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, errConcurrentModification(snap.TypeRN, rec.TypeRN)
	} else {
		snap.TypeRN = rn.Next(snap.TypeRN)
	}
	curSnap, err := s.retrieveSnap(rec)
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, err
	}
	s.operator.Explicit(ctx, func(ds sd.Source) error {
		if CheckSpec(snap.TypeTS, curSnap.TypeTS) != nil {
			newTerm := ConvertSpecToRec(snap.TypeTS)
			err = s.types.InsertTerm(ds, newTerm)
			if err != nil {
				return err
			}
			rec.TermID = newTerm.Ident()
			rec.TypeRN = snap.TypeRN
		}
		if rec.TypeRN == snap.TypeRN {
			err = s.types.UpdateType(ds, rec)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, err
	}
	s.log.Debug("modification succeeded", idAttr)
	return snap, nil
}

func (s *service) Retrieve(recID id.ADT) (_ TypeSnap, err error) {
	ctx := context.Background()
	var root TypeRec
	s.operator.Implicit(ctx, func(ds sd.Source) error {
		root, err = s.types.SelectTypeRecByID(ds, recID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", recID))
		return TypeSnap{}, err
	}
	return s.retrieveSnap(root)
}

func (s *service) retrieveSnap(typeRec TypeRec) (_ TypeSnap, err error) {
	ctx := context.Background()
	var termRec TermRec
	s.operator.Implicit(ctx, func(ds sd.Source) error {
		termRec, err = s.types.SelectTermRecByID(ds, typeRec.TermID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", typeRec.TypeID))
		return TypeSnap{}, err
	}
	return TypeSnap{
		TypeID: typeRec.TypeID,
		TypeRN: typeRec.TypeRN,
		Title:  typeRec.Title,
		TypeTS: ConvertRecToSpec(termRec),
	}, nil
}

func (s *service) RetreiveRefs() (refs []TypeRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds sd.Source) error {
		refs, err = s.types.SelectTypeRefs(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(recs []TypeRec) []id.ADT {
	termIDs := []id.ADT{}
	for _, r := range recs {
		termIDs = append(termIDs, r.TermID)
	}
	return termIDs
}

func ErrSymMissingInEnv(want sym.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func errConcurrentModification(got rn.ADT, want rn.ADT) error {
	return fmt.Errorf("entity concurrent modification: want revision %v, got revision %v", want, got)
}

func errOptimisticUpdate(got rn.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}

func CheckRef(got, want id.ADT) error {
	if got != want {
		return fmt.Errorf("type mismatch: want %+v, got %+v", want, got)
	}
	return nil
}

// aka eqtp
func CheckSpec(got, want TermSpec) error {
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
		err := CheckSpec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckSpec(gotSt.Z, wantSt.Z)
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
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
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
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
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
func CheckRec(got, want TermRec) error {
	switch wantSt := want.(type) {
	case OneRec:
		_, ok := got.(OneRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		return nil
	case TensorRec:
		gotSt, ok := got.(TensorRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		err := CheckRec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckRec(gotSt.Z, wantSt.Z)
	case LolliRec:
		gotSt, ok := got.(LolliRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		err := CheckRec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckRec(gotSt.Z, wantSt.Z)
	case PlusRec:
		gotSt, ok := got.(PlusRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckRec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	case WithRec:
		gotSt, ok := got.(WithRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckRec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		panic(ErrRecTypeUnexpected(want))
	}
}

func ErrSpecTypeUnexpected(got TermSpec) error {
	return fmt.Errorf("spec type unexpected: %T", got)
}

func ErrRefTypeUnexpected(got TermRef) error {
	return fmt.Errorf("ref type unexpected: %T", got)
}

func ErrDoesNotExist(want id.ADT) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrMissingInEnv(want id.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func ErrMissingInCfg(want id.ADT) error {
	return fmt.Errorf("root missing in cfg: %v", want)
}

func ErrMissingInCtx(want sym.ADT) error {
	return fmt.Errorf("root missing in ctx: %v", want)
}

func ErrRecTypeUnexpected(got TermRec) error {
	return fmt.Errorf("rec type unexpected: %T", got)
}

func ErrSpecTypeMismatch(got, want TermSpec) error {
	return fmt.Errorf("spec type mismatch: want %T, got %T", want, got)
}

func ErrSnapTypeMismatch(got, want TermRec) error {
	return fmt.Errorf("root type mismatch: want %T, got %T", want, got)
}

func ErrPolarityUnexpected(got TermRec) error {
	return fmt.Errorf("root polarity unexpected: %v", got.Pol())
}

func ErrPolarityMismatch(a, b TermRec) error {
	return fmt.Errorf("root polarity mismatch: %v != %v", a.Pol(), b.Pol())
}
