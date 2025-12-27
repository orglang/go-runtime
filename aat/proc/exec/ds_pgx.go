package exec

import (
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/avt/core"
	"orglang/orglang/avt/data"
	"orglang/orglang/avt/id"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo() repo {
	return &repoPgx{}
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	return &repoPgx{l}
}

func (r *repoPgx) SelectMain(data.Source, id.ADT) (MainCfg, error) {
	return MainCfg{}, nil
}

func (r *repoPgx) UpdateMain(data.Source, MainMod) error {
	return nil
}

type repoPgx2 struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo2() SemRepo {
	return &repoPgx2{}
}
func (r *repoPgx2) InsertSem(source data.Source, roots ...SemRec) error {
	ds := data.MustConform[data.SourcePgx](source)
	dtos, err := DataFromSemRecs(roots)
	if err != nil {
		r.log.Error("model mapping failed")
		return err
	}
	batch := pgx.Batch{}
	for _, dto := range dtos {
		args := pgx.NamedArgs{
			"id":   dto.ID,
			"kind": dto.K,
			"pid":  dto.PID,
			"vid":  dto.VID,
			"spec": dto.TR,
		}
		batch.Queue(insertRoot, args)
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	for _, dto := range dtos {
		_, err = br.Exec()
		if err != nil {
			r.log.Error("query execution failed", slog.Any("id", dto.ID), slog.String("q", insertRoot))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPgx2) SelectSemByID(source data.Source, rid id.ADT) (SemRec, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE id = $1`
	return r.execute(source, query, rid.String())
}

func (r *repoPgx2) execute(source data.Source, query string, arg string) (SemRec, error) {
	ds := data.MustConform[data.SourcePgx](source)
	rows, err := ds.Conn.Query(ds.Ctx, query, arg)
	if err != nil {
		r.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SemRecData])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.log.Error("row collection failed")
		return nil, err
	}
	root, err := dataToSemRec(dto)
	if err != nil {
		r.log.Error("model mapping failed")
		return nil, err
	}
	r.log.Log(ds.Ctx, core.LevelTrace, "entity selection succeeded", slog.Any("root", root))
	return root, nil
}

const (
	insertRoot = `
		insert into pool_steps (
			id, kind, pid, vid, spec
		) values (
			@id, @kind, @pid, @vid, @spec
		)`
)
