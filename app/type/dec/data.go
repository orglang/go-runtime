package dec

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
