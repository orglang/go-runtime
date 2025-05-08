package xact

import (
	"fmt"

	"smecalculus/rolevod/lib/sym"
)

type SemRec interface {
	sem()
}

type MsgRec struct {
}

func (s MsgRec) sem() {}

type SvcRec struct {
}

func (s SvcRec) sem() {}

type TermSpec interface {
	ConnPH() sym.ADT
}

// микс LabSpec и балкового SendSpec
type CallSpec struct {
	MainPH sym.ADT
	X      sym.ADT
	SigSN  sym.ADT
	Ys     []sym.ADT
	Cont   TermSpec
}

func (s CallSpec) ConnPH() sym.ADT { return s.MainPH }

// микс CaseSpec и балкового RecvSpec
type SpawnSpec struct {
	MainPH sym.ADT
}

func (s SpawnSpec) ConnPH() sym.ADT { return s.MainPH }

type AcqureSpec struct {
	MainPH sym.ADT
}

func (s AcqureSpec) ConnPH() sym.ADT { return s.MainPH }

type AcceptSpec struct {
	MainPH sym.ADT
}

func (s AcceptSpec) ConnPH() sym.ADT { return s.MainPH }

type DetachSpec struct {
	MainPH sym.ADT
}

func (s DetachSpec) ConnPH() sym.ADT { return s.MainPH }

type ReleaseSpec struct {
	MainPH sym.ADT
}

func (s ReleaseSpec) ConnPH() sym.ADT { return s.MainPH }

func ErrTermTypeUnexpected(got TermSpec) error {
	return fmt.Errorf("term type unexpected: %T", got)
}
