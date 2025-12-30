package dec

import (
	"errors"
	"log/slog"
	"math"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
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

func (r *daoPgx) Insert(source sd.Source, mod ProcRec) error {
	ds := sd.MustConform[sd.SourcePgx](source)
	idAttr := slog.Any("id", mod.DecID)
	dto, err := DataFromSigRec(mod)
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
		"sig_id": dto.SigID,
		"rev":    dto.SigRN,
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
		"sig_id":   dto.SigID,
		"rev_from": dto.SigRN,
		"rev_to":   math.MaxInt64,
		"chnl_key": dto.X.ChnlPH,
		"role_fqn": dto.X.TypeQN,
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
	for _, ce := range dto.Ys {
		args := pgx.NamedArgs{
			"sig_id":   dto.SigID,
			"rev_from": dto.SigRN,
			"rev_to":   math.MaxInt64,
			"chnl_key": ce.ChnlPH,
			"role_fqn": ce.TypeQN,
		}
		batch.Queue(insertCE, args)
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	for range dto.Ys {
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

func (r *daoPgx) SelectByID(source sd.Source, rid id.ADT) (ProcSnap, error) {
	ds := sd.MustConform[sd.SourcePgx](source)
	idAttr := slog.Any("id", rid)
	rows, err := ds.Conn.Query(ds.Ctx, selectById, rid.String())
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", selectById))
		return ProcSnap{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sigSnapDS])
	if err != nil {
		r.log.Error("row collection failed", idAttr)
		return ProcSnap{}, err
	}
	r.log.Log(ds.Ctx, lf.LevelTrace, "entitiy selection succeed", slog.Any("dto", dto))
	return DataToSigSnap(dto)
}

func (r *daoPgx) SelectEnv(source sd.Source, ids []id.ADT) (map[id.ADT]ProcRec, error) {
	sigs, err := r.SelectByIDs(source, ids)
	if err != nil {
		return nil, err
	}
	env := make(map[id.ADT]ProcRec, len(sigs))
	for _, s := range sigs {
		env[s.DecID] = s
	}
	return env, nil
}

func (r *daoPgx) SelectByIDs(source sd.Source, ids []id.ADT) (_ []ProcRec, err error) {
	ds := sd.MustConform[sd.SourcePgx](source)
	if len(ids) == 0 {
		return []ProcRec{}, nil
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
	var dtos []sigRecDS
	for _, rid := range ids {
		rows, err := br.Query()
		if err != nil {
			r.log.Error("query execution failed", slog.Any("id", rid), slog.String("q", selectById))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sigRecDS])
		if err != nil {
			r.log.Error("row collection failed", slog.Any("id", rid))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	r.log.Log(ds.Ctx, lf.LevelTrace, "entities selection succeed", slog.Any("dtos", dtos))
	return DataToSigRecs(dtos)
}

func (r *daoPgx) SelectAll(source sd.Source) ([]ProcRef, error) {
	ds := sd.MustConform[sd.SourcePgx](source)
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
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[sigRefDS])
	if err != nil {
		r.log.Error("rows collection failed")
		return nil, err
	}
	return DataToSigRefs(dtos)
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
