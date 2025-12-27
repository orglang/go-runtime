package elab

import (
	"errors"
	"fmt"
	"maps"
	a "orglang/orglang/proto/rast2/ast"
	tc "orglang/orglang/proto/rast2/typecheck"
	"slices"
)

func ElabTps(env a.Environment, dcls map[string]a.Decl) error {
	for _, d := range maps.Clone(dcls) {
		switch dcl := d.(type) {
		case a.TpDef:
			if !tc.Contractive(dcl.A) {
				return ErrTypeNotContractive(dcl.A)
			}
			err := tc.EsyncTp(env, dcl.A)
			if err != nil {
				return err
			}
			env[dcl.V] = dcl
		case a.ExpDecDef:
			delta := dcl.Ctx.Linear
			if dups(append(delta, dcl.Zc)) {
				return ErrDuplicateVariable
			}
			err := tc.EsyncTp(env, dcl.Zc.A)
			if err != nil {
				return err
			}
			env[dcl.F] = dcl
		default:
			continue
		}
	}
	return nil
}

func ElabExps(env a.Environment, dcls map[string]a.Decl) error {
	for _, d := range maps.Clone(dcls) {
		switch dcl := d.(type) {
		case a.TpDef:
			// already checked validity during first pass
			env[dcl.V] = dcl
		case a.ExpDecDef:
			err := tc.CheckExp(env, dcl.Ctx, dcl.P, dcl.Zc)
			if err != nil {
				return err
			}
			env[dcl.F] = dcl
		case a.Exec:
			dec, ok := env[dcl.F].(a.ExpDecDef)
			if !ok {
				return ErrProcessUndefined(dcl.F)
			}
			if len(dec.Ctx.Ordered) > 0 {
				return ErrProcessNonEmptyContext(dcl.F)
			}
			env[dcl.F] = dcl
		default:
			panic(ErrUnexpectedDecl)
		}
	}
	return nil
}

func ElabDecls(env a.Environment, dcls map[string]a.Decl) error {
	// first pass: check validity of types and create internal names
	err1 := ElabTps(env, dcls)
	if err1 != nil {
		return err1
	}
	// second pass: perform reconstruction and type checking
	err2 := ElabExps(env, env)
	if err2 != nil {
		return err2
	}
	return nil
}

func dups(delta []a.ChanTp) bool {
	var xs []string
	for _, d := range delta {
		xs = append(xs, d.X.V)
	}
	slices.Sort(xs)
	for i, x := range xs {
		if slices.Contains(xs[i:], x) {
			return true
		}
	}
	return false
}

var (
	ErrElabImpossible    = errors.New("elab impossible")
	ErrDuplicateVariable = errors.New("duplicate variable in process declaration")
	ErrUnexpectedDecl    = errors.New("unexpected decl")
)

func ErrProcessUndefined(f string) error {
	return fmt.Errorf("process %q udefined", f)
}

func ErrProcessNonEmptyContext(f string) error {
	return fmt.Errorf("process %q has a non-empty context, cannot be executed", f)
}

func ErrTypeNotContractive(a a.Stype) error {
	return fmt.Errorf("type not contractive: %v", a)
}
