package typedef

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/typedef:Convert.*
var (
	ConvertRecToRef  func(TypeRec) TypeRef
	ConvertSnapToRef func(TypeSnap) TypeRef
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/revnum:Convert.*
// goverter:extend orglang/orglang/adt/typedef:Msg.*
// goverter:extend Msg.*
var (
	MsgFromTypeSpec  func(TypeSpec) TypeSpecME
	MsgToTypeSpec    func(TypeSpecME) (TypeSpec, error)
	MsgFromTypeRef   func(TypeRef) TypeRefME
	MsgToTypeRef     func(TypeRefME) (TypeRef, error)
	MsgFromTypeRefs  func([]TypeRef) []TypeRefME
	MsgToTypeRefs    func([]TypeRefME) ([]TypeRef, error)
	MsgFromTypeSnap  func(TypeSnap) TypeSnapME
	MsgToTypeSnap    func(TypeSnapME) (TypeSnap, error)
	MsgFromTypeSnaps func([]TypeSnap) []TypeSnapME
	MsgToTypeSnaps   func([]TypeSnapME) ([]TypeSnap, error)
	MsgFromTermRefs  func([]TermRef) []TermRefME
	MsgToTermRefs    func([]TermRefME) ([]TermRef, error)
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/revnum:Convert.*
// goverter:extend orglang/orglang/adt/typedef:Msg.*
var (
	ViewFromTypeRef  func(TypeRef) TypeRefVP
	ViewToTypeRef    func(TypeRefVP) (TypeRef, error)
	ViewFromTypeRefs func([]TypeRef) []TypeRefVP
	ViewToTypeRefs   func([]TypeRefVP) ([]TypeRef, error)
	ViewFromTypeSnap func(TypeSnap) TypeSnapVP
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend orglang/orglang/adt/identity:Convert.*
// goverter:extend orglang/orglang/adt/revnum:Convert.*
// goverter:extend orglang/orglang/adt/typedef:Data.*
// goverter:extend data.*
// goverter:extend DataToTermRef
// goverter:extend DataFromTermRef
var (
	DataToTypeRef     func(typeRefDS) (TypeRef, error)
	DataFromTypeRef   func(TypeRef) (typeRefDS, error)
	DataToTypeRefs    func([]typeRefDS) ([]TypeRef, error)
	DataFromTypeRefs  func([]TypeRef) ([]typeRefDS, error)
	DataToTypeRec     func(typeRecDS) (TypeRec, error)
	DataFromTypeRec   func(TypeRec) (typeRecDS, error)
	DataToTypeRecs    func([]typeRecDS) ([]TypeRec, error)
	DataFromTypeRecs  func([]TypeRec) ([]typeRecDS, error)
	DataToTermRefs    func([]*TermRefDS) ([]TermRef, error)
	DataFromTermRefs  func([]TermRef) []*TermRefDS
	DataToTermRoots   func([]*termRecDS) ([]TermRec, error)
	DataFromTermRoots func([]TermRec) []*termRecDS
)
