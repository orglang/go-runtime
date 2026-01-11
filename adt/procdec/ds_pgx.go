package procdec

import (
	"errors"
	"log/slog"
	"math"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"
	"orglang/orglang/lib/lf"

	"orglang/orglang/adt/identity"
)

// Adapter
type pgxDAO struct {
	log *slog.Logger
}

func newPgxDAO(l *slog.Logger) *pgxDAO {
	name := slog.String("name", reflect.TypeFor[pgxDAO]().Name())
	return &pgxDAO{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &pgxDAO{}
}

func (dao *pgxDAO) Insert(source db.Source, mod DecRec) error {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("id", mod.DecID)
	dto, err := DataFromDecRec(mod)
	if err != nil {
		dao.log.Error("model conversion failed", idAttr)
		return err
	}
	insertRoot := `
		insert into dec_roots (
			dec_id, rev, title
		) VALUES (
			@dec_id, @rev, @title
		)`
	rootArgs := pgx.NamedArgs{
		"dec_id": dto.DecID,
		"rev":    dto.DecRN,
		"title":  dto.Title,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertRoot, rootArgs)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", insertRoot))
		return err
	}
	insertPE := `
		insert into dec_pes (
			dec_id, from_rn, to_rn, chnl_key, role_fqn
		) VALUES (
			@dec_id, @from_rn, @to_rn, @chnl_key, @role_fqn
		)`
	peArgs := pgx.NamedArgs{
		"dec_id":   dto.DecID,
		"from_rn":  dto.DecRN,
		"to_rn":    math.MaxInt64,
		"chnl_key": dto.X.BindPH,
		"role_fqn": dto.X.TypeQN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertPE, peArgs)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", insertPE))
		return err
	}
	insertCE := `
		insert into dec_ces (
			dec_id, from_rn, to_rn, chnl_key, role_fqn
		) VALUES (
			@dec_id, @from_rn, @to_rn, @chnl_key, @role_fqn
		)`
	batch := pgx.Batch{}
	for _, ce := range dto.Ys {
		args := pgx.NamedArgs{
			"dec_id":   dto.DecID,
			"from_rn":  dto.DecRN,
			"to_rn":    math.MaxInt64,
			"chnl_key": ce.BindPH,
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
			dao.log.Error("query execution failed", idAttr, slog.String("q", insertCE))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (dao *pgxDAO) SelectByID(source db.Source, rid identity.ADT) (DecSnap, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("id", rid)
	rows, err := ds.Conn.Query(ds.Ctx, selectById, rid.String())
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", selectById))
		return DecSnap{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[decSnapDS])
	if err != nil {
		dao.log.Error("row collection failed", idAttr)
		return DecSnap{}, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entitiy selection succeed", slog.Any("dto", dto))
	return DataToDecSnap(dto)
}

func (dao *pgxDAO) SelectEnv(source db.Source, ids []identity.ADT) (map[identity.ADT]DecRec, error) {
	decs, err := dao.SelectByIDs(source, ids)
	if err != nil {
		return nil, err
	}
	env := make(map[identity.ADT]DecRec, len(decs))
	for _, s := range decs {
		env[s.DecID] = s
	}
	return env, nil
}

func (dao *pgxDAO) SelectByIDs(source db.Source, ids []identity.ADT) (_ []DecRec, err error) {
	ds := db.MustConform[db.SourcePgx](source)
	if len(ids) == 0 {
		return []DecRec{}, nil
	}
	batch := pgx.Batch{}
	for _, rid := range ids {
		if rid.IsEmpty() {
			return nil, identity.ErrEmpty
		}
		batch.Queue(selectById, rid.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var dtos []decRecDS
	for _, rid := range ids {
		rows, err := br.Query()
		if err != nil {
			dao.log.Error("query execution failed", slog.Any("id", rid), slog.String("q", selectById))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[decRecDS])
		if err != nil {
			dao.log.Error("row collection failed", slog.Any("id", rid))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entities selection succeed", slog.Any("dtos", dtos))
	return DataToDecRecs(dtos)
}

func (dao *pgxDAO) SelectAll(source db.Source) ([]DecRef, error) {
	ds := db.MustConform[db.SourcePgx](source)
	query := `
		select
			dec_id, rev, title
		from dec_roots`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		dao.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[decRefDS])
	if err != nil {
		dao.log.Error("rows collection failed")
		return nil, err
	}
	return DataToDecRefs(dtos)
}

const (
	selectById = `
		select
			sr.dec_id,
			sr.rev,
			(array_agg(sr.title))[1] as title,
			(jsonb_agg(to_jsonb((select ep from (select sp.chnl_key, sp.role_fqn) ep))))[0] as pe,
			jsonb_agg(to_jsonb((select ep from (select sc.chnl_key, sc.role_fqn) ep))) filter (where sc.dec_id is not null) as ces
		from dec_roots sr
		left join dec_pes sp
			on sp.dec_id = sr.dec_id
			and sp.from_rn >= sr.rev
			and sp.to_rn > sr.rev
		left join dec_ces sc
			on sc.dec_id = sr.dec_id
			and sc.from_rn >= sr.rev
			and sc.to_rn > sr.rev
		where sr.dec_id = $1
		group by sr.dec_id, sr.rev`
)
