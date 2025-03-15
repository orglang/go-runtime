package sig

import (
	"errors"
	"log/slog"
	"math"

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
	name := slog.String("name", "sigRepoPgx")
	return &repoPgx{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &repoPgx{}
}

func (r *repoPgx) Insert(source data.Source, root Root) error {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", root.SigID)
	dto, err := DataFromRoot(root)
	if err != nil {
		r.log.Error("model mapping failed", idAttr)
		return err
	}
	insertRoot := `
		insert into sig_roots (
			sig_id, rev, title
		) VALUES (
			@sig_id, @rev, @title
		)`
	rootArgs := pgx.NamedArgs{
		"sig_id": dto.ID,
		"rev":    dto.Rev,
		"title":  dto.Title,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertRoot, rootArgs)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", insertRoot))
		return err
	}
	insertPE := `
		insert into sig_pes (
			sig_id, rev_from, rev_to, chnl_key, role_fqn
		) VALUES (
			@sig_id, @rev_from, @rev_to, @chnl_key, @role_fqn
		)`
	peArgs := pgx.NamedArgs{
		"sig_id":   dto.ID,
		"rev_from": dto.Rev,
		"rev_to":   math.MaxInt64,
		"chnl_key": dto.PE.Key,
		"role_fqn": dto.PE.Link,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertPE, peArgs)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", insertPE))
		return err
	}
	insertCE := `
		insert into sig_ces (
			sig_id, rev_from, rev_to, chnl_key, role_fqn
		) VALUES (
			@sig_id, @rev_from, @rev_to, @chnl_key, @role_fqn
		)`
	batch := pgx.Batch{}
	for _, ce := range dto.CEs {
		args := pgx.NamedArgs{
			"sig_id":   dto.ID,
			"rev_from": dto.Rev,
			"rev_to":   math.MaxInt64,
			"chnl_key": ce.Key,
			"role_fqn": ce.Link,
		}
		batch.Queue(insertCE, args)
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	for range dto.CEs {
		_, err = br.Exec()
		if err != nil {
			r.log.Error("query execution failed", idAttr, slog.String("q", insertCE))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPgx) SelectByID(source data.Source, rid id.ADT) (Root, error) {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("id", rid)
	rows, err := ds.Conn.Query(ds.Ctx, selectById, rid.String())
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", selectById))
		return Root{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[rootData])
	if err != nil {
		r.log.Error("row collection failed", idAttr)
		return Root{}, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entitiy selection succeeded", slog.Any("dto", dto))
	return DataToRoot(dto)
}

func (r *repoPgx) SelectEnv(source data.Source, ids []ID) (map[ID]Root, error) {
	sigs, err := r.SelectByIDs(source, ids)
	if err != nil {
		return nil, err
	}
	env := make(map[ID]Root, len(sigs))
	for _, s := range sigs {
		env[s.SigID] = s
	}
	return env, nil
}

func (r *repoPgx) SelectByIDs(source data.Source, ids []ID) (_ []Root, err error) {
	ds := data.MustConform[data.SourcePgx](source)
	if len(ids) == 0 {
		return []Root{}, nil
	}
	batch := pgx.Batch{}
	for _, rid := range ids {
		if rid.IsEmpty() {
			return nil, id.ErrEmpty
		}
		batch.Queue(selectById, rid.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var dtos []rootData
	for _, rid := range ids {
		rows, err := br.Query()
		if err != nil {
			r.log.Error("query execution failed", slog.Any("id", rid), slog.String("q", selectById))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[rootData])
		if err != nil {
			r.log.Error("row collection failed", slog.Any("id", rid))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entities selection succeeded", slog.Any("dtos", dtos))
	return DataToRoots(dtos)
}

func (r *repoPgx) SelectAll(source data.Source) ([]Ref, error) {
	ds := data.MustConform[data.SourcePgx](source)
	query := `
		select
			sig_id, rev, title
		from sig_roots`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		r.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[refData])
	if err != nil {
		r.log.Error("rows collection failed")
		return nil, err
	}
	return DataToRefs(dtos)
}

const (
	selectById = `
		select
			sr.sig_id,
			sr.rev,
			(array_agg(sr.title))[1] as title,
			(jsonb_agg(to_jsonb((select ep from (select sp.chnl_key, sp.role_fqn) ep))))[0] as pe,
			jsonb_agg(to_jsonb((select ep from (select sc.chnl_key, sc.role_fqn) ep))) filter (where sc.sig_id is not null) as ces
		from sig_roots sr
		left join sig_pes sp
			on sp.sig_id = sr.sig_id
			and sp.rev_from >= sr.rev
			and sp.rev_to > sr.rev
		left join sig_ces sc
			on sc.sig_id = sr.sig_id
			and sc.rev_from >= sr.rev
			and sc.rev_to > sr.rev
		where sr.sig_id = $1
		group by sr.sig_id, sr.rev`
)
