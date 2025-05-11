package def

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/jackc/pgx/v5"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	return &repoPgx{l}
}

// for compilation purposes
func newRepo() Repo {
	return &repoPgx{}
}

func (r *repoPgx) InsertType(source data.Source, rec TypeRec) error {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", rec.TypeID)
	r.log.Log(ds.Ctx, core.LevelTrace, "entity insertion started", idAttr)
	dto, err := DataFromTypeRec(rec)
	if err != nil {
		r.log.Error("model mapping failed", idAttr)
		return err
	}
	insertRoot := `
		insert into role_roots (
			role_id, rev, title
		) values (
			@role_id, @rev, @title
		)`
	rootArgs := pgx.NamedArgs{
		"role_id": dto.TypeID,
		"rev":     dto.TypeRN,
		"title":   dto.Title,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertRoot, rootArgs)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", insertRoot))
		return err
	}
	insertState := `
		insert into role_states (
			role_id, state_id, rev_from, rev_to
		) values (
			@role_id, @state_id, @rev_from, @rev_to
		)`
	stateArgs := pgx.NamedArgs{
		"role_id":  dto.TypeID,
		"rev_from": dto.TypeRN,
		"rev_to":   math.MaxInt64,
		"state_id": dto.TermID,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertState, stateArgs)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", insertState))
		return err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entity insertion succeeded", idAttr)
	return nil
}

func (r *repoPgx) UpdateType(source data.Source, rec TypeRec) error {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", rec.TypeID)
	r.log.Log(ds.Ctx, core.LevelTrace, "entity update started", idAttr)
	dto, err := DataFromTypeRec(rec)
	if err != nil {
		r.log.Error("model mapping failed", idAttr)
		return err
	}
	updateRoot := `
		update role_roots
		set rev = @rev,
			state_id = @state_id
		where role_id = @role_id
			and rev = @rev - 1`
	insertSnap := `
		insert into role_snaps (
			role_id, rev, title, state_id
		) values (
			@role_id, @rev, @title, @state_id
		)`
	args := pgx.NamedArgs{
		"role_id":  dto.TypeID,
		"rev":      dto.TypeRN,
		"title":    dto.Title,
		"state_id": dto.TermID,
	}
	ct, err := ds.Conn.Exec(ds.Ctx, updateRoot, args)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", updateRoot))
		return err
	}
	if ct.RowsAffected() == 0 {
		r.log.Error("entity update failed", idAttr)
		return errOptimisticUpdate(rec.TypeRN - 1)
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertSnap, args)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", insertSnap))
		return err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entity update succeeded", idAttr)
	return nil
}

func (r *repoPgx) SelectTypeRefs(source data.Source) ([]TypeRef, error) {
	ds := data.MustConform[data.SourcePgx](source)
	query := `
		SELECT
			role_id, rev, title
		FROM role_roots`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		r.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[typeRefData])
	if err != nil {
		r.log.Error("rows collection failed")
		return nil, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entities selection succeeded", slog.Any("dtos", dtos))
	return DataToTypeRefs(dtos)
}

func (r *repoPgx) SelectTypeRecByID(source data.Source, recID id.ADT) (TypeRec, error) {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", recID)
	rows, err := ds.Conn.Query(ds.Ctx, selectById, recID.String())
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", selectById))
		return TypeRec{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[typeRecData])
	if err != nil {
		r.log.Error("row collection failed", idAttr)
		return TypeRec{}, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entity selection succeeded", idAttr)
	return DataToTypeRec(dto)
}

func (r *repoPgx) SelectTypeRecByQN(source data.Source, recQN sym.ADT) (TypeRec, error) {
	ds := data.MustConform[data.SourcePgx](source)
	fqnAttr := slog.Any("qn", recQN)
	rows, err := ds.Conn.Query(ds.Ctx, selectByFQN, sym.ConvertToString(recQN))
	if err != nil {
		r.log.Error("query execution failed", fqnAttr, slog.String("q", selectByFQN))
		return TypeRec{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[typeRecData])
	if err != nil {
		r.log.Error("row collection failed", fqnAttr)
		return TypeRec{}, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entity selection succeeded", fqnAttr)
	return DataToTypeRec(dto)
}

func (r *repoPgx) SelectTypeRecsByIDs(source data.Source, recIDs []id.ADT) (_ []TypeRec, err error) {
	ds := data.MustConform[data.SourcePgx](source)
	if len(recIDs) == 0 {
		return []TypeRec{}, nil
	}
	query := `
		select
			role_id, rev, title, state_id, whole_id
		from role_roots
		where role_id = $1`
	batch := pgx.Batch{}
	for _, rid := range recIDs {
		if rid.IsEmpty() {
			return nil, id.ErrEmpty
		}
		batch.Queue(query, rid.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var dtos []typeRecData
	for _, rid := range recIDs {
		rows, err := br.Query()
		if err != nil {
			r.log.Error("query execution failed", slog.Any("id", rid), slog.String("q", query))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[typeRecData])
		if err != nil {
			r.log.Error("row collection failed", slog.Any("id", rid))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entities selection succeeded", slog.Any("dtos", dtos))
	return DataToTypeRecs(dtos)
}

func (r *repoPgx) SelectTypeEnv(source data.Source, recQNs []sym.ADT) (map[sym.ADT]TypeRec, error) {
	recs, err := r.SelectTypeRecsByQNs(source, recQNs)
	if err != nil {
		return nil, err
	}
	env := make(map[sym.ADT]TypeRec, len(recs))
	for i, root := range recs {
		env[recQNs[i]] = root
	}
	return env, nil
}

func (r *repoPgx) SelectTypeRecsByQNs(source data.Source, recQNs []sym.ADT) (_ []TypeRec, err error) {
	ds := data.MustConform[data.SourcePgx](source)
	if len(recQNs) == 0 {
		return []TypeRec{}, nil
	}
	batch := pgx.Batch{}
	for _, fqn := range recQNs {
		batch.Queue(selectByFQN, sym.ConvertToString(fqn))
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var dtos []typeRecData
	for _, fqn := range recQNs {
		rows, err := br.Query()
		if err != nil {
			r.log.Error("query execution failed", slog.Any("fqn", fqn), slog.String("q", selectByFQN))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[typeRecData])
		if err != nil {
			r.log.Error("row collection failed", slog.Any("fqn", fqn))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entities selection succeeded", slog.Any("dtos", dtos))
	return DataToTypeRecs(dtos)
}

func (r *repoPgx) InsertTerm(source data.Source, rec TermRec) (err error) {
	ds := data.MustConform[data.SourcePgx](source)
	dto := dataFromTermRec(rec)
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
			r.log.Error("query execution failed", slog.Any("id", rec.Ident()), slog.String("q", query))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPgx) SelectTermRecByID(source data.Source, recID id.ADT) (TermRec, error) {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", recID)
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
	rows, err := ds.Conn.Query(ds.Ctx, query, recID.String())
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
	return statesToTermRec(states, states[recID.String()])
}

func (r *repoPgx) SelectTermEnv(source data.Source, recIDs []id.ADT) (map[id.ADT]TermRec, error) {
	recs, err := r.SelectTermRecsByIDs(source, recIDs)
	if err != nil {
		return nil, err
	}
	env := make(map[id.ADT]TermRec, len(recs))
	for _, rec := range recs {
		env[rec.Ident()] = rec
	}
	return env, nil
}

func (r *repoPgx) SelectTermRecsByIDs(source data.Source, recIDs []id.ADT) (_ []TermRec, err error) {
	ds := data.MustConform[data.SourcePgx](source)
	batch := pgx.Batch{}
	for _, rid := range recIDs {
		batch.Queue(selectByID, rid.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var recs []TermRec
	for _, recID := range recIDs {
		idAttr := slog.Any("id", recID)
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
			r.log.Error("entity selection failed", idAttr)
			return nil, ErrDoesNotExist(recID)
		}
		rec, err := dataToTermRec(&termRecData{recID.String(), dtos})
		if err != nil {
			r.log.Error("model mapping failed", idAttr)
			return nil, err
		}
		recs = append(recs, rec)
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entities selection succeeded", slog.Any("recs", recs))
	return recs, err
}

const (
	selectByFQN = `
		select
			rr.role_id,
			rr.rev,
			rr.title,
			rs.state_id,
			null as whole_id
		from role_roots rr
		left join aliases a
			on a.id = rr.role_id
			and a.rev_from >= rr.rev
			and a.rev_to > rr.rev
		left join role_states rs
			on rs.role_id = rr.role_id
			and rs.rev_from >= rr.rev
			and rs.rev_to > rr.rev
		where a.sym = $1`

	selectById = `
		select
			rr.role_id,
			rr.rev,
			rr.title,
			rs.state_id,
			null as whole_id
		from role_roots rr
		left join role_states rs
			on rs.role_id = rr.role_id
			and rs.rev_from >= rr.rev
			and rs.rev_to > rr.rev
		where rr.role_id = $1`

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
		SELECT * FROM state_tree`
)
