package procexec

import "github.com/orglang/go-sdk/adt/procexec"

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/procdef:Msg.*
var (
	MsgToExecRef    func(procexec.ExecRef) (ExecRef, error)
	MsgFromExecRef  func(ExecRef) procexec.ExecRef
	MsgToExecSnap   func(procexec.ExecSnap) (ExecSnap, error)
	MsgFromExecSnap func(ExecSnap) procexec.ExecSnap
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/go-runtime/adt/identity:Convert.*
// goverter:extend orglang/go-runtime/adt/revnum:Convert.*
// goverter:extend orglang/go-runtime/adt/procdef:Data.*
// goverter:extend data.*
var (
	DataFromMod  func(ExecMod) (execModDS, error)
	DataToLiab   func(liabDS) (Liab, error)
	DataFromLiab func(Liab) liabDS
)
