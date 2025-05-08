package dec

import (
	"context"
	"log/slog"
	"testing"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	typedef "smecalculus/rolevod/app/type/def"
	"smecalculus/rolevod/internal/alias"
)

func TestKinshipEstalish(t *testing.T) {
	newService(&roleRepoStub{}, &stateRepoStub{}, &aliasRepoStub{}, &operatorStub{}, slog.Default())
}

type roleRepoStub struct {
}

func (r *roleRepoStub) Insert(source data.Source, root TypeRec) error {
	return nil
}
func (r *roleRepoStub) Update(source data.Source, root TypeRec) error {
	return nil
}
func (r *roleRepoStub) SelectRefs(source data.Source) ([]TypeRef, error) {
	return []TypeRef{}, nil
}
func (r *roleRepoStub) SelectByID(source data.Source, id id.ADT) (TypeRec, error) {
	return TypeRec{}, nil
}
func (r *roleRepoStub) SelectByIDs(source data.Source, ids []id.ADT) ([]TypeRec, error) {
	return []TypeRec{}, nil
}
func (r *roleRepoStub) SelectByRef(source data.Source, ref TypeRef) (TypeSnap, error) {
	return TypeSnap{}, nil
}
func (r *roleRepoStub) SelectByQN(source data.Source, fqn sym.ADT) (TypeRec, error) {
	return TypeRec{}, nil
}
func (r *roleRepoStub) SelectByQNs(source data.Source, fqns []sym.ADT) ([]TypeRec, error) {
	return []TypeRec{}, nil
}
func (r *roleRepoStub) SelectEnv(source data.Source, fqns []sym.ADT) (map[sym.ADT]TypeRec, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectParts(id id.ADT) ([]TypeRef, error) {
	return []TypeRef{}, nil
}

type stateRepoStub struct {
}

func (r *stateRepoStub) Insert(source data.Source, root typedef.TermRec) error {
	return nil
}
func (r *stateRepoStub) SelectAll(source data.Source) ([]typedef.TermRef, error) {
	return []typedef.TermRef{}, nil
}
func (r *stateRepoStub) SelectByID(source data.Source, sid id.ADT) (typedef.TermRec, error) {
	return nil, nil
}
func (r *stateRepoStub) SelectEnv(source data.Source, ids []id.ADT) (map[id.ADT]typedef.TermRec, error) {
	return nil, nil
}
func (r *stateRepoStub) SelectByIDs(source data.Source, ids []id.ADT) ([]typedef.TermRec, error) {
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
