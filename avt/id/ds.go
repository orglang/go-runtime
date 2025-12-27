package id

import (
	"database/sql"
)

func ConvertToNullString(id ADT) sql.NullString {
	if id.IsEmpty() {
		return sql.NullString{}
	}
	return sql.NullString{String: id.String(), Valid: true}
}

func ConvertFromNullString(dto sql.NullString) (ADT, error) {
	if dto.Valid {
		return ConvertFromString(dto.String)
	}
	return Empty(), nil
}

func ConvertPtrToNullString(id *ADT) sql.NullString {
	if id == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: id.String(), Valid: true}
}
