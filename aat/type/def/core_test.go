package def

import (
	"context"
	"log/slog"
	"testing"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"

	"orglang/orglang/aet/alias"
)

func TestKinshipEstalish(t *testing.T) {
	newService(&roleRepoStub{}, &aliasRepoStub{}, &operatorStub{}, slog.Default())
}

type roleRepoStub struct {
}

func (r *roleRepoStub) InsertType(source sd.Source, root TypeRec) error {
	return nil
}
func (r *roleRepoStub) UpdateType(source sd.Source, root TypeRec) error {
	return nil
}
func (r *roleRepoStub) SelectTypeRefs(source sd.Source) ([]TypeRef, error) {
	return []TypeRef{}, nil
}
func (r *roleRepoStub) SelectTypeRecByID(source sd.Source, id id.ADT) (TypeRec, error) {
	return TypeRec{}, nil
}
func (r *roleRepoStub) SelectTypeRecsByIDs(source sd.Source, ids []id.ADT) ([]TypeRec, error) {
	return []TypeRec{}, nil
}
func (r *roleRepoStub) SelectByRef(source sd.Source, ref TypeRef) (TypeSnap, error) {
	return TypeSnap{}, nil
}
func (r *roleRepoStub) SelectTypeRecByQN(source sd.Source, fqn sym.ADT) (TypeRec, error) {
	return TypeRec{}, nil
}
func (r *roleRepoStub) SelectTypeRecsByQNs(source sd.Source, fqns []sym.ADT) ([]TypeRec, error) {
	return []TypeRec{}, nil
}
func (r *roleRepoStub) SelectTypeEnv(source sd.Source, fqns []sym.ADT) (map[sym.ADT]TypeRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectParts(id id.ADT) ([]TypeRef, error) {
	return []TypeRef{}, nil
}
func (r *roleRepoStub) InsertTerm(source sd.Source, root TermRec) error {
	return nil
}
func (r *roleRepoStub) SelectAll(source sd.Source) ([]TermRef, error) {
	return []TermRef{}, nil
}
func (r *roleRepoStub) SelectTermRecByID(source sd.Source, sid id.ADT) (TermRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectTermEnv(source sd.Source, ids []id.ADT) (map[id.ADT]TermRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectTermRecsByIDs(source sd.Source, ids []id.ADT) ([]TermRec, error) {
	return nil, nil
}

type aliasRepoStub struct {
}

func (r *aliasRepoStub) Insert(ds sd.Source, ar alias.Root) error {
	return nil
}

type operatorStub struct {
}

func (o *operatorStub) Explicit(ctx context.Context, op func(sd.Source) error) error {
	return nil
}
func (o *operatorStub) Implicit(ctx context.Context, op func(sd.Source) error) error {
	return nil
}
