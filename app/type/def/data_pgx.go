package def

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	name := slog.String("name", "stateRepoPgx")
	return &repoPgx{l.With(name)}
}

// for compilation purposes
func newRepo() TermRepo {
	return &repoPgx{}
}

func (r *repoPgx) Insert(source data.Source, root TermRec) (err error) {
	ds := data.MustConform[data.SourcePgx](source)
	dto := dataFromTermRec(root)
	query := `
		INSERT INTO states (
			id, kind, from_id, spec
		) VALUES (
			@id, @kind, @from_id, @spec
		)`
	batch := pgx.Batch{}
	for _, st := range dto.States {
		sa := pgx.NamedArgs{
			"id":      st.ID,
			"kind":    st.K,
			"from_id": st.FromID,
			"spec":    st.Spec,
		}
		batch.Queue(query, sa)
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	for range dto.States {
		_, err = br.Exec()
		if err != nil {
			r.log.Error("query execution failed", slog.Any("id", root.Ident()), slog.String("q", query))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPgx) SelectAll(source data.Source) ([]TermRef, error) {
	ds := data.MustConform[data.SourcePgx](source)
	query := `
		SELECT
			kind, id
		FROM states`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[*TermRefData])
	if err != nil {
		return nil, err
	}
	return DataToTermRefs(dtos)
}

func (r *repoPgx) SelectByID(source data.Source, typeID id.ADT) (TermRec, error) {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", typeID)
	query := `
		WITH RECURSIVE top_states AS (
			SELECT
				rs.*
			FROM states rs
			WHERE id = $1
			UNION ALL
			SELECT
				bs.*
			FROM states bs, top_states ts
			WHERE bs.from_id = ts.id
		)
		SELECT * FROM top_states`
	rows, err := ds.Conn.Query(ds.Ctx, query, typeID.String())
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[stateData])
	if err != nil {
		r.log.Error("row collection failed", idAttr)
		return nil, err
	}
	if len(dtos) == 0 {
		r.log.Error("entity selection failed", idAttr)
		return nil, fmt.Errorf("no rows selected")
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entity selection succeeded", slog.Any("dtos", dtos))
	states := make(map[string]stateData, len(dtos))
	for _, dto := range dtos {
		states[dto.ID] = dto
	}
	return statesToTermRec(states, states[typeID.String()])
}

func (r *repoPgx) SelectEnv(source data.Source, ids []id.ADT) (map[id.ADT]TermRec, error) {
	states, err := r.SelectByIDs(source, ids)
	if err != nil {
		return nil, err
	}
	env := make(map[id.ADT]TermRec, len(states))
	for _, st := range states {
		env[st.Ident()] = st
	}
	return env, nil
}

func (r *repoPgx) SelectByIDs(source data.Source, ids []id.ADT) (_ []TermRec, err error) {
	ds := data.MustConform[data.SourcePgx](source)
	batch := pgx.Batch{}
	for _, rid := range ids {
		batch.Queue(selectByID, rid.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var roots []TermRec
	for _, rid := range ids {
		idAttr := slog.Any("id", rid)
		rows, err := br.Query()
		if err != nil {
			r.log.Error("query execution failed", idAttr, slog.String("q", selectByID))
		}
		defer rows.Close()
		dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[stateData])
		if err != nil {
			r.log.Error("rows collection failed", idAttr)
		}
		if len(dtos) == 0 {
			err = ErrDoesNotExist(rid)
			r.log.Error("entity selection failed", idAttr)
		}
		root, err := dataToTermRec(&termRecData{rid.String(), dtos})
		if err != nil {
			r.log.Error("model mapping failed", idAttr)
		}
		roots = append(roots, root)
	}
	if err != nil {
		return nil, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entities selection succeeded", slog.Any("roots", roots))
	return roots, err
}

const (
	selectByID = `
		WITH RECURSIVE state_tree AS (
			SELECT root.*
			FROM states root
			WHERE id = $1
			UNION ALL
			SELECT child.*
			FROM states child, state_tree parent
			WHERE child.from_id = parent.id
		)
		SELECT * FROM state_tree
	`
)
