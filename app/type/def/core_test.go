package def

import (
	"context"
	"log/slog"
	"testing"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"
)

func TestKinshipEstalish(t *testing.T) {
	newService(&roleRepoStub{}, &aliasRepoStub{}, &operatorStub{}, slog.Default())
}

type roleRepoStub struct {
}

func (r *roleRepoStub) InsertType(source data.Source, root TypeRec) error {
	return nil
}
func (r *roleRepoStub) UpdateType(source data.Source, root TypeRec) error {
	return nil
}
func (r *roleRepoStub) SelectTypeRefs(source data.Source) ([]TypeRef, error) {
	return []TypeRef{}, nil
}
func (r *roleRepoStub) SelectTypeRecByID(source data.Source, id id.ADT) (TypeRec, error) {
	return TypeRec{}, nil
}
func (r *roleRepoStub) SelectTypeRecsByIDs(source data.Source, ids []id.ADT) ([]TypeRec, error) {
	return []TypeRec{}, nil
}
func (r *roleRepoStub) SelectByRef(source data.Source, ref TypeRef) (TypeSnap, error) {
	return TypeSnap{}, nil
}
func (r *roleRepoStub) SelectTypeRecByQN(source data.Source, fqn sym.ADT) (TypeRec, error) {
	return TypeRec{}, nil
}
func (r *roleRepoStub) SelectTypeRecsByQNs(source data.Source, fqns []sym.ADT) ([]TypeRec, error) {
	return []TypeRec{}, nil
}
func (r *roleRepoStub) SelectTypeEnv(source data.Source, fqns []sym.ADT) (map[sym.ADT]TypeRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectParts(id id.ADT) ([]TypeRef, error) {
	return []TypeRef{}, nil
}
func (r *roleRepoStub) InsertTerm(source data.Source, root TermRec) error {
	return nil
}
func (r *roleRepoStub) SelectAll(source data.Source) ([]TermRef, error) {
	return []TermRef{}, nil
}
func (r *roleRepoStub) SelectTermRecByID(source data.Source, sid id.ADT) (TermRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectTermEnv(source data.Source, ids []id.ADT) (map[id.ADT]TermRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectTermRecsByIDs(source data.Source, ids []id.ADT) ([]TermRec, error) {
	return nil, nil
}

type aliasRepoStub struct {
}

func (r *aliasRepoStub) Insert(ds data.Source, ar alias.Root) error {
	return nil
}

type operatorStub struct {
}

func (o *operatorStub) Explicit(ctx context.Context, op func(data.Source) error) error {
	return nil
}
func (o *operatorStub) Implicit(ctx context.Context, op func(data.Source) error) error {
	return nil
}
