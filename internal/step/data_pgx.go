package step

import (
	"errors"
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
	name := slog.String("name", "stepRepoPgx")
	return &repoPgx{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &repoPgx{}
}

func (r *repoPgx) Insert(source data.Source, roots ...Root) error {
	ds := data.MustConform[data.SourcePgx](source)
	dtos, err := DataFromRoots(roots)
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
			"spec": dto.Spec,
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

func (r *repoPgx) SelectAll(source data.Source) ([]Ref, error) {
	return nil, nil
}

func (r *repoPgx) SelectByID(source data.Source, rid ID) (Root, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE id = $1`
	return r.execute(source, query, rid.String())
}

func (r *repoPgx) SelectByPID(source data.Source, pid id.ADT) (Root, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE pid = $1`
	return r.execute(source, query, pid.String())
}

func (r *repoPgx) SelectByVID(source data.Source, vid id.ADT) (Root, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE vid = $1`
	return r.execute(source, query, vid.String())
}

func (r *repoPgx) execute(source data.Source, query string, arg string) (Root, error) {
	ds := data.MustConform[data.SourcePgx](source)
	rows, err := ds.Conn.Query(ds.Ctx, query, arg)
	if err != nil {
		r.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RootData])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.log.Error("row collection failed")
		return nil, err
	}
	root, err := dataToRoot(dto)
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
