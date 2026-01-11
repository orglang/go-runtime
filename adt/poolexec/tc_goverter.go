package poolexec

import (
	"orglang/orglang/adt/procexec"
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
var (
	ConvertRecToRef func(ExecRec) ExecRef
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/procdef:Msg.*
var (
	MsgToExecSpec   func(ExecSpecME) (ExecSpec, error)
	MsgFromExecSpec func(ExecSpec) ExecSpecME
	MsgToExecRef    func(ExecRefME) (ExecRef, error)
	MsgFromExecRef  func(ExecRef) ExecRefME
	MsgToExecSnap   func(ExecSnapME) (ExecSnap, error)
	MsgFromExecSnap func(ExecSnap) ExecSnapME
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
var (
	DataToExecRef    func(execRefDS) (ExecRef, error)
	DataFromExecRef  func(ExecRef) execRefDS
	DataToExecRefs   func([]execRefDS) ([]ExecRef, error)
	DataFromExecRefs func([]ExecRef) []execRefDS
	DataToExecRec    func(execRecDS) (ExecRec, error)
	DataFromExecRec  func(ExecRec) execRecDS
	DataToLiab       func(liabDS) (procexec.Liab, error)
	DataFromLiab     func(procexec.Liab) liabDS
	DataToExecSnap   func(execSnapDS) (ExecSnap, error)
	DataFromExecSnap func(ExecSnap) execSnapDS
	DataToEPs        func([]epDS) ([]procexec.EP, error)
)
