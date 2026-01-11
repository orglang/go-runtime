package typeexp

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"
	"orglang/orglang/lib/lf"

	"orglang/orglang/adt/identity"
)

// Adapter
type daoPgx struct {
	log *slog.Logger
}

func newDaoPgx(l *slog.Logger) *daoPgx {
	return &daoPgx{l}
}

// for compilation purposes
func newRepo() Repo {
	return &daoPgx{}
}

func (d *daoPgx) Insert(source db.Source, rec ExpRec) (err error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("termID", rec.Ident())
	dto := DataFromTermRec(rec)
	query := `
		INSERT INTO type_term_states (
			exp_id, kind, from_id, spec
		) VALUES (
			@exp_id, @kind, @from_id, @spec
		)`
	batch := pgx.Batch{}
	for _, st := range dto.States {
		sa := pgx.NamedArgs{
			"exp_id":  st.ExpID,
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
			d.log.Error("query execution failed", idAttr, slog.String("q", query))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *daoPgx) SelectRecByID(source db.Source, termID identity.ADT) (ExpRec, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("termID", termID)
	query := `
		WITH RECURSIVE top_states AS (
			SELECT rs.*
			FROM type_term_states rs
			WHERE id = $1
			UNION ALL
			SELECT bs.*
			FROM type_term_states bs, top_states ts
			WHERE bs.from_id = ts.id
		)
		SELECT * FROM top_states`
	rows, err := ds.Conn.Query(ds.Ctx, query, termID.String())
	if err != nil {
		d.log.Error("query execution failed", idAttr, slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[stateDS])
	if err != nil {
		d.log.Error("row collection failed", idAttr)
		return nil, err
	}
	if len(dtos) == 0 {
		d.log.Error("entity selection failed", idAttr)
		return nil, fmt.Errorf("no rows selected")
	}
	d.log.Log(ds.Ctx, lf.LevelTrace, "entity selection succeed", slog.Any("dtos", dtos))
	type_term_states := make(map[string]stateDS, len(dtos))
	for _, dto := range dtos {
		type_term_states[dto.ExpID] = dto
	}
	return statesToTermRec(type_term_states, type_term_states[termID.String()])
}

func (d *daoPgx) SelectEnv(source db.Source, termIDs []identity.ADT) (map[identity.ADT]ExpRec, error) {
	recs, err := d.SelectRecsByIDs(source, termIDs)
	if err != nil {
		return nil, err
	}
	env := make(map[identity.ADT]ExpRec, len(recs))
	for _, rec := range recs {
		env[rec.Ident()] = rec
	}
	return env, nil
}

func (d *daoPgx) SelectRecsByIDs(source db.Source, termIDs []identity.ADT) (_ []ExpRec, err error) {
	ds := db.MustConform[db.SourcePgx](source)
	batch := pgx.Batch{}
	for _, termID := range termIDs {
		batch.Queue(selectByID, termID.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var recs []ExpRec
	for _, termID := range termIDs {
		idAttr := slog.Any("termID", termID)
		rows, err := br.Query()
		if err != nil {
			d.log.Error("query execution failed", idAttr, slog.String("q", selectByID))
		}
		defer rows.Close()
		dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[stateDS])
		if err != nil {
			d.log.Error("rows collection failed", idAttr)
		}
		if len(dtos) == 0 {
			d.log.Error("entity selection failed", idAttr)
			return nil, ErrDoesNotExist(termID)
		}
		rec, err := DataToTermRec(&expRecDS{termID.String(), dtos})
		if err != nil {
			d.log.Error("model mapping failed", idAttr)
			return nil, err
		}
		recs = append(recs, rec)
	}
	d.log.Log(ds.Ctx, lf.LevelTrace, "entities selection succeed", slog.Any("recs", recs))
	return recs, err
}

const (
	selectByID = `
		WITH RECURSIVE state_tree AS (
			SELECT root.*
			FROM type_term_states root
			WHERE id = $1
			UNION ALL
			SELECT child.*
			FROM type_term_states child, state_tree parent
			WHERE child.from_id = parent.id
		)
		SELECT * FROM state_tree`
)
