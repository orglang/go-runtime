package xact

import (
	"fmt"
	"smecalculus/rolevod/lib/sym"
)

type Sem interface {
	sem()
}

type MsgSem struct {
}

func (s MsgSem) sem() {}

type SvcSem struct {
}

func (s SvcSem) sem() {}

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
	ViaPH sym.ADT
}

func (s AcqureSpec) ConnPH() sym.ADT { return s.ViaPH }

type AcceptSpec struct {
	ViaPH sym.ADT
}

func (s AcceptSpec) ConnPH() sym.ADT { return s.ViaPH }

type DetachSpec struct {
	ViaPH sym.ADT
}

func (s DetachSpec) ConnPH() sym.ADT { return s.ViaPH }

type ReleaseSpec struct {
	ViaPH sym.ADT
}

func (s ReleaseSpec) ConnPH() sym.ADT { return s.ViaPH }

func ErrTermTypeUnexpected(got TermSpec) error {
	return fmt.Errorf("term type unexpected: %T", got)
}
