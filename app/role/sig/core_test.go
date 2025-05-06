package role

import (
	"context"
	"log/slog"
	"testing"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"
	"smecalculus/rolevod/internal/state"
)

func TestKinshipEstalish(t *testing.T) {
	newService(&roleRepoStub{}, &stateRepoStub{}, &aliasRepoStub{}, &operatorStub{}, slog.Default())
}

type roleRepoStub struct {
}

func (r *roleRepoStub) Insert(source data.Source, root Impl) error {
	return nil
}
func (r *roleRepoStub) Update(source data.Source, root Impl) error {
	return nil
}
func (r *roleRepoStub) SelectRefs(source data.Source) ([]Ref, error) {
	return []Ref{}, nil
}
func (r *roleRepoStub) SelectByID(source data.Source, id id.ADT) (Impl, error) {
	return Impl{}, nil
}
func (r *roleRepoStub) SelectByIDs(source data.Source, ids []id.ADT) ([]Impl, error) {
	return []Impl{}, nil
}
func (r *roleRepoStub) SelectByRef(source data.Source, ref Ref) (Snap, error) {
	return Snap{}, nil
}
func (r *roleRepoStub) SelectByFQN(source data.Source, fqn sym.ADT) (Impl, error) {
	return Impl{}, nil
}
func (r *roleRepoStub) SelectByFQNs(source data.Source, fqns []sym.ADT) ([]Impl, error) {
	return []Impl{}, nil
}
func (r *roleRepoStub) SelectEnv(source data.Source, fqns []sym.ADT) (map[sym.ADT]Impl, error) {
	return nil, nil
}
func (r *roleRepoStub) SelectParts(id id.ADT) ([]Ref, error) {
	return []Ref{}, nil
}

type stateRepoStub struct {
}

func (r *stateRepoStub) Insert(source data.Source, root state.Root) error {
	return nil
}
func (r *stateRepoStub) SelectAll(source data.Source) ([]state.Ref, error) {
	return []state.Ref{}, nil
}
func (r *stateRepoStub) SelectByID(source data.Source, sid id.ADT) (state.Root, error) {
	return nil, nil
}
func (r *stateRepoStub) SelectEnv(source data.Source, ids []id.ADT) (map[state.ID]state.Root, error) {
	return nil, nil
}
func (r *stateRepoStub) SelectByIDs(source data.Source, ids []id.ADT) ([]state.Root, error) {
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
