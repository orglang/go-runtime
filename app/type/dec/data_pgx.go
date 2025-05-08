package dec

import (
	"errors"
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

func (r *repoPgx) Insert(source data.Source, rec TypeRec) error {
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

func (r *repoPgx) Update(source data.Source, rec TypeRec) error {
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

func (r *repoPgx) SelectRefs(source data.Source) ([]TypeRef, error) {
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

func (r *repoPgx) SelectByID(source data.Source, recID id.ADT) (TypeRec, error) {
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

func (r *repoPgx) SelectByQN(source data.Source, recQN sym.ADT) (TypeRec, error) {
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

func (r *repoPgx) SelectByIDs(source data.Source, recIDs []id.ADT) (_ []TypeRec, err error) {
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

func (r *repoPgx) SelectEnv(source data.Source, recQNs []sym.ADT) (map[sym.ADT]TypeRec, error) {
	recs, err := r.SelectByQNs(source, recQNs)
	if err != nil {
		return nil, err
	}
	env := make(map[sym.ADT]TypeRec, len(recs))
	for i, root := range recs {
		env[recQNs[i]] = root
	}
	return env, nil
}

func (r *repoPgx) SelectByQNs(source data.Source, recQNs []sym.ADT) (_ []TypeRec, err error) {
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
)
