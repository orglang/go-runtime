package exec

import (
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/lf"
	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
)

// Adapter
type daoPgx struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo() repo {
	return &daoPgx{}
}

func newDaoPgx(l *slog.Logger) *daoPgx {
	return &daoPgx{l}
}

func (r *daoPgx) SelectMain(sd.Source, id.ADT) (MainCfg, error) {
	return MainCfg{}, nil
}

func (r *daoPgx) UpdateMain(sd.Source, MainMod) error {
	return nil
}

type repoPgx2 struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo2() SemRepo {
	return &repoPgx2{}
}
func (r *repoPgx2) InsertSem(source sd.Source, roots ...SemRec) error {
	ds := sd.MustConform[sd.SourcePgx](source)
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

func (r *repoPgx2) SelectSemByID(source sd.Source, rid id.ADT) (SemRec, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE id = $1`
	return r.execute(source, query, rid.String())
}

func (r *repoPgx2) execute(source sd.Source, query string, arg string) (SemRec, error) {
	ds := sd.MustConform[sd.SourcePgx](source)
	rows, err := ds.Conn.Query(ds.Ctx, query, arg)
	if err != nil {
		r.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SemRecDS])
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
	r.log.Log(ds.Ctx, lf.LevelTrace, "entity selection succeed", slog.Any("root", root))
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
